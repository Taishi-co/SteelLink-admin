package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-admin/backend/internal/database"
	"github.com/gin-gonic/gin"
)

type groupProcurementItem struct {
	ID                 int64                  `json:"id"`
	ProcurementNo      string                 `json:"procurementNo"`
	Title              string                 `json:"title"`
	LeaderName         string                 `json:"leaderName"`
	Status             string                 `json:"status"`
	ProductID          int64                  `json:"productId"`
	ProductNo          string                 `json:"productNo"`
	ProductName        string                 `json:"productName"`
	SellerEnterpriseID int64                  `json:"sellerEnterpriseId"`
	SellerName         string                 `json:"sellerName"`
	GroupPrice         float64                `json:"groupPrice"`
	MinGroupQuantity   float64                `json:"minGroupQuantity"`
	MaxGroupQuantity   float64                `json:"maxGroupQuantity"`
	Category           string                 `json:"category"`
	Target             string                 `json:"target"`
	TargetQuantity     float64                `json:"targetQuantity"`
	Unit               string                 `json:"unit"`
	BudgetMin          float64                `json:"budgetMin"`
	BudgetMax          float64                `json:"budgetMax"`
	BudgetDisplay      string                 `json:"budgetDisplay"`
	Stage              string                 `json:"stage"`
	Progress           int                    `json:"progress"`
	ParticipatingCount int                    `json:"participatingCount"`
	DemandQuantity     float64                `json:"demandQuantity"`
	DeliveryDays       int                    `json:"deliveryDays"`
	AdminRemark        string                 `json:"adminRemark"`
	SellerRemark       string                 `json:"sellerRemark"`
	Source             string                 `json:"source"`
	StartedAt          *time.Time             `json:"startedAt,omitempty"`
	EndedAt            *time.Time             `json:"endedAt,omitempty"`
	SellerRespondedAt  *time.Time             `json:"sellerRespondedAt,omitempty"`
	ClosedAt           *time.Time             `json:"closedAt,omitempty"`
	MasterOrderID      int64                  `json:"masterOrderId"`
	MasterOrderNo      string                 `json:"masterOrderNo"`
	Participants       []groupParticipantItem `json:"participants,omitempty"`
	CreatedAt          time.Time              `json:"createdAt"`
	UpdatedAt          time.Time              `json:"updatedAt"`
}

