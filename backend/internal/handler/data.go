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
	IDCard         string    `json:"idCard,omitempty"`
	IDCardFrontURL string    `json:"idCardFrontUrl,omitempty"`
	IDCardBackURL  string    `json:"idCardBackUrl,omitempty"`
	IdentityReject string    `json:"identityReject,omitempty"`
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
	ID                    int64                 `json:"id"`
	Name                  string                `json:"name"`
	CategoryID            *int64                `json:"categoryId,omitempty"`
	EnterpriseID          int64                 `json:"enterpriseId"`
	Category              string                `json:"category,omitempty"`
	SubCategory           string                `json:"subCategory,omitempty"`
	SteelMill             string                `json:"steelMill,omitempty"`
	GoodsType             string                `json:"goodsType,omitempty"`
	WarehouseServiceID    int64                 `json:"warehouseServiceId,omitempty"`
	WarehouseEnterpriseID int64                 `json:"warehouseEnterpriseId,omitempty"`
	Spec                  string                `json:"spec,omitempty"`
	Material              string                `json:"material,omitempty"`
	WarehouseName         string                `json:"warehouseName,omitempty"`
	WarehouseCode         string                `json:"warehouseCode,omitempty"`
	Province              string                `json:"province,omitempty"`
	City                  string                `json:"city,omitempty"`
	District              string                `json:"district,omitempty"`
	Address               string                `json:"address,omitempty"`
	BundleNo              string                `json:"bundleNo,omitempty"`
	ProductNo             string                `json:"productNo,omitempty"`
	Unit                  string                `json:"unit,omitempty"`
	Price                 float64               `json:"price"`
	Stock                 float64               `json:"stock"`
	MinOrderQty           float64               `json:"minOrderQty"`
	DeliveryDays          int                   `json:"deliveryDays"`
	QuoteValidity         int                   `json:"quoteValidity"`
	Description           string                `json:"description,omitempty"`
	Images                []string              `json:"images,omitempty"`
	Status                int                   `json:"status"`
	AuditRemark           string                `json:"auditRemark,omitempty"`
	AuditorID             int64                 `json:"auditorId,omitempty"`
	AuditedAt             *time.Time            `json:"auditedAt,omitempty"`
	AuditHistory          []productAuditLogItem `json:"auditHistory,omitempty"`
	CreatedAt             time.Time             `json:"createdAt"`
	UpdatedAt             time.Time             `json:"updatedAt"`
}

type productAuditSnapshot map[string]any

type productAuditLogItem struct {
	ID             int64                `json:"id"`
	ProductID      int64                `json:"productId"`
	ProductSkuID   string               `json:"productSkuId,omitempty"`
	Action         string               `json:"action,omitempty"`
	Status         string               `json:"status"`
	FromStatus     int                  `json:"fromStatus,omitempty"`
	ToStatus       int                  `json:"toStatus,omitempty"`
	OperatorID     int64                `json:"operatorId,omitempty"`
	OperatorName   string               `json:"operatorName,omitempty"`
	Remark         string               `json:"remark,omitempty"`
	BeforeSnapshot productAuditSnapshot `json:"beforeSnapshot,omitempty"`
	AfterSnapshot  productAuditSnapshot `json:"afterSnapshot,omitempty"`
	CreatedAt      time.Time            `json:"createdAt"`
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
	Remark             map[string]any      `json:"remark,omitempty"`
	TerminationReason  *terminationReason  `json:"terminationReason,omitempty"`
	Events             []orderEventItem    `json:"events,omitempty"`
	FinancingInfo      map[string]any      `json:"financingInfo,omitempty"`
	LogisticsBidding   map[string]any      `json:"logisticsBidding,omitempty"`
	LogisticsQuotes    []map[string]any    `json:"logisticsQuotes,omitempty"`
	ShipmentTask       map[string]any      `json:"shipmentTask,omitempty"`
	PledgeEvents       []map[string]any    `json:"pledgeEvents,omitempty"`
	BreachRecords      []breachRecordItem  `json:"breachRecords,omitempty"`
	LitigationRecords  []litigationItem    `json:"litigationRecords,omitempty"`
}

type orderTimelineItem struct {
	Key         string     `json:"key"`
	Label       string     `json:"label"`
	Status      string     `json:"status"`
	Time        *time.Time `json:"time,omitempty"`
	Description string     `json:"description,omitempty"`
}

