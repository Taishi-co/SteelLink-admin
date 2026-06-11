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
	pendingProducts, _ := queryCount(ctx, db, `SELECT COUNT(*) FROM products WHERE deleted_at IS NULL AND status = 1`)
	pendingSupplyDemands := 0
	if tableExists(ctx, db, "supply_demand_posts") {
		pendingSupplyDemands, _ = queryCount(ctx, db, `SELECT COUNT(*) FROM supply_demand_posts WHERE deleted_at IS NULL AND status = 4`)
	}
	pendingOrders, _ := queryCount(ctx, db, `SELECT COUNT(*) FROM orders WHERE deleted_at IS NULL AND status IN (2, 3, 5, 12, 13, 14, 15, 22, 23, 24, 25, 26, 27)`)
	stats.PendingAudits = pendingEnterprises + pendingProducts + pendingSupplyDemands + pendingOrders

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
	if intRoleID != 1 && intRoleID != 2 && intRoleID != 3 {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": []interface{}{}})
		return
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
	if !rejectReasonProvided(req.Status, 3, req.Reject) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "驳回原因不能为空"})
		return
	}

	var result sql.Result
	var err error
	if req.Status == 2 {
		result, err = db.ExecContext(ctx, `UPDATE users SET identity_status = 2, identity_reject = '', updated_at = NOW() WHERE deleted_at IS NULL AND id = ? AND identity_status = 1`, id)
	} else {
		result, err = db.ExecContext(ctx, `UPDATE users SET identity_status = 3, identity_reject = ?, updated_at = NOW() WHERE deleted_at IS NULL AND id = ? AND identity_status = 1`, strings.TrimSpace(req.Reject), id)
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
	if !isEnterpriseAuditDecision(req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid enterprise audit decision"})
		return
	}
	if !rejectReasonProvided(req.Status, 2, req.AuditRemark) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "驳回原因不能为空"})
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
		args = append(args, strings.TrimSpace(req.AuditRemark))
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
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		kw := "%" + keyword + "%"
		whereParts = append(whereParts, `(name LIKE ? OR COALESCE(category, '') LIKE ? OR COALESCE(sub_category, '') LIKE ? OR COALESCE(steel_mill, '') LIKE ? OR COALESCE(bundle_no, '') LIKE ? OR COALESCE(product_no, '') LIKE ? OR COALESCE(description, '') LIKE ?)`)
		args = append(args, kw, kw, kw, kw, kw, kw, kw)
	}
	if status := c.Query("status"); status != "" {
		statusValue, err := strconv.Atoi(status)
		if err == nil && statusValue >= 0 && statusValue <= 4 {
			whereParts = append(whereParts, "status = ?")
			args = append(args, statusValue)
		}
	}
	if enterpriseID := c.Query("enterpriseId"); enterpriseID == "" {
		if enterpriseID = c.Query("enterprise_id"); enterpriseID != "" {
			whereParts = append(whereParts, "enterprise_id = ?")
			args = append(args, enterpriseID)
		}
	} else {
		whereParts = append(whereParts, "enterprise_id = ?")
		args = append(args, enterpriseID)
	}
	if category := strings.TrimSpace(c.Query("category")); category != "" {
		whereParts = append(whereParts, "(category = ? OR sub_category = ?)")
		args = append(args, category, category)
	}
	if subCategory := strings.TrimSpace(c.Query("subCategory")); subCategory != "" {
		whereParts = append(whereParts, "sub_category = ?")
		args = append(args, subCategory)
	}
	if goodsType := strings.TrimSpace(c.Query("goodsType")); goodsType != "" && columnExists(ctx, db, "products", "goods_type") {
		whereParts = append(whereParts, "goods_type = ?")
		args = append(args, goodsType)
	}
	if warehouseServiceID := c.Query("warehouseServiceId"); warehouseServiceID == "" {
		warehouseServiceID = c.Query("warehouse_service_id")
		if warehouseServiceID != "" && columnExists(ctx, db, "products", "warehouse_service_id") {
			whereParts = append(whereParts, "warehouse_service_id = ?")
			args = append(args, warehouseServiceID)
		}
	} else if columnExists(ctx, db, "products", "warehouse_service_id") {
		whereParts = append(whereParts, "warehouse_service_id = ?")
		args = append(args, warehouseServiceID)
	}
	whereClause := strings.Join(whereParts, " AND ")
	goodsTypeExpr := "'spot'"
	if columnExists(ctx, db, "products", "goods_type") {
		goodsTypeExpr = "COALESCE(goods_type, 'spot')"
	}
	warehouseServiceIDExpr := "0"
	if columnExists(ctx, db, "products", "warehouse_service_id") {
		warehouseServiceIDExpr = "COALESCE(warehouse_service_id, 0)"
	}
	warehouseEnterpriseIDExpr := "0"
	if columnExists(ctx, db, "products", "warehouse_enterprise_id") {
		warehouseEnterpriseIDExpr = "COALESCE(warehouse_enterprise_id, 0)"
	}
	auditorIDExpr := "0"
	if columnExists(ctx, db, "products", "auditor_id") {
		auditorIDExpr = "COALESCE(auditor_id, 0)"
	}
	availableQtyExpr := "COALESCE(quantity, 0)"
	if columnExists(ctx, db, "products", "available_qty") {
		availableQtyExpr = "COALESCE(available_qty, quantity, 0)"
	}

	total, err := queryCount(ctx, db, fmt.Sprintf(`SELECT COUNT(*) FROM products WHERE %s`, whereClause), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, name, enterprise_id, COALESCE(category, ''), COALESCE(sub_category, ''), COALESCE(steel_mill, ''),
		       %s, %s, %s,
		       spec, warehouse_info, COALESCE(bundle_no, ''), COALESCE(product_no, ''),
		       COALESCE(unit, ''), COALESCE(unit_price, 0), COALESCE(quantity, 0), %s, COALESCE(min_order_qty, 0),
		       COALESCE(delivery_days, 0), COALESCE(quote_validity, 0),
		       COALESCE(description, ''), images, status, created_at, COALESCE(audit_remark, ''), %s, approved_at, updated_at
		FROM products
		WHERE %s
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, goodsTypeExpr, warehouseServiceIDExpr, warehouseEnterpriseIDExpr, availableQtyExpr, auditorIDExpr, whereClause), queryArgs...)
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
		var availableQty float64
		if err := rows.Scan(&item.ID, &item.Name, &item.EnterpriseID, &item.Category, &item.SubCategory, &item.SteelMill,
			&item.GoodsType, &item.WarehouseServiceID, &item.WarehouseEnterpriseID,
			&rawSpec, &rawWarehouse, &item.BundleNo, &item.ProductNo,
			&item.Unit, &item.Price, &item.Stock, &availableQty, &item.MinOrderQty,
			&item.DeliveryDays, &item.QuoteValidity,
			&item.Description, &imageJSON, &rawStatus, &item.CreatedAt, &auditRemark, &item.AuditorID, &auditedAt, &updatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		item.AvailableQty = &availableQty
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
		if item.WarehouseServiceID == 0 {
			item.WarehouseServiceID = warehouse.WarehouseServiceID
		}
		if item.WarehouseEnterpriseID == 0 {
			item.WarehouseEnterpriseID = warehouse.WarehouseEnterpriseID
		}
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
	if c.Query("intervention") == "1" {
		interventionWhere := "status IN (10, 11, 24, 25)"
		if tableExists(ctx, db, "breach_records") {
			interventionWhere += " OR EXISTS (SELECT 1 FROM breach_records br WHERE br.order_id = orders.id AND br.deleted_at IS NULL AND br.status IN ('submitted', 'under_review', 'mediating', 'litigating'))"
		}
		if tableExists(ctx, db, "litigation_records") {
			interventionWhere += " OR EXISTS (SELECT 1 FROM litigation_records lr WHERE lr.order_id = orders.id AND lr.deleted_at IS NULL AND lr.status IN ('filed', 'accepted', 'in_progress'))"
		}
		where += " AND (" + interventionWhere + ")"
	}

	countArgs := make([]any, len(args))
	copy(countArgs, args)
	total, err := queryCount(ctx, db, fmt.Sprintf(`SELECT COUNT(*) FROM orders WHERE %s`, where), countArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	remarkExpr := "NULL"
	if columnExists(ctx, db, "orders", "remark") {
		remarkExpr = "remark"
	}
	queryArgs := append(args, limit, offset)
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, order_no, buyer_id, COALESCE(buyer_name, ''), seller_id, COALESCE(seller_enterprise_id, 0),
		       COALESCE(seller_name, ''), COALESCE(total_quantity, 0), COALESCE(total_amount, 0),
		       COALESCE(tax_total_amount, 0), COALESCE(currency, 'CNY'), status,
		       COALESCE(delivery_voucher, ''), COALESCE(receipt_voucher, ''), receipt_at, expires_at, created_at, %s
		FROM orders
		WHERE %s
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, remarkExpr, where), queryArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer rows.Close()

	items := make([]orderItem, 0)
	for rows.Next() {
		item, err := scanOrder(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		if err := fillOrderTerminationSummary(ctx, db, &item); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    pageResult[orderItem]{List: items, Total: total},
	})
}

