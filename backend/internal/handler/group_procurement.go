package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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
	MasterOrderStatus  int                    `json:"masterOrderStatus"`
	MasterOrderUpdated *time.Time             `json:"masterOrderUpdatedAt,omitempty"`
	Participants       []groupParticipantItem `json:"participants,omitempty"`
	CreatedAt          time.Time              `json:"createdAt"`
	UpdatedAt          time.Time              `json:"updatedAt"`
}

type groupParticipantItem struct {
	ID                        int64      `json:"id"`
	ActivityID                int64      `json:"activityId"`
	BuyerEnterpriseID         int64      `json:"buyerEnterpriseId"`
	BuyerUserID               int64      `json:"buyerUserId"`
	BuyerName                 string     `json:"buyerName"`
	Quantity                  float64    `json:"quantity"`
	Unit                      string     `json:"unit"`
	UnitPrice                 float64    `json:"unitPrice"`
	TotalAmount               float64    `json:"totalAmount"`
	ContactName               string     `json:"contactName"`
	ContactPhone              string     `json:"contactPhone"`
	DeliveryAddressSource     string     `json:"deliveryAddressSource"`
	DestWarehouseServiceID    int64      `json:"destWarehouseServiceId"`
	DestWarehouseEnterpriseID int64      `json:"destWarehouseEnterpriseId"`
	DestWarehouseName         string     `json:"destWarehouseName"`
	DestWarehouseProvince     string     `json:"destWarehouseProvince"`
	DestWarehouseCity         string     `json:"destWarehouseCity"`
	DestWarehouseDistrict     string     `json:"destWarehouseDistrict"`
	DestWarehouseAddress      string     `json:"destWarehouseAddress"`
	DeliveryAddress           string     `json:"deliveryAddress"`
	DeliveryMethod            string     `json:"deliveryMethod"`
	PaymentMethod             string     `json:"paymentMethod"`
	DeliveryNote              string     `json:"deliveryNote"`
	AgreementSigned           bool       `json:"agreementSigned"`
	AgreementSignedAt         *time.Time `json:"agreementSignedAt,omitempty"`
	Status                    string     `json:"status"`
	ChildOrderID              int64      `json:"childOrderId"`
	ChildOrderNo              string     `json:"childOrderNo"`
	ChildOrderStatus          int        `json:"childOrderStatus"`
	ChildOrderUpdated         *time.Time `json:"childOrderUpdatedAt,omitempty"`
	Remark                    string     `json:"remark"`
	CreatedAt                 time.Time  `json:"createdAt"`
	UpdatedAt                 time.Time  `json:"updatedAt"`
}

type groupProcurementRequestItem struct {
	ID                        int64     `json:"id"`
	RequestNo                 string    `json:"requestNo"`
	Source                    string    `json:"source"`
	SourceProductID           int64     `json:"sourceProductId"`
	AIProcurementID           int64     `json:"aiProcurementId"`
	Category                  string    `json:"category"`
	SubCategory               string    `json:"subCategory"`
	ProductName               string    `json:"productName"`
	Material                  string    `json:"material"`
	Spec                      string    `json:"spec"`
	Quantity                  float64   `json:"quantity"`
	Unit                      string    `json:"unit"`
	TargetPrice               float64   `json:"targetPrice"`
	ContactName               string    `json:"contactName"`
	ContactPhone              string    `json:"contactPhone"`
	DeliveryAddressSource     string    `json:"deliveryAddressSource"`
	DestWarehouseServiceID    int64     `json:"destWarehouseServiceId"`
	DestWarehouseEnterpriseID int64     `json:"destWarehouseEnterpriseId"`
	DestWarehouseName         string    `json:"destWarehouseName"`
	DestWarehouseProvince     string    `json:"destWarehouseProvince"`
	DestWarehouseCity         string    `json:"destWarehouseCity"`
	DestWarehouseDistrict     string    `json:"destWarehouseDistrict"`
	DestWarehouseAddress      string    `json:"destWarehouseAddress"`
	DeliveryAddress           string    `json:"deliveryAddress"`
	DeliveryMethod            string    `json:"deliveryMethod"`
	PaymentMethod             string    `json:"paymentMethod"`
	DeliveryNote              string    `json:"deliveryNote"`
	BuyerEnterpriseID         int64     `json:"buyerEnterpriseId"`
	BuyerUserID               int64     `json:"buyerUserId"`
	BuyerName                 string    `json:"buyerName"`
	Status                    string    `json:"status"`
	ActivityID                int64     `json:"activityId"`
	Remark                    string    `json:"remark"`
	CreatedAt                 time.Time `json:"createdAt"`
	UpdatedAt                 time.Time `json:"updatedAt"`
}

