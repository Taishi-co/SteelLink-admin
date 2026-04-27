package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
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

type adminUpsertRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password,omitempty"`
	RealName string `json:"realName"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Status   int    `json:"status"`
	RoleID   *int64 `json:"roleId"`
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
	ID             int64     `json:"id"`
	Username       string    `json:"username"`
	Phone          string    `json:"phone,omitempty"`
	Email          string    `json:"email,omitempty"`
	RealName       string    `json:"realName,omitempty"`
	EnterpriseID   *int64    `json:"enterpriseId,omitempty"`
	IdentityStatus int       `json:"identityStatus"`
	Status         int       `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
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
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	CategoryID    *int64     `json:"categoryId,omitempty"`
	EnterpriseID  int64      `json:"enterpriseId"`
	Category      string     `json:"category,omitempty"`
	SubCategory   string     `json:"subCategory,omitempty"`
	SteelMill     string     `json:"steelMill,omitempty"`
	Spec          string     `json:"spec,omitempty"`
	Material      string     `json:"material,omitempty"`
	WarehouseName string     `json:"warehouseName,omitempty"`
	WarehouseCode string     `json:"warehouseCode,omitempty"`
	Province      string     `json:"province,omitempty"`
	City          string     `json:"city,omitempty"`
	District      string     `json:"district,omitempty"`
	Address       string     `json:"address,omitempty"`
	BundleNo      string     `json:"bundleNo,omitempty"`
	Unit          string     `json:"unit,omitempty"`
	Price         float64    `json:"price"`
	Stock         float64    `json:"stock"`
	MinOrderQty   float64    `json:"minOrderQty"`
	DeliveryDays  int        `json:"deliveryDays"`
	QuoteValidity int        `json:"quoteValidity"`
	Description   string     `json:"description,omitempty"`
	Images        []string   `json:"images,omitempty"`
	Status        int        `json:"status"`
	AuditRemark   string     `json:"auditRemark,omitempty"`
	AuditedAt     *time.Time `json:"auditedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

type orderItem struct {
	ID                 int64               `json:"id"`
	OrderNo            string              `json:"orderNo"`
	BuyerID            int64               `json:"buyerId"`
	BuyerName          string              `json:"buyerName"`
	SellerID           int64               `json:"sellerId"`
	SellerEnterpriseID int64               `json:"sellerEnterpriseId"`
	SellerName         string              `json:"sellerName"`
	TotalQuantity      float64             `json:"totalQuantity"`
	TotalAmount        float64             `json:"totalAmount"`
	TaxTotalAmount     float64             `json:"taxTotalAmount"`
	Currency           string              `json:"currency"`
	Status             int                 `json:"status"`
	DeliveryVoucher    string              `json:"deliveryVoucher,omitempty"`
	ReceiptVoucher     string              `json:"receiptVoucher,omitempty"`
	ReceiptAt          *time.Time          `json:"receiptAt,omitempty"`
	ExpiresAt          *time.Time          `json:"expiresAt,omitempty"`
	CreatedAt          time.Time           `json:"createdAt"`
	Timeline           []orderTimelineItem `json:"timeline"`
}

type orderTimelineItem struct {
	Key         string     `json:"key"`
	Label       string     `json:"label"`
	Status      string     `json:"status"`
	Time        *time.Time `json:"time,omitempty"`
	Description string     `json:"description,omitempty"`
}

type categoryItem struct {
	ID       int64  `json:"id"`
	ParentID int64  `json:"parentId"`
	Name     string `json:"name"`
	Code     string `json:"code,omitempty"`
	Icon     string `json:"icon,omitempty"`
	Sort     int    `json:"sort"`
	Status   int    `json:"status"`
}

type roleUpsertRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Status      int    `json:"status"`
	Sort        int    `json:"sort"`
}

type menuUpsertRequest struct {
	ParentID   int64  `json:"parentId"`
	Name       string `json:"name" binding:"required"`
	Path       string `json:"path"`
	Component  string `json:"component"`
	Icon       string `json:"icon"`
	Type       int    `json:"type"`
	Permission string `json:"permission"`
	Status     int    `json:"status"`
	Sort       int    `json:"sort"`
}

type roleMenusRequest struct {
	MenuIDs []int64 `json:"menuIds"`
}

type enterpriseVerificationInfo struct {
	LegalPerson     string `json:"legal_person"`
	BusinessLicense string `json:"business_license"`
	ContactName     string `json:"contact_name"`
	ContactPhone    string `json:"contact_phone"`
	Address         string `json:"address"`
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

type operationLogItem struct {
	ID           int64     `json:"id"`
	AdminID      int64     `json:"adminId"`
	AdminName    string    `json:"adminName"`
	Module       string    `json:"module"`
	Action       string    `json:"action"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	IP           string    `json:"ip"`
	RequestBody  string    `json:"requestBody,omitempty"`
	ResponseCode int       `json:"responseCode"`
	CreatedAt    time.Time `json:"createdAt"`
}