type sqlScanner interface {
	Scan(dest ...any) error
}

func scanOrder(scanner sqlScanner) (orderItem, error) {
	var item orderItem
	var rawRemark sql.NullString
	err := scanner.Scan(
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
		&item.DeliveryVoucher,
		&item.ReceiptVoucher,
		&item.ReceiptAt,
		&item.ExpiresAt,
		&item.CreatedAt,
		&rawRemark,
	)
	if err != nil {
		return item, err
	}
	item.Remark = parseJSONMap(rawRemark)
	item.Timeline = buildOrderTimeline(item)
	return item, nil
}

// GetOrder returns one order with fulfillment and operation detail.
func GetOrder(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid order id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	remarkExpr := "NULL"
	if columnExists(ctx, db, "orders", "remark") {
		remarkExpr = "remark"
	}
	row := db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT id, order_no, buyer_id, COALESCE(buyer_name, ''), seller_id, COALESCE(seller_enterprise_id, 0),
		       COALESCE(seller_name, ''), COALESCE(total_quantity, 0), COALESCE(total_amount, 0),
		       COALESCE(tax_total_amount, 0), COALESCE(currency, 'CNY'), status,
		       COALESCE(delivery_voucher, ''), COALESCE(receipt_voucher, ''), receipt_at, expires_at, created_at, %s
		FROM orders
		WHERE deleted_at IS NULL AND id = ?
		LIMIT 1`, remarkExpr), id)

	item, err := scanOrder(row)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if err := fillOrderDetail(ctx, db, &item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": item})
}

func fillOrderDetail(ctx context.Context, db *sql.DB, item *orderItem) error {
	if tableExists(ctx, db, "order_events") {
		events, err := listOrderEvents(ctx, db, item.ID)
		if err != nil {
			return err
		}
		item.Events = events
	}
	if tableExists(ctx, db, "order_financing_requests") {
		info, err := getOrderFinancingInfo(ctx, db, item.ID)
		if err != nil {
			return err
		}
		item.FinancingInfo = info
	}
	if tableExists(ctx, db, "logistics_biddings") {
		info, err := getLogisticsBiddingInfo(ctx, db, item.ID)
		if err != nil {
			return err
		}
		item.LogisticsBidding = info
	}
	if tableExists(ctx, db, "logistics_bid_quotes") {
		quotes, err := listLogisticsQuotes(ctx, db, item.ID)
		if err != nil {
			return err
		}
		item.LogisticsQuotes = quotes
	}
	if tableExists(ctx, db, "shipment_tasks") {
		task, err := getShipmentTaskInfo(ctx, db, item.ID)
		if err != nil {
			return err
		}
		item.ShipmentTask = task
	}
	if tableExists(ctx, db, "pledge_release_events") {
		events, err := listPledgeEvents(ctx, db, item.ID)
		if err != nil {
			return err
		}
		item.PledgeEvents = events
	}
	if tableExists(ctx, db, "breach_records") {
		records, err := listBreachRecords(ctx, db, item.ID)
		if err != nil {
			return err
		}
		item.BreachRecords = records
	}
	if tableExists(ctx, db, "litigation_records") {
		records, err := listLitigationRecords(ctx, db, item.ID)
		if err != nil {
			return err
		}
		item.LitigationRecords = records
	}
	item.TerminationReason = buildOrderTerminationReason(*item)
	return nil
}

func fillOrderTerminationSummary(ctx context.Context, db *sql.DB, item *orderItem) error {
	summary := *item
	if tableExists(ctx, db, "order_events") {
		events, err := listOrderTerminationEvents(ctx, db, item.ID)
		if err != nil {
			return err
		}
		summary.Events = events
	}
	if tableExists(ctx, db, "order_financing_requests") {
		info, err := getOrderFinancingInfo(ctx, db, item.ID)
		if err != nil {
			return err
		}
		summary.FinancingInfo = info
	}
	item.TerminationReason = buildOrderTerminationReason(summary)
	return nil
}

func buildOrderTerminationReason(order orderItem) *terminationReason {
	needsIntervention := truthyValue(order.Remark["admin_attention"])
	if event := latestMatchingOrderEvent(order.Events, isLogisticsNoCapacityEvent); order.Status == 25 || event != nil || mapString(order.Remark, "logistics_no_capacity_reason", "logisticsNoCapacityReason") != "" {
		return &terminationReason{
			Code:              "logistics_no_capacity",
			Label:             "物流无运力-已终止",
			Party:             "logistics",
			PartyLabel:        "物流方",
			Reason:            firstNonEmpty(mapString(order.Remark, "logistics_no_capacity_reason", "logisticsNoCapacityReason"), eventPayloadString(event, "reason", "note", "message"), "无合规物流可承接"),
			Source:            "logistics",
			EventType:         eventTypeOf(event),
			NeedsIntervention: needsIntervention,
			CreatedAt:         eventCreatedAt(event),
		}
	}
	if event := latestMatchingOrderEvent(order.Events, isFinancingRiskRejectedEvent); financingInfoRejected(order.FinancingInfo) || event != nil {
		return &terminationReason{
			Code:              "financing_risk_rejected",
			Label:             "资金驳回-已终止",
			Party:             "funding",
			PartyLabel:        "资金方",
			Reason:            firstNonEmpty(mapString(order.FinancingInfo, "risk_remark", "riskRemark"), eventPayloadString(event, "reason", "note", "message"), "风控不通过或额度不足"),
			Source:            "financing",
			EventType:         eventTypeOf(event),
			NeedsIntervention: needsIntervention,
			CreatedAt:         eventCreatedAt(event),
		}
	}
	if event := latestMatchingOrderEvent(order.Events, isSellerRejectedEvent); event != nil || mapString(order.Remark, "seller_reject_reason", "sellerRejectReason", "seller_rejected_reason", "sellerRejectedReason", "no_stock_reason", "noStockReason") != "" {
		return &terminationReason{
			Code:              "seller_rejected",
			Label:             "卖方拒单-已终止",
			Party:             "seller",
			PartyLabel:        "卖方",
			Reason:            firstNonEmpty(mapString(order.Remark, "seller_reject_reason", "sellerRejectReason", "seller_rejected_reason", "sellerRejectedReason", "no_stock_reason", "noStockReason"), eventPayloadString(event, "reason", "note", "message"), "卖方拒绝确认订单"),
			Source:            "order_event",
			EventType:         eventTypeOf(event),
			NeedsIntervention: needsIntervention,
			CreatedAt:         eventCreatedAt(event),
		}
	}
	if event := latestMatchingOrderEvent(order.Events, isBuyerCancelledEvent); event != nil || mapString(order.Remark, "buyer_cancel_reason", "buyerCancelReason") != "" {
		return &terminationReason{
			Code:              "buyer_cancelled",
			Label:             "买方取消-已终止",
			Party:             "buyer",
			PartyLabel:        "买方",
			Reason:            firstNonEmpty(mapString(order.Remark, "buyer_cancel_reason", "buyerCancelReason", "cancel_reason", "cancelReason", "cancellation_reason", "cancellationReason"), eventPayloadString(event, "reason", "note", "message"), "买方取消订单"),
			Source:            "order_event",
			EventType:         eventTypeOf(event),
			NeedsIntervention: needsIntervention,
			CreatedAt:         eventCreatedAt(event),
		}
	}
	if order.Status == 9 {
		return cancelledOrderTerminationReason(order, needsIntervention)
	}
	return nil
}

func listOrderEvents(ctx context.Context, db *sql.DB, orderID int64) ([]orderEventItem, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, order_id, COALESCE(order_no, ''), event_type, COALESCE(actor_user_id, 0),
		       COALESCE(actor_enterprise_id, 0), COALESCE(actor_service, ''), COALESCE(from_status, 0),
		       COALESCE(to_status, 0), payload, created_at
		FROM order_events
		WHERE order_id = ?
		ORDER BY id DESC
		LIMIT 80`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]orderEventItem, 0)
	for rows.Next() {
		var item orderEventItem
		var payload sql.NullString
		if err := rows.Scan(&item.ID, &item.OrderID, &item.OrderNo, &item.EventType, &item.ActorUserID, &item.ActorEnterpriseID, &item.ActorService, &item.FromStatus, &item.ToStatus, &payload, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Payload = parseJSONMap(payload)
		items = append(items, item)
	}
	return items, rows.Err()
}