type createGroupProcurementRequest struct {
	RequestID      int64   `json:"requestId"`
	RequestIDs     []int64 `json:"requestIds"`
	ProductID      int64   `json:"productId" binding:"required"`
	Title          string  `json:"title"`
	TargetQuantity float64 `json:"targetQuantity"`
	DurationHours  int     `json:"durationHours"`
	AdminRemark    string  `json:"adminRemark"`
}

type endGroupProcurementRequest struct {
	Remark string `json:"remark"`
}

type rejectGroupProcurementRequest struct {
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

func StartGroupProcurementAutoFinalizer(interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		run := func() {
			db := database.GetDB()
			if db == nil {
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := ensureGroupProcurementSchema(ctx, db); err != nil {
				log.Printf("group procurement auto finalizer schema error: %v", err)
				return
			}
			if err := reconcileExpiredGroupProcurements(ctx, db); err != nil {
				log.Printf("group procurement auto finalizer reconcile error: %v", err)
				return
			}
			if _, err := finalizeDueGroupProcurements(ctx, db, 0); err != nil {
				log.Printf("group procurement auto finalizer error: %v", err)
			}
		}
		run()
		for range ticker.C {
			run()
		}
	}()
}

func ListGroupProcurements(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := ensureGroupProcurementSchema(ctx, db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if err := reconcileExpiredGroupProcurements(ctx, db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if _, err := finalizeDueGroupProcurements(ctx, db, currentAdminID(c)); err != nil {
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
	if err := reconcileExpiredGroupProcurements(ctx, db); err != nil {
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
	requestIDs := uniquePositiveIDs(req.RequestIDs)
	if req.RequestID > 0 {
		requestIDs = uniquePositiveIDs(append(requestIDs, req.RequestID))
	}
	sourceRequests := make([]groupProcurementRequestItem, 0)
	var sourceRequest groupProcurementRequestItem
	var err error
	if len(requestIDs) > 0 {
		sourceRequests, err = listGroupProcurementRequestsByIDs(ctx, db, requestIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		if len(sourceRequests) != len(requestIDs) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "部分集采申请不存在"})
			return
		}
		for _, item := range sourceRequests {
			if item.Status != "pending" {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "存在已处理的集采申请"})
				return
			}
		}
		if len(sourceRequests) > 0 {
			sourceRequest = sourceRequests[0]
		}
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
		if sourceRequest.ID > 0 && sourceRequest.ProductName != "" {
			title = "平台集采：" + sourceRequest.ProductName
		} else {
			title = "平台集采：" + product.Name
		}
	}
	targetQuantity := req.TargetQuantity
	if targetQuantity <= 0 && len(sourceRequests) > 0 {
		for _, item := range sourceRequests {
			targetQuantity += item.Quantity
		}
	}
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
	adminRemark := strings.TrimSpace(req.AdminRemark)
	if adminRemark == "" && len(sourceRequests) > 0 {
		adminRemark = strings.TrimSpace(sourceRequests[0].Remark)
	}
	source := "admin"
	if len(sourceRequests) > 0 {
		source = firstNonEmpty(sourceRequest.Source, "manual") + "_request"
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO procurement_activities
		(procurement_no, title, leader_name, status, product_id, product_no, product_name, seller_enterprise_id, seller_name,
		 category, target, target_quantity, unit, budget_min, budget_max, budget_display, stage, progress, region, open_date,
		 participating_count, demand_quantity, qualified_suppliers, quote_count, contract_count, delivery_days, quote_mode,
		 evaluation_mode, admin_remark, source, created_by, updated_by, created_at, updated_at)
		VALUES (?, ?, ?, 'pending_seller', ?, ?, ?, ?, ?, ?, ?, ?, '吨', 0, 0, '', '待卖家确认', 5, '', ?, 0, 0, 0, 0, 0, ?, '卖家确认价', '最小成团量', ?, ?, ?, ?, ?, ?)`,
		procurementNo, title, "平台运营", product.ID, product.ProductNo, product.Name, product.SellerEnterpriseID, product.SellerName,
		firstNonEmpty(product.SubCategory, product.Category), product.Name, targetQuantity, openDate, product.DeliveryDays, adminRemark, source, adminID, adminID, now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	var id int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM procurement_activities WHERE procurement_no = ?`, procurementNo).Scan(&id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO procurement_logs (activity_id, content, color, created_at) VALUES (?, ?, 'blue', ?)`, id, "管理端向卖家发起集采申请", now); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	for _, item := range sourceRequests {
		if _, err := tx.ExecContext(ctx, `INSERT IGNORE INTO procurement_activity_requests (activity_id, request_id, created_at) VALUES (?, ?, ?)`, id, item.ID, now); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		invitationNo := generateAdminProcurementInvitationNo()
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO procurement_invitations
			(invitation_no, activity_id, request_id, buyer_enterprise_id, buyer_user_id, buyer_name, status, expires_at, remark, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, 'pending', NULL, ?, ?, ?)`,
			invitationNo, id, item.ID, item.BuyerEnterpriseID, item.BuyerUserID, item.BuyerName, strings.TrimSpace(item.Remark), now, now); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		result, err := tx.ExecContext(ctx, `UPDATE procurement_requests SET status = 'invited', activity_id = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL AND status = 'pending'`, id, now, item.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		if affected, err := result.RowsAffected(); err == nil && affected == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "存在已处理的集采申请"})
			return
		}
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	committed = true
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
	if err := reconcileExpiredGroupProcurements(ctx, db); err != nil {
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

func RejectGroupProcurementRequest(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := ensureGroupProcurementSchema(ctx, db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request id"})
		return
	}
	var req rejectGroupProcurementRequest
	_ = c.ShouldBindJSON(&req)
	existing, err := getGroupProcurementRequestByID(ctx, db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "集采申请不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if existing.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该集采申请已处理"})
		return
	}
	remark := strings.TrimSpace(req.Remark)
	updateSQL := `UPDATE procurement_requests SET status = 'rejected', updated_at = ? WHERE id = ? AND deleted_at IS NULL`
	args := []any{time.Now(), id}
	if remark != "" {
		updateSQL = `UPDATE procurement_requests SET status = 'rejected', remark = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`
		args = []any{remark, time.Now(), id}
	}
	if _, err := db.ExecContext(ctx, updateSQL, args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	logOpFromContext(c, db, "group-procurement-request", "reject", int(id), fmt.Sprintf(`{"requestId":%d}`, id))
	item, _ := getGroupProcurementRequestByID(ctx, db, id)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "集采申请已驳回", "data": item})
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
		SELECT id, COALESCE(request_no,''), COALESCE(source,''), COALESCE(source_product_id,0), COALESCE(ai_procurement_id,0),
		       COALESCE(category,''), COALESCE(sub_category,''), COALESCE(product_name,''), COALESCE(material,''), COALESCE(spec,''),
		       COALESCE(quantity,0), COALESCE(unit,'吨'), COALESCE(target_price,0), COALESCE(contact_name,''), COALESCE(contact_phone,''),
		       COALESCE(delivery_address_source,''), COALESCE(dest_warehouse_service_id,0), COALESCE(dest_warehouse_enterprise_id,0),
		       COALESCE(dest_warehouse_name,''), COALESCE(dest_warehouse_province,''), COALESCE(dest_warehouse_city,''), COALESCE(dest_warehouse_district,''), COALESCE(dest_warehouse_address,''),
		       COALESCE(delivery_address,''), COALESCE(delivery_method,''), COALESCE(payment_method,''), COALESCE(delivery_note,''),
		       COALESCE(buyer_enterprise_id,0), COALESCE(buyer_user_id,0), COALESCE(buyer_name,''), COALESCE(status,'pending'),
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
		if err := scanGroupProcurementRequest(rows, &item); err != nil {
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
			       COALESCE(master_order_id,0), COALESCE(master_order_no,''),
			       COALESCE((SELECT o.status FROM orders o WHERE o.id = procurement_activities.master_order_id AND o.deleted_at IS NULL LIMIT 1), 0),
			       (SELECT o.updated_at FROM orders o WHERE o.id = procurement_activities.master_order_id AND o.deleted_at IS NULL LIMIT 1),
			       created_at, updated_at
			FROM procurement_activities ` + tail
}

func scanGroupProcurement(scanner sqlScanner) (groupProcurementItem, error) {
	var item groupProcurementItem
	var startedAt, endedAt, sellerRespondedAt, closedAt, masterOrderUpdated sql.NullTime
	err := scanner.Scan(
		&item.ID, &item.ProcurementNo, &item.Title, &item.LeaderName, &item.Status,
		&item.ProductID, &item.ProductNo, &item.ProductName, &item.SellerEnterpriseID, &item.SellerName,
		&item.GroupPrice, &item.MinGroupQuantity, &item.MaxGroupQuantity, &item.Category, &item.Target,
		&item.TargetQuantity, &item.Unit, &item.BudgetMin, &item.BudgetMax, &item.BudgetDisplay,
		&item.Stage, &item.Progress, &item.ParticipatingCount, &item.DemandQuantity, &item.DeliveryDays,
		&item.AdminRemark, &item.SellerRemark, &item.Source, &startedAt, &endedAt, &sellerRespondedAt, &closedAt,
		&item.MasterOrderID, &item.MasterOrderNo, &item.MasterOrderStatus, &masterOrderUpdated, &item.CreatedAt, &item.UpdatedAt,
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
	if masterOrderUpdated.Valid {
		item.MasterOrderUpdated = &masterOrderUpdated.Time
	}
	return item, err
}

func reconcileExpiredGroupProcurements(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		UPDATE procurement_activities
		SET status = 'ended',
		    stage = '已截止',
		    progress = 100,
		    closed_at = COALESCE(closed_at, NOW()),
		    updated_at = NOW()
		WHERE deleted_at IS NULL
		  AND status = 'active'
		  AND ended_at IS NOT NULL
		  AND ended_at <= NOW()`)
	return err
}