type terminationReason struct {
	Code              string     `json:"code"`
	Label             string     `json:"label"`
	Party             string     `json:"party,omitempty"`
	PartyLabel        string     `json:"partyLabel,omitempty"`
	Reason            string     `json:"reason,omitempty"`
	Source            string     `json:"source,omitempty"`
	EventType         string     `json:"eventType,omitempty"`
	NeedsIntervention bool       `json:"needsIntervention"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
}

type orderEventItem struct {
	ID                int64          `json:"id"`
	OrderID           int64          `json:"orderId"`
	OrderNo           string         `json:"orderNo,omitempty"`
	EventType         string         `json:"eventType"`
	ActorUserID       int64          `json:"actorUserId"`
	ActorEnterpriseID int64          `json:"actorEnterpriseId"`
	ActorService      string         `json:"actorService,omitempty"`
	FromStatus        int            `json:"fromStatus"`
	ToStatus          int            `json:"toStatus"`
	Payload           map[string]any `json:"payload,omitempty"`
	CreatedAt         time.Time      `json:"createdAt"`
}

type breachRecordItem struct {
	ID                   int64      `json:"id"`
	OrderID              int64      `json:"orderId"`
	OrderNo              string     `json:"orderNo,omitempty"`
	ReporterEnterpriseID int64      `json:"reporterEnterpriseId"`
	BreachParty          string     `json:"breachParty"`
	BreachEnterpriseID   int64      `json:"breachEnterpriseId"`
	BreachType           string     `json:"breachType"`
	Description          string     `json:"description"`
	EvidenceURLs         []string   `json:"evidenceUrls,omitempty"`
	Status               string     `json:"status"`
	MediationResult      string     `json:"mediationResult,omitempty"`
	MediatedBy           int64      `json:"mediatedBy,omitempty"`
	MediatedAt           *time.Time `json:"mediatedAt,omitempty"`
	ResolutionResult     string     `json:"resolutionResult,omitempty"`
	ResolvedBy           int64      `json:"resolvedBy,omitempty"`
	ResolvedAt           *time.Time `json:"resolvedAt,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

type litigationItem struct {
	ID                    int64      `json:"id"`
	BreachRecordID        int64      `json:"breachRecordId"`
	OrderID               int64      `json:"orderId"`
	OrderNo               string     `json:"orderNo,omitempty"`
	PlaintiffEnterpriseID int64      `json:"plaintiffEnterpriseId"`
	DefendantEnterpriseID int64      `json:"defendantEnterpriseId"`
	CaseNo                string     `json:"caseNo,omitempty"`
	Court                 string     `json:"court,omitempty"`
	ClaimAmount           float64    `json:"claimAmount"`
	Description           string     `json:"description"`
	EvidenceURLs          []string   `json:"evidenceUrls,omitempty"`
	Status                string     `json:"status"`
	JudgmentResult        string     `json:"judgmentResult,omitempty"`
	JudgmentURL           string     `json:"judgmentUrl,omitempty"`
	JudgedAt              *time.Time `json:"judgedAt,omitempty"`
	ClosedAt              *time.Time `json:"closedAt,omitempty"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
}

type categoryItem struct {
	ID           int64  `json:"id"`
	ParentID     int64  `json:"parentId"`
	Name         string `json:"name"`
	DisplayName  string `json:"displayName,omitempty"`
	Code         string `json:"code,omitempty"`
	Icon         string `json:"icon,omitempty"`
	ImageURL     string `json:"imageUrl,omitempty"`
	Sort         int    `json:"sort"`
	Status       int    `json:"status"`
	ProductCount int    `json:"productCount"`
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
	WarehouseServiceID    int64  `json:"warehouse_service_id"`
	WarehouseEnterpriseID int64  `json:"warehouse_enterprise_id"`
	WarehouseName         string `json:"warehouse_name"`
	WarehouseCode         string `json:"warehouse_code"`
	Province              string `json:"province"`
	City                  string `json:"city"`
	District              string `json:"district"`
	Address               string `json:"address"`
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

func columnsExist(ctx context.Context, db *sql.DB, table string) map[string]bool {
	m := make(map[string]bool)
	rows, err := db.QueryContext(ctx,
		`SELECT column_name FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ?`,
		table,
	)
	if err != nil {
		return m
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			m[name] = true
		}
	}
	return m
}

func mapProductIntStatus(raw int) int {
	switch raw {
	case 0, 1, 2, 3, 4:
		return raw
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
		return 2
	case "4", "rejected":
		return 4
	case "3", "offline":
		return 3
	case "1", "pending_review":
		return 1
	case "0", "draft":
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

func isEnterpriseAuditDecision(status int) bool {
	return status == 1 || status == 2
}

func normalizeProductAuditDecision(status int) (statusValue int, logStatus string, ok bool) {
	switch status {
	case 1:
		return 2, "approved", true
	case 2:
		return 4, "rejected", true
	default:
		return 0, "", false
	}
}

func rejectReasonProvided(status, rejectStatus int, reason string) bool {
	return status != rejectStatus || strings.TrimSpace(reason) != ""
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

func parseStringList(raw sql.NullString) []string {
	if !raw.Valid || raw.String == "" {
		return nil
	}
	var values []string
	_ = json.Unmarshal([]byte(raw.String), &values)
	return values
}

func parseJSONMap(raw sql.NullString) map[string]any {
	if !raw.Valid || raw.String == "" {
		return nil
	}
	var value map[string]any
	_ = json.Unmarshal([]byte(raw.String), &value)
	return value
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
			(6, '服务权限审核', '/enterprise/services', 'AuditOutlined', 1, 'business:enterprise-service', 1, 3),
			(6, '仓储资源审核', '/service-offers/warehouse', 'ContainerOutlined', 1, 'service:warehouse:audit', 1, 4),
			(6, '物流资源审核', '/service-offers/logistics', 'TruckOutlined', 1, 'service:logistics:audit', 1, 5),
			(6, '金融服务审核', '/service-offers/finance', 'BankOutlined', 1, 'service:finance:audit', 1, 6),
			(0, '商品中心', '', 'ShoppingOutlined', 1, '', 1, 4),
			(13, '商品管理', '/product', 'ShopOutlined', 1, 'product:manage', 1, 1),
			(13, '订单管理', '/product/order', 'ShoppingOutlined', 1, 'order:manage', 1, 2),
			(13, '类目配置', '/product/category', 'AppstoreOutlined', 1, 'product:category', 1, 3),
			(0, 'AI采购管理', '', 'ThunderboltOutlined', 1, '', 1, 5),
			(17, '采购配置', '/ai-procurement/config', 'SettingOutlined', 1, 'ai-procurement:config', 1, 1),
			(0, '操作日志', '/log', 'FileTextOutlined', 1, 'log', 1, 6),
			(0, '站内信', '/notification', 'MailOutlined', 1, 'notification', 1, 7)`)
	}

	ensureProductCategoriesTable(ctx, db)
	ensureProductAuditLogsTable(ctx, db)
	_ = ensureNotificationsTable(ctx, db)

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
	ensureDefaultAdminMenus(ctx, db)
}

