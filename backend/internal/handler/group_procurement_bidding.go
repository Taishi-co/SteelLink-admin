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

type groupProcurementProjectItem struct {
	ID                   int64                               `json:"id"`
	ProjectNo            string                              `json:"projectNo"`
	Title                string                              `json:"title"`
	Status               string                              `json:"status"`
	Category             string                              `json:"category"`
	SubCategory          string                              `json:"subCategory"`
	ProductName          string                              `json:"productName"`
	Material             string                              `json:"material"`
	Spec                 string                              `json:"spec"`
	TargetQuantity       float64                             `json:"targetQuantity"`
	DemandQuantity       float64                             `json:"demandQuantity"`
	Unit                 string                              `json:"unit"`
	TargetPrice          float64                             `json:"targetPrice"`
	DeliveryAddress      string                              `json:"deliveryAddress"`
	BuyerRequestCount    int                                 `json:"buyerRequestCount"`
	BuyerEnterpriseCount int                                 `json:"buyerEnterpriseCount"`
	InvitedSupplierCount int                                 `json:"invitedSupplierCount"`
	QuotedSupplierCount  int                                 `json:"quotedSupplierCount"`
	QuoteDeadline        *time.Time                          `json:"quoteDeadline,omitempty"`
	AwardedAt            *time.Time                          `json:"awardedAt,omitempty"`
	OrderedAt            *time.Time                          `json:"orderedAt,omitempty"`
	AwardQuoteID         int64                               `json:"awardQuoteId"`
	AwardSupplierID      int64                               `json:"awardSupplierId"`
	AwardSupplierName    string                              `json:"awardSupplierName"`
	AwardProductID       int64                               `json:"awardProductId"`
	AwardProductNo       string                              `json:"awardProductNo"`
	AwardProductName     string                              `json:"awardProductName"`
	AwardUnitPrice       float64                             `json:"awardUnitPrice"`
	AwardDeliveryDays    int                                 `json:"awardDeliveryDays"`
	OrderID              int64                               `json:"orderId"`
	OrderNo              string                              `json:"orderNo"`
	AdminRemark          string                              `json:"adminRemark"`
	Requests             []groupProcurementRequestItem       `json:"requests,omitempty"`
	Enrollments          []groupProcurementProjectSellerItem `json:"enrollments,omitempty"`
	Quotes               []groupProcurementProjectQuoteItem  `json:"quotes,omitempty"`
	CreatedAt            time.Time                           `json:"createdAt"`
	UpdatedAt            time.Time                           `json:"updatedAt"`
}

