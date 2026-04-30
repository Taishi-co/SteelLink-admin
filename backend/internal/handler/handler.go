package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-admin/backend/internal/auth"
	"github.com/Taishi-co/SteelLink-admin/backend/internal/database"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func defaultAdminUsername() string {
	if value := os.Getenv("ADMIN_DEFAULT_USERNAME"); value != "" {
		return value
	}
	return "admin"
}

func defaultAdminPassword() string {
	if value := os.Getenv("ADMIN_DEFAULT_PASSWORD"); value != "" {
		return value
	}
	return "admin123"
}

func defaultAdminRealName() string {
	if value := os.Getenv("ADMIN_DEFAULT_REAL_NAME"); value != "" {
		return value
	}
	return "Platform Admin"
}

// HealthCheck returns server health status
func HealthCheck(c *gin.Context) {
	dbStatus := "ok"
	if err := database.Ping(); err != nil {
		dbStatus = err.Error()
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "degraded",
			"message":  "SteelLink Admin API database check failed",
			"database": dbStatus,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"message":  "SteelLink Admin API is running",
		"database": dbStatus,
	})
}

func GetDashboardStats(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	stats := dashboardStats{}
	var err error

	stats.TotalUsers, err = queryCount(ctx, db, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	stats.TotalEnterprises, err = queryCount(ctx, db, `SELECT COUNT(*) FROM enterprises WHERE deleted_at IS NULL`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	stats.TotalProducts, err = queryCount(ctx, db, `SELECT COUNT(*) FROM products WHERE deleted_at IS NULL`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	pendingEnterprises, _ := queryCount(ctx, db, `SELECT COUNT(*) FROM enterprises WHERE deleted_at IS NULL AND status = 0`)
	pendingProducts, _ := queryCount(ctx, db, `SELECT COUNT(*) FROM products WHERE deleted_at IS NULL AND status IN ('draft', 'pending_review', '0')`)
	pendingOrders, _ := queryCount(ctx, db, `SELECT COUNT(*) FROM orders WHERE deleted_at IS NULL AND status IN ('pending_contract', 'pending_payment', 'pending_payment_audit', 'pending_pickup')`)
	stats.PendingAudits = pendingEnterprises + pendingProducts + pendingOrders

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

const (
	maxLoginFails   = 5
	lockoutDuration = 60 * time.Minute
)

func ensureAdminSecurityColumns(ctx context.Context, db *sql.DB) {
	if !columnExists(ctx, db, "admin_users", "login_fail_count") {
		db.ExecContext(ctx, `ALTER TABLE admin_users ADD COLUMN login_fail_count INT NOT NULL DEFAULT 0`)
	}
	if !columnExists(ctx, db, "admin_users", "locked_until") {
		db.ExecContext(ctx, `ALTER TABLE admin_users ADD COLUMN locked_until DATETIME NULL DEFAULT NULL`)
	}
}

// Login handles admin login
func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid login payload",
		})
		return
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db != nil && tableExists(ctx, db, "admin_users") {
		ensureAdminSecurityColumns(ctx, db)

		var id int64
		var username, realName, hashedPwd string
		var roleID int
		var failCount int
		var lockedUntil sql.NullTime

		err := db.QueryRowContext(ctx, `
			SELECT id, username, real_name, password, role_id, COALESCE(login_fail_count, 0), locked_until
			FROM admin_users WHERE deleted_at IS NULL AND username = ? AND status = 1 LIMIT 1`, req.Username,
		).Scan(&id, &username, &realName, &hashedPwd, &roleID, &failCount, &lockedUntil)
		if err == nil {
			if lockedUntil.Valid && lockedUntil.Time.After(time.Now()) {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": fmt.Sprintf("账号已锁定，请于 %s 后再试", lockedUntil.Time.Format("15:04:05")),
				})
				return
			}

			if err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(req.Password)); err != nil {
				failCount++
				if failCount >= maxLoginFails {
					lockTime := time.Now().Add(lockoutDuration)
					db.ExecContext(ctx, `UPDATE admin_users SET login_fail_count = ?, locked_until = ?, updated_at = NOW() WHERE id = ?`, failCount, lockTime, id)
					c.JSON(http.StatusForbidden, gin.H{
						"code":    403,
						"message": fmt.Sprintf("连续登录失败%d次，账号已锁定1小时", maxLoginFails),
					})
				} else {
					db.ExecContext(ctx, `UPDATE admin_users SET login_fail_count = ?, updated_at = NOW() WHERE id = ?`, failCount, id)
					c.JSON(http.StatusUnauthorized, gin.H{
						"code":    401,
						"message": fmt.Sprintf("用户名或密码错误（已失败%d次，%d次后锁定）", failCount, maxLoginFails),
					})
				}
				return
			}

			db.ExecContext(ctx, `UPDATE admin_users SET login_fail_count = 0, locked_until = NULL, updated_at = NOW() WHERE id = ?`, id)

			token, err := auth.GenerateAdminToken(id, username, realName, roleID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to generate token"})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data": gin.H{
					"token": token,
					"adminInfo": gin.H{
						"id":       id,
						"username": username,
						"realName": realName,
						"roleId":   roleID,
					},
				},
			})

			logOperation(ctx, db, id, username, "auth", "login", "POST", c.Request.URL.Path, c.ClientIP(), "", 0)
			return
		}
	}

	if req.Username != defaultAdminUsername() || req.Password != defaultAdminPassword() {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "invalid username or password",
		})
		return
	}

	token, err := auth.GenerateAdminToken(1, defaultAdminUsername(), defaultAdminRealName(), 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"token": token,
			"adminInfo": gin.H{
				"id":       1,
				"username": defaultAdminUsername(),
				"realName": defaultAdminRealName(),
				"roleId":   1,
			},
		},
	})

	if db != nil {
		logOperation(c.Request.Context(), db, 1, defaultAdminUsername(), "auth", "login", "POST", c.Request.URL.Path, c.ClientIP(), "", 0)
	}
}

// Logout clears current session on client side
func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    true,
	})
}

// GetAdminInfo returns current admin info from token claims
func GetAdminInfo(c *gin.Context) {
	adminID, _ := c.Get("admin_id")
	username, _ := c.Get("admin_username")
	realName, _ := c.Get("admin_real_name")
	roleID, _ := c.Get("admin_role_id")

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id":       adminID,
			"username": username,
			"realName": realName,
			"roleId":   roleID,
		},
	})
}

// ListAdmins returns list of admins
func ListAdmins(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db != nil && tableExists(ctx, db, "admin_users") {
		limit, offset := parsePageParams(c.Query("page"), c.Query("pageSize"))
		total, err := queryCount(ctx, db, `SELECT COUNT(*) FROM admin_users WHERE deleted_at IS NULL`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}

		rows, err := db.QueryContext(ctx, `
			SELECT id, username, COALESCE(real_name, ''), COALESCE(email, ''), COALESCE(phone, ''), status, role_id, created_at
			FROM admin_users
			WHERE deleted_at IS NULL
			ORDER BY id DESC
			LIMIT ? OFFSET ?`, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		defer rows.Close()

		items := make([]adminItem, 0)
		for rows.Next() {
			var item adminItem
			if err := rows.Scan(&item.ID, &item.Username, &item.RealName, &item.Email, &item.Phone, &item.Status, &item.RoleID, &item.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
				return
			}
			items = append(items, item)
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": pageResult[adminItem]{List: items, Total: total}})
		return
	}

	items := listLocalAdmins()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    pageResult[adminItem]{List: items, Total: len(items)},
	})
}