func listOrderTerminationEvents(ctx context.Context, db *sql.DB, orderID int64) ([]orderEventItem, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, order_id, COALESCE(order_no, ''), event_type, COALESCE(actor_user_id, 0),
		       COALESCE(actor_enterprise_id, 0), COALESCE(actor_service, ''), COALESCE(from_status, 0),
		       COALESCE(to_status, 0), payload, created_at
		FROM order_events
		WHERE order_id = ?
		  AND (
			event_type IN ('logistics_no_capacity', 'financing_risk_rejected')
			OR event_type LIKE '%cancel%'
			OR event_type LIKE '%reject%'
			OR event_type LIKE '%refuse%'
			OR event_type LIKE '%no_stock%'
		  )
		ORDER BY id DESC
		LIMIT 20`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]orderEventItem, 0)
	for rows.Next() {
		var item orderEventItem
		var payload sql.NullString
		if err := rows.Scan(&item.ID, &item.OrderID, &item.OrderNo, &item.EventType, &item.ActorUserID, &item.ActorEnterpriseID, &item.ActorService, &item.FromStatus, &item.ToStatus, &payload, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Payload = parseJSONMap(payload)
		items = append(items, item)
	}
	return items, rows.Err()
}

func latestMatchingOrderEvent(events []orderEventItem, match func(orderEventItem) bool) *orderEventItem {
	for index := range events {
		if match(events[index]) {
			return &events[index]
		}
	}
	return nil
}

func isLogisticsNoCapacityEvent(event orderEventItem) bool {
	eventType := strings.ToLower(event.EventType)
	return eventType == "logistics_no_capacity" || strings.Contains(eventType, "no_capacity")
}

func isFinancingRiskRejectedEvent(event orderEventItem) bool {
	eventType := strings.ToLower(event.EventType)
	return eventType == "financing_risk_rejected" || (strings.Contains(eventType, "financ") && strings.Contains(eventType, "reject"))
}

func isSellerRejectedEvent(event orderEventItem) bool {
	eventType := strings.ToLower(event.EventType)
	party := strings.ToLower(firstNonEmpty(mapString(event.Payload, "party", "cancel_party", "cancelParty", "actor_role", "actorRole"), event.ActorService))
	if strings.Contains(party, "seller") && (strings.Contains(eventType, "reject") || strings.Contains(eventType, "refuse") || strings.Contains(eventType, "cancel") || strings.Contains(eventType, "no_stock")) {
		return true
	}
	return strings.Contains(eventType, "seller") && (strings.Contains(eventType, "reject") || strings.Contains(eventType, "refuse") || strings.Contains(eventType, "cancel") || strings.Contains(eventType, "no_stock"))
}

func isBuyerCancelledEvent(event orderEventItem) bool {
	eventType := strings.ToLower(event.EventType)
	party := strings.ToLower(firstNonEmpty(mapString(event.Payload, "party", "cancel_party", "cancelParty", "actor_role", "actorRole"), event.ActorService))
	if strings.Contains(party, "buyer") && strings.Contains(eventType, "cancel") {
		return true
	}
	return strings.Contains(eventType, "buyer") && (strings.Contains(eventType, "cancel") || strings.Contains(eventType, "withdraw"))
}

func financingInfoRejected(info map[string]any) bool {
	return rejectedValue(mapString(info, "risk_result", "riskResult")) || rejectedValue(mapString(info, "status"))
}

func rejectedValue(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return false
	}
	for _, keyword := range []string{"rejected", "reject", "failed", "fail", "denied", "declined", "refused", "not_pass", "not passed", "不通过", "未通过", "拒绝", "驳回", "额度不足"} {
		if strings.Contains(normalized, keyword) {
			return true
		}
	}
	return false
}

func cancelledOrderTerminationReason(order orderItem, needsIntervention bool) *terminationReason {
	party := strings.ToLower(firstNonEmpty(mapString(order.Remark, "termination_party", "terminationParty", "cancel_party", "cancelParty", "party"), ""))
	reason := firstNonEmpty(mapString(order.Remark, "termination_reason", "terminationReason", "cancel_reason", "cancelReason", "cancellation_reason", "cancellationReason"), "订单已取消或交易终止")
	if strings.Contains(party, "seller") {
		return &terminationReason{Code: "seller_rejected", Label: "卖方拒单-已终止", Party: "seller", PartyLabel: "卖方", Reason: reason, Source: "order_status", NeedsIntervention: needsIntervention}
	}
	if strings.Contains(party, "buyer") {
		return &terminationReason{Code: "buyer_cancelled", Label: "买方取消-已终止", Party: "buyer", PartyLabel: "买方", Reason: reason, Source: "order_status", NeedsIntervention: needsIntervention}
	}
	return &terminationReason{Code: "cancelled", Label: "订单取消-已终止", Reason: reason, Source: "order_status", NeedsIntervention: needsIntervention}
}

func mapString(values map[string]any, keys ...string) string {
	if values == nil {
		return ""
	}
	for _, key := range keys {
		switch value := values[key].(type) {
		case string:
			if strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		case fmt.Stringer:
			if strings.TrimSpace(value.String()) != "" {
				return strings.TrimSpace(value.String())
			}
		case float64, int, int64, bool:
			return fmt.Sprint(value)
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func truthyValue(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		normalized := strings.ToLower(strings.TrimSpace(typed))
		return normalized == "true" || normalized == "1" || normalized == "yes" || normalized == "y"
	case float64:
		return typed != 0
	case int:
		return typed != 0
	case int64:
		return typed != 0
	default:
		return false
	}
}

func eventPayloadString(event *orderEventItem, keys ...string) string {
	if event == nil {
		return ""
	}
	return mapString(event.Payload, keys...)
}

func eventTypeOf(event *orderEventItem) string {
	if event == nil {
		return ""
	}
	return event.EventType
}

func eventCreatedAt(event *orderEventItem) *time.Time {
	if event == nil {
		return nil
	}
	return &event.CreatedAt
}

func getOrderFinancingInfo(ctx context.Context, db *sql.DB, orderID int64) (map[string]any, error) {
	row := db.QueryRowContext(ctx, `
		SELECT id, status, COALESCE(risk_result, ''), COALESCE(risk_remark, ''), COALESCE(approved_amount, 0),
		       COALESCE(requested_amount, 0), COALESCE(funding_enterprise_id, 0), COALESCE(disbursement_voucher, ''),
		       COALESCE(pledge_voucher_url, ''), COALESCE(commission_voucher, ''), COALESCE(debt_contract_url, ''),
		       COALESCE(pledge_remark, ''), created_at, updated_at
		FROM order_financing_requests
		WHERE order_id = ?
		ORDER BY id DESC
		LIMIT 1`, orderID)
	var id, fundingEnterpriseID int64
	var status, riskResult, riskRemark, disbursementVoucher, pledgeVoucherURL, commissionVoucher, debtContractURL, pledgeRemark string
	var approvedAmount, requestedAmount float64
	var createdAt, updatedAt time.Time
	if err := row.Scan(&id, &status, &riskResult, &riskRemark, &approvedAmount, &requestedAmount, &fundingEnterpriseID, &disbursementVoucher, &pledgeVoucherURL, &commissionVoucher, &debtContractURL, &pledgeRemark, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return map[string]any{
		"id": id, "status": status, "risk_result": riskResult, "riskRemark": riskRemark,
		"approved_amount": approvedAmount, "approvedAmount": approvedAmount,
		"requested_amount": requestedAmount, "requestedAmount": requestedAmount,
		"funding_enterprise_id": fundingEnterpriseID, "fundingEnterpriseId": fundingEnterpriseID,
		"disbursement_voucher": disbursementVoucher, "disbursementVoucher": disbursementVoucher,
		"pledge_voucher_url": pledgeVoucherURL, "pledgeVoucherUrl": pledgeVoucherURL,
		"commission_voucher": commissionVoucher, "commissionVoucher": commissionVoucher,
		"debt_contract_url": debtContractURL, "debtContractUrl": debtContractURL,
		"pledge_remark": pledgeRemark, "pledgeRemark": pledgeRemark,
		"createdAt": createdAt, "updatedAt": updatedAt,
	}, nil
}

func getLogisticsBiddingInfo(ctx context.Context, db *sql.DB, orderID int64) (map[string]any, error) {
	row := db.QueryRowContext(ctx, `
		SELECT id, status, COALESCE(freight_payer, ''), COALESCE(selected_quote_id, 0),
		       COALESCE(pickup_address, ''), COALESCE(destination_address, ''), COALESCE(requirement, ''),
		       COALESCE(initiator_role, ''), COALESCE(responder_confirmed, 0), created_at, updated_at
		FROM logistics_biddings
		WHERE order_id = ?
		ORDER BY id DESC
		LIMIT 1`, orderID)
	var id, selectedQuoteID int64
	var responderConfirmed bool
	var status, freightPayer, pickupAddress, destinationAddress, requirement, initiatorRole string
	var createdAt, updatedAt time.Time
	if err := row.Scan(&id, &status, &freightPayer, &selectedQuoteID, &pickupAddress, &destinationAddress, &requirement, &initiatorRole, &responderConfirmed, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return map[string]any{
		"id": id, "status": status, "freight_payer": freightPayer, "freightPayer": freightPayer,
		"selected_quote_id": selectedQuoteID, "selectedQuoteId": selectedQuoteID,
		"pickup_address": pickupAddress, "pickupAddress": pickupAddress,
		"destination_address": destinationAddress, "destinationAddress": destinationAddress,
		"requirement": requirement, "initiator_role": initiatorRole, "initiatorRole": initiatorRole,
		"responder_confirmed": responderConfirmed, "responderConfirmed": responderConfirmed,
		"createdAt": createdAt, "updatedAt": updatedAt,
	}, nil
}

func listLogisticsQuotes(ctx context.Context, db *sql.DB, orderID int64) ([]map[string]any, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, COALESCE(logistics_name, ''), COALESCE(logistics_enterprise_id, 0), COALESCE(amount, 0),
		       COALESCE(vehicle_info, ''), COALESCE(driver_info, ''), COALESCE(estimated_pickup_at, ''),
		       COALESCE(estimated_delivery_at, ''), COALESCE(remark, ''), created_at
		FROM logistics_bid_quotes
		WHERE order_id = ?
		ORDER BY id DESC`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]map[string]any, 0)
	for rows.Next() {
		var id, logisticsEnterpriseID int64
		var logisticsName, vehicleInfo, driverInfo, estimatedPickupAt, estimatedDeliveryAt, remark string
		var amount float64
		var createdAt time.Time
		if err := rows.Scan(&id, &logisticsName, &logisticsEnterpriseID, &amount, &vehicleInfo, &driverInfo, &estimatedPickupAt, &estimatedDeliveryAt, &remark, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"id": id, "logistics_name": logisticsName, "logisticsName": logisticsName,
			"logistics_enterprise_id": logisticsEnterpriseID, "logisticsEnterpriseId": logisticsEnterpriseID,
			"amount": amount, "vehicle_info": vehicleInfo, "vehicleInfo": vehicleInfo,
			"driver_info": driverInfo, "driverInfo": driverInfo,
			"estimated_pickup_at": estimatedPickupAt, "estimatedPickupAt": estimatedPickupAt,
			"estimated_delivery_at": estimatedDeliveryAt, "estimatedDeliveryAt": estimatedDeliveryAt,
			"remark": remark, "createdAt": createdAt,
		})
	}
	return items, rows.Err()
}