type groupProcurementProjectSellerItem struct {
	ID                 int64      `json:"id"`
	ProjectID          int64      `json:"projectId"`
	SellerEnterpriseID int64      `json:"sellerEnterpriseId"`
	SellerUserID       int64      `json:"sellerUserId"`
	SellerName         string     `json:"sellerName"`
	ProductID          int64      `json:"productId"`
	ProductNo          string     `json:"productNo"`
	ProductName        string     `json:"productName"`
	Status             string     `json:"status"`
	Remark             string     `json:"remark"`
	ReviewedBy         int64      `json:"reviewedBy"`
	ReviewedAt         *time.Time `json:"reviewedAt,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

type groupProcurementProjectQuoteItem struct {
	ID                 int64     `json:"id"`
	ProjectID          int64     `json:"projectId"`
	EnrollmentID       int64     `json:"enrollmentId"`
	SellerEnterpriseID int64     `json:"sellerEnterpriseId"`
	SellerUserID       int64     `json:"sellerUserId"`
	SellerName         string    `json:"sellerName"`
	ProductID          int64     `json:"productId"`
	ProductNo          string    `json:"productNo"`
	ProductName        string    `json:"productName"`
	UnitPrice          float64   `json:"unitPrice"`
	SupplyQuantity     float64   `json:"supplyQuantity"`
	Unit               string    `json:"unit"`
	DeliveryDays       int       `json:"deliveryDays"`
	DeliveryWarehouse  string    `json:"deliveryWarehouse"`
	TaxIncluded        bool      `json:"taxIncluded"`
	TaxRate            float64   `json:"taxRate"`
	PaymentTerms       string    `json:"paymentTerms"`
	Remark             string    `json:"remark"`
	IsLatest           bool      `json:"isLatest"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type createGroupProcurementProjectRequest struct {
	RequestIDs     []int64 `json:"requestIds"`
	Title          string  `json:"title"`
	TargetQuantity float64 `json:"targetQuantity"`
	DurationHours  int     `json:"durationHours"`
	AdminRemark    string  `json:"adminRemark"`
}

type inviteGroupProcurementProjectSellersRequest struct {
	ProductIDs []int64 `json:"productIds"`
}

type awardGroupProcurementProjectRequest struct {
	QuoteID int64  `json:"quoteId"`
	Remark  string `json:"remark"`
}

type createGroupProcurementProjectOrderRequest struct {
	Remark string `json:"remark"`
}

func ListGroupProcurementProjectsV2(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := ensureGroupProcurementBiddingSchema(ctx, db); err != nil {
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
	if category := strings.TrimSpace(c.Query("category")); category != "" {
		where += " AND (category = ? OR sub_category = ?)"
		args = append(args, category, category)
	}
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		kw := "%" + keyword + "%"
		where += " AND (project_no LIKE ? OR title LIKE ? OR product_name LIKE ?)"
		args = append(args, kw, kw, kw)
	}
	total, err := queryCount(ctx, db, "SELECT COUNT(*) FROM procurement_projects WHERE "+where, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	rows, err := db.QueryContext(ctx, procurementProjectSelectSQL("WHERE "+where+" ORDER BY updated_at DESC, id DESC LIMIT ? OFFSET ?"), append(args, limit, offset)...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer rows.Close()
	list := make([]groupProcurementProjectItem, 0)
	for rows.Next() {
		item, err := scanProcurementProject(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		list = append(list, item)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": pageResult[groupProcurementProjectItem]{List: list, Total: total}})
}

func GetGroupProcurementProjectV2(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := ensureGroupProcurementBiddingSchema(ctx, db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid project id"})
		return
	}
	item, err := getGroupProcurementProjectByID(ctx, db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "项目不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": item})
}

func CreateGroupProcurementProjectV2(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := ensureGroupProcurementBiddingSchema(ctx, db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	var req createGroupProcurementProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	requestIDs := uniquePositiveIDs(req.RequestIDs)
	if len(requestIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "至少选择一条集采申请"})
		return
	}
	requests, err := listGroupProcurementRequestsByIDs(ctx, db, requestIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if len(requests) != len(requestIDs) {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "部分集采申请不存在"})
		return
	}
	primary := requests[0]
	for _, item := range requests {
		if item.Status != "pending" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "存在已处理的集采申请"})
			return
		}
		if item.Category != primary.Category || item.SubCategory != primary.SubCategory || item.ProductName != primary.ProductName || item.Material != primary.Material || item.Spec != primary.Spec {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "所选采购申请不是同一类商品，不能归并到同一项目"})
			return
		}
	}
	totalQuantity := req.TargetQuantity
	if totalQuantity <= 0 {
		for _, item := range requests {
			totalQuantity += item.Quantity
		}
	}
	durationHours := req.DurationHours
	if durationHours <= 0 {
		durationHours = 48
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "集采竞价项目：" + primary.ProductName
	}
	now := time.Now()
	deadline := now.Add(time.Duration(durationHours) * time.Hour)
	projectNo := "GPP" + time.Now().Format("20060102150405") + randomNumeric(4)
	adminID := currentAdminID(c)
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
		INSERT INTO procurement_projects
		(project_no, title, status, category, sub_category, product_name, material, spec, target_quantity, demand_quantity, unit, target_price, delivery_address, buyer_request_count, buyer_enterprise_count, invited_supplier_count, quoted_supplier_count, quote_deadline, admin_remark, created_by, updated_by, created_at, updated_at)
		VALUES (?, ?, 'collecting', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, 0, 0, ?, ?, ?, ?, ?, ?)`,
		projectNo, title, primary.Category, primary.SubCategory, primary.ProductName, primary.Material, primary.Spec, totalQuantity, totalQuantity, defaultAdminUnit(primary.Unit), primary.TargetPrice, primary.DeliveryAddress, deadline, strings.TrimSpace(req.AdminRemark), adminID, adminID, now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	var projectID int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM procurement_projects WHERE project_no = ?`, projectNo).Scan(&projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	buyerSet := make(map[int64]struct{})
	sumQuantity := 0.0
	for _, item := range requests {
		buyerSet[item.BuyerEnterpriseID] = struct{}{}
		sumQuantity += item.Quantity
		if _, err := tx.ExecContext(ctx, `INSERT INTO procurement_project_requests (project_id, request_id, created_at) VALUES (?, ?, ?)`, projectID, item.ID, now); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		if _, err := tx.ExecContext(ctx, `UPDATE procurement_requests SET status = 'selected', updated_at = ? WHERE id = ? AND deleted_at IS NULL`, now, item.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE procurement_projects SET buyer_request_count = ?, buyer_enterprise_count = ?, demand_quantity = ?, target_quantity = ?, updated_at = ? WHERE id = ?`, len(requests), len(buyerSet), sumQuantity, sumQuantity, now, projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO procurement_logs (project_id, content, color, created_at) VALUES (?, ?, 'blue', ?)`, projectID, "管理端创建集采竞价项目并完成需求归并", now); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	committed = true
	logOpFromContext(c, db, "group-procurement-project", "create", 0, fmt.Sprintf(`{"projectNo":"%s"}`, projectNo))
	item, err := getGroupProcurementProjectByID(ctx, db, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": item})
}

func InviteGroupProcurementProjectSellers(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := ensureGroupProcurementBiddingSchema(ctx, db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || projectID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid project id"})
		return
	}
	var req inviteGroupProcurementProjectSellersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	productIDs := uniquePositiveIDs(req.ProductIDs)
	if len(productIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "至少选择一个卖家商品"})
		return
	}
	project, err := getGroupProcurementProjectByID(ctx, db, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "项目不存在"})
		return
	}
	if project.Status == "awarded" || project.Status == "order_created" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "当前项目状态不可再邀请卖家"})
		return
	}
	adminID := currentAdminID(c)
	now := time.Now()
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
	inserted := 0
	for _, productID := range productIDs {
		product, err := getGroupProductSnapshot(ctx, db, productID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "卖家商品不存在"})
			return
		}
		var existingID int64
		err = tx.QueryRowContext(ctx, `SELECT id FROM procurement_supplier_enrollments WHERE project_id = ? AND seller_enterprise_id = ? AND deleted_at IS NULL LIMIT 1`, projectID, product.SellerEnterpriseID).Scan(&existingID)
		if err == nil && existingID > 0 {
			continue
		}
		if err != nil && err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO procurement_supplier_enrollments
			(project_id, seller_enterprise_id, seller_name, product_id, product_no, product_name, status, remark, reviewed_by, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, 'invited', ?, ?, ?, ?)`,
			projectID, product.SellerEnterpriseID, product.SellerName, product.ID, product.ProductNo, product.Name, "管理端邀请卖家参与报价", adminID, now, now); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		inserted++
	}
	if inserted == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "没有新增可邀请卖家"})
		return
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE procurement_projects
		SET status = 'quoting',
		    invited_supplier_count = (SELECT COUNT(*) FROM procurement_supplier_enrollments WHERE project_id = ? AND deleted_at IS NULL),
		    updated_by = ?,
		    updated_at = ?
		WHERE id = ?`, projectID, adminID, now, projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO procurement_logs (project_id, content, color, created_at) VALUES (?, ?, 'green', ?)`, projectID, fmt.Sprintf("管理端新增邀请 %d 个卖家参与报价", inserted), now); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	committed = true
	item, err := getGroupProcurementProjectByID(ctx, db, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": item})
}

func AwardGroupProcurementProject(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := ensureGroupProcurementBiddingSchema(ctx, db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || projectID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid project id"})
		return
	}
	var req awardGroupProcurementProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if req.QuoteID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请选择中标报价"})
		return
	}
	project, err := getGroupProcurementProjectByID(ctx, db, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "项目不存在"})
		return
	}
	if project.Status == "order_created" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "项目已成单，不能重复定标"})
		return
	}
	quote, err := getGroupProcurementProjectQuoteByID(ctx, db, req.QuoteID)
	if err != nil || quote.ProjectID != projectID || !quote.IsLatest {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "中标报价不存在"})
		return
	}
	if quote.SupplyQuantity < project.DemandQuantity {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "中标报价可供数量不足，无法覆盖项目总需求"})
		return
	}
	now := time.Now()
	adminID := currentAdminID(c)
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
	if _, err := tx.ExecContext(ctx, `UPDATE procurement_supplier_quotes SET status = CASE WHEN id = ? THEN 'awarded' ELSE 'submitted' END, updated_at = ? WHERE project_id = ? AND is_latest = 1 AND deleted_at IS NULL`, req.QuoteID, now, projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO procurement_project_awards
		(project_id, quote_id, seller_enterprise_id, seller_name, product_id, product_no, product_name, unit_price, award_quantity, delivery_days, remark, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			quote_id = VALUES(quote_id),
			seller_enterprise_id = VALUES(seller_enterprise_id),
			seller_name = VALUES(seller_name),
			product_id = VALUES(product_id),
			product_no = VALUES(product_no),
			product_name = VALUES(product_name),
			unit_price = VALUES(unit_price),
			award_quantity = VALUES(award_quantity),
			delivery_days = VALUES(delivery_days),
			remark = VALUES(remark),
			updated_at = VALUES(updated_at)`,
		projectID, quote.ID, quote.SellerEnterpriseID, quote.SellerName, quote.ProductID, quote.ProductNo, quote.ProductName, quote.UnitPrice, project.DemandQuantity, quote.DeliveryDays, strings.TrimSpace(req.Remark), adminID, now, now); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE procurement_projects
		SET status = 'awarded',
		    awarded_at = ?,
		    award_quote_id = ?,
		    award_supplier_id = ?,
		    award_supplier_name = ?,
		    award_product_id = ?,
		    award_product_no = ?,
		    award_product_name = ?,
		    award_unit_price = ?,
		    award_delivery_days = ?,
		    updated_by = ?,
		    updated_at = ?
		WHERE id = ?`,
		now, quote.ID, quote.SellerEnterpriseID, quote.SellerName, quote.ProductID, quote.ProductNo, quote.ProductName, quote.UnitPrice, quote.DeliveryDays, adminID, now, projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO procurement_logs (project_id, content, color, created_at) VALUES (?, ?, 'green', ?)`, projectID, fmt.Sprintf("管理端完成定标，中标卖家：%s", quote.SellerName), now); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	committed = true
	item, err := getGroupProcurementProjectByID(ctx, db, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": item})
}

