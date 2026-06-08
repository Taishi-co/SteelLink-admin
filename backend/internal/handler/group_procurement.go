package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/Taishi-co/SteelLink-admin/backend/internal/database"
)

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
