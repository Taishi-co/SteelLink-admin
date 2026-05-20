package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-admin/backend/internal/database"
	"github.com/gin-gonic/gin"
)

type sendNotificationRequest struct {
	RecipientUserIDs       []int64                `json:"recipient_user_ids"`
	RecipientEnterpriseIDs []int64                `json:"recipient_enterprise_ids"`
	RecipientRoles         []string               `json:"recipient_roles"`
	SendToAll              bool                   `json:"send_to_all"`
	Type                   string                 `json:"type"`
	Title                  string                 `json:"title" binding:"required"`
	Content                string                 `json:"content" binding:"required"`
	LinkURL                string                 `json:"link_url"`
	Metadata               map[string]interface{} `json:"metadata"`
}

type notificationRecipient struct {
	UserID         int64
	EnterpriseID   int64
	EnterpriseCode string
}

func SendNotification(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	var req sendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	req.Type = strings.TrimSpace(req.Type)
	if req.Type == "" {
		req.Type = "admin"
	}
	if !validNotificationType(req.Type) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "消息类型无效"})
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	req.Content = strings.TrimSpace(req.Content)
	req.LinkURL = strings.TrimSpace(req.LinkURL)
	if req.Title == "" || req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "消息标题和内容不能为空"})
		return
	}
	if len(req.Title) > 120 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "消息标题最多 120 字"})
		return
	}
	if len(req.Content) > 5000 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "消息内容最多 5000 字"})
		return
	}
	if len(req.LinkURL) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "站内链接最多 500 字"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	if err := ensureNotificationsTable(ctx, db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	recipients, err := resolveNotificationRecipients(ctx, db, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if len(recipients) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "没有可接收站内信的用户"})
		return
	}

	metadata := req.Metadata
	if metadata == nil {
		metadata = map[string]interface{}{}
	}
	metadata["admin_id"] = currentAdminID(c)
	if adminName, _ := c.Get("admin_real_name"); adminName != nil {
		metadata["admin_name"] = fmt.Sprint(adminName)
	}
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "metadata 不是有效 JSON"})
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO notifications (
			recipient_user_id, recipient_enterprise_id, sender_user_id,
			type, title, content, link_url, status, metadata, created_at, updated_at
		) VALUES (?, ?, 0, ?, ?, ?, ?, 0, ?, NOW(), NOW())`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	defer stmt.Close()

	created := 0
	for _, recipient := range recipients {
		if _, err := stmt.ExecContext(ctx, recipient.UserID, recipient.EnterpriseID, req.Type, req.Title, req.Content, req.LinkURL, string(metadataBytes)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
			return
		}
		created++
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	logOpFromContext(c, db, "notification", "send", 0, fmt.Sprintf(`{"created":%d}`, created))
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    gin.H{"created": created},
	})
}

func validNotificationType(value string) bool {
	switch value {
	case "admin", "order", "audit", "system":
		return true
	default:
		return false
	}
}

func ensureNotificationsTable(ctx context.Context, db *sql.DB) error {
	if !tableExists(ctx, db, "notifications") {
		_, err := db.ExecContext(ctx, `CREATE TABLE notifications (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			recipient_user_id BIGINT UNSIGNED NOT NULL,
			recipient_enterprise_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
			sender_user_id BIGINT UNSIGNED NOT NULL DEFAULT 0,
			type VARCHAR(32) NOT NULL,
			title VARCHAR(120) NOT NULL,
			content TEXT NOT NULL,
			link_url VARCHAR(500) NOT NULL DEFAULT '',
			status TINYINT NOT NULL DEFAULT 0,
			read_at DATETIME NULL DEFAULT NULL,
			metadata JSON NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME NULL DEFAULT NULL,
			KEY idx_notifications_recipient_user_id (recipient_user_id),
			KEY idx_notifications_recipient_enterprise_id (recipient_enterprise_id),
			KEY idx_notifications_sender_user_id (sender_user_id),
			KEY idx_notifications_type (type),
			KEY idx_notifications_status (status),
			KEY idx_notifications_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
		return err
	}

	columns := map[string]string{
		"recipient_enterprise_id": "ALTER TABLE notifications ADD COLUMN recipient_enterprise_id BIGINT UNSIGNED NOT NULL DEFAULT 0",
		"sender_user_id":          "ALTER TABLE notifications ADD COLUMN sender_user_id BIGINT UNSIGNED NOT NULL DEFAULT 0",
		"link_url":                "ALTER TABLE notifications ADD COLUMN link_url VARCHAR(500) NOT NULL DEFAULT ''",
		"read_at":                 "ALTER TABLE notifications ADD COLUMN read_at DATETIME NULL DEFAULT NULL",
		"metadata":                "ALTER TABLE notifications ADD COLUMN metadata JSON NULL",
		"updated_at":              "ALTER TABLE notifications ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP",
		"deleted_at":              "ALTER TABLE notifications ADD COLUMN deleted_at DATETIME NULL DEFAULT NULL",
	}
	for column, stmt := range columns {
		if !columnExists(ctx, db, "notifications", column) {
			if _, err := db.ExecContext(ctx, stmt); err != nil {
				return err
			}
		}
	}
	return nil
}

