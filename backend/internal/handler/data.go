package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type pageResult[T any] struct {
	List  []T `json:"list"`
	Total int `json:"total"`
}

type dashboardStats struct {
	TotalUsers       int `json:"totalUsers"`
	TotalEnterprises int `json:"totalEnterprises"`
	TotalProducts    int `json:"totalProducts"`
	PendingAudits    int `json:"pendingAudits"`
}

type adminItem struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	RealName  string    `json:"realName"`
	Email     string    `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Status    int       `json:"status"`
	RoleID    *int64    `json:"roleId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type roleItem struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description,omitempty"`
	Status      int       `json:"status"`
	Sort        int       `json:"sort"`
	CreatedAt   time.Time `json:"createdAt"`
}

type menuItem struct {
	ID         int64      `json:"id"`
	ParentID   int64      `json:"parentId"`
	Name       string     `json:"name"`
	Path       string     `json:"path"`
	Component  string     `json:"component,omitempty"`
	Icon       string     `json:"icon,omitempty"`
	Type       int        `json:"type"`
	Permission string     `json:"permission,omitempty"`
	Status     int        `json:"status"`
	Sort       int        `json:"sort"`
	Children   []menuItem `json:"children,omitempty"`
}

type userItem struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Phone        string    `json:"phone,omitempty"`
	Email        string    `json:"email,omitempty"`
	RealName     string    `json:"realName,omitempty"`
	EnterpriseID *int64    `json:"enterpriseId,omitempty"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
}

type enterpriseItem struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	ShortName       string     `json:"shortName,omitempty"`
	CreditCode      string     `json:"creditCode,omitempty"`
	LegalPerson     string     `json:"legalPerson,omitempty"`
	ContactPerson   string     `json:"contactPerson,omitempty"`
	ContactPhone    string     `json:"contactPhone,omitempty"`
	Address         string     `json:"address,omitempty"`
	BusinessLicense string     `json:"businessLicense,omitempty"`
	Status          int        `json:"status"`
	AuditRemark     string     `json:"auditRemark,omitempty"`
	AuditedAt       *time.Time `json:"auditedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type productItem struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	CategoryID   *int64     `json:"categoryId,omitempty"`
	EnterpriseID int64      `json:"enterpriseId"`
	Category     string     `json:"category,omitempty"`
	SubCategory  string     `json:"subCategory,omitempty"`
	SteelMill    string     `json:"steelMill,omitempty"`
	Spec         string     `json:"spec,omitempty"`
	Material     string     `json:"material,omitempty"`
	WarehouseName string    `json:"warehouseName,omitempty"`
	WarehouseCode string    `json:"warehouseCode,omitempty"`
	Province     string     `json:"province,omitempty"`
	City         string     `json:"city,omitempty"`
	District     string     `json:"district,omitempty"`
	Address      string     `json:"address,omitempty"`
	BundleNo     string     `json:"bundleNo,omitempty"`
	Unit         string     `json:"unit,omitempty"`
	Price        float64    `json:"price"`
	Stock        float64    `json:"stock"`
	MinOrderQty  float64    `json:"minOrderQty"`
	DeliveryDays int        `json:"deliveryDays"`
	QuoteValidity int       `json:"quoteValidity"`
	Description  string     `json:"description,omitempty"`
	Images       []string   `json:"images,omitempty"`
	Status       int        `json:"status"`
	AuditRemark  string     `json:"auditRemark,omitempty"`
	AuditedAt    *time.Time `json:"auditedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

type orderItem struct {
	ID                 int64      `json:"id"`
	OrderNo            string     `json:"orderNo"`
	BuyerID            int64      `json:"buyerId"`
	BuyerName          string     `json:"buyerName"`
	SellerID           int64      `json:"sellerId"`
	SellerEnterpriseID int64      `json:"sellerEnterpriseId"`
	SellerName         string     `json:"sellerName"`
	TotalQuantity      float64    `json:"totalQuantity"`
	TotalAmount        float64    `json:"totalAmount"`
	TaxTotalAmount     float64    `json:"taxTotalAmount"`
	Currency           string     `json:"currency"`
	Status             int        `json:"status"`
	DeliveryVoucher    string     `json:"deliveryVoucher,omitempty"`
	ReceiptVoucher     string     `json:"receiptVoucher,omitempty"`
	ReceiptAt          *time.Time `json:"receiptAt,omitempty"`
	ExpiresAt          *time.Time `json:"expiresAt,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
}

type categoryItem struct {
	ID       int64 `json:"id"`
	ParentID int64 `json:"parentId"`
	Name     string `json:"name"`
	Code     string `json:"code,omitempty"`
	Icon     string `json:"icon,omitempty"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status"`
}