// GetAdmin returns a single admin
func GetAdmin(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db != nil && tableExists(ctx, db, "admin_users") {
		var item adminItem
		row := db.QueryRowContext(ctx, `
			SELECT id, username, COALESCE(real_name, ''), COALESCE(email, ''), COALESCE(phone, ''), status, role_id, created_at
			FROM admin_users
			WHERE deleted_at IS NULL AND id = ?
			LIMIT 1`, c.Param("id"))
		if err := row.Scan(&item.ID, &item.Username, &item.RealName, &item.Email, &item.Phone, &item.Status, &item.RoleID, &item.CreatedAt); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "admin not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": item})
		return
	}

	var id int64
	fmt.Sscanf(c.Param("id"), "%d", &id)
	item, ok := getLocalAdminByID(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "admin not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": item})
}

// CreateAdmin creates a new admin
func CreateAdmin(c *gin.Context) {
	var req adminUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid admin payload"})
		return
	}
	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "password is required"})
		return
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db != nil && tableExists(ctx, db, "admin_users") {
		hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败"})
			return
		}
		now := time.Now()
		result, err := db.ExecContext(ctx, `
			INSERT INTO admin_users (username, password, real_name, email, phone, status, role_id, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			req.Username,
			string(hashedPwd),
			req.RealName,
			req.Email,
			req.Phone,
			req.Status,
			req.RoleID,
			now,
			now,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		id, _ := result.LastInsertId()
		logOpFromContext(c, db, "admin", "create", 0, fmt.Sprintf(`{"username":"%s","realName":"%s"}`, req.Username, req.RealName))
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": adminItem{
			ID:        id,
			Username:  req.Username,
			RealName:  req.RealName,
			Email:     req.Email,
			Phone:     req.Phone,
			Status:    req.Status,
			RoleID:    req.RoleID,
			CreatedAt: now,
		}})
		return
	}

	item := createLocalAdmin(req)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": item})
}

// UpdateAdmin updates an admin
func UpdateAdmin(c *gin.Context) {
	var req adminUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid admin payload"})
		return
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db != nil && tableExists(ctx, db, "admin_users") {
		if req.Password != "" {
			hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败"})
				return
			}
			result, err := db.ExecContext(ctx, `
				UPDATE admin_users
				SET username = ?, password = ?, real_name = ?, email = ?, phone = ?, status = ?, role_id = ?, updated_at = ?
				WHERE id = ? AND deleted_at IS NULL`,
				req.Username,
				string(hashedPwd),
				req.RealName,
				req.Email,
				req.Phone,
				req.Status,
				req.RoleID,
				time.Now(),
				c.Param("id"),
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
				return
			}
			affected, _ := result.RowsAffected()
			if affected == 0 {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "admin not found"})
				return
			}
		} else {
			result, err := db.ExecContext(ctx, `
				UPDATE admin_users
				SET username = ?, real_name = ?, email = ?, phone = ?, status = ?, role_id = ?, updated_at = ?
				WHERE id = ? AND deleted_at IS NULL`,
				req.Username,
				req.RealName,
				req.Email,
				req.Phone,
				req.Status,
				req.RoleID,
				time.Now(),
				c.Param("id"),
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
				return
			}
			affected, _ := result.RowsAffected()
			if affected == 0 {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "admin not found"})
				return
			}
		}
		logOpFromContext(c, db, "admin", "update", 0, fmt.Sprintf(`{"id":"%s","username":"%s"}`, c.Param("id"), req.Username))
		GetAdmin(c)
		return
	}

	var id int64
	fmt.Sscanf(c.Param("id"), "%d", &id)
	item, ok := updateLocalAdmin(id, req)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "admin not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": item})
}

// DeleteAdmin deletes an admin
func DeleteAdmin(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db != nil && tableExists(ctx, db, "admin_users") {
		result, err := db.ExecContext(ctx, `UPDATE admin_users SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL`, time.Now(), c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		affected, _ := result.RowsAffected()
		if affected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "admin not found"})
			return
		}
		logOpFromContext(c, db, "admin", "delete", 0, fmt.Sprintf(`{"id":"%s"}`, c.Param("id")))
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": true})
		return
	}

	var id int64
	fmt.Sscanf(c.Param("id"), "%d", &id)
	if !deleteLocalAdmin(id) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "default admin cannot be deleted"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": true})
}

func GetMyMenus(c *gin.Context) {
	roleID, _ := c.Get("admin_role_id")
	intRoleID := 1
	if v, ok := roleID.(int); ok {
		intRoleID = v
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db == nil || !tableExists(ctx, db, "admin_menus") {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": []interface{}{}})
		return
	}

	if intRoleID == 1 {
		rows, err := db.QueryContext(ctx, `
			SELECT id, parent_id, name, COALESCE(path, ''), COALESCE(component, ''), COALESCE(icon, ''), type, COALESCE(permission, ''), status, sort
			FROM admin_menus WHERE deleted_at IS NULL AND status = 1
			ORDER BY sort ASC, id ASC`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		defer rows.Close()
		menus := scanMenuRows(rows)
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": buildMenuTree(menus)})
		return
	}

	if !tableExists(ctx, db, "admin_role_menus") {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": []interface{}{}})
		return
	}

	rows, err := db.QueryContext(ctx, `
		WITH RECURSIVE role_menus AS (
			SELECT m.id, m.parent_id, m.name, m.path, m.component, m.icon, m.type, m.permission, m.status, m.sort
			FROM admin_menus m
			INNER JOIN admin_role_menus rm ON rm.menu_id = m.id
			WHERE rm.role_id = ? AND m.deleted_at IS NULL AND m.status = 1
			UNION
			SELECT p.id, p.parent_id, p.name, p.path, p.component, p.icon, p.type, p.permission, p.status, p.sort
			FROM admin_menus p
			INNER JOIN role_menus c ON c.parent_id = p.id
			WHERE p.deleted_at IS NULL AND p.status = 1
		)
		SELECT id, parent_id, name, COALESCE(path, ''), COALESCE(component, ''), COALESCE(icon, ''), type, COALESCE(permission, ''), status, sort
		FROM role_menus
		ORDER BY sort ASC, id ASC`, intRoleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer rows.Close()
	menus := scanMenuRows(rows)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": buildMenuTree(menus)})
}

func ListOperationLogs(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db == nil || !tableExists(ctx, db, "admin_operation_logs") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	limit, offset := parsePageParams(c.Query("page"), c.Query("pageSize"))

	whereParts := []string{"1=1"}
	args := make([]any, 0)

	if module := c.Query("module"); module != "" {
		whereParts = append(whereParts, "module = ?")
		args = append(args, module)
	}
	if action := c.Query("action"); action != "" {
		whereParts = append(whereParts, "action = ?")
		args = append(args, action)
	}
	if adminName := c.Query("adminName"); adminName != "" {
		whereParts = append(whereParts, "admin_name LIKE ?")
		args = append(args, "%"+adminName+"%")
	}
	if startDate := c.Query("startDate"); startDate != "" {
		whereParts = append(whereParts, "created_at >= ?")
		args = append(args, startDate+" 00:00:00")
	}
	if endDate := c.Query("endDate"); endDate != "" {
		whereParts = append(whereParts, "created_at <= ?")
		args = append(args, endDate+" 23:59:59")
	}

	whereClause := strings.Join(whereParts, " AND ")

	total, err := queryCount(ctx, db, fmt.Sprintf(`SELECT COUNT(*) FROM admin_operation_logs WHERE %s`, whereClause), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, admin_id, COALESCE(admin_name, ''), COALESCE(module, ''), COALESCE(action, ''),
		       COALESCE(method, ''), COALESCE(path, ''), COALESCE(ip, ''),
		       COALESCE(request_body, ''), COALESCE(response_code, 0), created_at
		FROM admin_operation_logs
		WHERE %s
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, whereClause), queryArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer rows.Close()

	items := make([]operationLogItem, 0)
	for rows.Next() {
		var item operationLogItem
		if err := rows.Scan(&item.ID, &item.AdminID, &item.AdminName, &item.Module, &item.Action,
			&item.Method, &item.Path, &item.IP, &item.RequestBody, &item.ResponseCode, &item.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    pageResult[operationLogItem]{List: items, Total: total},
	})
}

// ListRoles returns list of roles
func ListRoles(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	limit, offset := parsePageParams(c.Query("page"), c.Query("pageSize"))

	if db != nil && tableExists(ctx, db, "admin_roles") {
		total, err := queryCount(ctx, db, `SELECT COUNT(*) FROM admin_roles WHERE deleted_at IS NULL`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		rows, err := db.QueryContext(ctx, `
			SELECT id, name, code, COALESCE(description, ''), status, sort, created_at
			FROM admin_roles
			WHERE deleted_at IS NULL
			ORDER BY sort ASC, id ASC
			LIMIT ? OFFSET ?`, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		defer rows.Close()
		items := make([]roleItem, 0)
		for rows.Next() {
			var item roleItem
			if err := rows.Scan(&item.ID, &item.Name, &item.Code, &item.Description, &item.Status, &item.Sort, &item.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
				return
			}
			items = append(items, item)
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": pageResult[roleItem]{List: items, Total: total}})
		return
	}

	if db != nil && tableExists(ctx, db, "roles") {
		total, err := queryCount(ctx, db, `SELECT COUNT(*) FROM roles WHERE deleted_at IS NULL`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		rows, err := db.QueryContext(ctx, `
			SELECT id, name, code, COALESCE(description, ''), created_at
			FROM roles
			WHERE deleted_at IS NULL
			ORDER BY id ASC
			LIMIT ? OFFSET ?`, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		defer rows.Close()
		items := make([]roleItem, 0)
		for rows.Next() {
			var item roleItem
			if err := rows.Scan(&item.ID, &item.Name, &item.Code, &item.Description, &item.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
				return
			}
			item.Status = 1
			item.Sort = int(item.ID)
			items = append(items, item)
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": pageResult[roleItem]{List: items, Total: total}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    pageResult[roleItem]{List: []roleItem{}, Total: 0},
	})
}

// GetRole returns a single role
func GetRole(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db == nil || !tableExists(ctx, db, "admin_roles") {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "role not found"})
		return
	}

	var item roleItem
	row := db.QueryRowContext(ctx, `
		SELECT id, name, code, COALESCE(description, ''), status, sort, created_at
		FROM admin_roles WHERE deleted_at IS NULL AND id = ? LIMIT 1`, c.Param("id"))
	if err := row.Scan(&item.ID, &item.Name, &item.Code, &item.Description, &item.Status, &item.Sort, &item.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "role not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": item})
}

// CreateRole creates a new role
func CreateRole(c *gin.Context) {
	var req roleUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid role payload"})
		return
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db == nil || !tableExists(ctx, db, "admin_roles") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	result, err := db.ExecContext(ctx, `
		INSERT INTO admin_roles (name, code, description, status, sort) VALUES (?, ?, ?, ?, ?)`,
		req.Name, req.Code, req.Description, req.Status, req.Sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": roleItem{
		ID: id, Name: req.Name, Code: req.Code, Description: req.Description,
		Status: req.Status, Sort: req.Sort, CreatedAt: time.Now(),
	}})
}

// UpdateRole updates a role
func UpdateRole(c *gin.Context) {
	var req roleUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid role payload"})
		return
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db == nil || !tableExists(ctx, db, "admin_roles") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	result, err := db.ExecContext(ctx, `
		UPDATE admin_roles SET name=?, code=?, description=?, status=?, sort=?, updated_at=?
		WHERE id=? AND deleted_at IS NULL`,
		req.Name, req.Code, req.Description, req.Status, req.Sort, time.Now(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "role not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// DeleteRole deletes a role
func DeleteRole(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db == nil || !tableExists(ctx, db, "admin_roles") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	result, err := db.ExecContext(ctx, `UPDATE admin_roles SET deleted_at=? WHERE id=? AND deleted_at IS NULL`, time.Now(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "role not found"})
		return
	}
	db.ExecContext(ctx, `DELETE FROM admin_role_menus WHERE role_id = ?`, c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": true})
}

// GetRoleMenus returns menu IDs assigned to a role
func GetRoleMenus(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db == nil || !tableExists(ctx, db, "admin_role_menus") {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": []int64{}})
		return
	}

	rows, err := db.QueryContext(ctx, `SELECT menu_id FROM admin_role_menus WHERE role_id = ?`, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		ids = append(ids, id)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": ids})
}

// UpdateRoleMenus sets menu IDs for a role
func UpdateRoleMenus(c *gin.Context) {
	var req roleMenusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid payload"})
		return
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db == nil || !tableExists(ctx, db, "admin_role_menus") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	roleID := c.Param("id")
	db.ExecContext(ctx, `DELETE FROM admin_role_menus WHERE role_id = ?`, roleID)

	for _, menuID := range req.MenuIDs {
		db.ExecContext(ctx, `INSERT INTO admin_role_menus (role_id, menu_id) VALUES (?, ?)`, roleID, menuID)
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// ListMenus returns list of menus
func ListMenus(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db != nil && tableExists(ctx, db, "admin_menus") {
		rows, err := db.QueryContext(ctx, `
			SELECT id, parent_id, name, COALESCE(path, ''), COALESCE(component, ''), COALESCE(icon, ''), type, COALESCE(permission, ''), status, sort
			FROM admin_menus
			WHERE deleted_at IS NULL
			ORDER BY sort ASC, id ASC`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		defer rows.Close()
		items := make([]menuItem, 0)
		for rows.Next() {
			var item menuItem
			if err := rows.Scan(&item.ID, &item.ParentID, &item.Name, &item.Path, &item.Component, &item.Icon, &item.Type, &item.Permission, &item.Status, &item.Sort); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
				return
			}
			items = append(items, item)
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": pageResult[menuItem]{List: items, Total: len(items)}})
		return
	}

	menus := []menuItem{
		{ID: 1, ParentID: 0, Name: "Dashboard", Path: "/dashboard", Icon: "DashboardOutlined", Type: 1, Status: 1, Sort: 1},
		{ID: 2, ParentID: 0, Name: "Admin Management", Path: "/system/admin", Icon: "SettingOutlined", Type: 1, Status: 1, Sort: 2},
		{ID: 3, ParentID: 0, Name: "Role Management", Path: "/system/role", Icon: "SettingOutlined", Type: 1, Status: 1, Sort: 3},
		{ID: 4, ParentID: 0, Name: "Menu Management", Path: "/system/menu", Icon: "SettingOutlined", Type: 1, Status: 1, Sort: 4},
		{ID: 5, ParentID: 0, Name: "User Management", Path: "/user", Icon: "UserOutlined", Type: 1, Status: 1, Sort: 5},
		{ID: 6, ParentID: 0, Name: "Enterprise Management", Path: "/enterprise", Icon: "BankOutlined", Type: 1, Status: 1, Sort: 6},
		{ID: 7, ParentID: 0, Name: "Product Management", Path: "/product", Icon: "ShoppingOutlined", Type: 1, Status: 1, Sort: 7},
		{ID: 8, ParentID: 0, Name: "Category Management", Path: "/product/category", Icon: "ShoppingOutlined", Type: 1, Status: 1, Sort: 8},
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    pageResult[menuItem]{List: menus, Total: len(menus)},
	})
}

// GetMenu returns a single menu
func GetMenu(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db == nil || !tableExists(ctx, db, "admin_menus") {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "menu not found"})
		return
	}

	var item menuItem
	row := db.QueryRowContext(ctx, `
		SELECT id, parent_id, name, COALESCE(path,''), COALESCE(component,''), COALESCE(icon,''), type, COALESCE(permission,''), status, sort
		FROM admin_menus WHERE deleted_at IS NULL AND id = ? LIMIT 1`, c.Param("id"))
	if err := row.Scan(&item.ID, &item.ParentID, &item.Name, &item.Path, &item.Component, &item.Icon, &item.Type, &item.Permission, &item.Status, &item.Sort); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "menu not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": item})
}

// CreateMenu creates a new menu
func CreateMenu(c *gin.Context) {
	var req menuUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid menu payload"})
		return
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db == nil || !tableExists(ctx, db, "admin_menus") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	result, err := db.ExecContext(ctx, `
		INSERT INTO admin_menus (parent_id, name, path, component, icon, type, permission, status, sort)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.ParentID, req.Name, req.Path, req.Component, req.Icon, req.Type, req.Permission, req.Status, req.Sort)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": menuItem{
		ID: id, ParentID: req.ParentID, Name: req.Name, Path: req.Path,
		Component: req.Component, Icon: req.Icon, Type: req.Type,
		Permission: req.Permission, Status: req.Status, Sort: req.Sort,
	}})
}