func getShipmentTaskInfo(ctx context.Context, db *sql.DB, orderID int64) (map[string]any, error) {
	row := db.QueryRowContext(ctx, `
		SELECT id, status, COALESCE(logistics_enterprise_id, 0), COALESCE(logistics_quote_id, 0),
		       COALESCE(pickup_voucher, ''), COALESCE(delivery_voucher, ''), COALESCE(signature_voucher, ''),
		       picked_up_at, delivered_at, signed_at, created_at, updated_at
		FROM shipment_tasks
		WHERE order_id = ?
		ORDER BY id DESC
		LIMIT 1`, orderID)
	var id, logisticsEnterpriseID, logisticsQuoteID int64
	var status, pickupVoucher, deliveryVoucher, signatureVoucher string
	var pickedUpAt, deliveredAt, signedAt sql.NullTime
	var createdAt, updatedAt time.Time
	if err := row.Scan(&id, &status, &logisticsEnterpriseID, &logisticsQuoteID, &pickupVoucher, &deliveryVoucher, &signatureVoucher, &pickedUpAt, &deliveredAt, &signedAt, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return map[string]any{
		"id": id, "status": status,
		"logistics_enterprise_id": logisticsEnterpriseID, "logisticsEnterpriseId": logisticsEnterpriseID,
		"logistics_quote_id": logisticsQuoteID, "logisticsQuoteId": logisticsQuoteID,
		"pickup_voucher": pickupVoucher, "pickupVoucher": pickupVoucher,
		"delivery_voucher": deliveryVoucher, "deliveryVoucher": deliveryVoucher,
		"signature_voucher": signatureVoucher, "signatureVoucher": signatureVoucher,
		"pickedUpAt": nullTimeValue(pickedUpAt), "deliveredAt": nullTimeValue(deliveredAt), "signedAt": nullTimeValue(signedAt),
		"createdAt": createdAt, "updatedAt": updatedAt,
	}, nil
}

func listPledgeEvents(ctx context.Context, db *sql.DB, orderID int64) ([]map[string]any, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, financing_request_id, event_type, COALESCE(voucher_url, ''), COALESCE(reason, ''),
		       COALESCE(pledge_release_events.condition, ''), created_at
		FROM pledge_release_events
		WHERE order_id = ?
		ORDER BY id DESC`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]map[string]any, 0)
	for rows.Next() {
		var id, financingRequestID int64
		var eventType, voucherURL, reason, condition string
		var createdAt time.Time
		if err := rows.Scan(&id, &financingRequestID, &eventType, &voucherURL, &reason, &condition, &createdAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"id": id, "financing_request_id": financingRequestID, "financingRequestId": financingRequestID,
			"event_type": eventType, "eventType": eventType, "voucher_url": voucherURL, "voucherUrl": voucherURL,
			"reason": reason, "condition": condition, "created_at": createdAt, "createdAt": createdAt,
		})
	}
	return items, rows.Err()
}

func listBreachRecords(ctx context.Context, db *sql.DB, orderID int64) ([]breachRecordItem, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, order_id, COALESCE(order_no, ''), COALESCE(reporter_enterprise_id, 0), COALESCE(breach_party, ''),
		       COALESCE(breach_enterprise_id, 0), COALESCE(breach_type, ''), COALESCE(description, ''),
		       evidence_urls, COALESCE(status, ''), COALESCE(mediation_result, ''), COALESCE(mediated_by, 0),
		       mediated_at, COALESCE(resolution_result, ''), COALESCE(resolved_by, 0), resolved_at, created_at, updated_at
		FROM breach_records
		WHERE order_id = ? AND deleted_at IS NULL
		ORDER BY id DESC`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]breachRecordItem, 0)
	for rows.Next() {
		var item breachRecordItem
		var evidence sql.NullString
		if err := rows.Scan(&item.ID, &item.OrderID, &item.OrderNo, &item.ReporterEnterpriseID, &item.BreachParty, &item.BreachEnterpriseID, &item.BreachType, &item.Description, &evidence, &item.Status, &item.MediationResult, &item.MediatedBy, &item.MediatedAt, &item.ResolutionResult, &item.ResolvedBy, &item.ResolvedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.EvidenceURLs = parseStringList(evidence)
		items = append(items, item)
	}
	return items, rows.Err()
}