func CreateGroupProcurementProjectOrder(c *gin.Context) {
	db := database.GetDB()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	if err := ensureGroupProcurementBiddingSchema(ctx, db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	projectID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || projectID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid project id"})
		return
	}
	var req createGroupProcurementProjectOrderRequest
	_ = c.ShouldBindJSON(&req)
	project, err := getGroupProcurementProjectByID(ctx, db, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "项目不存在"})
		return
	}
	if project.OrderID > 0 || project.Status == "order_created" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "项目已成单"})
		return
	}
	if project.Status != "awarded" || project.AwardProductID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请先完成定标"})
		return
	}
	requests, err := listProjectRequests(ctx, db, projectID)
	if err != nil || len(requests) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "项目缺少归集申请，不能成单"})
		return
	}
	product, err := getGroupProductSnapshot(ctx, db, project.AwardProductID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "中标商品不存在"})
		return
	}
	adminID := currentAdminID(c)
	orderID, orderNo, err := createAwardProjectOrders(ctx, db, project, product, requests, adminID, strings.TrimSpace(req.Remark))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	logOpFromContext(c, db, "group-procurement-project", "order", 0, fmt.Sprintf(`{"projectId":%d,"orderId":%d}`, projectID, orderID))
	item, err := getGroupProcurementProjectByID(ctx, db, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	item.OrderID = orderID
	item.OrderNo = orderNo
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": item})
}