func finalizeDueGroupProcurements(ctx context.Context, db *sql.DB, operatorID int64) (int, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id
		FROM procurement_activities
		WHERE deleted_at IS NULL
		  AND status = 'ended'
		  AND COALESCE(master_order_id, 0) = 0
		  AND ended_at IS NOT NULL
		  AND ended_at <= NOW()
		ORDER BY ended_at ASC, id ASC
		LIMIT 20`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	finalized := 0
	for _, id := range ids {
		item, err := getGroupProcurementByID(ctx, db, id)
		if err != nil {
			return finalized, err
		}
		if item.MasterOrderID > 0 || item.Status != "ended" {
			continue
		}
		participants, err := listGroupParticipants(ctx, db, id)
		if err != nil {
			return finalized, err
		}
		totalQuantity := 0.0
		for _, participant := range participants {
			if participant.Status == "joined" {
				totalQuantity += participant.Quantity
			}
		}
		if totalQuantity <= 0 || (item.MinGroupQuantity > 0 && totalQuantity < item.MinGroupQuantity) {
			continue
		}
		product, err := getGroupProductSnapshot(ctx, db, item.ProductID)
		if err != nil {
			if logErr := insertGroupProcurementFailureLog(ctx, db, id, "系统到期自动成单失败：商品信息不可用"); logErr != nil {
				return finalized, logErr
			}
			continue
		}
		_, _, err = createGroupOrders(ctx, db, item, product, participants, totalQuantity, operatorID, "系统到期自动成单")
		if err != nil {
			if logErr := insertGroupProcurementFailureLog(ctx, db, id, "系统到期自动成单失败："+truncateLogContent(err.Error(), 180)); logErr != nil {
				return finalized, logErr
			}
			continue
		}
		finalized++
	}
	return finalized, nil
}

func insertGroupProcurementFailureLog(ctx context.Context, db *sql.DB, activityID int64, content string) error {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM procurement_logs WHERE activity_id = ? AND content = ?`, activityID, content).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	_, err := db.ExecContext(ctx, `INSERT INTO procurement_logs (activity_id, content, color, created_at) VALUES (?, ?, 'red', ?)`, activityID, content, time.Now())
	return err
}