type enterpriseVerificationInfo struct {
	LegalPerson    string `json:"legal_person"`
	BusinessLicense string `json:"business_license"`
	ContactName    string `json:"contact_name"`
	ContactPhone   string `json:"contact_phone"`
	Address        string `json:"address"`
}

type productSpecInfo struct {
	Material string `json:"material"`
	Spec     string `json:"spec"`
}

type warehouseInfo struct {
	WarehouseName string `json:"warehouse_name"`
	WarehouseCode string `json:"warehouse_code"`
	Province      string `json:"province"`
	City          string `json:"city"`
	District      string `json:"district"`
	Address       string `json:"address"`
}

func parsePageParams(pageStr, pageSizeStr string) (limit, offset int) {
	page := 1
	pageSize := 10
	if pageStr != "" {
		fmt.Sscanf(pageStr, "%d", &page)
	}
	if pageSizeStr != "" {
		fmt.Sscanf(pageSizeStr, "%d", &pageSize)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return pageSize, (page - 1) * pageSize
}

func queryCount(ctx context.Context, db *sql.DB, query string, args ...any) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func tableExists(ctx context.Context, db *sql.DB, table string) bool {
	var exists int
	err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?`,
		table,
	).Scan(&exists)
	return err == nil && exists > 0
}

func columnExists(ctx context.Context, db *sql.DB, table, column string) bool {
	var exists int
	err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?`,
		table, column,
	).Scan(&exists)
	return err == nil && exists > 0
}

func mapProductStatus(raw string) int {
	switch strings.ToLower(raw) {
	case "2", "approved":
		return 1
	case "4", "rejected":
		return 2
	case "3", "offline":
		return 2
	case "1", "0", "pending_review", "draft":
		return 0
	default:
		return 0
	}
}

func mapOrderPending(raw string) bool {
	switch strings.ToLower(raw) {
	case "pending_contract", "pending_payment", "pending_payment_audit", "pending_pickup":
		return true
	default:
		return false
	}
}

func mapOrderStatus(raw string) int {
	switch strings.ToLower(raw) {
	case "pending":
		return 1
	case "pending_contract":
		return 2
	case "pending_payment":
		return 3
	case "pending_payment_audit":
		return 4
	case "pending_pickup":
		return 5
	case "shipped":
		return 6
	case "cancelled":
		return 7
	case "completed":
		return 8
	default:
		return 0
	}
}

func parseLegacyAuditRemark(raw string) string {
	if raw == "" {
		return ""
	}
	payload := map[string]string{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return raw
	}
	if value := payload["audit_reason"]; value != "" {
		return value
	}
	if value := payload["remark"]; value != "" {
		return value
	}
	return ""
}

func parseVerificationInfo(raw sql.NullString) enterpriseVerificationInfo {
	if !raw.Valid || raw.String == "" {
		return enterpriseVerificationInfo{}
	}
	var info enterpriseVerificationInfo
	_ = json.Unmarshal([]byte(raw.String), &info)
	return info
}

func parseImages(raw sql.NullString) []string {
	if !raw.Valid || raw.String == "" {
		return nil
	}
	var images []string
	_ = json.Unmarshal([]byte(raw.String), &images)
	return images
}

func parseProductSpec(raw sql.NullString) productSpecInfo {
	if !raw.Valid || raw.String == "" {
		return productSpecInfo{}
	}
	var spec productSpecInfo
	_ = json.Unmarshal([]byte(raw.String), &spec)
	return spec
}

func parseWarehouseInfo(raw sql.NullString) warehouseInfo {
	if !raw.Valid || raw.String == "" {
		return warehouseInfo{}
	}
	var info warehouseInfo
	_ = json.Unmarshal([]byte(raw.String), &info)
	return info
}