func listLitigationRecords(ctx context.Context, db *sql.DB, orderID int64) ([]litigationItem, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, breach_record_id, order_id, COALESCE(order_no, ''), COALESCE(plaintiff_enterprise_id, 0),
		       COALESCE(defendant_enterprise_id, 0), COALESCE(case_no, ''), COALESCE(court, ''),
		       COALESCE(claim_amount, 0), COALESCE(description, ''), evidence_urls, COALESCE(status, ''),
		       COALESCE(judgment_result, ''), COALESCE(judgment_url, ''), judged_at, closed_at, created_at, updated_at
		FROM litigation_records
		WHERE order_id = ? AND deleted_at IS NULL
		ORDER BY id DESC`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]litigationItem, 0)
	for rows.Next() {
		var item litigationItem
		var evidence sql.NullString
		if err := rows.Scan(&item.ID, &item.BreachRecordID, &item.OrderID, &item.OrderNo, &item.PlaintiffEnterpriseID, &item.DefendantEnterpriseID, &item.CaseNo, &item.Court, &item.ClaimAmount, &item.Description, &evidence, &item.Status, &item.JudgmentResult, &item.JudgmentURL, &item.JudgedAt, &item.ClosedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.EvidenceURLs = parseStringList(evidence)
		items = append(items, item)
	}
	return items, rows.Err()
}

func nullTimeValue(value sql.NullTime) any {
	if value.Valid {
		return value.Time
	}
	return nil
}

type orderFreezeRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type breachActionRequest struct {
	Result string `json:"result" binding:"required"`
}

type litigationStatusRequest struct {
	Status         string  `json:"status" binding:"required"`
	JudgmentResult string  `json:"judgmentResult"`
	JudgmentURL    string  `json:"judgmentUrl"`
	ClaimAmount    float64 `json:"claimAmount"`
}

func currentAdminID(c *gin.Context) int64 {
	adminID, _ := c.Get("admin_id")
	switch v := adminID.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	default:
		return 0
	}
}

func insertOrderEvent(ctx context.Context, db *sql.DB, orderID int64, orderNo string, eventType string, adminID int64, fromStatus, toStatus int, payload map[string]any) error {
	if !tableExists(ctx, db, "order_events") {
		return nil
	}
	var rawPayload any
	if payload != nil {
		bytes, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		rawPayload = string(bytes)
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO order_events (order_id, order_no, event_type, actor_user_id, actor_enterprise_id, actor_service, from_status, to_status, payload, created_at)
		VALUES (?, ?, ?, ?, 0, 'admin', ?, ?, ?, NOW())`,
		orderID, orderNo, eventType, adminID, fromStatus, toStatus, rawPayload)
	return err
}

func getOrderStatusForAction(ctx context.Context, db *sql.DB, orderID int64) (string, int, error) {
	var orderNo string
	var status int
	err := db.QueryRowContext(ctx, `SELECT order_no, status FROM orders WHERE deleted_at IS NULL AND id = ? LIMIT 1`, orderID).Scan(&orderNo, &status)
	return orderNo, status, err
}

func restoreStatusFromFreezePayload(payload sql.NullString) (int, bool) {
	if !payload.Valid || strings.TrimSpace(payload.String) == "" {
		return 0, false
	}
	parsed := parseJSONMap(payload)
	if parsed == nil {
		return 0, false
	}
	var targetStatus int
	switch value := parsed["from_status"].(type) {
	case float64:
		targetStatus = int(value)
	case int:
		targetStatus = value
	default:
		return 0, false
	}
	if targetStatus < 0 || targetStatus > 27 || targetStatus == 10 {
		return 0, false
	}
	return targetStatus, true
}

func isTerminalOrderStatus(status int) bool {
	switch status {
	case 8, 9, 21:
		return true
	default:
		return false
	}
}

func canFreezeOrderStatus(status int) bool {
	return status != 10 && !isTerminalOrderStatus(status)
}

// FreezeOrder freezes an order for platform operation intervention.
func FreezeOrder(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || orderID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid order id"})
		return
	}
	var req orderFreezeRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Reason) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "freeze reason is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	orderNo, fromStatus, err := getOrderStatusForAction(ctx, db, orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if fromStatus == 10 {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": true})
		return
	}
	if !canFreezeOrderStatus(fromStatus) {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "terminal order cannot be frozen"})
		return
	}
	if _, err := db.ExecContext(ctx, `UPDATE orders SET status = 10, updated_at = NOW() WHERE id = ?`, orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	payload := map[string]any{"reason": strings.TrimSpace(req.Reason), "from_status": fromStatus}
	if err := insertOrderEvent(ctx, db, orderID, orderNo, "order_frozen", currentAdminID(c), fromStatus, 10, payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	logOpFromContext(c, db, "order", "freeze", 0, req.Reason)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": true})
}

// UnfreezeOrder restores an order to its status before the latest freeze event when possible.
func UnfreezeOrder(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || orderID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid order id"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	orderNo, fromStatus, err := getOrderStatusForAction(ctx, db, orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if fromStatus != 10 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "order is not frozen"})
		return
	}
	targetStatus := 0
	foundFreezeEvent := false
	if tableExists(ctx, db, "order_events") {
		var payload sql.NullString
		if err := db.QueryRowContext(ctx, `SELECT payload FROM order_events WHERE order_id = ? AND event_type = 'order_frozen' ORDER BY id DESC LIMIT 1`, orderID).Scan(&payload); err == nil {
			var ok bool
			targetStatus, ok = restoreStatusFromFreezePayload(payload)
			if !ok {
				c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "missing valid freeze event, cannot safely unfreeze order"})
				return
			}
			foundFreezeEvent = true
		}
	}
	if !foundFreezeEvent {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "missing freeze event, cannot safely unfreeze order"})
		return
	}
	if _, err := db.ExecContext(ctx, `UPDATE orders SET status = ?, updated_at = NOW() WHERE id = ?`, targetStatus, orderID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if err := insertOrderEvent(ctx, db, orderID, orderNo, "order_unfrozen", currentAdminID(c), fromStatus, targetStatus, map[string]any{"restored_status": targetStatus}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	logOpFromContext(c, db, "order", "unfreeze", 0)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": true})
}

// MediateBreach records platform mediation for a breach report.
func MediateBreach(c *gin.Context) {
	updateBreachRecord(c, "mediate")
}

// ResolveBreach records the final platform handling result for a breach report.
func ResolveBreach(c *gin.Context) {
	updateBreachRecord(c, "resolve")
}

func updateBreachRecord(c *gin.Context, action string) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}
	if !tableExists(c.Request.Context(), db, "breach_records") {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "breach_records table not found"})
		return
	}
	breachID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || breachID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid breach id"})
		return
	}
	var req breachActionRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Result) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "result is required"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var orderID int64
	var orderNo string
	var orderStatus int
	if err := db.QueryRowContext(ctx, `
		SELECT br.order_id, COALESCE(br.order_no, ''), o.status
		FROM breach_records br
		JOIN orders o ON o.id = br.order_id
		WHERE br.id = ? AND br.deleted_at IS NULL
		LIMIT 1`, breachID).Scan(&orderID, &orderNo, &orderStatus); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "breach record not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	adminID := currentAdminID(c)
	eventType := "breach_mediation_started"
	status := "mediating"
	query := `UPDATE breach_records SET status = ?, mediation_result = ?, mediated_by = ?, mediated_at = NOW(), updated_by = ?, updated_at = NOW() WHERE id = ?`
	args := []any{status, strings.TrimSpace(req.Result), adminID, adminID, breachID}
	if action == "resolve" {
		eventType = "breach_resolved"
		status = "resolved"
		query = `UPDATE breach_records SET status = ?, resolution_result = ?, resolved_by = ?, resolved_at = NOW(), updated_by = ?, updated_at = NOW() WHERE id = ?`
		args = []any{status, strings.TrimSpace(req.Result), adminID, adminID, breachID}
	}
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if err := insertOrderEvent(ctx, db, orderID, orderNo, eventType, adminID, orderStatus, orderStatus, map[string]any{"breach_record_id": breachID, "result": strings.TrimSpace(req.Result)}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	logOpFromContext(c, db, "order", "breach_"+action, 0, req.Result)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": true})
}