func getGroupProcurementProjectByID(ctx context.Context, db *sql.DB, id int64) (groupProcurementProjectItem, error) {
	row := db.QueryRowContext(ctx, procurementProjectSelectSQL("WHERE deleted_at IS NULL AND id = ? LIMIT 1"), id)
	item, err := scanProcurementProject(row)
	if err != nil {
		return item, err
	}
	item.Requests, _ = listProjectRequests(ctx, db, id)
	item.Enrollments, _ = listProjectEnrollments(ctx, db, id)
	item.Quotes, _ = listProjectQuotes(ctx, db, id)
	return item, nil
}

func procurementProjectSelectSQL(tail string) string {
	return `
		SELECT id, COALESCE(project_no,''), COALESCE(title,''), COALESCE(status,''), COALESCE(category,''), COALESCE(sub_category,''),
		       COALESCE(product_name,''), COALESCE(material,''), COALESCE(spec,''), COALESCE(target_quantity,0), COALESCE(demand_quantity,0),
		       COALESCE(unit,'吨'), COALESCE(target_price,0), COALESCE(delivery_address,''), COALESCE(buyer_request_count,0), COALESCE(buyer_enterprise_count,0),
		       COALESCE(invited_supplier_count,0), COALESCE(quoted_supplier_count,0), quote_deadline, awarded_at, ordered_at,
		       COALESCE(award_quote_id,0), COALESCE(award_supplier_id,0), COALESCE(award_supplier_name,''), COALESCE(award_product_id,0),
		       COALESCE(award_product_no,''), COALESCE(award_product_name,''), COALESCE(award_unit_price,0), COALESCE(award_delivery_days,0),
		       COALESCE(order_id,0), COALESCE(order_no,''), COALESCE(admin_remark,''), created_at, updated_at
		FROM procurement_projects ` + tail
}