// UpdateMenu updates a menu
func UpdateMenu(c *gin.Context) {
	var req menuUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid menu payload"})
		return
	}

	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db == nil || !tableExists(ctx, db, "admin_menus") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	result, err := db.ExecContext(ctx, `
		UPDATE admin_menus SET parent_id=?, name=?, path=?, component=?, icon=?, type=?, permission=?, status=?, sort=?, updated_at=?
		WHERE id=? AND deleted_at IS NULL`,
		req.ParentID, req.Name, req.Path, req.Component, req.Icon, req.Type, req.Permission, req.Status, req.Sort, time.Now(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "menu not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// DeleteMenu deletes a menu
func DeleteMenu(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db == nil || !tableExists(ctx, db, "admin_menus") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	result, err := db.ExecContext(ctx, `UPDATE admin_menus SET deleted_at=? WHERE id=? AND deleted_at IS NULL`, time.Now(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "menu not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": true})
}

// ListUsers returns list of users
func ListUsers(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	limit, offset := parsePageParams(c.Query("page"), c.Query("pageSize"))

	total, err := queryCount(ctx, db, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	userCols := columnsExist(ctx, db, "users")
	hasEnterpriseID := userCols["enterprise_id"]
	hasIdentityStatus := userCols["identity_status"]
	hasIDCard := userCols["id_card"]
	hasIDCardFrontURL := userCols["id_card_front_url"]
	hasIDCardBackURL := userCols["id_card_back_url"]
	hasIdentityReject := userCols["identity_reject"]
	query := `
		SELECT id, COALESCE(username, ''), COALESCE(phone, ''), COALESCE(email, ''), status, created_at`
		if hasEnterpriseID {
			query += `, enterprise_id`
		}
		if hasIdentityStatus {
			query += `, COALESCE(identity_status, 0)`
		}
		if hasIDCard {
			query += `, COALESCE(id_card, '')`
		}
		if hasIDCardFrontURL {
		query += `, COALESCE(id_card_front_url, '')`
	}
	if hasIDCardBackURL {
		query += `, COALESCE(id_card_back_url, '')`
	}
	if hasIdentityReject {
		query += `, COALESCE(identity_reject, '')`
	}
	query += ` FROM users WHERE deleted_at IS NULL ORDER BY id DESC LIMIT ? OFFSET ?`

	rows, err := db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer rows.Close()

	items := make([]userItem, 0)
	for rows.Next() {
		var item userItem
		item.RealName = ""
		scanTargets := []any{&item.ID, &item.Username, &item.Phone, &item.Email, &item.Status, &item.CreatedAt}
		if hasEnterpriseID {
			scanTargets = append(scanTargets, &item.EnterpriseID)
		}
		if hasIdentityStatus {
			scanTargets = append(scanTargets, &item.IdentityStatus)
		}
		if hasIDCard {
			scanTargets = append(scanTargets, &item.IDCard)
		}
		if hasIDCardFrontURL {
			scanTargets = append(scanTargets, &item.IDCardFrontURL)
		}
		if hasIDCardBackURL {
			scanTargets = append(scanTargets, &item.IDCardBackURL)
		}
		if hasIdentityReject {
			scanTargets = append(scanTargets, &item.IdentityReject)
		}
		if err := rows.Scan(scanTargets...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		if item.RealName == "" {
			item.RealName = item.Username
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    pageResult[userItem]{List: items, Total: total},
	})
}

// GetUser returns a single user
func GetUser(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id := c.Param("id")
	userCols := columnsExist(ctx, db, "users")
	hasEnterpriseID := userCols["enterprise_id"]
	hasRealName := userCols["real_name"]
	hasIdentityStatus := userCols["identity_status"]
	hasIDCard := userCols["id_card"]
	hasIDCardFrontURL := userCols["id_card_front_url"]
	hasIDCardBackURL := userCols["id_card_back_url"]
	hasIdentityReject := userCols["identity_reject"]

	query := `SELECT id, COALESCE(username, ''), COALESCE(phone, ''), COALESCE(email, ''), status, created_at`
	if hasEnterpriseID {
		query += `, enterprise_id`
	}
	if hasRealName {
		query += `, COALESCE(real_name, '')`
	}
	if hasIdentityStatus {
		query += `, COALESCE(identity_status, 0)`
	}
	if hasIDCard {
		query += `, COALESCE(id_card, '')`
	}
	if hasIDCardFrontURL {
		query += `, COALESCE(id_card_front_url, '')`
	}
	if hasIDCardBackURL {
		query += `, COALESCE(id_card_back_url, '')`
	}
	if hasIdentityReject {
		query += `, COALESCE(identity_reject, '')`
	}
	query += ` FROM users WHERE deleted_at IS NULL AND id = ? LIMIT 1`

	row := db.QueryRowContext(ctx, query, id)

	var item userItem
	var realName sql.NullString
	scanTargets := []any{&item.ID, &item.Username, &item.Phone, &item.Email, &item.Status, &item.CreatedAt}
	if hasEnterpriseID {
		scanTargets = append(scanTargets, &item.EnterpriseID)
	}
	if hasRealName {
		scanTargets = append(scanTargets, &realName)
	}
	if hasIdentityStatus {
		scanTargets = append(scanTargets, &item.IdentityStatus)
	}
	if hasIDCard {
		scanTargets = append(scanTargets, &item.IDCard)
	}
	if hasIDCardFrontURL {
		scanTargets = append(scanTargets, &item.IDCardFrontURL)
	}
	if hasIDCardBackURL {
		scanTargets = append(scanTargets, &item.IDCardBackURL)
	}
	if hasIdentityReject {
		scanTargets = append(scanTargets, &item.IdentityReject)
	}

	if err := row.Scan(scanTargets...); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	if hasRealName && realName.Valid && realName.String != "" {
		item.RealName = realName.String
	} else if item.RealName == "" {
		item.RealName = item.Username
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    item,
	})
}

// UpdateUserStatus updates user status
func UpdateUserStatus(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	id := c.Param("id")

	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if req.Status != 0 && req.Status != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "状态值无效，仅支持 0(停用) 或 1(启用)"})
		return
	}

	result, err := db.ExecContext(ctx, `UPDATE users SET status = ?, updated_at = NOW() WHERE deleted_at IS NULL AND id = ?`, req.Status, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	logOpFromContext(c, db, "user", "update_status", 0, fmt.Sprintf(`{"id":"%s","status":%d}`, id, req.Status))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

func AuditUserIdentity(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db == nil || !columnExists(ctx, db, "users", "identity_status") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	id := c.Param("id")

	var req struct {
		Status int    `json:"status"`
		Reject string `json:"reject"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if req.Status != 2 && req.Status != 3 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "状态值无效，仅支持 2(通过) 或 3(驳回)"})
		return
	}

	var result sql.Result
	var err error
	if req.Status == 2 {
		result, err = db.ExecContext(ctx, `UPDATE users SET identity_status = 2, identity_reject = '', updated_at = NOW() WHERE deleted_at IS NULL AND id = ? AND identity_status = 1`, id)
	} else {
		result, err = db.ExecContext(ctx, `UPDATE users SET identity_status = 3, identity_reject = ?, updated_at = NOW() WHERE deleted_at IS NULL AND id = ? AND identity_status = 1`, req.Reject, id)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户不存在或实名状态不可审核"})
		return
	}

	logOpFromContext(c, db, "user", "audit_identity", 0, fmt.Sprintf(`{"id":"%s","identity_status":%d}`, id, req.Status))
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// ListEnterprises returns list of enterprises
func ListEnterprises(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	limit, offset := parsePageParams(c.Query("page"), c.Query("pageSize"))
	whereParts := []string{"deleted_at IS NULL"}
	args := make([]any, 0)
	if status := c.Query("status"); status != "" {
		statusValue, err := strconv.Atoi(status)
		if err == nil {
			whereParts = append(whereParts, "status = ?")
			args = append(args, statusValue)
		}
	}
	whereClause := strings.Join(whereParts, " AND ")

	total, err := queryCount(ctx, db, fmt.Sprintf(`SELECT COUNT(*) FROM enterprises WHERE %s`, whereClause), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, name, COALESCE(credit_code, ''), status, created_at, verification_info, COALESCE(audit_remark, ''), COALESCE(remark, '')
		FROM enterprises
		WHERE %s
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, whereClause), queryArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer rows.Close()

	items := make([]enterpriseItem, 0)
	for rows.Next() {
		var item enterpriseItem
		var verification sql.NullString
		var createdAt sql.NullTime
		var auditRemark sql.NullString
		var legacyRemark sql.NullString
		if err := rows.Scan(&item.ID, &item.Name, &item.CreditCode, &item.Status, &createdAt, &verification, &auditRemark, &legacyRemark); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time
		}
		info := parseVerificationInfo(verification)
		item.LegalPerson = info.LegalPerson
		item.ContactPerson = info.ContactName
		item.ContactPhone = info.ContactPhone
		item.Address = info.Address
		if auditRemark.Valid && auditRemark.String != "" {
			item.AuditRemark = auditRemark.String
		} else if legacyRemark.Valid {
			item.AuditRemark = parseLegacyAuditRemark(legacyRemark.String)
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    pageResult[enterpriseItem]{List: items, Total: total},
	})
}

// GetEnterprise returns a single enterprise
func GetEnterprise(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var item enterpriseItem
	var verification sql.NullString
	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	var auditedAt sql.NullTime
	var auditRemark sql.NullString
	row := db.QueryRowContext(ctx, `
		SELECT id, name, COALESCE(credit_code, ''), status, COALESCE(audit_remark, ''), audited_at, created_at, updated_at, verification_info
		FROM enterprises
		WHERE deleted_at IS NULL AND id = ?
		LIMIT 1`, c.Param("id"))

	if err := row.Scan(&item.ID, &item.Name, &item.CreditCode, &item.Status, &auditRemark, &auditedAt, &createdAt, &updatedAt, &verification); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "enterprise not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	if createdAt.Valid {
		item.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		item.UpdatedAt = updatedAt.Time
	}
	if auditedAt.Valid {
		item.AuditedAt = &auditedAt.Time
	}
	if auditRemark.Valid {
		item.AuditRemark = auditRemark.String
	}

	info := parseVerificationInfo(verification)
	item.LegalPerson = info.LegalPerson
	item.ContactPerson = info.ContactName
	item.ContactPhone = info.ContactPhone
	item.Address = info.Address
	item.BusinessLicense = info.BusinessLicense

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    item,
	})
}

// AuditEnterprise audits an enterprise
func AuditEnterprise(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	var req struct {
		Status      int    `json:"status" binding:"required"`
		AuditRemark string `json:"auditRemark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid audit payload"})
		return
	}
	if req.Status < 0 || req.Status > 2 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid enterprise status"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	adminID, _ := c.Get("admin_id")
	adminIDValue, _ := adminID.(int64)
	if adminIDValue == 0 {
		switch v := adminID.(type) {
		case int:
			adminIDValue = int64(v)
		case float64:
			adminIDValue = int64(v)
		}
	}

	setParts := []string{"status = ?"}
	args := []any{req.Status}
	if columnExists(ctx, db, "enterprises", "audit_remark") {
		setParts = append(setParts, "audit_remark = ?")
		args = append(args, req.AuditRemark)
	}
	if columnExists(ctx, db, "enterprises", "auditor_id") {
		setParts = append(setParts, "auditor_id = ?")
		args = append(args, adminIDValue)
	}
	if columnExists(ctx, db, "enterprises", "audited_at") {
		setParts = append(setParts, "audited_at = ?")
		args = append(args, time.Now())
	}

	args = append(args, c.Param("id"))
	query := fmt.Sprintf("UPDATE enterprises SET %s WHERE id = ? AND deleted_at IS NULL", strings.Join(setParts, ", "))
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "enterprise not found"})
		return
	}

	var enterpriseCode string
	row := db.QueryRowContext(ctx, `SELECT enterprise_code FROM enterprises WHERE id = ? AND deleted_at IS NULL`, c.Param("id"))
	if err := row.Scan(&enterpriseCode); err == nil && enterpriseCode != "" {
		if req.Status == 1 {
			db.ExecContext(ctx, `UPDATE users SET role = 'enterprise', updated_at = NOW() WHERE enterprise_code = ? AND deleted_at IS NULL`, enterpriseCode)
		} else if req.Status == 2 {
			db.ExecContext(ctx, `UPDATE users SET role = 'personal', updated_at = NOW() WHERE enterprise_code = ? AND deleted_at IS NULL`, enterpriseCode)
		}
	}

	logOpFromContext(c, db, "enterprise", "audit", 0, fmt.Sprintf(`{"id":"%s","status":%d}`, c.Param("id"), req.Status))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "审核完成",
		"data":    true,
	})
}

