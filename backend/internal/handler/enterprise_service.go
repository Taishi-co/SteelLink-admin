package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-admin/backend/internal/database"
	"github.com/gin-gonic/gin"
)

type serviceApplicationItem struct {
	ID             int64          `json:"id"`
	EnterpriseID   int64          `json:"enterpriseId"`
	EnterpriseCode string         `json:"enterpriseCode"`
	EnterpriseName string         `json:"enterpriseName"`
	Service        string         `json:"service"`
	EnterpriseRole string         `json:"enterpriseRole"`
	Status         int            `json:"status"`
	AuditReason    string         `json:"auditReason"`
	SubmittedBy    int64          `json:"submittedBy"`
	AuditedBy      int64          `json:"auditedBy"`
	AuditedAt      *time.Time     `json:"auditedAt"`
	Documents      []documentItem `json:"documents"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

type documentItem struct {
	ID            int64     `json:"id"`
	ApplicationID int64     `json:"applicationId"`
	DocumentType  string    `json:"documentType"`
	FileURL       string    `json:"fileUrl"`
	CreatedAt     time.Time `json:"createdAt"`
}

type servicePermissionItem struct {
	ID             int64      `json:"id"`
	EnterpriseID   int64      `json:"enterpriseId"`
	EnterpriseCode string     `json:"enterpriseCode"`
	Service        string     `json:"service"`
	CapabilityKind string     `json:"capabilityKind"`
	EnterpriseRole string     `json:"enterpriseRole"`
	Status         int        `json:"status"`
	AuditReason    string     `json:"auditReason"`
	ApprovedAt     *time.Time `json:"approvedAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func ensureEnterpriseServiceTables(ctx context.Context, db *sql.DB) {
	if !tableExists(ctx, db, "enterprise_service_permissions") {
		db.ExecContext(ctx, `CREATE TABLE enterprise_service_permissions (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			enterprise_id BIGINT UNSIGNED NOT NULL,
			enterprise_code VARCHAR(32) NOT NULL,
			service VARCHAR(32) NOT NULL,
			status TINYINT NOT NULL DEFAULT 1,
			audit_reason VARCHAR(500) NOT NULL DEFAULT '',
			approved_at DATETIME NULL,
			created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL,
			UNIQUE KEY uk_enterprise_service (enterprise_id, service),
			INDEX idx_enterprise_code (enterprise_code),
			INDEX idx_status (status),
			INDEX idx_deleted_at (deleted_at)
		)`)
	}
	if !tableExists(ctx, db, "enterprise_service_applications") {
		db.ExecContext(ctx, `CREATE TABLE enterprise_service_applications (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			enterprise_id BIGINT UNSIGNED NOT NULL,
			enterprise_code VARCHAR(32) NOT NULL,
			service VARCHAR(32) NOT NULL,
			status TINYINT NOT NULL DEFAULT 1,
			audit_reason VARCHAR(500) NOT NULL DEFAULT '',
			submitted_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
			audited_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
			audited_at DATETIME NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL,
			INDEX idx_enterprise_code (enterprise_code),
			INDEX idx_service_status (service, status),
			INDEX idx_deleted_at (deleted_at)
		)`)
	}
	if !tableExists(ctx, db, "enterprise_application_documents") {
		db.ExecContext(ctx, `CREATE TABLE enterprise_application_documents (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			application_id BIGINT UNSIGNED NOT NULL,
			document_type VARCHAR(64) NOT NULL,
			file_url VARCHAR(500) NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL,
			INDEX idx_application_id (application_id),
			INDEX idx_deleted_at (deleted_at)
		)`)
	}
	ensureCapabilityPermissionTable(ctx, db, "enterprise_trade_capability_permissions")
	ensureCapabilityPermissionTable(ctx, db, "enterprise_provider_service_permissions")
	migrateLegacyEnterpriseServicePermissions(ctx, db)
}

func ensureCapabilityPermissionTable(ctx context.Context, db *sql.DB, table string) {
	if tableExists(ctx, db, table) {
		return
	}
	db.ExecContext(ctx, fmt.Sprintf(`CREATE TABLE %s (
		id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
		enterprise_id BIGINT UNSIGNED NOT NULL,
		enterprise_code VARCHAR(32) NOT NULL,
		service VARCHAR(32) NOT NULL,
		status TINYINT NOT NULL DEFAULT 1,
		audit_reason VARCHAR(500) NOT NULL DEFAULT '',
		approved_at DATETIME NULL,
		created_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_by BIGINT UNSIGNED NOT NULL DEFAULT 0,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		deleted_at DATETIME NULL,
		UNIQUE KEY uk_enterprise_service (enterprise_id, service),
		INDEX idx_enterprise_code (enterprise_code),
		INDEX idx_status (status),
		INDEX idx_deleted_at (deleted_at)
	)`, table))
}