type groupParticipantItem struct {
	ID                int64      `json:"id"`
	ActivityID        int64      `json:"activityId"`
	BuyerEnterpriseID int64      `json:"buyerEnterpriseId"`
	BuyerUserID       int64      `json:"buyerUserId"`
	BuyerName         string     `json:"buyerName"`
	Quantity          float64    `json:"quantity"`
	Unit              string     `json:"unit"`
	UnitPrice         float64    `json:"unitPrice"`
	TotalAmount       float64    `json:"totalAmount"`
	ContactName       string     `json:"contactName"`
	ContactPhone      string     `json:"contactPhone"`
	DeliveryAddress   string     `json:"deliveryAddress"`
	AgreementSigned   bool       `json:"agreementSigned"`
	AgreementSignedAt *time.Time `json:"agreementSignedAt,omitempty"`
	Status            string     `json:"status"`
	ChildOrderID      int64      `json:"childOrderId"`
	ChildOrderNo      string     `json:"childOrderNo"`
	Remark            string     `json:"remark"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

type groupProcurementRequestItem struct {
	ID                int64     `json:"id"`
	RequestNo         string    `json:"requestNo"`
	Source            string    `json:"source"`
	Category          string    `json:"category"`
	ProductName       string    `json:"productName"`
	Quantity          float64   `json:"quantity"`
	Unit              string    `json:"unit"`
	DeliveryAddress   string    `json:"deliveryAddress"`
	BuyerEnterpriseID int64     `json:"buyerEnterpriseId"`
	BuyerUserID       int64     `json:"buyerUserId"`
	BuyerName         string    `json:"buyerName"`
	Status            string    `json:"status"`
	ActivityID        int64     `json:"activityId"`
	Remark            string    `json:"remark"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type createGroupProcurementRequest struct {
	ProductID      int64   `json:"productId" binding:"required"`
	Title          string  `json:"title"`
	TargetQuantity float64 `json:"targetQuantity"`
	DurationHours  int     `json:"durationHours"`
	AdminRemark    string  `json:"adminRemark"`
}

type endGroupProcurementRequest struct {
	Remark string `json:"remark"`
}

type groupProductSnapshot struct {
	ID                 int64
	ProductNo          string
	Name               string
	Category           string
	SubCategory        string
	SteelMill          string
	UnitPrice          float64
	AvailableQty       float64
	MinOrderQty        float64
	Currency           string
	TaxIncluded        bool
	TaxRate            float64
	SellerEnterpriseID int64
	SellerName         string
	Status             int
	WarehouseInfo      map[string]any
	DeliveryDays       int
}

func ListGroupProcurements(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := ensureGroupProcurementSchema(ctx, db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	limit, offset := parsePageParams(c.Query("page"), c.Query("pageSize"))
	whereParts := []string{"deleted_at IS NULL"}
	args := make([]any, 0)
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		kw := "%" + keyword + "%"
		whereParts = append(whereParts, "(title LIKE ? OR procurement_no LIKE ? OR product_name LIKE ? OR seller_name LIKE ?)")
		args = append(args, kw, kw, kw, kw)
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		whereParts = append(whereParts, "status = ?")
		args = append(args, status)
	}
	if category := strings.TrimSpace(c.Query("category")); category != "" {
		whereParts = append(whereParts, "category = ?")
		args = append(args, category)
	}
	whereClause := strings.Join(whereParts, " AND ")
	total, err := queryCount(ctx, db, fmt.Sprintf("SELECT COUNT(*) FROM procurement_activities WHERE %s", whereClause), args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	rows, err := db.QueryContext(ctx, groupProcurementSelectSQL("WHERE "+whereClause+" ORDER BY id DESC LIMIT ? OFFSET ?"), append(args, limit, offset)...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer rows.Close()
	items := make([]groupProcurementItem, 0)
	for rows.Next() {
		item, err := scanGroupProcurement(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		items = append(items, item)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": pageResult[groupProcurementItem]{List: items, Total: total}})
}

func GetGroupProcurement(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := ensureGroupProcurementSchema(ctx, db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid procurement id"})
		return
	}
	row := db.QueryRowContext(ctx, groupProcurementSelectSQL("WHERE deleted_at IS NULL AND id = ? LIMIT 1"), id)
	item, err := scanGroupProcurement(row)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "group procurement not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	participants, err := listGroupParticipants(ctx, db, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	item.Participants = participants
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": item})
}

func CreateGroupProcurement(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := ensureGroupProcurementSchema(ctx, db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	var req createGroupProcurementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	product, err := getGroupProductSnapshot(ctx, db, req.ProductID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "product not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if product.Status != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "只有已上架商品可以发起集采"})
		return
	}
	adminID := currentAdminID(c)
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "平台集采：" + product.Name
	}
	targetQuantity := req.TargetQuantity
	if targetQuantity <= 0 {
		targetQuantity = product.MinOrderQty
	}
	if targetQuantity <= 0 {
		targetQuantity = 1
	}
	durationHours := req.DurationHours
	if durationHours <= 0 {
		durationHours = 24
	}
	procurementNo := generateAdminProcurementNo()
	now := time.Now()
	openDate := now.Add(time.Duration(durationHours) * time.Hour).Format("2006-01-02")
	_, err = db.ExecContext(ctx, `
		INSERT INTO procurement_activities
		(procurement_no, title, leader_name, status, product_id, product_no, product_name, seller_enterprise_id, seller_name,
		 category, target, target_quantity, unit, budget_min, budget_max, budget_display, stage, progress, region, open_date,
		 participating_count, demand_quantity, qualified_suppliers, quote_count, contract_count, delivery_days, quote_mode,
		 evaluation_mode, admin_remark, source, created_by, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, 'pending_seller', ?, ?, ?, ?, ?, ?, ?, ?, '吨', 0, 0, '', '待卖家确认', 5, '', ?, 0, 0, 0, 0, 0, ?, '卖家确认价', '最小成团量', ?, 'admin', ?, ?, ?, ?)`,
		procurementNo, title, "平台运营", product.ID, product.ProductNo, product.Name, product.SellerEnterpriseID, product.SellerName,
		firstNonEmpty(product.SubCategory, product.Category), product.Name, targetQuantity, openDate, product.DeliveryDays, strings.TrimSpace(req.AdminRemark), adminID, adminID, now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	var id int64
	if err := db.QueryRowContext(ctx, `SELECT id FROM procurement_activities WHERE procurement_no = ?`, procurementNo).Scan(&id); err == nil {
		_, _ = db.ExecContext(ctx, `INSERT INTO procurement_logs (activity_id, content, color, created_at) VALUES (?, ?, 'blue', ?)`, id, "管理端向卖家发起集采申请", now)
	}
	logOpFromContext(c, db, "group-procurement", "create", 0, fmt.Sprintf(`{"productId":%d,"procurementNo":"%s"}`, product.ID, procurementNo))
	GetGroupProcurementByID(c, id)
}

func EndGroupProcurement(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	if err := ensureGroupProcurementSchema(ctx, db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid procurement id"})
		return
	}
	var req endGroupProcurementRequest
	_ = c.ShouldBindJSON(&req)
	item, err := getGroupProcurementByID(ctx, db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "group procurement not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if item.MasterOrderID > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该集采已生成订单"})
		return
	}
	if item.Status != "active" && item.Status != "ended" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "只有已公开的集采可以结束生成订单"})
		return
	}
	participants, err := listGroupParticipants(ctx, db, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	totalQuantity := 0.0
	for _, participant := range participants {
		if participant.Status == "joined" {
			totalQuantity += participant.Quantity
		}
	}
	if totalQuantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "暂无买家参团，不能生成订单"})
		return
	}
	if item.MinGroupQuantity > 0 && totalQuantity < item.MinGroupQuantity {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "未达到最小成团量，不能生成订单"})
		return
	}
	product, err := getGroupProductSnapshot(ctx, db, item.ProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	adminID := currentAdminID(c)
	masterOrderID, masterOrderNo, err := createGroupOrders(ctx, db, item, product, participants, totalQuantity, adminID, strings.TrimSpace(req.Remark))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	logOpFromContext(c, db, "group-procurement", "end", 0, fmt.Sprintf(`{"id":%d,"masterOrderId":%d}`, id, masterOrderID))
	result, _ := getGroupProcurementByID(ctx, db, id)
	result.MasterOrderID = masterOrderID
	result.MasterOrderNo = masterOrderNo
	result.Participants, _ = listGroupParticipants(ctx, db, id)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "集采已结束并生成订单", "data": result})
}

