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

const (
	supplyDemandStatusPublished     = 1
	supplyDemandStatusMatched       = 2
	supplyDemandStatusPendingReview = 4
	supplyDemandStatusRejected      = 5
)

type supplyDemandPostItem struct {
	ID                int64      `json:"id"`
	PostNo            string     `json:"postNo"`
	Type              string     `json:"type"`
	Title             string     `json:"title"`
	Category          string     `json:"category"`
	Material          string     `json:"material,omitempty"`
	Spec              string     `json:"spec,omitempty"`
	Quantity          float64    `json:"quantity"`
	Unit              string     `json:"unit"`
	TargetPrice       float64    `json:"targetPrice"`
	Budget            float64    `json:"budget"`
	Currency          string     `json:"currency"`
	Province          string     `json:"province,omitempty"`
	City              string     `json:"city,omitempty"`
	District          string     `json:"district,omitempty"`
	DeliveryAddress   string     `json:"deliveryAddress,omitempty"`
	Description       string     `json:"description,omitempty"`
	Tags              string     `json:"tags,omitempty"`
	ContactName       string     `json:"contactName,omitempty"`
	ContactPhone      string     `json:"contactPhone,omitempty"`
	FinancingRequired bool       `json:"financingRequired"`
	MatchCount        int        `json:"matchCount"`
	MatchAmount       float64    `json:"matchAmount"`
	Status            int        `json:"status"`
	AuditRemark       string     `json:"auditRemark,omitempty"`
	AuditorID         int64      `json:"auditorId,omitempty"`
	AuditedAt         *time.Time `json:"auditedAt,omitempty"`
	EnterpriseID      int64      `json:"enterpriseId,omitempty"`
	CreatedBy         int64      `json:"createdBy"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

func ensureSupplyDemandPostTable(ctx context.Context, db *sql.DB) {
	if !tableExists(ctx, db, "supply_demand_posts") {
		db.ExecContext(ctx, `CREATE TABLE supply_demand_posts (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			post_no VARCHAR(32) NOT NULL UNIQUE,
			type VARCHAR(32) NOT NULL,
			title VARCHAR(200) NOT NULL,
			category VARCHAR(50) NOT NULL,
			material VARCHAR(100) DEFAULT '',
			spec VARCHAR(100) DEFAULT '',
			quantity DECIMAL(12,2) DEFAULT 0,
			unit VARCHAR(20) DEFAULT '吨',
			target_price DECIMAL(12,2) DEFAULT 0,
			budget DECIMAL(14,2) DEFAULT 0,
			currency VARCHAR(10) DEFAULT 'CNY',
			province VARCHAR(50) DEFAULT '',
			city VARCHAR(50) DEFAULT '',
			district VARCHAR(50) DEFAULT '',
			delivery_address VARCHAR(255) DEFAULT '',
			description TEXT,
			tags JSON NULL,
			contact_name VARCHAR(50) DEFAULT '',
			contact_phone VARCHAR(30) DEFAULT '',
			financing_required TINYINT(1) DEFAULT 0,
			match_count INT DEFAULT 0,
			match_amount DECIMAL(14,2) DEFAULT 0,
			status TINYINT DEFAULT 4,
			audit_remark VARCHAR(500) DEFAULT '',
			auditor_id BIGINT UNSIGNED DEFAULT 0,
			audited_at DATETIME NULL,
			enterprise_id BIGINT UNSIGNED DEFAULT 0,
			created_by BIGINT UNSIGNED DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_by BIGINT UNSIGNED DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL,
			INDEX idx_type (type),
			INDEX idx_status (status),
			INDEX idx_enterprise_id (enterprise_id),
			INDEX idx_created_by (created_by),
			INDEX idx_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	}
	if !columnExists(ctx, db, "supply_demand_posts", "audit_remark") {
		db.ExecContext(ctx, `ALTER TABLE supply_demand_posts ADD COLUMN audit_remark VARCHAR(500) DEFAULT ''`)
	}
	if !columnExists(ctx, db, "supply_demand_posts", "auditor_id") {
		db.ExecContext(ctx, `ALTER TABLE supply_demand_posts ADD COLUMN auditor_id BIGINT UNSIGNED DEFAULT 0`)
	}
	if !columnExists(ctx, db, "supply_demand_posts", "audited_at") {
		db.ExecContext(ctx, `ALTER TABLE supply_demand_posts ADD COLUMN audited_at DATETIME NULL`)
	}
}

