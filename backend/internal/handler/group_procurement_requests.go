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

type rejectGroupProcurementRequest struct {
	Remark string `json:"remark"`
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