func migrateLegacyEnterpriseServicePermissions(ctx context.Context, db *sql.DB) {
	if !tableExists(ctx, db, "enterprise_service_permissions") {
		return
	}
	rows, err := db.QueryContext(ctx, `
		SELECT enterprise_id, enterprise_code, service, status, COALESCE(audit_reason, ''), approved_at, created_by, updated_by
		FROM enterprise_service_permissions
		WHERE deleted_at IS NULL`)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var enterpriseID, createdBy, updatedBy int64
		var enterpriseCode, service, reason string
		var status int
		var approvedAt sql.NullTime
		if err := rows.Scan(&enterpriseID, &enterpriseCode, &service, &status, &reason, &approvedAt, &createdBy, &updatedBy); err != nil {
			continue
		}
		table := capabilityPermissionTable(service)
		if table == "" {
			continue
		}
		var count int
		if err := db.QueryRowContext(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE enterprise_id=? AND service=? AND deleted_at IS NULL`, table), enterpriseID, service).Scan(&count); err != nil || count > 0 {
			continue
		}
		db.ExecContext(ctx, fmt.Sprintf(`INSERT INTO %s (enterprise_id, enterprise_code, service, status, audit_reason, approved_at, created_by, updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, table),
			enterpriseID, enterpriseCode, service, status, reason, approvedAt, createdBy, updatedBy)
	}
}

func ListEnterpriseServiceApplications(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	ensureEnterpriseServiceTables(ctx, db)

	limit, offset := parsePageParams(c.Query("page"), c.Query("pageSize"))
	where := []string{"a.deleted_at IS NULL"}
	args := []any{}
	if service := c.Query("service"); service != "" {
		where = append(where, "a.service = ?")
		args = append(args, service)
	}
	if status := c.Query("status"); status != "" {
		where = append(where, "a.status = ?")
		args = append(args, status)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		where = append(where, "(a.enterprise_code LIKE ? OR e.name LIKE ?)")
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}
	whereClause := strings.Join(where, " AND ")

	total, err := queryCount(ctx, db, fmt.Sprintf(`SELECT COUNT(*) FROM enterprise_service_applications a LEFT JOIN enterprises e ON e.id = a.enterprise_id WHERE %s`, whereClause), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT a.id, a.enterprise_id, a.enterprise_code, COALESCE(e.name, ''), a.service, a.status,
		       COALESCE(a.audit_reason, ''), a.submitted_by, COALESCE(a.audited_by, 0), a.audited_at, a.created_at, a.updated_at
		FROM enterprise_service_applications a
		LEFT JOIN enterprises e ON e.id = a.enterprise_id
		WHERE %s
		ORDER BY a.id DESC
		LIMIT ? OFFSET ?`, whereClause), queryArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer rows.Close()

	items := make([]serviceApplicationItem, 0)
	for rows.Next() {
		var item serviceApplicationItem
		var auditedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.EnterpriseID, &item.EnterpriseCode, &item.EnterpriseName, &item.Service, &item.Status, &item.AuditReason, &item.SubmittedBy, &item.AuditedBy, &auditedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		item.EnterpriseRole = serviceToEnterpriseRole(item.Service)
		if auditedAt.Valid {
			item.AuditedAt = &auditedAt.Time
		}
		item.Documents = listApplicationDocuments(ctx, db, item.ID)
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": pageResult[serviceApplicationItem]{List: items, Total: total}})
}

func AuditEnterpriseServiceApplication(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	ensureEnterpriseServiceTables(ctx, db)

	var req struct {
		Status int    `json:"status"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.Status != 2 && req.Status != 3) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "状态值无效，仅支持 2(通过) 或 3(驳回)"})
		return
	}

	adminID, _ := c.Get("admin_id")
	adminIDValue := normalizeAdminID(adminID)
	id := c.Param("id")

	var enterpriseID int64
	var enterpriseCode, service string
	if err := db.QueryRowContext(ctx, `SELECT enterprise_id, enterprise_code, service FROM enterprise_service_applications WHERE id = ? AND deleted_at IS NULL`, id).
		Scan(&enterpriseID, &enterpriseCode, &service); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "服务申请不存在"})
		return
	}

	now := time.Now()
	if _, err := db.ExecContext(ctx, `UPDATE enterprise_service_applications SET status=?, audit_reason=?, audited_by=?, audited_at=?, updated_at=NOW() WHERE id=? AND deleted_at IS NULL`,
		req.Status, req.Reason, adminIDValue, now, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	permissionStatus := req.Status
	approvedAt := sql.NullTime{}
	if req.Status == 2 {
		approvedAt = sql.NullTime{Time: now, Valid: true}
	}
	upsertEnterpriseServicePermission(ctx, db, enterpriseID, enterpriseCode, service, permissionStatus, req.Reason, approvedAt, adminIDValue)

	logOpFromContext(c, db, "enterprise_service", "audit", 0, fmt.Sprintf(`{"id":"%s","status":%d}`, id, req.Status))
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": true})
}