func ensureDefaultAdminMenus(ctx context.Context, db *sql.DB) {
	if db == nil || !tableExists(ctx, db, "admin_menus") {
		return
	}
	ensureAdminMenu(ctx, db, 0, "数据概览", "/dashboard", "DashboardOutlined", "", 1)
	systemID := ensureAdminMenu(ctx, db, 0, "系统管理", "", "SettingOutlined", "", 2)
	ensureAdminMenu(ctx, db, systemID, "管理员管理", "/system/admin", "TeamOutlined", "system:admin", 1)
	ensureAdminMenu(ctx, db, systemID, "角色权限", "/system/role", "SafetyOutlined", "system:role", 2)
	ensureAdminMenu(ctx, db, systemID, "菜单配置", "/system/menu", "MenuOutlined", "system:menu", 3)

	businessID := ensureAdminMenu(ctx, db, 0, "业务管理", "", "SolutionOutlined", "", 3)
	ensureAdminMenu(ctx, db, businessID, "用户管理", "/user", "UserOutlined", "business:user", 1)
	ensureAdminMenu(ctx, db, businessID, "企业审核", "/enterprise", "BankOutlined", "business:enterprise", 2)
	ensureAdminMenu(ctx, db, businessID, "服务权限审核", "/enterprise/services", "AuditOutlined", "business:enterprise-service", 3)
	ensureAdminMenu(ctx, db, businessID, "仓储资源审核", "/service-offers/warehouse", "ContainerOutlined", "service:warehouse:audit", 4)
	ensureAdminMenu(ctx, db, businessID, "物流资源审核", "/service-offers/logistics", "TruckOutlined", "service:logistics:audit", 5)
	ensureAdminMenu(ctx, db, businessID, "金融服务审核", "/service-offers/finance", "BankOutlined", "service:finance:audit", 6)

	productID := ensureAdminMenu(ctx, db, 0, "商品中心", "", "ShoppingOutlined", "", 4)
	ensureAdminMenu(ctx, db, productID, "商品管理", "/product", "ShopOutlined", "product:manage", 1)
	ensureAdminMenu(ctx, db, productID, "订单管理", "/product/order", "ShoppingOutlined", "order:manage", 2)
	ensureAdminMenu(ctx, db, productID, "类目配置", "/product/category", "AppstoreOutlined", "product:category", 3)

	aiID := ensureAdminMenu(ctx, db, 0, "AI采购管理", "", "ThunderboltOutlined", "", 5)
	ensureAdminMenu(ctx, db, aiID, "采购配置", "/ai-procurement/config", "SettingOutlined", "ai-procurement:config", 1)
	ensureAdminMenu(ctx, db, 0, "操作日志", "/log", "FileTextOutlined", "log", 6)
	ensureAdminMenu(ctx, db, 0, "站内信", "/notification", "MailOutlined", "notification", 7)

	if tableExists(ctx, db, "admin_role_menus") {
		_, _ = db.ExecContext(ctx, `INSERT IGNORE INTO admin_role_menus (role_id, menu_id)
			SELECT 1, id FROM admin_menus WHERE deleted_at IS NULL`)
		ensureRoleMenusByPaths(ctx, db, 2, []string{
			"/dashboard",
			"/user",
			"/enterprise",
			"/enterprise/services",
			"/service-offers/warehouse",
			"/service-offers/logistics",
			"/service-offers/finance",
			"/product",
			"/product/order",
			"/product/category",
			"/log",
			"/notification",
		})
		ensureRoleMenusByPaths(ctx, db, 3, []string{
			"/dashboard",
			"/user",
			"/enterprise",
			"/enterprise/services",
			"/service-offers/warehouse",
			"/service-offers/logistics",
			"/service-offers/finance",
			"/product",
			"/product/order",
			"/product/category",
			"/log",
		})
	}
}