func resolveNotificationRecipients(ctx context.Context, db *sql.DB, req sendNotificationRequest) ([]notificationRecipient, error) {
	userCols := columnsExist(ctx, db, "users")
	if len(userCols) == 0 {
		return nil, fmt.Errorf("users table unavailable")
	}

	selectParts := []string{"id"}
	if userCols["enterprise_id"] {
		selectParts = append(selectParts, "COALESCE(enterprise_id, 0)")
	} else {
		selectParts = append(selectParts, "0")
	}
	if userCols["enterprise_code"] {
		selectParts = append(selectParts, "COALESCE(enterprise_code, '')")
	} else {
		selectParts = append(selectParts, "''")
	}

	where := make([]string, 0, 3)
	args := make([]any, 0)
	if userCols["deleted_at"] {
		where = append(where, "deleted_at IS NULL")
	}
	if userCols["status"] {
		where = append(where, "status = 1")
	}

	if !req.SendToAll {
		filters := make([]string, 0, 3)
		if ids := positiveInt64s(req.RecipientUserIDs); len(ids) > 0 {
			filters = append(filters, "id IN ("+sqlPlaceholders(len(ids))+")")
			for _, id := range ids {
				args = append(args, id)
			}
		}
		if enterpriseFilter, enterpriseArgs, err := notificationEnterpriseFilter(ctx, db, userCols, req.RecipientEnterpriseIDs); err != nil {
			return nil, err
		} else if enterpriseFilter != "" {
			filters = append(filters, enterpriseFilter)
			args = append(args, enterpriseArgs...)
		}
		if roles := nonEmptyStrings(req.RecipientRoles); len(roles) > 0 && userCols["role"] {
			filters = append(filters, "role IN ("+sqlPlaceholders(len(roles))+")")
			for _, role := range roles {
				args = append(args, role)
			}
		}
		if len(filters) == 0 {
			return nil, nil
		}
		where = append(where, "("+strings.Join(filters, " OR ")+")")
	}
	if len(where) == 0 {
		where = append(where, "1 = 1")
	}

	rows, err := db.QueryContext(ctx, fmt.Sprintf(
		"SELECT %s FROM users WHERE %s ORDER BY id ASC",
		strings.Join(selectParts, ", "),
		strings.Join(where, " AND "),
	), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recipients := make([]notificationRecipient, 0)
	seen := map[int64]struct{}{}
	codes := make([]string, 0)
	for rows.Next() {
		var recipient notificationRecipient
		if err := rows.Scan(&recipient.UserID, &recipient.EnterpriseID, &recipient.EnterpriseCode); err != nil {
			return nil, err
		}
		if recipient.UserID <= 0 {
			continue
		}
		if _, ok := seen[recipient.UserID]; ok {
			continue
		}
		seen[recipient.UserID] = struct{}{}
		if recipient.EnterpriseID == 0 && recipient.EnterpriseCode != "" {
			codes = append(codes, recipient.EnterpriseCode)
		}
		recipients = append(recipients, recipient)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	codeToID, err := enterpriseIDsByCode(ctx, db, codes)
	if err != nil {
		return nil, err
	}
	for i := range recipients {
		if recipients[i].EnterpriseID == 0 && recipients[i].EnterpriseCode != "" {
			recipients[i].EnterpriseID = codeToID[recipients[i].EnterpriseCode]
		}
	}
	return recipients, nil
}

func notificationEnterpriseFilter(ctx context.Context, db *sql.DB, userCols map[string]bool, rawIDs []int64) (string, []any, error) {
	ids := positiveInt64s(rawIDs)
	if len(ids) == 0 {
		return "", nil, nil
	}
	if userCols["enterprise_id"] {
		args := make([]any, 0, len(ids))
		for _, id := range ids {
			args = append(args, id)
		}
		return "enterprise_id IN (" + sqlPlaceholders(len(ids)) + ")", args, nil
	}
	if !userCols["enterprise_code"] || !columnExists(ctx, db, "enterprises", "enterprise_code") {
		return "", nil, nil
	}
	codes, err := enterpriseCodesByID(ctx, db, ids)
	if err != nil {
		return "", nil, err
	}
	if len(codes) == 0 {
		return "", nil, nil
	}
	args := make([]any, 0, len(codes))
	for _, code := range codes {
		args = append(args, code)
	}
	return "enterprise_code IN (" + sqlPlaceholders(len(codes)) + ")", args, nil
}

func enterpriseCodesByID(ctx context.Context, db *sql.DB, ids []int64) ([]string, error) {
	if len(ids) == 0 || !tableExists(ctx, db, "enterprises") || !columnExists(ctx, db, "enterprises", "enterprise_code") {
		return nil, nil
	}
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}
	rows, err := db.QueryContext(ctx, "SELECT enterprise_code FROM enterprises WHERE id IN ("+sqlPlaceholders(len(ids))+")", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	codes := make([]string, 0)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		if strings.TrimSpace(code) != "" {
			codes = append(codes, code)
		}
	}
	return codes, rows.Err()
}

func enterpriseIDsByCode(ctx context.Context, db *sql.DB, codes []string) (map[string]int64, error) {
	result := map[string]int64{}
	codes = nonEmptyStrings(codes)
	if len(codes) == 0 || !tableExists(ctx, db, "enterprises") || !columnExists(ctx, db, "enterprises", "enterprise_code") {
		return result, nil
	}
	args := make([]any, 0, len(codes))
	for _, code := range codes {
		args = append(args, code)
	}
	rows, err := db.QueryContext(ctx, "SELECT id, enterprise_code FROM enterprises WHERE enterprise_code IN ("+sqlPlaceholders(len(codes))+")", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var code string
		if err := rows.Scan(&id, &code); err != nil {
			return nil, err
		}
		result[code] = id
	}
	return result, rows.Err()
}

func positiveInt64s(values []int64) []int64 {
	seen := map[int64]struct{}{}
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func nonEmptyStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func sqlPlaceholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.TrimRight(strings.Repeat("?,", count), ",")
}