func ListGroupProcurementRequests(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := ensureGroupProcurementSchema(ctx, db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	limit, offset := parsePageParams(c.Query("page"), c.Query("pageSize"))
	where := "deleted_at IS NULL"
	args := make([]any, 0)
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		where += " AND (request_no LIKE ? OR product_name LIKE ? OR category LIKE ? OR buyer_name LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw, kw)
	}
	total, err := queryCount(ctx, db, "SELECT COUNT(*) FROM procurement_requests WHERE "+where, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id, COALESCE(request_no,''), COALESCE(source,''), COALESCE(category,''), COALESCE(product_name,''), COALESCE(quantity,0), COALESCE(unit,'吨'),
		       COALESCE(delivery_address,''), COALESCE(buyer_enterprise_id,0), COALESCE(buyer_user_id,0), COALESCE(buyer_name,''), COALESCE(status,'pending'),
		       COALESCE(activity_id,0), COALESCE(remark,''), created_at, updated_at
		FROM procurement_requests WHERE `+where+` ORDER BY id DESC LIMIT ? OFFSET ?`, append(args, limit, offset)...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer rows.Close()
	items := make([]groupProcurementRequestItem, 0)
	for rows.Next() {
		var item groupProcurementRequestItem
		if err := rows.Scan(&item.ID, &item.RequestNo, &item.Source, &item.Category, &item.ProductName, &item.Quantity, &item.Unit, &item.DeliveryAddress, &item.BuyerEnterpriseID, &item.BuyerUserID, &item.BuyerName, &item.Status, &item.ActivityID, &item.Remark, &item.CreatedAt, &item.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		items = append(items, item)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": pageResult[groupProcurementRequestItem]{List: items, Total: total}})
}

func GetGroupProcurementByID(c *gin.Context, id int64) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	item, err := getGroupProcurementByID(ctx, db, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": item})
}

func groupProcurementSelectSQL(tail string) string {
	return `
		SELECT id, COALESCE(procurement_no,''), COALESCE(title,''), COALESCE(leader_name,''), COALESCE(status,''),
		       COALESCE(product_id,0), COALESCE(product_no,''), COALESCE(product_name,''), COALESCE(seller_enterprise_id,0), COALESCE(seller_name,''),
		       COALESCE(group_price,0), COALESCE(min_group_quantity,0), COALESCE(max_group_quantity,0), COALESCE(category,''), COALESCE(target,''),
		       COALESCE(target_quantity,0), COALESCE(unit,'吨'), COALESCE(budget_min,0), COALESCE(budget_max,0), COALESCE(budget_display,''),
		       COALESCE(stage,''), COALESCE(progress,0), COALESCE(participating_count,0), COALESCE(demand_quantity,0), COALESCE(delivery_days,0),
		       COALESCE(admin_remark,''), COALESCE(seller_remark,''), COALESCE(source,''), started_at, ended_at, seller_responded_at, closed_at,
		       COALESCE(master_order_id,0), COALESCE(master_order_no,''), created_at, updated_at
		FROM procurement_activities ` + tail
}

func scanGroupProcurement(scanner sqlScanner) (groupProcurementItem, error) {
	var item groupProcurementItem
	var startedAt, endedAt, sellerRespondedAt, closedAt sql.NullTime
	err := scanner.Scan(
		&item.ID, &item.ProcurementNo, &item.Title, &item.LeaderName, &item.Status,
		&item.ProductID, &item.ProductNo, &item.ProductName, &item.SellerEnterpriseID, &item.SellerName,
		&item.GroupPrice, &item.MinGroupQuantity, &item.MaxGroupQuantity, &item.Category, &item.Target,
		&item.TargetQuantity, &item.Unit, &item.BudgetMin, &item.BudgetMax, &item.BudgetDisplay,
		&item.Stage, &item.Progress, &item.ParticipatingCount, &item.DemandQuantity, &item.DeliveryDays,
		&item.AdminRemark, &item.SellerRemark, &item.Source, &startedAt, &endedAt, &sellerRespondedAt, &closedAt,
		&item.MasterOrderID, &item.MasterOrderNo, &item.CreatedAt, &item.UpdatedAt,
	)
	if startedAt.Valid {
		item.StartedAt = &startedAt.Time
	}
	if endedAt.Valid {
		item.EndedAt = &endedAt.Time
	}
	if sellerRespondedAt.Valid {
		item.SellerRespondedAt = &sellerRespondedAt.Time
	}
	if closedAt.Valid {
		item.ClosedAt = &closedAt.Time
	}
	return item, err
}

func getGroupProcurementByID(ctx context.Context, db *sql.DB, id int64) (groupProcurementItem, error) {
	row := db.QueryRowContext(ctx, groupProcurementSelectSQL("WHERE deleted_at IS NULL AND id = ? LIMIT 1"), id)
	item, err := scanGroupProcurement(row)
	if err != nil {
		return item, err
	}
	item.Participants, _ = listGroupParticipants(ctx, db, id)
	return item, nil
}

func listGroupParticipants(ctx context.Context, db *sql.DB, activityID int64) ([]groupParticipantItem, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, activity_id, COALESCE(buyer_enterprise_id,0), COALESCE(buyer_user_id,0), COALESCE(buyer_name,''), COALESCE(quantity,0),
		       COALESCE(unit,'吨'), COALESCE(unit_price,0), COALESCE(total_amount,0), COALESCE(contact_name,''), COALESCE(contact_phone,''),
		       COALESCE(delivery_address,''), COALESCE(agreement_signed,0), agreement_signed_at, COALESCE(status,'joined'),
		       COALESCE(child_order_id,0), COALESCE(child_order_no,''), COALESCE(remark,''), created_at, updated_at
		FROM procurement_participants
		WHERE deleted_at IS NULL AND activity_id = ? AND status <> 'cancelled'
		ORDER BY id ASC`, activityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]groupParticipantItem, 0)
	for rows.Next() {
		var item groupParticipantItem
		var signedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.ActivityID, &item.BuyerEnterpriseID, &item.BuyerUserID, &item.BuyerName, &item.Quantity, &item.Unit, &item.UnitPrice, &item.TotalAmount, &item.ContactName, &item.ContactPhone, &item.DeliveryAddress, &item.AgreementSigned, &signedAt, &item.Status, &item.ChildOrderID, &item.ChildOrderNo, &item.Remark, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		if signedAt.Valid {
			item.AgreementSignedAt = &signedAt.Time
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func getGroupProductSnapshot(ctx context.Context, db *sql.DB, productID int64) (groupProductSnapshot, error) {
	var product groupProductSnapshot
	var warehouseRaw sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT p.id, COALESCE(p.product_no,''), COALESCE(p.name,''), COALESCE(p.category,''), COALESCE(p.sub_category,''), COALESCE(p.steel_mill,''),
		       COALESCE(p.unit_price,0), COALESCE(p.available_qty, p.quantity, 0), COALESCE(p.min_order_qty,0), COALESCE(p.currency,'CNY'),
		       COALESCE(p.tax_included,1), COALESCE(p.tax_rate,13), COALESCE(p.enterprise_id,0), COALESCE(e.name,''), COALESCE(p.status,0),
		       p.warehouse_info, COALESCE(p.delivery_days,0)
		FROM products p
		LEFT JOIN enterprises e ON e.id = p.enterprise_id
		WHERE p.deleted_at IS NULL AND p.id = ?
		LIMIT 1`, productID).
		Scan(&product.ID, &product.ProductNo, &product.Name, &product.Category, &product.SubCategory, &product.SteelMill, &product.UnitPrice, &product.AvailableQty, &product.MinOrderQty, &product.Currency, &product.TaxIncluded, &product.TaxRate, &product.SellerEnterpriseID, &product.SellerName, &product.Status, &warehouseRaw, &product.DeliveryDays)
	if err != nil {
		return product, err
	}
	product.WarehouseInfo = parseJSONMap(warehouseRaw)
	return product, nil
}

func createGroupOrders(ctx context.Context, db *sql.DB, activity groupProcurementItem, product groupProductSnapshot, participants []groupParticipantItem, totalQuantity float64, adminID int64, remark string) (int64, string, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, "", err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE products SET available_qty = available_qty - ?, locked_qty = locked_qty + ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL AND available_qty >= ?`, totalQuantity, totalQuantity, product.ID, totalQuantity)
	if err != nil {
		return 0, "", err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return 0, "", fmt.Errorf("商品可用库存不足，无法生成集采订单")
	}
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	masterNo := generateAdminOrderNo()
	masterItems := buildGroupOrderItems(product, activity.GroupPrice, totalQuantity)
	masterRemark := map[string]any{"source": "group_procurement_master", "group_procurement_id": activity.ID, "participant_count": len(participants), "admin_remark": remark}
	masterDelivery := map[string]any{"address": "按各子订单收货地址履约"}
	masterID, err := insertGroupOrder(ctx, tx, masterNo, 0, "集采母订单", activity.SellerEnterpriseID, activity.SellerName, masterItems, activity.GroupPrice*totalQuantity, 1, masterDelivery, masterRemark, adminID, now, expiresAt)
	if err != nil {
		return 0, "", err
	}
	if err := insertGroupOrderEvent(ctx, tx, masterID, masterNo, "group_procurement_master_created", adminID, 0, 0, 1, masterRemark, now); err != nil {
		return 0, "", err
	}
	for _, participant := range participants {
		if participant.Status != "joined" {
			continue
		}
		childNo := generateAdminOrderNo()
		childItems := buildGroupOrderItems(product, activity.GroupPrice, participant.Quantity)
		delivery := productDeliveryInfo(product.WarehouseInfo)
		delivery["address"] = participant.DeliveryAddress
		childRemark := map[string]any{
			"source":               "group_procurement_child",
			"group_procurement_id": activity.ID,
			"master_order_id":      masterID,
			"master_order_no":      masterNo,
			"contact_name":         participant.ContactName,
			"contact_phone":        participant.ContactPhone,
			"agreement_signed":     participant.AgreementSigned,
		}
		childID, err := insertGroupOrder(ctx, tx, childNo, participant.BuyerEnterpriseID, participant.BuyerName, activity.SellerEnterpriseID, activity.SellerName, childItems, activity.GroupPrice*participant.Quantity, 12, delivery, childRemark, adminID, now, expiresAt)
		if err != nil {
			return 0, "", err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE procurement_participants SET status = 'order_created', child_order_id = ?, child_order_no = ?, updated_at = ? WHERE id = ?`, childID, childNo, now, participant.ID); err != nil {
			return 0, "", err
		}
		if err := insertGroupOrderEvent(ctx, tx, childID, childNo, "group_procurement_child_created", adminID, participant.BuyerEnterpriseID, 0, 12, childRemark, now); err != nil {
			return 0, "", err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE procurement_activities SET status = 'order_created', stage = '已生成订单', progress = 100, contract_count = ?, master_order_id = ?, master_order_no = ?, closed_at = ?, updated_at = ? WHERE id = ?`, len(participants), masterID, masterNo, now, now, activity.ID); err != nil {
		return 0, "", err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO procurement_logs (activity_id, content, color, created_at) VALUES (?, ?, 'green', ?)`, activity.ID, "管理端结束集采并生成母子订单", now); err != nil {
		return 0, "", err
	}
	if err := tx.Commit(); err != nil {
		return 0, "", err
	}
	return masterID, masterNo, nil
}

func insertGroupOrder(ctx context.Context, tx *sql.Tx, orderNo string, buyerID int64, buyerName string, sellerID int64, sellerName string, items []map[string]any, totalAmount float64, status int, delivery map[string]any, remark map[string]any, adminID int64, now, expiresAt time.Time) (int64, error) {
	itemsJSON, _ := json.Marshal(items)
	deliveryJSON, _ := json.Marshal(delivery)
	remarkJSON, _ := json.Marshal(remark)
	result, err := tx.ExecContext(ctx, `
		INSERT INTO orders
		(order_no, buyer_id, buyer_name, seller_id, seller_enterprise_id, seller_name, items, total_quantity, total_amount, tax_total_amount, currency, status, delivery_info, expires_at, remark, created_by, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'CNY', ?, ?, ?, ?, ?, ?, ?, ?)`,
		orderNo, buyerID, buyerName, sellerID, sellerID, sellerName, string(itemsJSON), items[0]["quantity"], totalAmount, totalAmount, status, string(deliveryJSON), expiresAt, string(remarkJSON), adminID, adminID, now, now)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func insertGroupOrderEvent(ctx context.Context, tx *sql.Tx, orderID int64, orderNo, eventType string, adminID, actorEnterpriseID int64, fromStatus, toStatus int, payload map[string]any, now time.Time) error {
	if !tableExists(ctx, database.GetDB(), "order_events") {
		return nil
	}
	payloadJSON, _ := json.Marshal(payload)
	_, err := tx.ExecContext(ctx, `
		INSERT INTO order_events
		(order_id, order_no, event_type, actor_user_id, actor_enterprise_id, actor_service, from_status, to_status, payload, created_at)
		VALUES (?, ?, ?, ?, ?, 'admin', ?, ?, ?, ?)`,
		orderID, orderNo, eventType, adminID, actorEnterpriseID, fromStatus, toStatus, string(payloadJSON), now)
	return err
}

func buildGroupOrderItems(product groupProductSnapshot, unitPrice, quantity float64) []map[string]any {
	return []map[string]any{{
		"product_id":   product.ID,
		"product_no":   product.ProductNo,
		"product_name": product.Name,
		"steel_mill":   product.SteelMill,
		"category":     firstNonEmpty(product.SubCategory, product.Category),
		"quantity":     quantity,
		"unit_price":   unitPrice,
		"total_amount": unitPrice * quantity,
		"tax_rate":     product.TaxRate,
		"tax_amount":   0,
	}}
}

func productDeliveryInfo(warehouse map[string]any) map[string]any {
	result := make(map[string]any)
	copyField := func(from, to string) {
		if value, ok := warehouse[from]; ok {
			result[to] = value
		}
	}
	copyField("warehouse_id", "warehouse_id")
	copyField("warehouse_service_id", "warehouse_service_id")
	copyField("warehouse_enterprise_id", "warehouse_enterprise_id")
	copyField("warehouse_name", "warehouse_name")
	copyField("warehouse_code", "warehouse_code")
	copyField("province", "warehouse_province")
	copyField("city", "warehouse_city")
	copyField("district", "warehouse_district")
	copyField("address", "warehouse_address")
	return result
}

func ensureGroupProcurementSchema(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database unavailable")
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS procurement_activities (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			procurement_no VARCHAR(40) UNIQUE,
			title VARCHAR(200),
			leader_name VARCHAR(120),
			status VARCHAR(32),
			product_id BIGINT UNSIGNED DEFAULT 0,
			product_no VARCHAR(64),
			product_name VARCHAR(200),
			seller_enterprise_id BIGINT UNSIGNED DEFAULT 0,
			seller_name VARCHAR(200),
			group_price DECIMAL(16,2) DEFAULT 0,
			min_group_quantity DECIMAL(14,2) DEFAULT 0,
			max_group_quantity DECIMAL(14,2) DEFAULT 0,
			category VARCHAR(80),
			target VARCHAR(120),
			target_quantity DECIMAL(14,2) DEFAULT 0,
			unit VARCHAR(20) DEFAULT '吨',
			budget_min DECIMAL(16,2) DEFAULT 0,
			budget_max DECIMAL(16,2) DEFAULT 0,
			budget_display VARCHAR(60),
			stage VARCHAR(60),
			progress INT DEFAULT 0,
			region VARCHAR(120),
			open_date VARCHAR(40),
			participating_count INT DEFAULT 0,
			demand_quantity DECIMAL(14,2) DEFAULT 0,
			qualified_suppliers INT DEFAULT 0,
			quote_count INT DEFAULT 0,
			contract_count INT DEFAULT 0,
			delivery_days INT DEFAULT 0,
			quote_mode VARCHAR(40),
			evaluation_mode VARCHAR(40),
			documents JSON NULL,
			admin_remark VARCHAR(500),
			seller_remark VARCHAR(500),
			source VARCHAR(40),
			started_at DATETIME NULL,
			ended_at DATETIME NULL,
			seller_responded_at DATETIME NULL,
			closed_at DATETIME NULL,
			master_order_id BIGINT UNSIGNED DEFAULT 0,
			master_order_no VARCHAR(64),
			created_by BIGINT UNSIGNED DEFAULT 0,
			updated_by BIGINT UNSIGNED DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL
		)`,
		`CREATE TABLE IF NOT EXISTS procurement_logs (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			activity_id BIGINT UNSIGNED,
			content VARCHAR(255),
			color VARCHAR(20) DEFAULT 'gray',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS procurement_participants (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			activity_id BIGINT UNSIGNED NOT NULL,
			buyer_enterprise_id BIGINT UNSIGNED DEFAULT 0,
			buyer_user_id BIGINT UNSIGNED DEFAULT 0,
			buyer_name VARCHAR(200),
			quantity DECIMAL(14,2) DEFAULT 0,
			unit VARCHAR(20) DEFAULT '吨',
			unit_price DECIMAL(16,2) DEFAULT 0,
			total_amount DECIMAL(16,2) DEFAULT 0,
			contact_name VARCHAR(80),
			contact_phone VARCHAR(40),
			delivery_address VARCHAR(500),
			agreement_signed TINYINT(1) DEFAULT 0,
			agreement_signed_at DATETIME NULL,
			status VARCHAR(32) DEFAULT 'joined',
			child_order_id BIGINT UNSIGNED DEFAULT 0,
			child_order_no VARCHAR(64),
			remark VARCHAR(500),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL
		)`,
		`CREATE TABLE IF NOT EXISTS procurement_requests (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			request_no VARCHAR(40) UNIQUE,
			source VARCHAR(40),
			category VARCHAR(80),
			product_name VARCHAR(200),
			quantity DECIMAL(14,2) DEFAULT 0,
			unit VARCHAR(20) DEFAULT '吨',
			delivery_address VARCHAR(500),
			buyer_enterprise_id BIGINT UNSIGNED DEFAULT 0,
			buyer_user_id BIGINT UNSIGNED DEFAULT 0,
			buyer_name VARCHAR(200),
			status VARCHAR(32) DEFAULT 'pending',
			activity_id BIGINT UNSIGNED DEFAULT 0,
			remark VARCHAR(500),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL
		)`,
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	columns := map[string]string{
		"product_id":           "BIGINT UNSIGNED DEFAULT 0",
		"product_no":           "VARCHAR(64)",
		"product_name":         "VARCHAR(200)",
		"seller_enterprise_id": "BIGINT UNSIGNED DEFAULT 0",
		"seller_name":          "VARCHAR(200)",
		"group_price":          "DECIMAL(16,2) DEFAULT 0",
		"min_group_quantity":   "DECIMAL(14,2) DEFAULT 0",
		"max_group_quantity":   "DECIMAL(14,2) DEFAULT 0",
		"admin_remark":         "VARCHAR(500)",
		"seller_remark":        "VARCHAR(500)",
		"source":               "VARCHAR(40)",
		"started_at":           "DATETIME NULL",
		"ended_at":             "DATETIME NULL",
		"seller_responded_at":  "DATETIME NULL",
		"closed_at":            "DATETIME NULL",
		"master_order_id":      "BIGINT UNSIGNED DEFAULT 0",
		"master_order_no":      "VARCHAR(64)",
		"created_by":           "BIGINT UNSIGNED DEFAULT 0",
		"updated_by":           "BIGINT UNSIGNED DEFAULT 0",
	}
	for name, definition := range columns {
		if !columnExists(ctx, db, "procurement_activities", name) {
			if _, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE procurement_activities ADD COLUMN %s %s", name, definition)); err != nil {
				return err
			}
		}
	}
	return nil
}

func generateAdminProcurementNo() string {
	return "GP" + time.Now().Format("20060102150405") + randomNumeric(4)
}

func generateAdminOrderNo() string {
	return "O" + time.Now().Format("20060102150405") + randomNumeric(4)
}

func randomNumeric(length int) string {
	const digits = "0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		result[i] = digits[n.Int64()]
	}
	return string(result)
}