func ensureRoleMenusByPaths(ctx context.Context, db *sql.DB, roleID int64, paths []string) {
	for _, path := range paths {
		_, _ = db.ExecContext(ctx, `INSERT IGNORE INTO admin_role_menus (role_id, menu_id)
			SELECT ?, id FROM admin_menus WHERE path = ? AND deleted_at IS NULL`,
			roleID, path)
	}
}

func ensureAdminMenu(ctx context.Context, db *sql.DB, parentID int64, name, path, icon, permission string, sort int) int64 {
	var id int64
	var err error
	if path != "" {
		err = db.QueryRowContext(ctx, `SELECT id FROM admin_menus WHERE path = ? AND deleted_at IS NULL ORDER BY id ASC LIMIT 1`, path).Scan(&id)
	} else {
		err = db.QueryRowContext(ctx, `SELECT id FROM admin_menus WHERE parent_id = ? AND name = ? AND deleted_at IS NULL ORDER BY id ASC LIMIT 1`, parentID, name).Scan(&id)
	}
	if err == nil {
		_, _ = db.ExecContext(ctx, `UPDATE admin_menus SET parent_id=?, name=?, path=?, icon=?, permission=?, type=1, status=1, sort=?, updated_at=NOW() WHERE id=?`,
			parentID, name, path, icon, permission, sort, id)
		return id
	}
	result, err := db.ExecContext(ctx, `INSERT INTO admin_menus (parent_id, name, path, icon, type, permission, status, sort) VALUES (?, ?, ?, ?, 1, ?, 1, ?)`,
		parentID, name, path, icon, permission, sort)
	if err != nil {
		return 0
	}
	id, _ = result.LastInsertId()
	return id
}