// ListProducts returns list of products
func ListProducts(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	limit, offset := parsePageParams(c.Query("page"), c.Query("pageSize"))
	whereParts := []string{"deleted_at IS NULL"}
	args := make([]any, 0)
	if status := c.Query("status"); status != "" {
		statusValue, err := strconv.Atoi(status)
		if err == nil {
			switch statusValue {
			case 0:
				whereParts = append(whereParts, "status IN (?, ?)")
				args = append(args, 0, 1)
			case 1:
				whereParts = append(whereParts, "status = ?")
				args = append(args, 2)
			case 2:
				whereParts = append(whereParts, "status = ?")
				args = append(args, 4)
			}
		}
	}
	whereClause := strings.Join(whereParts, " AND ")

	total, err := queryCount(ctx, db, fmt.Sprintf(`SELECT COUNT(*) FROM products WHERE %s`, whereClause), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, name, enterprise_id, COALESCE(category, ''), COALESCE(sub_category, ''), COALESCE(steel_mill, ''),
		       spec, warehouse_info, COALESCE(bundle_no, ''), COALESCE(product_no, ''),
		       COALESCE(unit_price, 0), COALESCE(quantity, 0), COALESCE(min_order_qty, 0),
		       COALESCE(delivery_days, 0), COALESCE(quote_validity, 0),
		       COALESCE(description, ''), images, status, created_at, COALESCE(audit_remark, ''), approved_at, updated_at
		FROM products
		WHERE %s
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, whereClause), queryArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer rows.Close()

	items := make([]productItem, 0)
	for rows.Next() {
		var item productItem
		var imageJSON sql.NullString
		var rawSpec sql.NullString
		var rawWarehouse sql.NullString
		var rawStatus int
		var auditRemark sql.NullString
		var auditedAt sql.NullTime
		var updatedAt sql.NullTime
		var productNo string
		if err := rows.Scan(&item.ID, &item.Name, &item.EnterpriseID, &item.Category, &item.SubCategory, &item.SteelMill,
			&rawSpec, &rawWarehouse, &item.BundleNo, &productNo,
			&item.Price, &item.Stock, &item.MinOrderQty,
			&item.DeliveryDays, &item.QuoteValidity,
			&item.Description, &imageJSON, &rawStatus, &item.CreatedAt, &auditRemark, &auditedAt, &updatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		item.Images = parseImages(imageJSON)
		item.Status = mapProductIntStatus(rawStatus)
		specInfo := parseProductSpec(rawSpec)
		item.Spec = specInfo.Spec
		item.Material = specInfo.Material
		warehouse := parseWarehouseInfo(rawWarehouse)
		item.WarehouseName = warehouse.WarehouseName
		item.WarehouseCode = warehouse.WarehouseCode
		item.Province = warehouse.Province
		item.City = warehouse.City
		item.District = warehouse.District
		item.Address = warehouse.Address
		if auditRemark.Valid {
			item.AuditRemark = auditRemark.String
		}
		if auditedAt.Valid {
			item.AuditedAt = &auditedAt.Time
		}
		if updatedAt.Valid {
			item.UpdatedAt = updatedAt.Time
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    pageResult[productItem]{List: items, Total: total},
	})
}

// ListOrders returns list of orders with search support
func ListOrders(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	limit, offset := parsePageParams(c.Query("page"), c.Query("pageSize"))

	where := "deleted_at IS NULL"
	args := make([]any, 0)

	if keyword := c.Query("keyword"); keyword != "" {
		where += " AND (order_no LIKE ? OR COALESCE(buyer_name,'') LIKE ? OR COALESCE(seller_name,'') LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw)
	}
	if buyerID := c.Query("buyerId"); buyerID != "" {
		where += " AND buyer_id = ?"
		args = append(args, buyerID)
	}
	if sellerID := c.Query("sellerId"); sellerID != "" {
		where += " AND (seller_id = ? OR seller_enterprise_id = ?)"
		args = append(args, sellerID, sellerID)
	}
	if status := c.Query("status"); status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}

	countArgs := make([]any, len(args))
	copy(countArgs, args)
	total, err := queryCount(ctx, db, fmt.Sprintf(`SELECT COUNT(*) FROM orders WHERE %s`, where), countArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	queryArgs := append(args, limit, offset)
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, order_no, buyer_id, COALESCE(buyer_name, ''), seller_id, COALESCE(seller_enterprise_id, 0),
		       COALESCE(seller_name, ''), COALESCE(total_quantity, 0), COALESCE(total_amount, 0),
		       COALESCE(tax_total_amount, 0), COALESCE(currency, 'CNY'), status,
		       COALESCE(payment_info, ''),
		       COALESCE(delivery_voucher, ''), COALESCE(receipt_voucher, ''), receipt_at, expires_at, created_at
		FROM orders
		WHERE %s
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, where), queryArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer rows.Close()

	items := make([]orderItem, 0)
	for rows.Next() {
		var item orderItem
		if err := rows.Scan(
			&item.ID,
			&item.OrderNo,
			&item.BuyerID,
			&item.BuyerName,
			&item.SellerID,
			&item.SellerEnterpriseID,
			&item.SellerName,
			&item.TotalQuantity,
			&item.TotalAmount,
			&item.TaxTotalAmount,
			&item.Currency,
			&item.Status,
			&item.PaymentInfoJSON,
			&item.DeliveryVoucher,
			&item.ReceiptVoucher,
			&item.ReceiptAt,
			&item.ExpiresAt,
			&item.CreatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		item.Timeline = buildOrderTimeline(item)
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    pageResult[orderItem]{List: items, Total: total},
	})
}