func UpdateEnterpriseServicePermission(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	ensureEnterpriseServiceTables(ctx, db)

	var req struct {
		Status int    `json:"status"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.Status != 2 && req.Status != 3 && req.Status != 4) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "状态值无效，仅支持 2/3/4"})
		return
	}

	enterpriseID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "企业ID无效"})
		return
	}
	service := c.Param("service")
	if serviceToEnterpriseRole(service) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "平台服务无效"})
		return
	}

	var enterpriseCode string
	if err := db.QueryRowContext(ctx, `SELECT enterprise_code FROM enterprises WHERE id=? AND deleted_at IS NULL`, enterpriseID).Scan(&enterpriseCode); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "企业不存在"})
		return
	}

	adminID, _ := c.Get("admin_id")
	now := time.Now()
	approvedAt := sql.NullTime{}
	if req.Status == 2 {
		approvedAt = sql.NullTime{Time: now, Valid: true}
	}
	upsertEnterpriseServicePermission(ctx, db, enterpriseID, enterpriseCode, service, req.Status, req.Reason, approvedAt, normalizeAdminID(adminID))
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": true})
}

func ListEnterpriseServicePermissions(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	ensureEnterpriseServiceTables(ctx, db)

	items := make([]servicePermissionItem, 0)
	for _, table := range []string{"enterprise_trade_capability_permissions", "enterprise_provider_service_permissions"} {
		rows, err := db.QueryContext(ctx, fmt.Sprintf(`
			SELECT id, enterprise_id, enterprise_code, service, status, COALESCE(audit_reason, ''), approved_at, updated_at
			FROM %s
			WHERE deleted_at IS NULL AND enterprise_id = ?
			ORDER BY service ASC`, table), c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		for rows.Next() {
			var item servicePermissionItem
			var approvedAt sql.NullTime
			if err := rows.Scan(&item.ID, &item.EnterpriseID, &item.EnterpriseCode, &item.Service, &item.Status, &item.AuditReason, &approvedAt, &item.UpdatedAt); err != nil {
				rows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
				return
			}
			item.EnterpriseRole = serviceToEnterpriseRole(item.Service)
			item.CapabilityKind = capabilityKind(item.Service)
			if approvedAt.Valid {
				item.ApprovedAt = &approvedAt.Time
			}
			items = append(items, item)
		}
		rows.Close()
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": items})
}

func listApplicationDocuments(ctx context.Context, db *sql.DB, applicationID int64) []documentItem {
	rows, err := db.QueryContext(ctx, `SELECT id, application_id, document_type, file_url, created_at FROM enterprise_application_documents WHERE deleted_at IS NULL AND application_id = ? ORDER BY id ASC`, applicationID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	result := make([]documentItem, 0)
	for rows.Next() {
		var item documentItem
		if err := rows.Scan(&item.ID, &item.ApplicationID, &item.DocumentType, &item.FileURL, &item.CreatedAt); err == nil {
			result = append(result, item)
		}
	}
	return result
}

func upsertEnterpriseServicePermission(ctx context.Context, db *sql.DB, enterpriseID int64, enterpriseCode, service string, status int, reason string, approvedAt sql.NullTime, adminID int64) {
	table := capabilityPermissionTable(service)
	if table == "" {
		return
	}
	var id int64
	err := db.QueryRowContext(ctx, fmt.Sprintf(`SELECT id FROM %s WHERE enterprise_id=? AND service=? AND deleted_at IS NULL`, table), enterpriseID, service).Scan(&id)
	if err == nil {
		db.ExecContext(ctx, fmt.Sprintf(`UPDATE %s SET status=?, audit_reason=?, approved_at=?, updated_by=?, updated_at=NOW() WHERE id=?`, table), status, reason, approvedAt, adminID, id)
		return
	}
	db.ExecContext(ctx, fmt.Sprintf(`INSERT INTO %s (enterprise_id, enterprise_code, service, status, audit_reason, approved_at, created_by, updated_by) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, table),
		enterpriseID, enterpriseCode, service, status, reason, approvedAt, adminID, adminID)
}

func capabilityKind(service string) string {
	switch service {
	case "sales", "procurement":
		return "trade"
	case "logistics", "warehouse", "finance":
		return "provider"
	default:
		return ""
	}
}

func capabilityPermissionTable(service string) string {
	switch capabilityKind(service) {
	case "trade":
		return "enterprise_trade_capability_permissions"
	case "provider":
		return "enterprise_provider_service_permissions"
	default:
		return ""
	}
}

func serviceToEnterpriseRole(service string) string {
	switch service {
	case "sales":
		return "seller"
	case "procurement":
		return "buyer"
	case "logistics":
		return "logistics"
	case "warehouse":
		return "warehouse"
	case "finance":
		return "fund"
	default:
		return ""
	}
}

func normalizeAdminID(value any) int64 {
	switch v := value.(type) {
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