func scanProcurementProject(scanner sqlScanner) (groupProcurementProjectItem, error) {
	var item groupProcurementProjectItem
	var quoteDeadline, awardedAt, orderedAt sql.NullTime
	err := scanner.Scan(
		&item.ID, &item.ProjectNo, &item.Title, &item.Status, &item.Category, &item.SubCategory,
		&item.ProductName, &item.Material, &item.Spec, &item.TargetQuantity, &item.DemandQuantity,
		&item.Unit, &item.TargetPrice, &item.DeliveryAddress, &item.BuyerRequestCount, &item.BuyerEnterpriseCount,
		&item.InvitedSupplierCount, &item.QuotedSupplierCount, &quoteDeadline, &awardedAt, &orderedAt,
		&item.AwardQuoteID, &item.AwardSupplierID, &item.AwardSupplierName, &item.AwardProductID,
		&item.AwardProductNo, &item.AwardProductName, &item.AwardUnitPrice, &item.AwardDeliveryDays,
		&item.OrderID, &item.OrderNo, &item.AdminRemark, &item.CreatedAt, &item.UpdatedAt,
	)
	if quoteDeadline.Valid {
		item.QuoteDeadline = &quoteDeadline.Time
	}
	if awardedAt.Valid {
		item.AwardedAt = &awardedAt.Time
	}
	if orderedAt.Valid {
		item.OrderedAt = &orderedAt.Time
	}
	return item, err
}