func effectiveOrderStatus(dbStatus int, paymentInfoJSON sql.NullString) int {
	if dbStatus != 3 {
		return dbStatus
	}
	if !paymentInfoJSON.Valid || paymentInfoJSON.String == "" {
		return dbStatus
	}
	var pi struct {
		VoucherURL string `json:"voucher_url"`
	}
	if json.Unmarshal([]byte(paymentInfoJSON.String), &pi) == nil && pi.VoucherURL != "" {
		return 4
	}
	return dbStatus
}

func buildOrderTimeline(order orderItem) []orderTimelineItem {
	timeline := []orderTimelineItem{
		{
			Key:         "created",
			Label:       "订单创建",
			Status:      "completed",
			Time:        &order.CreatedAt,
			Description: "买家提交订单，系统锁定库存。",
		},
	}

	if order.Status == 9 {
		timeline = append(timeline,
			orderTimelineItem{Key: "payment_uploaded", Label: "支付凭证上传", Status: "skipped", Description: "订单已关闭，未完成支付。"},
			orderTimelineItem{Key: "closed", Label: "订单关闭", Status: "completed", Time: order.ExpiresAt, Description: "订单已关闭，锁定库存已释放。"},
		)
		return timeline
	}

	paymentStatus := "pending"
	if order.Status == 4 {
		paymentStatus = "current"
	} else if order.Status >= 5 {
		paymentStatus = "completed"
	}
	timeline = append(timeline, orderTimelineItem{
		Key:         "payment_uploaded",
		Label:       "支付凭证上传",
		Status:      paymentStatus,
		Description: "买家上传支付凭证后，卖家确认收款。",
	})

	confirmStatus := "pending"
	if order.Status == 4 {
		confirmStatus = "current"
	} else if order.Status >= 5 {
		confirmStatus = "completed"
	}
	timeline = append(timeline, orderTimelineItem{
		Key:         "payment_confirmed",
		Label:       "卖家确认收款",
		Status:      confirmStatus,
		Description: "卖家核对支付凭证并确认收款。",
	})

	pickupStatus := "pending"
	if order.Status == 5 {
		pickupStatus = "current"
	} else if order.Status >= 6 {
		pickupStatus = "completed"
	}
	timeline = append(timeline, orderTimelineItem{
		Key:         "pickup_info",
		Label:       "提货信息录入",
		Status:      pickupStatus,
		Description: "买家录入承运车辆、司机和送达地址。",
	})

	outboundStatus := "pending"
	if order.Status == 5 {
		outboundStatus = "current"
	} else if order.Status >= 6 {
		outboundStatus = "completed"
	}
	timeline = append(timeline, orderTimelineItem{
		Key:         "outbound",
		Label:       "卖家确认出库",
		Status:      outboundStatus,
		Description: "卖家核验提货码并确认货物出库。",
	})

	receiptStatus := "pending"
	if order.Status == 6 || order.Status == 7 {
		receiptStatus = "current"
	} else if order.Status == 8 {
		receiptStatus = "completed"
	}
	timeline = append(timeline, orderTimelineItem{
		Key:         "receipt",
		Label:       "买家确认收货",
		Status:      receiptStatus,
		Time:        order.ReceiptAt,
		Description: "买家确认收货并上传签收凭证后，订单完成归档。",
	})

	return timeline
}