func logOperation(ctx context.Context, db *sql.DB, adminID int64, adminName, module, action, method, path, ip, requestBody string, responseCode int) {
	if !tableExists(ctx, db, "admin_operation_logs") {
		return
	}
	db.ExecContext(ctx, `
		INSERT INTO admin_operation_logs (admin_id, admin_name, module, action, method, path, ip, request_body, response_code, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		adminID, adminName, module, action, method, path, ip, requestBody, responseCode, time.Now())
}

func logOpFromContext(c *gin.Context, db *sql.DB, module, action string, responseCode int, reqBody ...string) {
	adminID, _ := c.Get("admin_id")
	adminName, _ := c.Get("admin_real_name")
	var aid int64
	switch v := adminID.(type) {
	case int64:
		aid = v
	case int:
		aid = int64(v)
	case float64:
		aid = int64(v)
	}
	name, _ := adminName.(string)
	ip := c.ClientIP()
	body := ""
	if len(reqBody) > 0 {
		body = reqBody[0]
		if len(body) > 2000 {
			body = body[:2000]
		}
	}
	logOperation(c.Request.Context(), db, aid, name, module, action, c.Request.Method, c.Request.URL.Path, ip, body, responseCode)
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

func mapProductIntStatus(raw int) int {
	switch raw {
	case 2:
		return 1
	case 3:
		return 1
	case 4:
		return 2
	default:
		return 0
	}
}

func mapProductStatus(raw string) int {
	if v, err := strconv.Atoi(raw); err == nil {
		return mapProductIntStatus(v)
	}
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
	case "delivered":
		return 7
	case "completed":
		return 8
	case "cancelled":
		return 9
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

var localAdminStore = struct {
	mu     sync.Mutex
	nextID int64
	items  []adminItem
}{
	nextID: 2,
}

func ensureLocalAdmins() {
	if len(localAdminStore.items) == 0 {
		localAdminStore.items = []adminItem{{
			ID:        1,
			Username:  defaultAdminUsername(),
			RealName:  defaultAdminRealName(),
			Status:    1,
			RoleID:    func() *int64 { value := int64(1); return &value }(),
			CreatedAt: time.Now(),
		}}
		return
	}

	localAdminStore.items[0].Username = defaultAdminUsername()
	localAdminStore.items[0].RealName = defaultAdminRealName()
}

func listLocalAdmins() []adminItem {
	localAdminStore.mu.Lock()
	defer localAdminStore.mu.Unlock()

	ensureLocalAdmins()
	items := make([]adminItem, len(localAdminStore.items))
	copy(items, localAdminStore.items)
	return items
}

func getLocalAdminByID(id int64) (adminItem, bool) {
	localAdminStore.mu.Lock()
	defer localAdminStore.mu.Unlock()

	ensureLocalAdmins()
	for _, item := range localAdminStore.items {
		if item.ID == id {
			return item, true
		}
	}
	return adminItem{}, false
}

func createLocalAdmin(req adminUpsertRequest) adminItem {
	localAdminStore.mu.Lock()
	defer localAdminStore.mu.Unlock()

	ensureLocalAdmins()
	item := adminItem{
		ID:        localAdminStore.nextID,
		Username:  req.Username,
		RealName:  req.RealName,
		Email:     req.Email,
		Phone:     req.Phone,
		Status:    req.Status,
		RoleID:    req.RoleID,
		CreatedAt: time.Now(),
	}
	localAdminStore.nextID += 1
	localAdminStore.items = append(localAdminStore.items, item)
	return item
}

func updateLocalAdmin(id int64, req adminUpsertRequest) (adminItem, bool) {
	localAdminStore.mu.Lock()
	defer localAdminStore.mu.Unlock()

	ensureLocalAdmins()
	for index, item := range localAdminStore.items {
		if item.ID != id {
			continue
		}

		if item.ID == 1 {
			localAdminStore.items[index].RealName = req.RealName
			localAdminStore.items[index].Email = req.Email
			localAdminStore.items[index].Phone = req.Phone
			localAdminStore.items[index].Status = req.Status
			localAdminStore.items[index].RoleID = req.RoleID
		} else {
			localAdminStore.items[index].Username = req.Username
			localAdminStore.items[index].RealName = req.RealName
			localAdminStore.items[index].Email = req.Email
			localAdminStore.items[index].Phone = req.Phone
			localAdminStore.items[index].Status = req.Status
			localAdminStore.items[index].RoleID = req.RoleID
		}
		return localAdminStore.items[index], true
	}

	return adminItem{}, false
}

func deleteLocalAdmin(id int64) bool {
	localAdminStore.mu.Lock()
	defer localAdminStore.mu.Unlock()

	ensureLocalAdmins()
	if id == 1 {
		return false
	}

	for index, item := range localAdminStore.items {
		if item.ID != id {
			continue
		}
		localAdminStore.items = append(localAdminStore.items[:index], localAdminStore.items[index+1:]...)
		return true
	}

	return false
}

func AutoMigrate(db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if !tableExists(ctx, db, "admin_roles") {
		db.ExecContext(ctx, `CREATE TABLE admin_roles (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(64) NOT NULL,
			code VARCHAR(64) NOT NULL,
			description VARCHAR(255) DEFAULT '',
			status TINYINT DEFAULT 1,
			sort INT DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME DEFAULT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)

		db.ExecContext(ctx, `INSERT INTO admin_roles (name, code, description, status, sort) VALUES
			('超级管理员', 'super_admin', '拥有所有权限', 1, 1),
			('运营管理员', 'operator', '负责日常运营管理', 1, 2),
			('审核员', 'auditor', '负责企业和商品审核', 1, 3)`)
	}

	if !tableExists(ctx, db, "admin_menus") {
		db.ExecContext(ctx, `CREATE TABLE admin_menus (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			parent_id BIGINT DEFAULT 0,
			name VARCHAR(64) NOT NULL,
			path VARCHAR(255) DEFAULT '',
			component VARCHAR(255) DEFAULT '',
			icon VARCHAR(64) DEFAULT '',
			type TINYINT DEFAULT 1,
			permission VARCHAR(128) DEFAULT '',
			status TINYINT DEFAULT 1,
			sort INT DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME DEFAULT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)

		db.ExecContext(ctx, `INSERT INTO admin_menus (parent_id, name, path, icon, type, permission, status, sort) VALUES
			(0, '数据概览', '/dashboard', 'DashboardOutlined', 1, '', 1, 1),
			(0, '系统管理', '', 'SettingOutlined', 1, '', 1, 2),
			(2, '管理员管理', '/system/admin', 'TeamOutlined', 1, 'system:admin', 1, 1),
			(2, '角色权限', '/system/role', 'SafetyOutlined', 1, 'system:role', 1, 2),
			(2, '菜单配置', '/system/menu', 'MenuOutlined', 1, 'system:menu', 1, 3),
			(0, '业务管理', '', 'SolutionOutlined', 1, '', 1, 3),
			(6, '用户管理', '/user', 'UserOutlined', 1, 'business:user', 1, 1),
			(6, '企业审核', '/enterprise', 'BankOutlined', 1, 'business:enterprise', 1, 2),
			(0, '商品中心', '', 'ShoppingOutlined', 1, '', 1, 4),
			(9, '商品管理', '/product', 'ShopOutlined', 1, 'product:manage', 1, 1),
			(9, '类目配置', '/product/category', 'AppstoreOutlined', 1, 'product:category', 1, 2)`)
	}

	if !tableExists(ctx, db, "admin_role_menus") {
		db.ExecContext(ctx, `CREATE TABLE admin_role_menus (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			role_id BIGINT NOT NULL,
			menu_id BIGINT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_role_menu (role_id, menu_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)

		db.ExecContext(ctx, `INSERT INTO admin_role_menus (role_id, menu_id)
			SELECT 1, id FROM admin_menus WHERE deleted_at IS NULL`)
	}
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

func scanMenuRows(rows *sql.Rows) []menuItem {
	items := make([]menuItem, 0)
	for rows.Next() {
		var m menuItem
		if err := rows.Scan(&m.ID, &m.ParentID, &m.Name, &m.Path, &m.Component, &m.Icon, &m.Type, &m.Permission, &m.Status, &m.Sort); err != nil {
			continue
		}
		items = append(items, m)
	}
	return items
}

func buildMenuTree(items []menuItem) []menuItem {
	nodeMap := make(map[int64]*menuItem)
	roots := make([]menuItem, 0)
	for i := range items {
		nodeMap[items[i].ID] = &items[i]
	}
	for i := range items {
		if items[i].ParentID == 0 {
			roots = append(roots, items[i])
		} else if parent, ok := nodeMap[items[i].ParentID]; ok {
			parent.Children = append(parent.Children, items[i])
		}
	}
	return roots
}