func ListSupplyDemandPosts(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	ensureSupplyDemandPostTable(ctx, db)

	limit, offset := parsePageParams(c.Query("page"), firstNonEmpty(c.Query("pageSize"), c.Query("page_size")))
	whereParts := []string{"deleted_at IS NULL"}
	args := make([]any, 0)

	if status := strings.TrimSpace(c.Query("status")); status != "" {
		statusValue, err := strconv.Atoi(status)
		if err == nil && isValidSupplyDemandStatus(statusValue) {
			whereParts = append(whereParts, "status = ?")
			args = append(args, statusValue)
		}
	}
	if postType := strings.TrimSpace(c.Query("type")); postType != "" && isValidSupplyDemandType(postType) {
		whereParts = append(whereParts, "type = ?")
		args = append(args, postType)
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		kw := "%" + keyword + "%"
		whereParts = append(whereParts, `(title LIKE ? OR category LIKE ? OR material LIKE ? OR spec LIKE ? OR description LIKE ? OR contact_name LIKE ? OR contact_phone LIKE ?)`)
		args = append(args, kw, kw, kw, kw, kw, kw, kw)
	}

	whereClause := strings.Join(whereParts, " AND ")
	total, err := queryCount(ctx, db, fmt.Sprintf(`SELECT COUNT(*) FROM supply_demand_posts WHERE %s`, whereClause), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, post_no, type, title, category, COALESCE(material, ''), COALESCE(spec, ''),
		       COALESCE(quantity, 0), COALESCE(unit, ''), COALESCE(target_price, 0), COALESCE(budget, 0), COALESCE(currency, 'CNY'),
		       COALESCE(province, ''), COALESCE(city, ''), COALESCE(district, ''), COALESCE(delivery_address, ''),
		       COALESCE(description, ''), COALESCE(CAST(tags AS CHAR), ''), COALESCE(contact_name, ''), COALESCE(contact_phone, ''),
		       COALESCE(financing_required, 0), COALESCE(match_count, 0), COALESCE(match_amount, 0), status,
		       COALESCE(audit_remark, ''), COALESCE(auditor_id, 0), audited_at,
		       COALESCE(enterprise_id, 0), COALESCE(created_by, 0), created_at, updated_at
		FROM supply_demand_posts
		WHERE %s
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, whereClause), queryArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer rows.Close()

	items := make([]supplyDemandPostItem, 0)
	for rows.Next() {
		var item supplyDemandPostItem
		var auditedAt sql.NullTime
		var financingRequired int
		if err := rows.Scan(
			&item.ID, &item.PostNo, &item.Type, &item.Title, &item.Category, &item.Material, &item.Spec,
			&item.Quantity, &item.Unit, &item.TargetPrice, &item.Budget, &item.Currency,
			&item.Province, &item.City, &item.District, &item.DeliveryAddress,
			&item.Description, &item.Tags, &item.ContactName, &item.ContactPhone,
			&financingRequired, &item.MatchCount, &item.MatchAmount, &item.Status,
			&item.AuditRemark, &item.AuditorID, &auditedAt,
			&item.EnterpriseID, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		if auditedAt.Valid {
			item.AuditedAt = &auditedAt.Time
		}
		item.FinancingRequired = financingRequired != 0
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": pageResult[supplyDemandPostItem]{List: items, Total: total}})
}

func AuditSupplyDemandPost(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	var req struct {
		Status           int    `json:"status"`
		AuditAction      string `json:"auditAction"`
		AuditActionSnake string `json:"audit_action"`
		AuditRemark      string `json:"auditRemark"`
		AuditRemarkSnake string `json:"audit_remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "审核参数无效"})
		return
	}

	auditAction := req.AuditAction
	if auditAction == "" {
		auditAction = req.AuditActionSnake
	}
	auditRemark := req.AuditRemark
	if auditRemark == "" {
		auditRemark = req.AuditRemarkSnake
	}
	statusValue, ok := normalizeSupplyDemandAuditStatus(req.Status, auditAction)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "审核状态无效，仅支持通过或驳回"})
		return
	}
	if !rejectReasonProvided(statusValue, supplyDemandStatusRejected, auditRemark) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "驳回原因不能为空"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	ensureSupplyDemandPostTable(ctx, db)

	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || postID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "供求信息ID无效"})
		return
	}

	adminID := currentAdminID(c)
	now := time.Now()
	result, err := db.ExecContext(ctx, `
		UPDATE supply_demand_posts
		SET status = ?, audit_remark = ?, auditor_id = ?, audited_at = ?, updated_by = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL`,
		statusValue, strings.TrimSpace(auditRemark), adminID, now, adminID, now, postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "供求信息不存在"})
		return
	}

	logOpFromContext(c, db, "supply-demand", "audit", 0, fmt.Sprintf(`{"id":"%d","status":%d}`, postID, statusValue))
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": true})
}

func normalizeSupplyDemandAuditStatus(status int, auditAction string) (int, bool) {
	switch strings.ToLower(strings.TrimSpace(auditAction)) {
	case "approve":
		return supplyDemandStatusPublished, true
	case "reject":
		return supplyDemandStatusRejected, true
	}
	switch status {
	case supplyDemandStatusPublished:
		return supplyDemandStatusPublished, true
	case supplyDemandStatusRejected:
		return supplyDemandStatusRejected, true
	}
	return 0, false
}

func isValidSupplyDemandStatus(status int) bool {
	return status >= 0 && status <= supplyDemandStatusRejected
}

func isValidSupplyDemandType(postType string) bool {
	switch postType {
	case "supply", "demand", "warehouse", "logistics":
		return true
	default:
		return false
	}
}