func listProjectRequests(ctx context.Context, db *sql.DB, projectID int64) ([]groupProcurementRequestItem, error) {
	rows, err := db.QueryContext(ctx, groupProcurementRequestSelectSQL("JOIN procurement_project_requests ppr ON ppr.request_id = procurement_requests.id AND ppr.deleted_at IS NULL WHERE procurement_requests.deleted_at IS NULL AND ppr.project_id = ? ORDER BY procurement_requests.id ASC"), projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]groupProcurementRequestItem, 0)
	for rows.Next() {
		var item groupProcurementRequestItem
		if err := scanGroupProcurementRequest(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func listProjectEnrollments(ctx context.Context, db *sql.DB, projectID int64) ([]groupProcurementProjectSellerItem, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, COALESCE(project_id,0), COALESCE(seller_enterprise_id,0), COALESCE(seller_user_id,0), COALESCE(seller_name,''), COALESCE(product_id,0),
		       COALESCE(product_no,''), COALESCE(product_name,''), COALESCE(status,''), COALESCE(remark,''), COALESCE(reviewed_by,0), reviewed_at, created_at, updated_at
		FROM procurement_supplier_enrollments
		WHERE deleted_at IS NULL AND project_id = ?
		ORDER BY id ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]groupProcurementProjectSellerItem, 0)
	for rows.Next() {
		var item groupProcurementProjectSellerItem
		var reviewedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.ProjectID, &item.SellerEnterpriseID, &item.SellerUserID, &item.SellerName, &item.ProductID, &item.ProductNo, &item.ProductName, &item.Status, &item.Remark, &item.ReviewedBy, &reviewedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		if reviewedAt.Valid {
			item.ReviewedAt = &reviewedAt.Time
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func listProjectQuotes(ctx context.Context, db *sql.DB, projectID int64) ([]groupProcurementProjectQuoteItem, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, COALESCE(project_id,0), COALESCE(enrollment_id,0), COALESCE(seller_enterprise_id,0), COALESCE(seller_user_id,0), COALESCE(seller_name,''),
		       COALESCE(product_id,0), COALESCE(product_no,''), COALESCE(product_name,''), COALESCE(unit_price,0), COALESCE(supply_quantity,0),
		       COALESCE(unit,'吨'), COALESCE(delivery_days,0), COALESCE(delivery_warehouse,''), COALESCE(tax_included,1), COALESCE(tax_rate,0),
		       COALESCE(payment_terms,''), COALESCE(remark,''), COALESCE(is_latest,1), COALESCE(status,''), created_at, updated_at
		FROM procurement_supplier_quotes
		WHERE deleted_at IS NULL AND project_id = ? AND is_latest = 1
		ORDER BY unit_price ASC, id ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]groupProcurementProjectQuoteItem, 0)
	for rows.Next() {
		var item groupProcurementProjectQuoteItem
		if err := rows.Scan(&item.ID, &item.ProjectID, &item.EnrollmentID, &item.SellerEnterpriseID, &item.SellerUserID, &item.SellerName, &item.ProductID, &item.ProductNo, &item.ProductName, &item.UnitPrice, &item.SupplyQuantity, &item.Unit, &item.DeliveryDays, &item.DeliveryWarehouse, &item.TaxIncluded, &item.TaxRate, &item.PaymentTerms, &item.Remark, &item.IsLatest, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func getGroupProcurementProjectQuoteByID(ctx context.Context, db *sql.DB, id int64) (groupProcurementProjectQuoteItem, error) {
	row := db.QueryRowContext(ctx, `
		SELECT id, COALESCE(project_id,0), COALESCE(enrollment_id,0), COALESCE(seller_enterprise_id,0), COALESCE(seller_user_id,0), COALESCE(seller_name,''),
		       COALESCE(product_id,0), COALESCE(product_no,''), COALESCE(product_name,''), COALESCE(unit_price,0), COALESCE(supply_quantity,0),
		       COALESCE(unit,'吨'), COALESCE(delivery_days,0), COALESCE(delivery_warehouse,''), COALESCE(tax_included,1), COALESCE(tax_rate,0),
		       COALESCE(payment_terms,''), COALESCE(remark,''), COALESCE(is_latest,1), COALESCE(status,''), created_at, updated_at
		FROM procurement_supplier_quotes
		WHERE deleted_at IS NULL AND id = ? LIMIT 1`, id)
	var item groupProcurementProjectQuoteItem
	err := row.Scan(&item.ID, &item.ProjectID, &item.EnrollmentID, &item.SellerEnterpriseID, &item.SellerUserID, &item.SellerName, &item.ProductID, &item.ProductNo, &item.ProductName, &item.UnitPrice, &item.SupplyQuantity, &item.Unit, &item.DeliveryDays, &item.DeliveryWarehouse, &item.TaxIncluded, &item.TaxRate, &item.PaymentTerms, &item.Remark, &item.IsLatest, &item.Status, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func createAwardProjectOrders(ctx context.Context, db *sql.DB, project groupProcurementProjectItem, product groupProductSnapshot, requests []groupProcurementRequestItem, adminID int64, remark string) (int64, string, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, "", err
	}
	defer tx.Rollback()
	var existingOrderID int64
	var currentStatus string
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(order_id,0), COALESCE(status,'') FROM procurement_projects WHERE id = ? AND deleted_at IS NULL FOR UPDATE`, project.ID).Scan(&existingOrderID, &currentStatus); err != nil {
		return 0, "", err
	}
	if existingOrderID > 0 || currentStatus == "order_created" {
		return 0, "", fmt.Errorf("项目已生成订单")
	}
	totalQuantity := 0.0
	for _, item := range requests {
		totalQuantity += item.Quantity
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
	masterItems := buildGroupOrderItems(product, project.AwardUnitPrice, totalQuantity)
	masterRemark := map[string]any{"source": "group_procurement_project_master", "project_id": project.ID, "request_count": len(requests), "admin_remark": remark}
	masterDelivery := map[string]any{"address": "按各买家申请地址履约"}
	masterID, err := insertGroupOrder(ctx, tx, masterNo, 0, "集采竞价母订单", product.SellerEnterpriseID, product.SellerName, masterItems, project.AwardUnitPrice*totalQuantity, 1, masterDelivery, masterRemark, adminID, now, expiresAt)
	if err != nil {
		return 0, "", err
	}
	for _, req := range requests {
		childNo := generateAdminOrderNo()
		childItems := buildGroupOrderItems(product, project.AwardUnitPrice, req.Quantity)
		delivery := productDeliveryInfo(product.WarehouseInfo)
		delivery["address"] = req.DeliveryAddress
		if req.DestWarehouseServiceID > 0 {
			delivery["dest_warehouse_service_id"] = req.DestWarehouseServiceID
			delivery["dest_warehouse_enterprise_id"] = req.DestWarehouseEnterpriseID
			delivery["dest_warehouse_name"] = req.DestWarehouseName
			delivery["dest_warehouse_province"] = req.DestWarehouseProvince
			delivery["dest_warehouse_city"] = req.DestWarehouseCity
			delivery["dest_warehouse_district"] = req.DestWarehouseDistrict
			delivery["dest_warehouse_address"] = req.DestWarehouseAddress
		}
		childStatus := 12
		if req.DestWarehouseServiceID > 0 {
			childStatus = 13
		}
		childRemark := map[string]any{
			"source":                "group_procurement_project_child",
			"project_id":            project.ID,
			"project_no":            project.ProjectNo,
			"master_order_id":       masterID,
			"master_order_no":       masterNo,
			"request_id":            req.ID,
			"request_no":            req.RequestNo,
			"contact_name":          req.ContactName,
			"contact_phone":         req.ContactPhone,
			"delivery_address":      req.DeliveryAddress,
			"delivery_method":       req.DeliveryMethod,
			"payment_method":        req.PaymentMethod,
			"delivery_note":         req.DeliveryNote,
			"target_price":          req.TargetPrice,
			"group_award_unitprice": project.AwardUnitPrice,
		}
		if _, err := insertGroupOrder(ctx, tx, childNo, req.BuyerEnterpriseID, req.BuyerName, product.SellerEnterpriseID, product.SellerName, childItems, project.AwardUnitPrice*req.Quantity, childStatus, delivery, childRemark, adminID, now, expiresAt); err != nil {
			return 0, "", err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE procurement_requests SET status = 'accepted', updated_at = ? WHERE id = ?`, now, req.ID); err != nil {
			return 0, "", err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE procurement_projects SET status = 'order_created', order_id = ?, order_no = ?, ordered_at = ?, updated_at = ? WHERE id = ?`, masterID, masterNo, now, now, project.ID); err != nil {
		return 0, "", err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO procurement_logs (project_id, content, color, created_at) VALUES (?, ?, 'green', ?)`, project.ID, "管理端基于中标报价生成母子订单", now); err != nil {
		return 0, "", err
	}
	if err := tx.Commit(); err != nil {
		return 0, "", err
	}
	return masterID, masterNo, nil
}

func ensureGroupProcurementBiddingSchema(ctx context.Context, db *sql.DB) error {
	if err := ensureGroupProcurementSchema(ctx, db); err != nil {
		return err
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS procurement_projects (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			project_no VARCHAR(40) NOT NULL UNIQUE,
			title VARCHAR(200) NOT NULL,
			status VARCHAR(32) NOT NULL DEFAULT 'collecting',
			category VARCHAR(80) NOT NULL DEFAULT '',
			sub_category VARCHAR(80) NOT NULL DEFAULT '',
			product_name VARCHAR(200) NOT NULL DEFAULT '',
			material VARCHAR(120) NOT NULL DEFAULT '',
			spec VARCHAR(200) NOT NULL DEFAULT '',
			target_quantity DECIMAL(14,2) NOT NULL DEFAULT 0,
			demand_quantity DECIMAL(14,2) NOT NULL DEFAULT 0,
			unit VARCHAR(20) NOT NULL DEFAULT '吨',
			target_price DECIMAL(16,2) NOT NULL DEFAULT 0,
			delivery_address VARCHAR(500) NOT NULL DEFAULT '',
			buyer_request_count INT NOT NULL DEFAULT 0,
			buyer_enterprise_count INT NOT NULL DEFAULT 0,
			invited_supplier_count INT NOT NULL DEFAULT 0,
			quoted_supplier_count INT NOT NULL DEFAULT 0,
			quote_deadline DATETIME NULL,
			awarded_at DATETIME NULL,
			ordered_at DATETIME NULL,
			award_quote_id BIGINT NOT NULL DEFAULT 0,
			award_supplier_id BIGINT NOT NULL DEFAULT 0,
			award_supplier_name VARCHAR(200) NOT NULL DEFAULT '',
			award_product_id BIGINT NOT NULL DEFAULT 0,
			award_product_no VARCHAR(64) NOT NULL DEFAULT '',
			award_product_name VARCHAR(200) NOT NULL DEFAULT '',
			award_unit_price DECIMAL(16,2) NOT NULL DEFAULT 0,
			award_delivery_days INT NOT NULL DEFAULT 0,
			order_id BIGINT NOT NULL DEFAULT 0,
			order_no VARCHAR(64) NOT NULL DEFAULT '',
			admin_remark VARCHAR(500) NOT NULL DEFAULT '',
			created_by BIGINT NOT NULL DEFAULT 0,
			updated_by BIGINT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			deleted_at DATETIME NULL,
			INDEX idx_procurement_projects_status (status),
			INDEX idx_procurement_projects_category (category),
			INDEX idx_procurement_projects_updated_at (updated_at)
		)`,
		`CREATE TABLE IF NOT EXISTS procurement_project_requests (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			project_id BIGINT NOT NULL,
			request_id BIGINT NOT NULL,
			created_at DATETIME NOT NULL,
			deleted_at DATETIME NULL,
			UNIQUE KEY idx_procurement_project_request (project_id, request_id),
			INDEX idx_procurement_project_requests_request_id (request_id)
		)`,
		`CREATE TABLE IF NOT EXISTS procurement_supplier_enrollments (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			project_id BIGINT NOT NULL,
			seller_enterprise_id BIGINT NOT NULL,
			seller_user_id BIGINT NOT NULL DEFAULT 0,
			seller_name VARCHAR(200) NOT NULL DEFAULT '',
			product_id BIGINT NOT NULL DEFAULT 0,
			product_no VARCHAR(64) NOT NULL DEFAULT '',
			product_name VARCHAR(200) NOT NULL DEFAULT '',
			status VARCHAR(32) NOT NULL DEFAULT 'invited',
			remark VARCHAR(500) NOT NULL DEFAULT '',
			reviewed_by BIGINT NOT NULL DEFAULT 0,
			reviewed_at DATETIME NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			deleted_at DATETIME NULL,
			INDEX idx_procurement_enrollments_project_id (project_id),
			INDEX idx_procurement_enrollments_seller_id (seller_enterprise_id)
		)`,
		`CREATE TABLE IF NOT EXISTS procurement_supplier_quotes (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			project_id BIGINT NOT NULL,
			enrollment_id BIGINT NOT NULL,
			seller_enterprise_id BIGINT NOT NULL,
			seller_user_id BIGINT NOT NULL DEFAULT 0,
			seller_name VARCHAR(200) NOT NULL DEFAULT '',
			product_id BIGINT NOT NULL DEFAULT 0,
			product_no VARCHAR(64) NOT NULL DEFAULT '',
			product_name VARCHAR(200) NOT NULL DEFAULT '',
			unit_price DECIMAL(16,2) NOT NULL DEFAULT 0,
			supply_quantity DECIMAL(14,2) NOT NULL DEFAULT 0,
			unit VARCHAR(20) NOT NULL DEFAULT '吨',
			delivery_days INT NOT NULL DEFAULT 0,
			delivery_warehouse VARCHAR(200) NOT NULL DEFAULT '',
			tax_included TINYINT(1) NOT NULL DEFAULT 1,
			tax_rate DECIMAL(8,2) NOT NULL DEFAULT 0,
			payment_terms VARCHAR(200) NOT NULL DEFAULT '',
			remark VARCHAR(500) NOT NULL DEFAULT '',
			is_latest TINYINT(1) NOT NULL DEFAULT 1,
			status VARCHAR(32) NOT NULL DEFAULT 'submitted',
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			deleted_at DATETIME NULL,
			INDEX idx_procurement_quotes_project_id (project_id),
			INDEX idx_procurement_quotes_seller_id (seller_enterprise_id)
		)`,
		`CREATE TABLE IF NOT EXISTS procurement_project_awards (
			id BIGINT PRIMARY KEY AUTO_INCREMENT,
			project_id BIGINT NOT NULL UNIQUE,
			quote_id BIGINT NOT NULL,
			seller_enterprise_id BIGINT NOT NULL,
			seller_name VARCHAR(200) NOT NULL DEFAULT '',
			product_id BIGINT NOT NULL DEFAULT 0,
			product_no VARCHAR(64) NOT NULL DEFAULT '',
			product_name VARCHAR(200) NOT NULL DEFAULT '',
			unit_price DECIMAL(16,2) NOT NULL DEFAULT 0,
			award_quantity DECIMAL(14,2) NOT NULL DEFAULT 0,
			delivery_days INT NOT NULL DEFAULT 0,
			remark VARCHAR(500) NOT NULL DEFAULT '',
			created_by BIGINT NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			deleted_at DATETIME NULL,
			INDEX idx_procurement_project_awards_quote_id (quote_id)
		)`,
	}
	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	_, _ = db.ExecContext(ctx, `ALTER TABLE procurement_logs ADD COLUMN project_id BIGINT NOT NULL DEFAULT 0`)
	return nil
}

func defaultAdminUnit(unit string) string {
	unit = strings.TrimSpace(unit)
	if unit == "" {
		return "吨"
	}
	return unit
}