// GetProduct returns a single product
func GetProduct(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var item productItem
	var imageJSON sql.NullString
	var rawStatus int
	var auditRemark sql.NullString
	var auditedAt sql.NullTime
	var updatedAt sql.NullTime

	row := db.QueryRowContext(ctx, `
		SELECT id, name, enterprise_id, COALESCE(category, ''), COALESCE(sub_category, ''), COALESCE(steel_mill, ''),
		       spec, warehouse_info, COALESCE(bundle_no, ''), COALESCE(product_no, ''),
		       COALESCE(unit_price, 0), COALESCE(quantity, 0), COALESCE(min_order_qty, 0),
		       COALESCE(delivery_days, 0), COALESCE(quote_validity, 0),
		       COALESCE(description, ''), images, status, COALESCE(audit_remark, ''), approved_at, created_at, updated_at
		FROM products
		WHERE deleted_at IS NULL AND id = ?
		LIMIT 1`, c.Param("id"))

	var rawSpec sql.NullString
	var rawWarehouse sql.NullString
	var productNo string
	if err := row.Scan(
		&item.ID,
		&item.Name,
		&item.EnterpriseID,
		&item.Category,
		&item.SubCategory,
		&item.SteelMill,
		&rawSpec,
		&rawWarehouse,
		&item.BundleNo,
		&productNo,
		&item.Price,
		&item.Stock,
		&item.MinOrderQty,
		&item.DeliveryDays,
		&item.QuoteValidity,
		&item.Description,
		&imageJSON,
		&rawStatus,
		&auditRemark,
		&auditedAt,
		&item.CreatedAt,
		&updatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	item.Images = parseImages(imageJSON)
	item.Status = mapProductIntStatus(rawStatus)
	specInfo := parseProductSpec(rawSpec)
	item.Spec = specInfo.Spec
	item.Material = specInfo.Material
	warehouse := parseWarehouseInfo(rawWarehouse)
	item.WarehouseName = warehouse.WarehouseName
	item.WarehouseCode = warehouse.WarehouseCode
	item.Province = warehouse.Province
	item.City = warehouse.City
	item.District = warehouse.District
	item.Address = warehouse.Address
	if auditRemark.Valid {
		item.AuditRemark = auditRemark.String
	}
	if auditedAt.Valid {
		item.AuditedAt = &auditedAt.Time
	}
	if updatedAt.Valid {
		item.UpdatedAt = updatedAt.Time
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    item,
	})
}