// UpdateLitigationStatus updates platform litigation tracking status.
func UpdateLitigationStatus(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}
	if !tableExists(c.Request.Context(), db, "litigation_records") {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "litigation_records table not found"})
		return
	}
	litigationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || litigationID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid litigation id"})
		return
	}
	var req litigationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Status) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "litigation status is required"})
		return
	}
	allowed := map[string]bool{"filed": true, "accepted": true, "in_progress": true, "judged": true, "closed": true}
	if !allowed[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid litigation status"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var orderID int64
	var orderNo string
	var orderStatus int
	if err := db.QueryRowContext(ctx, `
		SELECT lr.order_id, COALESCE(lr.order_no, ''), o.status
		FROM litigation_records lr
		JOIN orders o ON o.id = lr.order_id
		WHERE lr.id = ? AND lr.deleted_at IS NULL
		LIMIT 1`, litigationID).Scan(&orderID, &orderNo, &orderStatus); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "litigation record not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	setParts := []string{"status = ?", "updated_by = ?", "updated_at = NOW()"}
	args := []any{req.Status, currentAdminID(c)}
	if req.JudgmentResult != "" {
		setParts = append(setParts, "judgment_result = ?")
		args = append(args, req.JudgmentResult)
	}
	if req.JudgmentURL != "" {
		setParts = append(setParts, "judgment_url = ?")
		args = append(args, req.JudgmentURL)
	}
	if req.ClaimAmount > 0 {
		setParts = append(setParts, "claim_amount = ?")
		args = append(args, req.ClaimAmount)
	}
	if req.Status == "judged" {
		setParts = append(setParts, "judged_at = COALESCE(judged_at, NOW())")
	}
	if req.Status == "closed" {
		setParts = append(setParts, "closed_at = COALESCE(closed_at, NOW())")
	}
	args = append(args, litigationID)
	if _, err := db.ExecContext(ctx, fmt.Sprintf(`UPDATE litigation_records SET %s WHERE id = ?`, strings.Join(setParts, ", ")), args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if err := insertOrderEvent(ctx, db, orderID, orderNo, "litigation_status_updated", currentAdminID(c), orderStatus, orderStatus, map[string]any{"litigation_record_id": litigationID, "status": req.Status}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	logOpFromContext(c, db, "order", "litigation_update", 0, req.Status)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": true})
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
		timeline = append(timeline, orderTimelineItem{Key: "cancelled", Label: "订单取消", Status: "completed", Time: order.ExpiresAt, Description: "订单已取消或交易终止。"})
		return timeline
	}
	if order.Status == 10 {
		timeline = append(timeline, orderTimelineItem{Key: "operation_intervention", Label: "平台冻结", Status: "current", Description: "订单已冻结，需平台运营处理后再恢复流程。"})
		return timeline
	}
	if order.Status == 11 {
		timeline = append(timeline, orderTimelineItem{Key: "operation_intervention", Label: "异常介入", Status: "current", Description: "订单处于异常状态，需平台运营协调处理。"})
		return timeline
	}

	position := orderFlowPosition(order.Status)
	steps := []orderTimelineItem{
		{Key: "warehouse_select", Label: "选择收货仓库", Description: "买方确认收货仓库或目的地仓储。"},
		{Key: "warehouse_check", Label: "仓库条件核验", Description: "仓储方核验库容、资质和承接条件。"},
		{Key: "seller_confirm", Label: "卖方确认订单", Description: "卖方确认价格、数量和交付条款。"},
		{Key: "buyer_confirm", Label: "买方最终确认", Description: "买方确认最终订单版本。"},
		{Key: "contract", Label: "买卖合同签署", Description: "买卖双方签署包含物流、仓储和货权责任的电子合同。"},
		{Key: "payment_finance", Label: "支付或融资", Description: "买方完成直采支付，或资金方完成托盘风控、合同和放款。"},
		{Key: "logistics_bidding", Label: "物流竞价", Description: "物流方提交司机、车辆资质与报价，平台生成候选车队。"},
		{Key: "logistics_confirm", Label: "确认物流", Description: "按运费承担方确认候选物流报价。"},
		{Key: "logistics_contract", Label: "物流合同签署", Description: "买方和物流方签署运输合同后锁定运输任务。"},
		{Key: "shipping", Label: "提货运输", Description: "物流方提货出库、上传单据并运输至目的地。"},
		{Key: "delivered", Label: "到货验收", Description: "目的地仓库或买方现场验货并确认封条。"},
		{Key: "receipt", Label: "确认收货", Time: order.ReceiptAt, Description: "买方或资金方确认收货并同步物流入库结果。"},
		{Key: "reconcile", Label: "卖方对账", Description: "卖方完成收款和履约对账确认。"},
		{Key: "archive", Label: "归档", Description: "订单履约完成，关键凭证和事件归档。"},
	}
	for index := range steps {
		if order.Status == 25 {
			switch {
			case steps[index].Key == "logistics_bidding":
				steps[index].Label = "物流无运力"
				steps[index].Status = "failed"
				steps[index].Description = "无合规物流可承接，订单履约终止。"
			case index+1 <= 6:
				steps[index].Status = "completed"
			default:
				steps[index].Status = "skipped"
			}
			timeline = append(timeline, steps[index])
			continue
		}
		steps[index].Status = timelineStepStatus(position, index+1)
		timeline = append(timeline, steps[index])
	}
	return timeline
}

func orderFlowPosition(status int) int {
	switch status {
	case 0:
		return 0
	case 12, 26:
		return 1
	case 13, 24:
		return 2
	case 14, 1:
		return 3
	case 15:
		return 4
	case 2:
		return 5
	case 3, 22, 23:
		return 6
	case 16, 25:
		return 7
	case 17:
		return 8
	case 27:
		return 9
	case 18, 5, 6:
		return 10
	case 7:
		return 11
	case 19:
		return 12
	case 20:
		return 13
	case 4:
		return 7
	case 8, 21:
		return 15
	default:
		return 0
	}
}

func timelineStepStatus(position, step int) string {
	if position > step {
		return "completed"
	}
	if position == step {
		return "current"
	}
	return "pending"
}

type productRowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type productAuditLogExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type productRowScanner interface {
	Scan(dest ...any) error
}

func getProductItemByID(ctx context.Context, querier productRowQuerier, schemaDB *sql.DB, productID int64) (productItem, error) {
	goodsTypeExpr := "'spot'"
	if columnExists(ctx, schemaDB, "products", "goods_type") {
		goodsTypeExpr = "COALESCE(goods_type, 'spot')"
	}
	warehouseServiceIDExpr := "0"
	if columnExists(ctx, schemaDB, "products", "warehouse_service_id") {
		warehouseServiceIDExpr = "COALESCE(warehouse_service_id, 0)"
	}
	warehouseEnterpriseIDExpr := "0"
	if columnExists(ctx, schemaDB, "products", "warehouse_enterprise_id") {
		warehouseEnterpriseIDExpr = "COALESCE(warehouse_enterprise_id, 0)"
	}
	auditorIDExpr := "0"
	if columnExists(ctx, schemaDB, "products", "auditor_id") {
		auditorIDExpr = "COALESCE(auditor_id, 0)"
	}

	row := querier.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT id, name, enterprise_id, COALESCE(category, ''), COALESCE(sub_category, ''), COALESCE(steel_mill, ''),
		       %s, %s, %s,
		       spec, warehouse_info, COALESCE(bundle_no, ''), COALESCE(product_no, ''),
		       COALESCE(unit, ''), COALESCE(unit_price, 0), COALESCE(quantity, 0), COALESCE(min_order_qty, 0),
		       COALESCE(delivery_days, 0), COALESCE(quote_validity, 0),
		       COALESCE(description, ''), images, status, COALESCE(audit_remark, ''), %s, approved_at, created_at, updated_at
		FROM products
		WHERE deleted_at IS NULL AND id = ?
		LIMIT 1`, goodsTypeExpr, warehouseServiceIDExpr, warehouseEnterpriseIDExpr, auditorIDExpr), productID)
	return scanProductItemRow(row)
}

func scanProductItemRow(row productRowScanner) (productItem, error) {
	var item productItem
	var imageJSON sql.NullString
	var rawStatus int
	var auditRemark sql.NullString
	var auditedAt sql.NullTime
	var updatedAt sql.NullTime
	var rawSpec sql.NullString
	var rawWarehouse sql.NullString
	if err := row.Scan(
		&item.ID,
		&item.Name,
		&item.EnterpriseID,
		&item.Category,
		&item.SubCategory,
		&item.SteelMill,
		&item.GoodsType,
		&item.WarehouseServiceID,
		&item.WarehouseEnterpriseID,
		&rawSpec,
		&rawWarehouse,
		&item.BundleNo,
		&item.ProductNo,
		&item.Unit,
		&item.Price,
		&item.Stock,
		&item.MinOrderQty,
		&item.DeliveryDays,
		&item.QuoteValidity,
		&item.Description,
		&imageJSON,
		&rawStatus,
		&auditRemark,
		&item.AuditorID,
		&auditedAt,
		&item.CreatedAt,
		&updatedAt,
	); err != nil {
		return productItem{}, err
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
	if item.WarehouseServiceID == 0 {
		item.WarehouseServiceID = warehouse.WarehouseServiceID
	}
	if item.WarehouseEnterpriseID == 0 {
		item.WarehouseEnterpriseID = warehouse.WarehouseEnterpriseID
	}
	if auditRemark.Valid {
		item.AuditRemark = auditRemark.String
	}
	if auditedAt.Valid {
		item.AuditedAt = &auditedAt.Time
	}
	if updatedAt.Valid {
		item.UpdatedAt = updatedAt.Time
	}
	return item, nil
}

func listProductAuditHistory(ctx context.Context, db *sql.DB, productID int64) ([]productAuditLogItem, error) {
	if db == nil || !tableExists(ctx, db, "product_audit_logs") {
		return nil, nil
	}
	columns := columnsExist(ctx, db, "product_audit_logs")
	fromStatusExpr := "0"
	if columns["from_status"] {
		fromStatusExpr = "COALESCE(from_status, 0)"
	}
	toStatusExpr := "0"
	if columns["to_status"] {
		toStatusExpr = "COALESCE(to_status, 0)"
	}
	beforeSnapshotExpr := "NULL"
	if columns["before_snapshot"] {
		beforeSnapshotExpr = "before_snapshot"
	}
	afterSnapshotExpr := "NULL"
	if columns["after_snapshot"] {
		afterSnapshotExpr = "after_snapshot"
	}

	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, product_id, COALESCE(product_sku_id, ''), COALESCE(action, ''), COALESCE(status, ''),
		       %s, %s, COALESCE(operator_id, 0), COALESCE(operator_name, ''), COALESCE(remark, ''),
		       %s, %s, created_at
		FROM product_audit_logs
		WHERE product_id = ?
		ORDER BY id DESC
		LIMIT 20`, fromStatusExpr, toStatusExpr, beforeSnapshotExpr, afterSnapshotExpr), productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]productAuditLogItem, 0)
	for rows.Next() {
		var item productAuditLogItem
		var beforeSnapshot sql.NullString
		var afterSnapshot sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.ProductSkuID,
			&item.Action,
			&item.Status,
			&item.FromStatus,
			&item.ToStatus,
			&item.OperatorID,
			&item.OperatorName,
			&item.Remark,
			&beforeSnapshot,
			&afterSnapshot,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.BeforeSnapshot = parseProductAuditSnapshot(beforeSnapshot)
		item.AfterSnapshot = parseProductAuditSnapshot(afterSnapshot)
		items = append(items, item)
	}
	return items, rows.Err()
}

