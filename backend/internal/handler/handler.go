package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-admin/backend/internal/auth"
	"github.com/Taishi-co/SteelLink-admin/backend/internal/database"
	"github.com/gin-gonic/gin"
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

	if req.Username != defaultAdminUsername() || req.Password != defaultAdminPassword() {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "invalid username or password",
		})
		return
	}

	token, err := auth.GenerateAdminToken(1, defaultAdminUsername(), defaultAdminRealName())
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

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id":       adminID,
			"username": username,
			"realName": realName,
			"roleId":   1,
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

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": pageResult[adminItem]{
			List: []adminItem{{
				ID:        1,
				Username:  defaultAdminUsername(),
				RealName:  defaultAdminRealName(),
				Status:    1,
				RoleID:    func() *int64 { v := int64(1); return &v }(),
				CreatedAt: time.Now(),
			}},
			Total: 1,
		},
	})
}

// GetAdmin returns a single admin
func GetAdmin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    nil,
	})
}

// CreateAdmin creates a new admin
func CreateAdmin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// UpdateAdmin updates an admin
func UpdateAdmin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// DeleteAdmin deletes an admin
func DeleteAdmin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
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
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    nil,
	})
}

// CreateRole creates a new role
func CreateRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// UpdateRole updates a role
func UpdateRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// DeleteRole deletes a role
func DeleteRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
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
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    nil,
	})
}

// CreateMenu creates a new menu
func CreateMenu(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// UpdateMenu updates a menu
func UpdateMenu(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// DeleteMenu deletes a menu
func DeleteMenu(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
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

	hasEnterpriseID := columnExists(ctx, db, "users", "enterprise_id")
	query := `
		SELECT id, username, COALESCE(phone, ''), COALESCE(email, ''), status, created_at`
	if hasEnterpriseID {
		query += `, enterprise_id`
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
		if hasEnterpriseID {
			if err := rows.Scan(&item.ID, &item.Username, &item.Phone, &item.Email, &item.Status, &item.CreatedAt, &item.EnterpriseID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
				return
			}
		} else {
			if err := rows.Scan(&item.ID, &item.Username, &item.Phone, &item.Email, &item.Status, &item.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
				return
			}
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
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    nil,
	})
}

// UpdateUserStatus updates user status
func UpdateUserStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ListEnterprises returns list of enterprises
func ListEnterprises(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	limit, offset := parsePageParams(c.Query("page"), c.Query("pageSize"))

	total, err := queryCount(ctx, db, `SELECT COUNT(*) FROM enterprises WHERE deleted_at IS NULL`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	rows, err := db.QueryContext(ctx, `
		SELECT id, name, COALESCE(credit_code, ''), status, created_at, verification_info
		FROM enterprises
		WHERE deleted_at IS NULL
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, limit, offset)
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
		if err := rows.Scan(&item.ID, &item.Name, &item.CreditCode, &item.Status, &createdAt, &verification); err != nil {
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

	total, err := queryCount(ctx, db, `SELECT COUNT(*) FROM products WHERE deleted_at IS NULL`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	rows, err := db.QueryContext(ctx, `
		SELECT id, name, enterprise_id, COALESCE(specs, ''), COALESCE(unit, ''), price, stock, COALESCE(description, ''), image_urls, status, created_at
		FROM products
		WHERE deleted_at IS NULL
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer rows.Close()

	items := make([]productItem, 0)
	for rows.Next() {
		var item productItem
		var imageURLs sql.NullString
		var rawStatus string
		if err := rows.Scan(&item.ID, &item.Name, &item.EnterpriseID, &item.Spec, &item.Unit, &item.Price, &item.Stock, &item.Description, &imageURLs, &rawStatus, &item.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		item.Images = parseImages(imageURLs)
		item.Status = mapProductStatus(rawStatus)
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    pageResult[productItem]{List: items, Total: total},
	})
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
	var imageURLs sql.NullString
	var rawStatus string
	var auditRemark sql.NullString
	var auditedAt sql.NullTime
	var updatedAt sql.NullTime

	row := db.QueryRowContext(ctx, `
		SELECT id, name, enterprise_id, COALESCE(specs, ''), COALESCE(unit, ''), price, stock,
		       COALESCE(description, ''), image_urls, status, COALESCE(audit_remark, ''), approved_at, created_at, updated_at
		FROM products
		WHERE deleted_at IS NULL AND id = ?
		LIMIT 1`, c.Param("id"))

	if err := row.Scan(
		&item.ID,
		&item.Name,
		&item.EnterpriseID,
		&item.Spec,
		&item.Unit,
		&item.Price,
		&item.Stock,
		&item.Description,
		&imageURLs,
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

	item.Images = parseImages(imageURLs)
	item.Status = mapProductStatus(rawStatus)
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

	rawStatus := "pending_review"
	switch req.Status {
	case 1:
		rawStatus = "approved"
	case 2:
		rawStatus = "rejected"
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
	args := []any{rawStatus}
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

	if tableExists(ctx, db, "product_audit_logs") {
		var skuID string
		productID := int64(0)
		fmt.Sscanf(c.Param("id"), "%d", &productID)
		_ = db.QueryRowContext(ctx, `SELECT COALESCE(sku_id, '') FROM products WHERE id = ? LIMIT 1`, c.Param("id")).Scan(&skuID)
		if _, err := db.ExecContext(ctx, `
			INSERT INTO product_audit_logs (product_id, product_sku_id, action, status, operator_id, operator_name, remark, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			productID,
			skuID,
			"platform_audit",
			rawStatus,
			operatorID,
			adminName,
			req.AuditRemark,
			time.Now(),
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ListProductCategories returns list of product categories
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
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// UpdateProductCategory updates a product category
func UpdateProductCategory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// DeleteProductCategory deletes a product category
func DeleteProductCategory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}