// AuditProduct audits a product
func AuditProduct(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	var req struct {
		Status      int    `json:"status" binding:"required"`
		AuditRemark string `json:"auditRemark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid audit payload"})
		return
	}
	if req.Status < 0 || req.Status > 2 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid product status"})
		return
	}

	statusValue := 1
	logStatus := "pending_review"
	switch req.Status {
	case 1:
		statusValue = 2
		logStatus = "approved"
	case 2:
		statusValue = 4
		logStatus = "rejected"
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	adminID, _ := c.Get("admin_id")
	adminUsername, _ := c.Get("admin_username")
	adminName := fmt.Sprint(adminUsername)
	operatorID := int64(0)
	switch v := adminID.(type) {
	case int64:
		operatorID = v
	case int:
		operatorID = int64(v)
	case float64:
		operatorID = int64(v)
	}

	setParts := []string{"status = ?"}
	args := []any{statusValue}
	if columnExists(ctx, db, "products", "audit_remark") {
		setParts = append(setParts, "audit_remark = ?")
		args = append(args, req.AuditRemark)
	}
	if columnExists(ctx, db, "products", "approved_at") {
		if req.Status == 1 {
			setParts = append(setParts, "approved_at = ?")
			args = append(args, time.Now())
		} else {
			setParts = append(setParts, "approved_at = NULL")
		}
	}
	if columnExists(ctx, db, "products", "updated_at") {
		setParts = append(setParts, "updated_at = ?")
		args = append(args, time.Now())
	}

	args = append(args, c.Param("id"))
	query := fmt.Sprintf("UPDATE products SET %s WHERE id = ? AND deleted_at IS NULL", strings.Join(setParts, ", "))
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "product not found"})
		return
	}

	var currentStatus int
	if err := db.QueryRowContext(ctx, `SELECT status FROM products WHERE id = ? LIMIT 1`, c.Param("id")).Scan(&currentStatus); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if currentStatus != statusValue {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "product audit status sync failed"})
		return
	}

	if tableExists(ctx, db, "product_audit_logs") {
		var productNo string
		productID := int64(0)
		fmt.Sscanf(c.Param("id"), "%d", &productID)
		_ = db.QueryRowContext(ctx, `SELECT COALESCE(product_no, '') FROM products WHERE id = ? LIMIT 1`, c.Param("id")).Scan(&productNo)
		if _, err := db.ExecContext(ctx, `
			INSERT INTO product_audit_logs (product_id, product_sku_id, action, status, operator_id, operator_name, remark, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			productID,
			productNo,
			"platform_audit",
			logStatus,
			operatorID,
			adminName,
			req.AuditRemark,
			time.Now(),
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
	}

	logOpFromContext(c, db, "product", "audit", 0, fmt.Sprintf(`{"id":"%s","status":%d}`, c.Param("id"), req.Status))

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ListProductCategories returns list of product categories
func UpdateProduct(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	id := c.Param("id")

	var req struct {
		Name          string  `json:"name"`
		Category      string  `json:"category"`
		SubCategory   string  `json:"subCategory"`
		SteelMill     string  `json:"steelMill"`
		UnitPrice     float64 `json:"unitPrice"`
		Quantity      float64 `json:"quantity"`
		MinOrderQty   float64 `json:"minOrderQty"`
		DeliveryDays  int     `json:"deliveryDays"`
		QuoteValidity int     `json:"quoteValidity"`
		Description   string  `json:"description"`
		AuditRemark   string  `json:"auditRemark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	setParts := []string{}
	args := []any{}

	if req.Name != "" {
		setParts = append(setParts, "name = ?")
		args = append(args, req.Name)
	}
	if req.Category != "" {
		setParts = append(setParts, "category = ?")
		args = append(args, req.Category)
	}
	if req.SubCategory != "" {
		setParts = append(setParts, "sub_category = ?")
		args = append(args, req.SubCategory)
	}
	if req.SteelMill != "" {
		setParts = append(setParts, "steel_mill = ?")
		args = append(args, req.SteelMill)
	}
	if req.UnitPrice > 0 {
		setParts = append(setParts, "unit_price = ?")
		args = append(args, req.UnitPrice)
	}
	if req.Quantity > 0 {
		setParts = append(setParts, "quantity = ?")
		args = append(args, req.Quantity)
	}
	if req.MinOrderQty > 0 {
		setParts = append(setParts, "min_order_qty = ?")
		args = append(args, req.MinOrderQty)
	}
	if req.DeliveryDays > 0 {
		setParts = append(setParts, "delivery_days = ?")
		args = append(args, req.DeliveryDays)
	}
	if req.QuoteValidity > 0 {
		setParts = append(setParts, "quote_validity = ?")
		args = append(args, req.QuoteValidity)
	}
	if req.Description != "" {
		setParts = append(setParts, "description = ?")
		args = append(args, req.Description)
	}
	if req.AuditRemark != "" {
		if columnExists(ctx, db, "products", "audit_remark") {
			setParts = append(setParts, "audit_remark = ?")
			args = append(args, req.AuditRemark)
		}
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "没有需要更新的字段"})
		return
	}

	setParts = append(setParts, "updated_at = ?")
	args = append(args, time.Now())
	args = append(args, id)

	query := fmt.Sprintf("UPDATE products SET %s WHERE deleted_at IS NULL AND id = ?", strings.Join(setParts, ", "))
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "商品不存在"})
		return
	}

	logOpFromContext(c, db, "product", "update", 0, fmt.Sprintf(`{"id":"%s"}`, id))
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": true})
}

func DeleteProduct(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db == nil || !tableExists(ctx, db, "products") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	result, err := db.ExecContext(ctx, `UPDATE products SET deleted_at = ?, status = 4, updated_at = ? WHERE deleted_at IS NULL AND id = ?`, time.Now(), time.Now(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "商品不存在"})
		return
	}

	logOpFromContext(c, db, "product", "delete", 0, fmt.Sprintf(`{"id":"%s"}`, c.Param("id")))
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": true})
}

func ListProductCategories(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if db != nil && tableExists(ctx, db, "product_categories") {
		rows, err := db.QueryContext(ctx, `
			SELECT id, COALESCE(parent_id, 0), name, COALESCE(code, ''), COALESCE(icon, ''), COALESCE(sort, 0), COALESCE(status, 1)
			FROM product_categories
			WHERE deleted_at IS NULL
			ORDER BY sort ASC, id ASC`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		defer rows.Close()
		items := make([]categoryItem, 0)
		for rows.Next() {
			var item categoryItem
			if err := rows.Scan(&item.ID, &item.ParentID, &item.Name, &item.Code, &item.Icon, &item.Sort, &item.Status); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
				return
			}
			items = append(items, item)
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": pageResult[categoryItem]{List: items, Total: len(items)}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    pageResult[categoryItem]{List: []categoryItem{}, Total: 0},
	})
}

// CreateProductCategory creates a new product category
func CreateProductCategory(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		Code   string `json:"code"`
		Sort   int    `json:"sort"`
		Status int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid payload"})
		return
	}
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if db == nil || !tableExists(ctx, db, "product_categories") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}
	result, err := db.ExecContext(ctx, `INSERT INTO product_categories (name, code, sort, status) VALUES (?, ?, ?, ?)`, req.Name, req.Code, req.Sort, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": categoryItem{ID: id, Name: req.Name, Code: req.Code, Sort: req.Sort, Status: req.Status}})
}

// UpdateProductCategory updates a product category
func UpdateProductCategory(c *gin.Context) {
	var req struct {
		Name   string `json:"name" binding:"required"`
		Code   string `json:"code"`
		Sort   int    `json:"sort"`
		Status int    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid payload"})
		return
	}
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if db == nil || !tableExists(ctx, db, "product_categories") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}
	result, err := db.ExecContext(ctx, `UPDATE product_categories SET name=?, code=?, sort=?, status=? WHERE id=? AND deleted_at IS NULL`, req.Name, req.Code, req.Sort, req.Status, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "category not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// DeleteProductCategory deletes a product category
func DeleteProductCategory(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if db == nil || !tableExists(ctx, db, "product_categories") {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}
	result, err := db.ExecContext(ctx, `UPDATE product_categories SET deleted_at=? WHERE id=? AND deleted_at IS NULL`, time.Now(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "category not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": true})
}