func truncateLogContent(value string, max int) string {
	value = strings.TrimSpace(value)
	if max <= 0 || len([]rune(value)) <= max {
		return value
	}
	runes := []rune(value)
	return string(runes[:max])
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

func getGroupProcurementRequestByID(ctx context.Context, db *sql.DB, id int64) (groupProcurementRequestItem, error) {
	var item groupProcurementRequestItem
	err := scanGroupProcurementRequest(db.QueryRowContext(ctx, groupProcurementRequestSelectSQL("WHERE deleted_at IS NULL AND id = ? LIMIT 1"), id), &item)
	return item, err
}

func listGroupProcurementRequestsByIDs(ctx context.Context, db *sql.DB, ids []int64) ([]groupProcurementRequestItem, error) {
	ids = uniquePositiveIDs(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	rows, err := db.QueryContext(ctx, groupProcurementRequestSelectSQL("WHERE deleted_at IS NULL AND id IN ("+strings.Join(placeholders, ",")+")"), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]groupProcurementRequestItem, 0, len(ids))
	for rows.Next() {
		var item groupProcurementRequestItem
		if err := scanGroupProcurementRequest(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func groupProcurementRequestSelectSQL(tail string) string {
	return `
		SELECT id, COALESCE(request_no,''), COALESCE(source,''), COALESCE(source_product_id,0), COALESCE(ai_procurement_id,0),
		       COALESCE(category,''), COALESCE(sub_category,''), COALESCE(product_name,''), COALESCE(material,''), COALESCE(spec,''),
		       COALESCE(quantity,0), COALESCE(unit,'吨'), COALESCE(target_price,0), COALESCE(contact_name,''), COALESCE(contact_phone,''),
		       COALESCE(delivery_address_source,''), COALESCE(dest_warehouse_service_id,0), COALESCE(dest_warehouse_enterprise_id,0),
		       COALESCE(dest_warehouse_name,''), COALESCE(dest_warehouse_province,''), COALESCE(dest_warehouse_city,''), COALESCE(dest_warehouse_district,''), COALESCE(dest_warehouse_address,''),
		       COALESCE(delivery_address,''), COALESCE(delivery_method,''), COALESCE(payment_method,''), COALESCE(delivery_note,''),
		       COALESCE(buyer_enterprise_id,0), COALESCE(buyer_user_id,0), COALESCE(buyer_name,''), COALESCE(status,'pending'),
		       COALESCE(activity_id,0), COALESCE(remark,''), created_at, updated_at
		FROM procurement_requests ` + tail
}

func scanGroupProcurementRequest(scanner sqlScanner, item *groupProcurementRequestItem) error {
	return scanner.Scan(
		&item.ID, &item.RequestNo, &item.Source, &item.SourceProductID, &item.AIProcurementID,
		&item.Category, &item.SubCategory, &item.ProductName, &item.Material, &item.Spec,
		&item.Quantity, &item.Unit, &item.TargetPrice, &item.ContactName, &item.ContactPhone,
		&item.DeliveryAddressSource, &item.DestWarehouseServiceID, &item.DestWarehouseEnterpriseID,
		&item.DestWarehouseName, &item.DestWarehouseProvince, &item.DestWarehouseCity, &item.DestWarehouseDistrict, &item.DestWarehouseAddress,
		&item.DeliveryAddress, &item.DeliveryMethod, &item.PaymentMethod, &item.DeliveryNote,
		&item.BuyerEnterpriseID, &item.BuyerUserID, &item.BuyerName, &item.Status,
		&item.ActivityID, &item.Remark, &item.CreatedAt, &item.UpdatedAt,
	)
}

func listGroupParticipants(ctx context.Context, db *sql.DB, activityID int64) ([]groupParticipantItem, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT p.id, p.activity_id, COALESCE(p.buyer_enterprise_id,0), COALESCE(p.buyer_user_id,0), COALESCE(p.buyer_name,''), COALESCE(p.quantity,0),
		       COALESCE(p.unit,'吨'), COALESCE(p.unit_price,0), COALESCE(p.total_amount,0), COALESCE(p.contact_name,''), COALESCE(p.contact_phone,''),
		       COALESCE(p.delivery_address_source,''), COALESCE(p.dest_warehouse_service_id,0), COALESCE(p.dest_warehouse_enterprise_id,0),
		       COALESCE(p.dest_warehouse_name,''), COALESCE(p.dest_warehouse_province,''), COALESCE(p.dest_warehouse_city,''), COALESCE(p.dest_warehouse_district,''), COALESCE(p.dest_warehouse_address,''),
		       COALESCE(p.delivery_address,''), COALESCE(p.agreement_signed,0), p.agreement_signed_at, COALESCE(p.status,'joined'),
		       COALESCE(p.delivery_method,''), COALESCE(p.payment_method,''), COALESCE(p.delivery_note,''),
		       COALESCE(p.child_order_id,0), COALESCE(p.child_order_no,''), COALESCE(o.status,0), o.updated_at,
		       COALESCE(p.remark,''), p.created_at, p.updated_at
		FROM procurement_participants p
		LEFT JOIN orders o ON o.id = p.child_order_id AND o.deleted_at IS NULL
		WHERE p.deleted_at IS NULL AND p.activity_id = ? AND p.status <> 'cancelled'
		ORDER BY p.id ASC`, activityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]groupParticipantItem, 0)
	for rows.Next() {
		var item groupParticipantItem
		var signedAt, childOrderUpdated sql.NullTime
		if err := rows.Scan(&item.ID, &item.ActivityID, &item.BuyerEnterpriseID, &item.BuyerUserID, &item.BuyerName, &item.Quantity, &item.Unit, &item.UnitPrice, &item.TotalAmount, &item.ContactName, &item.ContactPhone, &item.DeliveryAddressSource, &item.DestWarehouseServiceID, &item.DestWarehouseEnterpriseID, &item.DestWarehouseName, &item.DestWarehouseProvince, &item.DestWarehouseCity, &item.DestWarehouseDistrict, &item.DestWarehouseAddress, &item.DeliveryAddress, &item.AgreementSigned, &signedAt, &item.Status, &item.DeliveryMethod, &item.PaymentMethod, &item.DeliveryNote, &item.ChildOrderID, &item.ChildOrderNo, &item.ChildOrderStatus, &childOrderUpdated, &item.Remark, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		if signedAt.Valid {
			item.AgreementSignedAt = &signedAt.Time
		}
		if childOrderUpdated.Valid {
			item.ChildOrderUpdated = &childOrderUpdated.Time
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
	var existingMasterID int64
	var currentStatus string
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(master_order_id,0), COALESCE(status,'') FROM procurement_activities WHERE id = ? AND deleted_at IS NULL FOR UPDATE`, activity.ID).Scan(&existingMasterID, &currentStatus); err != nil {
		return 0, "", err
	}
	if existingMasterID > 0 {
		return 0, "", fmt.Errorf("该集采已生成订单")
	}
	if currentStatus != "active" && currentStatus != "ended" {
		return 0, "", fmt.Errorf("当前集采状态不可生成订单")
	}
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
	createdCount := 0
	for _, participant := range participants {
		if participant.Status != "joined" {
			continue
		}
		createdCount++
		childNo := generateAdminOrderNo()
		childItems := buildGroupOrderItems(product, activity.GroupPrice, participant.Quantity)
		delivery := productDeliveryInfo(product.WarehouseInfo)
		delivery["address"] = participant.DeliveryAddress
		if participant.DestWarehouseServiceID > 0 {
			delivery["dest_warehouse_service_id"] = participant.DestWarehouseServiceID
			delivery["dest_warehouse_enterprise_id"] = participant.DestWarehouseEnterpriseID
			delivery["dest_warehouse_name"] = participant.DestWarehouseName
			delivery["dest_warehouse_province"] = participant.DestWarehouseProvince
			delivery["dest_warehouse_city"] = participant.DestWarehouseCity
			delivery["dest_warehouse_district"] = participant.DestWarehouseDistrict
			delivery["dest_warehouse_address"] = participant.DestWarehouseAddress
		}
		childStatus := 12
		if participant.DestWarehouseServiceID > 0 {
			childStatus = 13
		}
		childRemark := map[string]any{
			"source":                  "group_procurement_child",
			"group_procurement_id":    activity.ID,
			"master_order_id":         masterID,
			"master_order_no":         masterNo,
			"contact_name":            participant.ContactName,
			"contact_phone":           participant.ContactPhone,
			"delivery_address_source": participant.DeliveryAddressSource,
			"delivery_address":        participant.DeliveryAddress,
			"delivery_method":         participant.DeliveryMethod,
			"payment_method":          participant.PaymentMethod,
			"delivery_note":           participant.DeliveryNote,
			"agreement_signed":        participant.AgreementSigned,
		}
		childID, err := insertGroupOrder(ctx, tx, childNo, participant.BuyerEnterpriseID, participant.BuyerName, activity.SellerEnterpriseID, activity.SellerName, childItems, activity.GroupPrice*participant.Quantity, childStatus, delivery, childRemark, adminID, now, expiresAt)
		if err != nil {
			return 0, "", err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE procurement_participants SET status = 'order_created', child_order_id = ?, child_order_no = ?, updated_at = ? WHERE id = ?`, childID, childNo, now, participant.ID); err != nil {
			return 0, "", err
		}
		if err := insertGroupOrderEvent(ctx, tx, childID, childNo, "group_procurement_child_created", adminID, participant.BuyerEnterpriseID, 0, childStatus, childRemark, now); err != nil {
			return 0, "", err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE procurement_activities SET status = 'order_created', stage = '已生成订单', progress = 100, contract_count = ?, master_order_id = ?, master_order_no = ?, closed_at = ?, updated_at = ? WHERE id = ?`, createdCount, masterID, masterNo, now, now, activity.ID); err != nil {
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
			delivery_address_source VARCHAR(32),
			dest_warehouse_service_id BIGINT UNSIGNED DEFAULT 0,
			dest_warehouse_enterprise_id BIGINT UNSIGNED DEFAULT 0,
			dest_warehouse_name VARCHAR(200),
			dest_warehouse_province VARCHAR(80),
			dest_warehouse_city VARCHAR(80),
			dest_warehouse_district VARCHAR(80),
			dest_warehouse_address VARCHAR(500),
			delivery_address VARCHAR(500),
			delivery_method VARCHAR(60),
			payment_method VARCHAR(60),
			delivery_note VARCHAR(500),
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
			source_product_id BIGINT UNSIGNED DEFAULT 0,
			ai_procurement_id BIGINT UNSIGNED DEFAULT 0,
			category VARCHAR(80),
			sub_category VARCHAR(80),
			product_name VARCHAR(200),
			material VARCHAR(120),
			spec VARCHAR(200),
			quantity DECIMAL(14,2) DEFAULT 0,
			unit VARCHAR(20) DEFAULT '吨',
			target_price DECIMAL(16,2) DEFAULT 0,
			contact_name VARCHAR(80),
			contact_phone VARCHAR(40),
			delivery_address_source VARCHAR(32),
			dest_warehouse_service_id BIGINT UNSIGNED DEFAULT 0,
			dest_warehouse_enterprise_id BIGINT UNSIGNED DEFAULT 0,
			dest_warehouse_name VARCHAR(200),
			dest_warehouse_province VARCHAR(80),
			dest_warehouse_city VARCHAR(80),
			dest_warehouse_district VARCHAR(80),
			dest_warehouse_address VARCHAR(500),
			delivery_address VARCHAR(500),
			delivery_method VARCHAR(60),
			payment_method VARCHAR(60),
			delivery_note VARCHAR(500),
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
		`CREATE TABLE IF NOT EXISTS procurement_activity_requests (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			activity_id BIGINT UNSIGNED NOT NULL,
			request_id BIGINT UNSIGNED NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY idx_procurement_activity_request (activity_id, request_id)
		)`,
		`CREATE TABLE IF NOT EXISTS procurement_invitations (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			invitation_no VARCHAR(40) UNIQUE,
			activity_id BIGINT UNSIGNED DEFAULT 0,
			request_id BIGINT UNSIGNED DEFAULT 0,
			buyer_enterprise_id BIGINT UNSIGNED DEFAULT 0,
			buyer_user_id BIGINT UNSIGNED DEFAULT 0,
			buyer_name VARCHAR(200),
			status VARCHAR(32) DEFAULT 'pending',
			participant_id BIGINT UNSIGNED DEFAULT 0,
			remark VARCHAR(500),
			expires_at DATETIME NULL,
			accepted_at DATETIME NULL,
			declined_at DATETIME NULL,
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
	participantColumns := map[string]string{
		"delivery_address_source":      "VARCHAR(32)",
		"dest_warehouse_service_id":    "BIGINT UNSIGNED DEFAULT 0",
		"dest_warehouse_enterprise_id": "BIGINT UNSIGNED DEFAULT 0",
		"dest_warehouse_name":          "VARCHAR(200)",
		"dest_warehouse_province":      "VARCHAR(80)",
		"dest_warehouse_city":          "VARCHAR(80)",
		"dest_warehouse_district":      "VARCHAR(80)",
		"dest_warehouse_address":       "VARCHAR(500)",
		"delivery_method":              "VARCHAR(60)",
		"payment_method":               "VARCHAR(60)",
		"delivery_note":                "VARCHAR(500)",
	}
	for name, definition := range participantColumns {
		if !columnExists(ctx, db, "procurement_participants", name) {
			if _, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE procurement_participants ADD COLUMN %s %s", name, definition)); err != nil {
				return err
			}
		}
	}
	requestColumns := map[string]string{
		"source_product_id":            "BIGINT UNSIGNED DEFAULT 0",
		"ai_procurement_id":            "BIGINT UNSIGNED DEFAULT 0",
		"sub_category":                 "VARCHAR(80)",
		"material":                     "VARCHAR(120)",
		"spec":                         "VARCHAR(200)",
		"target_price":                 "DECIMAL(16,2) DEFAULT 0",
		"contact_name":                 "VARCHAR(80)",
		"contact_phone":                "VARCHAR(40)",
		"delivery_address_source":      "VARCHAR(32)",
		"dest_warehouse_service_id":    "BIGINT UNSIGNED DEFAULT 0",
		"dest_warehouse_enterprise_id": "BIGINT UNSIGNED DEFAULT 0",
		"dest_warehouse_name":          "VARCHAR(200)",
		"dest_warehouse_province":      "VARCHAR(80)",
		"dest_warehouse_city":          "VARCHAR(80)",
		"dest_warehouse_district":      "VARCHAR(80)",
		"dest_warehouse_address":       "VARCHAR(500)",
		"delivery_method":              "VARCHAR(60)",
		"payment_method":               "VARCHAR(60)",
		"delivery_note":                "VARCHAR(500)",
	}
	for name, definition := range requestColumns {
		if !columnExists(ctx, db, "procurement_requests", name) {
			if _, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE procurement_requests ADD COLUMN %s %s", name, definition)); err != nil {
				return err
			}
		}
	}
	return nil
}

func generateAdminProcurementNo() string {
	return "GP" + time.Now().Format("20060102150405") + randomNumeric(4)
}

func generateAdminProcurementInvitationNo() string {
	return "GPI" + time.Now().Format("20060102150405") + randomNumeric(4)
}

func generateAdminOrderNo() string {
	return "O" + time.Now().Format("20060102150405") + randomNumeric(4)
}

func uniquePositiveIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
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