func insertProductAuditLog(ctx context.Context, executor productAuditLogExecutor, columns map[string]bool, beforeItem, afterItem productItem, status, operatorName, remark string, operatorID int64) error {
	fields := []string{"product_id", "product_sku_id", "action", "status", "operator_id", "operator_name", "remark", "created_at"}
	placeholders := []string{"?", "?", "?", "?", "?", "?", "?", "?"}
	values := []any{afterItem.ID, afterItem.ProductNo, "platform_audit", status, operatorID, operatorName, remark, time.Now()}
	if columns["from_status"] {
		fields = append(fields, "from_status")
		placeholders = append(placeholders, "?")
		values = append(values, beforeItem.Status)
	}
	if columns["to_status"] {
		fields = append(fields, "to_status")
		placeholders = append(placeholders, "?")
		values = append(values, afterItem.Status)
	}
	if columns["before_snapshot"] {
		fields = append(fields, "before_snapshot")
		placeholders = append(placeholders, "?")
		values = append(values, productAuditSnapshotJSON(buildProductAuditSnapshot(beforeItem)))
	}
	if columns["after_snapshot"] {
		fields = append(fields, "after_snapshot")
		placeholders = append(placeholders, "?")
		values = append(values, productAuditSnapshotJSON(buildProductAuditSnapshot(afterItem)))
	}

	query := fmt.Sprintf("INSERT INTO product_audit_logs (%s) VALUES (%s)", strings.Join(fields, ", "), strings.Join(placeholders, ", "))
	_, err := executor.ExecContext(ctx, query, values...)
	return err
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
	goodsTypeExpr := "'spot'"
	if columnExists(ctx, db, "products", "goods_type") {
		goodsTypeExpr = "COALESCE(goods_type, 'spot')"
	}
	warehouseServiceIDExpr := "0"
	if columnExists(ctx, db, "products", "warehouse_service_id") {
		warehouseServiceIDExpr = "COALESCE(warehouse_service_id, 0)"
	}
	warehouseEnterpriseIDExpr := "0"
	if columnExists(ctx, db, "products", "warehouse_enterprise_id") {
		warehouseEnterpriseIDExpr = "COALESCE(warehouse_enterprise_id, 0)"
	}
	auditorIDExpr := "0"
	if columnExists(ctx, db, "products", "auditor_id") {
		auditorIDExpr = "COALESCE(auditor_id, 0)"
	}

	row := db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT id, name, enterprise_id, COALESCE(category, ''), COALESCE(sub_category, ''), COALESCE(steel_mill, ''),
		       %s, %s, %s,
		       spec, warehouse_info, COALESCE(bundle_no, ''), COALESCE(product_no, ''),
		       COALESCE(unit, ''), COALESCE(unit_price, 0), COALESCE(quantity, 0), COALESCE(min_order_qty, 0),
		       COALESCE(delivery_days, 0), COALESCE(quote_validity, 0),
		       COALESCE(description, ''), images, status, COALESCE(audit_remark, ''), %s, approved_at, created_at, updated_at
		FROM products
		WHERE deleted_at IS NULL AND id = ?
		LIMIT 1`, goodsTypeExpr, warehouseServiceIDExpr, warehouseEnterpriseIDExpr, auditorIDExpr), c.Param("id"))

	var rawSpec sql.NullString
	var rawWarehouse sql.NullString
	if err := row.Scan(
		&item.ID,
		&item.Name,
		&item.EnterpriseID,
		&item.Category,
		&item.SubCategory,
		&item.SteelMill,
		&item.GoodsType,
		&item.WarehouseServiceID,
		&item.WarehouseEnterpriseID,
		&rawSpec,
		&rawWarehouse,
		&item.BundleNo,
		&item.ProductNo,
		&item.Unit,
		&item.Price,
		&item.Stock,
		&item.MinOrderQty,
		&item.DeliveryDays,
		&item.QuoteValidity,
		&item.Description,
		&imageJSON,
		&rawStatus,
		&auditRemark,
		&item.AuditorID,
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
	if item.WarehouseServiceID == 0 {
		item.WarehouseServiceID = warehouse.WarehouseServiceID
	}
	if item.WarehouseEnterpriseID == 0 {
		item.WarehouseEnterpriseID = warehouse.WarehouseEnterpriseID
	}
	if auditRemark.Valid {
		item.AuditRemark = auditRemark.String
	}
	if auditedAt.Valid {
		item.AuditedAt = &auditedAt.Time
	}
	if updatedAt.Valid {
		item.UpdatedAt = updatedAt.Time
	}
	history, err := listProductAuditHistory(ctx, db, item.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	item.AuditHistory = history

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
	statusValue, logStatus, ok := normalizeProductAuditDecision(req.Status)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid product audit decision"})
		return
	}
	if !rejectReasonProvided(req.Status, 2, req.AuditRemark) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "驳回原因不能为空"})
		return
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

	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid product id"})
		return
	}

	auditLogTableExists := tableExists(ctx, db, "product_audit_logs")
	auditLogColumns := map[string]bool{}
	if auditLogTableExists {
		auditLogColumns = columnsExist(ctx, db, "product_audit_logs")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer tx.Rollback()

	beforeItem, err := getProductItemByID(ctx, tx, db, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	setParts := []string{"status = ?"}
	args := []any{statusValue}
	if columnExists(ctx, db, "products", "audit_remark") {
		setParts = append(setParts, "audit_remark = ?")
		args = append(args, strings.TrimSpace(req.AuditRemark))
	}
	if columnExists(ctx, db, "products", "auditor_id") {
		setParts = append(setParts, "auditor_id = ?")
		args = append(args, operatorID)
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

	args = append(args, productID)
	query := fmt.Sprintf("UPDATE products SET %s WHERE id = ? AND deleted_at IS NULL", strings.Join(setParts, ", "))
	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "product not found"})
		return
	}

	afterItem, err := getProductItemByID(ctx, tx, db, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if afterItem.Status != statusValue {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "product audit status sync failed"})
		return
	}

	if auditLogTableExists {
		if err := insertProductAuditLog(ctx, tx, auditLogColumns, beforeItem, afterItem, logStatus, adminName, strings.TrimSpace(req.AuditRemark), operatorID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	logOpFromContext(c, db, "product", "audit", 0, fmt.Sprintf(`{"id":"%d","status":%d}`, productID, req.Status))

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
			SELECT id, COALESCE(parent_id, 0), name, COALESCE(display_name, ''), COALESCE(code, ''), COALESCE(icon, ''), COALESCE(image_url, ''), COALESCE(sort, 0), COALESCE(status, 1)
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
			if err := rows.Scan(&item.ID, &item.ParentID, &item.Name, &item.DisplayName, &item.Code, &item.Icon, &item.ImageURL, &item.Sort, &item.Status); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
				return
			}
			item.ProductCount = countProductsForCategory(ctx, db, item)
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

func countProductsForCategory(ctx context.Context, db *sql.DB, category categoryItem) int {
	if db == nil || !tableExists(ctx, db, "products") {
		return 0
	}
	conditions := []string{}
	args := []any{}
	productColumns := columnsExist(ctx, db, "products")
	if productColumns["category_id"] {
		conditions = append(conditions, "category_id = ?")
		args = append(args, category.ID)
	}
	names := []string{category.Name, category.DisplayName, category.Code}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if productColumns["category"] {
			conditions = append(conditions, "category = ?")
			args = append(args, name)
		}
		if productColumns["sub_category"] {
			conditions = append(conditions, "sub_category = ?")
			args = append(args, name)
		}
	}
	if len(conditions) == 0 {
		return 0
	}
	var total int
	query := fmt.Sprintf("SELECT COUNT(*) FROM products WHERE deleted_at IS NULL AND (%s)", strings.Join(conditions, " OR "))
	if err := db.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0
	}
	return total
}

// CreateProductCategory creates a new product category
func CreateProductCategory(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		DisplayName string `json:"displayName"`
		Code        string `json:"code"`
		Icon        string `json:"icon"`
		ImageURL    string `json:"imageUrl"`
		Sort        int    `json:"sort"`
		Status      int    `json:"status"`
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
	if req.DisplayName == "" {
		req.DisplayName = req.Name
	}
	result, err := db.ExecContext(ctx, `INSERT INTO product_categories (name, display_name, code, icon, image_url, sort, status) VALUES (?, ?, ?, ?, ?, ?, ?)`, req.Name, req.DisplayName, req.Code, req.Icon, req.ImageURL, req.Sort, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	id, _ := result.LastInsertId()
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": categoryItem{ID: id, Name: req.Name, DisplayName: req.DisplayName, Code: req.Code, Icon: req.Icon, ImageURL: req.ImageURL, Sort: req.Sort, Status: req.Status}})
}

// UpdateProductCategory updates a product category
func UpdateProductCategory(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		DisplayName string `json:"displayName"`
		Code        string `json:"code"`
		Icon        string `json:"icon"`
		ImageURL    string `json:"imageUrl"`
		Sort        int    `json:"sort"`
		Status      int    `json:"status"`
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
	if req.DisplayName == "" {
		req.DisplayName = req.Name
	}
	result, err := db.ExecContext(ctx, `UPDATE product_categories SET name=?, display_name=?, code=?, icon=?, image_url=?, sort=?, status=? WHERE id=? AND deleted_at IS NULL`, req.Name, req.DisplayName, req.Code, req.Icon, req.ImageURL, req.Sort, req.Status, c.Param("id"))
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
	var item categoryItem
	row := db.QueryRowContext(ctx, `SELECT id, COALESCE(parent_id, 0), name, COALESCE(display_name, ''), COALESCE(code, ''), COALESCE(icon, ''), COALESCE(image_url, ''), COALESCE(sort, 0), COALESCE(status, 1) FROM product_categories WHERE id=? AND deleted_at IS NULL LIMIT 1`, c.Param("id"))
	if err := row.Scan(&item.ID, &item.ParentID, &item.Name, &item.DisplayName, &item.Code, &item.Icon, &item.ImageURL, &item.Sort, &item.Status); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "category not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if count := countProductsForCategory(ctx, db, item); count > 0 {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": fmt.Sprintf("该类目下仍有 %d 个商品，不能删除", count)})
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
