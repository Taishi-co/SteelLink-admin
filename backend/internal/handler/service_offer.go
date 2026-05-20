package handler

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-admin/backend/internal/database"
	"github.com/gin-gonic/gin"
)

type serviceOfferItem struct {
	ID                 int64     `json:"id"`
	Type               string    `json:"type"`
	EnterpriseID       int64     `json:"enterpriseId"`
	Name               string    `json:"name"`
	Province           string    `json:"province,omitempty"`
	City               string    `json:"city,omitempty"`
	Address            string    `json:"address,omitempty"`
	Capacity           float64   `json:"capacity,omitempty"`
	AvailableCapacity  float64   `json:"availableCapacity,omitempty"`
	RouteStart         string    `json:"routeStart,omitempty"`
	RouteEnd           string    `json:"routeEnd,omitempty"`
	ServiceScope       string    `json:"serviceScope,omitempty"`
	VehicleType        string    `json:"vehicleType,omitempty"`
	LoadCapacity       float64   `json:"loadCapacity,omitempty"`
	PriceRule          string    `json:"priceRule,omitempty"`
	LoadingCapacity    string    `json:"loadingCapacity,omitempty"`
	BusinessHours      string    `json:"businessHours,omitempty"`
	PledgeSupervision  bool      `json:"pledgeSupervision,omitempty"`
	OutboundVoucherReq string    `json:"outboundVoucherReq,omitempty"`
	QuoteRules         string    `json:"quoteRules,omitempty"`
	FleetName          string    `json:"fleetName,omitempty"`
	VehicleCerts       string    `json:"vehicleCerts,omitempty"`
	DriverCerts        string    `json:"driverCerts,omitempty"`
	FinanceType        string    `json:"financeType,omitempty"`
	MinAmount          float64   `json:"minAmount,omitempty"`
	MaxAmount          float64   `json:"maxAmount,omitempty"`
	TermDays           int       `json:"termDays,omitempty"`
	AnnualRate         float64   `json:"annualRate,omitempty"`
	Status             int       `json:"status"`
	AuditRemark        string    `json:"auditRemark,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

func ensureServiceOfferTables(ctx context.Context, db *sql.DB) {
	if !tableExists(ctx, db, "warehouse_services") {
		db.ExecContext(ctx, `CREATE TABLE warehouse_services (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			enterprise_id BIGINT UNSIGNED NOT NULL,
			warehouse_name VARCHAR(200) NOT NULL,
			province VARCHAR(50) DEFAULT '',
			city VARCHAR(50) DEFAULT '',
			district VARCHAR(50) DEFAULT '',
			address VARCHAR(255) DEFAULT '',
			capacity DECIMAL(12,2) DEFAULT 0,
			available_capacity DECIMAL(12,2) DEFAULT 0,
			service_types JSON NULL,
			price_rule VARCHAR(500) DEFAULT '',
			loading_capacity VARCHAR(255) DEFAULT '',
			business_hours VARCHAR(255) DEFAULT '',
			pledge_supervision TINYINT(1) DEFAULT 0,
			outbound_voucher_req VARCHAR(500) DEFAULT '',
			images JSON NULL,
			certificates JSON NULL,
			description TEXT,
			status TINYINT DEFAULT 1,
			audit_remark VARCHAR(500) DEFAULT '',
			approved_at DATETIME NULL,
			created_by BIGINT UNSIGNED DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_by BIGINT UNSIGNED DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL,
			INDEX idx_enterprise_id (enterprise_id),
			INDEX idx_status (status),
			INDEX idx_deleted_at (deleted_at)
		)`)
	}
	if !tableExists(ctx, db, "logistics_services") {
		db.ExecContext(ctx, `CREATE TABLE logistics_services (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			enterprise_id BIGINT UNSIGNED NOT NULL,
			name VARCHAR(200) NOT NULL,
			route_start VARCHAR(100) DEFAULT '',
			route_end VARCHAR(100) DEFAULT '',
			service_scope VARCHAR(255) DEFAULT '',
			vehicle_type VARCHAR(100) DEFAULT '',
			load_capacity DECIMAL(12,2) DEFAULT 0,
			price_rule VARCHAR(500) DEFAULT '',
			quote_rules JSON NULL,
			fleet_name VARCHAR(200) DEFAULT '',
			vehicle_certs JSON NULL,
			driver_certs JSON NULL,
			images JSON NULL,
			certificates JSON NULL,
			description TEXT,
			status TINYINT DEFAULT 1,
			audit_remark VARCHAR(500) DEFAULT '',
			approved_at DATETIME NULL,
			created_by BIGINT UNSIGNED DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_by BIGINT UNSIGNED DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL,
			INDEX idx_enterprise_id (enterprise_id),
			INDEX idx_status (status),
			INDEX idx_deleted_at (deleted_at)
		)`)
	}
	if !tableExists(ctx, db, "finance_services") {
		db.ExecContext(ctx, `CREATE TABLE finance_services (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			enterprise_id BIGINT UNSIGNED NOT NULL,
			name VARCHAR(200) NOT NULL,
			finance_type VARCHAR(64) DEFAULT '',
			min_amount DECIMAL(14,2) DEFAULT 0,
			max_amount DECIMAL(14,2) DEFAULT 0,
			term_days INT DEFAULT 0,
			annual_rate DECIMAL(8,4) DEFAULT 0,
			fee_rule VARCHAR(500) DEFAULT '',
			risk_control_rule TEXT,
			required_documents JSON NULL,
			supported_scenarios JSON NULL,
			description TEXT,
			status TINYINT DEFAULT 1,
			audit_remark VARCHAR(500) DEFAULT '',
			approved_at DATETIME NULL,
			created_by BIGINT UNSIGNED DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_by BIGINT UNSIGNED DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL,
			INDEX idx_enterprise_id (enterprise_id),
			INDEX idx_status (status),
			INDEX idx_deleted_at (deleted_at)
		)`)
	}
	ensureServiceOfferColumns(ctx, db)
}

func ensureServiceOfferColumns(ctx context.Context, db *sql.DB) {
	if tableExists(ctx, db, "warehouse_services") {
		if !columnExists(ctx, db, "warehouse_services", "loading_capacity") {
			db.ExecContext(ctx, `ALTER TABLE warehouse_services ADD COLUMN loading_capacity VARCHAR(255) DEFAULT ''`)
		}
		if !columnExists(ctx, db, "warehouse_services", "business_hours") {
			db.ExecContext(ctx, `ALTER TABLE warehouse_services ADD COLUMN business_hours VARCHAR(255) DEFAULT ''`)
		}
		if !columnExists(ctx, db, "warehouse_services", "pledge_supervision") {
			db.ExecContext(ctx, `ALTER TABLE warehouse_services ADD COLUMN pledge_supervision TINYINT(1) DEFAULT 0`)
		}
		if !columnExists(ctx, db, "warehouse_services", "outbound_voucher_req") {
			db.ExecContext(ctx, `ALTER TABLE warehouse_services ADD COLUMN outbound_voucher_req VARCHAR(500) DEFAULT ''`)
		}
	}
	if tableExists(ctx, db, "logistics_services") {
		if !columnExists(ctx, db, "logistics_services", "quote_rules") {
			db.ExecContext(ctx, `ALTER TABLE logistics_services ADD COLUMN quote_rules JSON NULL`)
		}
		if !columnExists(ctx, db, "logistics_services", "fleet_name") {
			db.ExecContext(ctx, `ALTER TABLE logistics_services ADD COLUMN fleet_name VARCHAR(200) DEFAULT ''`)
		}
		if !columnExists(ctx, db, "logistics_services", "vehicle_certs") {
			db.ExecContext(ctx, `ALTER TABLE logistics_services ADD COLUMN vehicle_certs JSON NULL`)
		}
		if !columnExists(ctx, db, "logistics_services", "driver_certs") {
			db.ExecContext(ctx, `ALTER TABLE logistics_services ADD COLUMN driver_certs JSON NULL`)
		}
	}
}

func ListWarehouseServiceOffers(c *gin.Context) {
	listServiceOffers(c, "warehouse")
}

func ListLogisticsServiceOffers(c *gin.Context) {
	listServiceOffers(c, "logistics")
}

func ListFinanceServiceOffers(c *gin.Context) {
	listServiceOffers(c, "finance")
}

func AuditWarehouseServiceOffer(c *gin.Context) {
	auditServiceOffer(c, "warehouse_services")
}

func AuditLogisticsServiceOffer(c *gin.Context) {
	auditServiceOffer(c, "logistics_services")
}

func AuditFinanceServiceOffer(c *gin.Context) {
	auditServiceOffer(c, "finance_services")
}

func listServiceOffers(c *gin.Context, offerType string) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	ensureServiceOfferTables(ctx, db)
	limit, offset := parsePageParams(c.Query("page"), c.Query("pageSize"))

	table := "warehouse_services"
	nameExpr := "warehouse_name"
	selectCols := `id, enterprise_id, warehouse_name, COALESCE(province,''), COALESCE(city,''), COALESCE(address,''), COALESCE(capacity,0), COALESCE(available_capacity,0), '', '', '', '', 0, COALESCE(price_rule,''), COALESCE(loading_capacity,''), COALESCE(business_hours,''), COALESCE(pledge_supervision,0), COALESCE(outbound_voucher_req,''), '', '', '', '', '', 0, 0, 0, 0, status, COALESCE(audit_remark,''), created_at, updated_at`
	if offerType == "logistics" {
		table = "logistics_services"
		nameExpr = "name"
		selectCols = `id, enterprise_id, name, '', '', '', 0, 0, COALESCE(route_start,''), COALESCE(route_end,''), COALESCE(service_scope,''), COALESCE(vehicle_type,''), COALESCE(load_capacity,0), COALESCE(price_rule,''), '', '', 0, '', COALESCE(CAST(quote_rules AS CHAR),''), COALESCE(fleet_name,''), COALESCE(CAST(vehicle_certs AS CHAR),''), COALESCE(CAST(driver_certs AS CHAR),''), '', 0, 0, 0, 0, status, COALESCE(audit_remark,''), created_at, updated_at`
	}
	if offerType == "finance" {
		table = "finance_services"
		nameExpr = "name"
		selectCols = `id, enterprise_id, name, '', '', '', 0, 0, '', '', '', '', 0, COALESCE(fee_rule,''), '', '', 0, '', '', '', '', '', COALESCE(finance_type,''), COALESCE(min_amount,0), COALESCE(max_amount,0), COALESCE(term_days,0), COALESCE(annual_rate,0), status, COALESCE(audit_remark,''), created_at, updated_at`
	}

	where := []string{"deleted_at IS NULL"}
	args := []any{}
	if status := c.Query("status"); status != "" {
		where = append(where, "status = ?")
		args = append(args, status)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		where = append(where, nameExpr+" LIKE ?")
		args = append(args, "%"+keyword+"%")
	}
	whereClause := strings.Join(where, " AND ")
	total, err := queryCount(ctx, db, fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s`, table, whereClause), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`SELECT %s FROM %s WHERE %s ORDER BY id DESC LIMIT ? OFFSET ?`, selectCols, table, whereClause), queryArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer rows.Close()
	items := make([]serviceOfferItem, 0)
	for rows.Next() {
		var item serviceOfferItem
		item.Type = offerType
		if err := rows.Scan(
			&item.ID, &item.EnterpriseID, &item.Name, &item.Province, &item.City, &item.Address,
			&item.Capacity, &item.AvailableCapacity, &item.RouteStart, &item.RouteEnd, &item.ServiceScope,
			&item.VehicleType, &item.LoadCapacity, &item.PriceRule, &item.LoadingCapacity,
			&item.BusinessHours, &item.PledgeSupervision, &item.OutboundVoucherReq, &item.QuoteRules,
			&item.FleetName, &item.VehicleCerts, &item.DriverCerts, &item.FinanceType,
			&item.MinAmount, &item.MaxAmount, &item.TermDays, &item.AnnualRate, &item.Status,
			&item.AuditRemark, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		items = append(items, item)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": pageResult[serviceOfferItem]{List: items, Total: total}})
}

func auditServiceOffer(c *gin.Context, table string) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	ensureServiceOfferTables(ctx, db)
	var req struct {
		Status           int    `json:"status"`
		AuditAction      string `json:"auditAction"`
		AuditActionSnake string `json:"audit_action"`
		AuditRemark      string `json:"auditRemark"`
		AuditRemarkSnake string `json:"audit_remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "审核状态无效"})
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
	statusValue, ok := normalizeServiceOfferAuditStatus(req.Status, auditAction)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "审核状态无效，仅支持通过或驳回"})
		return
	}
	if !rejectReasonProvided(statusValue, 3, auditRemark) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "驳回原因不能为空"})
		return
	}
	setParts := []string{"status = ?", "audit_remark = ?", "updated_at = ?"}
	args := []any{statusValue, strings.TrimSpace(auditRemark), time.Now()}
	if statusValue == 2 {
		setParts = append(setParts, "approved_at = ?")
		args = append(args, time.Now())
	} else {
		setParts = append(setParts, "approved_at = NULL")
	}
	args = append(args, c.Param("id"))
	result, err := db.ExecContext(ctx, fmt.Sprintf(`UPDATE %s SET %s WHERE id = ? AND deleted_at IS NULL`, table, strings.Join(setParts, ", ")), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "资源不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": true})
}

func normalizeServiceOfferAuditStatus(status int, auditAction string) (int, bool) {
	switch strings.ToLower(strings.TrimSpace(auditAction)) {
	case "approve":
		return 2, true
	case "reject":
		return 3, true
	}
	switch status {
	case 1, 2:
		return 2, true
	case 3:
		return 3, true
	}
	return 0, false
}