func ensureProductCategoriesTable(ctx context.Context, db *sql.DB) {
	if !tableExists(ctx, db, "product_categories") {
		db.ExecContext(ctx, `CREATE TABLE product_categories (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			parent_id BIGINT DEFAULT 0,
			name VARCHAR(80) NOT NULL,
			display_name VARCHAR(80) NOT NULL DEFAULT '',
			code VARCHAR(80) DEFAULT '',
			icon VARCHAR(64) DEFAULT '',
			image_url VARCHAR(512) DEFAULT '',
			sort INT DEFAULT 0,
			status TINYINT DEFAULT 1,
			created_by BIGINT DEFAULT 0,
			updated_by BIGINT DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME DEFAULT NULL,
			UNIQUE KEY uk_product_categories_name (name),
			KEY idx_product_categories_status (status),
			KEY idx_product_categories_sort (sort),
			KEY idx_product_categories_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
		return
	}
	if !columnExists(ctx, db, "product_categories", "display_name") {
		db.ExecContext(ctx, `ALTER TABLE product_categories ADD COLUMN display_name VARCHAR(80) NOT NULL DEFAULT ''`)
	}
	if !columnExists(ctx, db, "product_categories", "image_url") {
		db.ExecContext(ctx, `ALTER TABLE product_categories ADD COLUMN image_url VARCHAR(512) DEFAULT ''`)
	}
	if !columnExists(ctx, db, "product_categories", "created_by") {
		db.ExecContext(ctx, `ALTER TABLE product_categories ADD COLUMN created_by BIGINT DEFAULT 0`)
	}
	if !columnExists(ctx, db, "product_categories", "updated_by") {
		db.ExecContext(ctx, `ALTER TABLE product_categories ADD COLUMN updated_by BIGINT DEFAULT 0`)
	}
}

func ensureProductAuditLogsTable(ctx context.Context, db *sql.DB) {
	if !tableExists(ctx, db, "product_audit_logs") {
		db.ExecContext(ctx, `CREATE TABLE product_audit_logs (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			product_id BIGINT UNSIGNED NOT NULL,
			product_sku_id VARCHAR(64) DEFAULT '',
			action VARCHAR(64) NOT NULL DEFAULT '',
			status VARCHAR(32) NOT NULL DEFAULT '',
			from_status TINYINT DEFAULT NULL,
			to_status TINYINT DEFAULT NULL,
			operator_id BIGINT UNSIGNED DEFAULT 0,
			operator_name VARCHAR(100) DEFAULT '',
			remark VARCHAR(500) DEFAULT '',
			before_snapshot JSON NULL,
			after_snapshot JSON NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			KEY idx_product_audit_logs_product_id_created_at (product_id, created_at),
			KEY idx_product_audit_logs_status (status)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
		return
	}
	if !columnExists(ctx, db, "product_audit_logs", "from_status") {
		db.ExecContext(ctx, `ALTER TABLE product_audit_logs ADD COLUMN from_status TINYINT DEFAULT NULL`)
	}
	if !columnExists(ctx, db, "product_audit_logs", "to_status") {
		db.ExecContext(ctx, `ALTER TABLE product_audit_logs ADD COLUMN to_status TINYINT DEFAULT NULL`)
	}
	if !columnExists(ctx, db, "product_audit_logs", "before_snapshot") {
		db.ExecContext(ctx, `ALTER TABLE product_audit_logs ADD COLUMN before_snapshot JSON NULL`)
	}
	if !columnExists(ctx, db, "product_audit_logs", "after_snapshot") {
		db.ExecContext(ctx, `ALTER TABLE product_audit_logs ADD COLUMN after_snapshot JSON NULL`)
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

func buildProductAuditSnapshot(item productItem) productAuditSnapshot {
	return productAuditSnapshot{
		"id":                    item.ID,
		"name":                  item.Name,
		"categoryId":            item.CategoryID,
		"enterpriseId":          item.EnterpriseID,
		"category":              item.Category,
		"subCategory":           item.SubCategory,
		"steelMill":             item.SteelMill,
		"goodsType":             item.GoodsType,
		"warehouseServiceId":    item.WarehouseServiceID,
		"warehouseEnterpriseId": item.WarehouseEnterpriseID,
		"spec":                  item.Spec,
		"material":              item.Material,
		"warehouseName":         item.WarehouseName,
		"warehouseCode":         item.WarehouseCode,
		"province":              item.Province,
		"city":                  item.City,
		"district":              item.District,
		"warehouseRegion":       strings.Join(nonEmptyProductSnapshotStrings(item.Province, item.City, item.District), " / "),
		"address":               item.Address,
		"bundleNo":              item.BundleNo,
		"productNo":             item.ProductNo,
		"unit":                  item.Unit,
		"price":                 item.Price,
		"stock":                 item.Stock,
		"minOrderQty":           item.MinOrderQty,
		"deliveryDays":          item.DeliveryDays,
		"quoteValidity":         item.QuoteValidity,
		"description":           item.Description,
		"images":                item.Images,
		"status":                item.Status,
		"auditRemark":           item.AuditRemark,
		"auditorId":             item.AuditorID,
		"auditedAt":             item.AuditedAt,
		"createdAt":             item.CreatedAt,
		"updatedAt":             item.UpdatedAt,
	}
}

func nonEmptyProductSnapshotStrings(values ...string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}
	return result
}

func productAuditSnapshotJSON(snapshot productAuditSnapshot) string {
	if snapshot == nil {
		return ""
	}
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return ""
	}
	return string(raw)
}

func parseProductAuditSnapshot(raw sql.NullString) productAuditSnapshot {
	if !raw.Valid || strings.TrimSpace(raw.String) == "" {
		return nil
	}
	var snapshot productAuditSnapshot
	if err := json.Unmarshal([]byte(raw.String), &snapshot); err != nil {
		return nil
	}
	return snapshot
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
