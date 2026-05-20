package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Taishi-co/SteelLink-admin/backend/internal/database"
	"github.com/gin-gonic/gin"
)

const aiProcurementMonitorStaleAfter = time.Hour

type aiProcurementMonitorResponse struct {
	TableReady   bool                                 `json:"tableReady"`
	GeneratedAt  time.Time                            `json:"generatedAt"`
	Config       aiProcurementMonitorConfigSummary    `json:"config"`
	Stats        aiProcurementMonitorStats            `json:"stats"`
	StageCounts  []aiProcurementMonitorCount          `json:"stageCounts"`
	StatusCounts []aiProcurementMonitorCount          `json:"statusCounts"`
	Records      pageResult[aiProcurementMonitorItem] `json:"records"`
}

type aiProcurementMonitorConfigSummary struct {
	Enabled                bool       `json:"enabled"`
	RuntimeStatus          string     `json:"runtimeStatus"`
	Warnings               []string   `json:"warnings"`
	UpdatedAt              *time.Time `json:"updatedAt,omitempty"`
	ParsingEnabled         bool       `json:"parsingEnabled"`
	ModelProvider          string     `json:"modelProvider"`
	Model                  string     `json:"model"`
	ModelAPIKeyConfigured  bool       `json:"modelApiKeyConfigured"`
	VisionAPIKeyConfigured bool       `json:"visionApiKeyConfigured"`
	VectorEnabled          bool       `json:"vectorEnabled"`
	VectorProvider         string     `json:"vectorProvider"`
	VectorTopK             int        `json:"vectorTopK"`
	VectorConnectionReady  bool       `json:"vectorConnectionReady"`
	FinanceEnabled         bool       `json:"financeEnabled"`
	AutoSubmitReview       bool       `json:"autoSubmitReview"`
	MaxRecommendations     int        `json:"maxRecommendations"`
}

type aiProcurementMonitorStats struct {
	Total             int `json:"total"`
	Today             int `json:"today"`
	Active            int `json:"active"`
	ReviewSubmitted   int `json:"reviewSubmitted"`
	Failed            int `json:"failed"`
	Stale             int `json:"stale"`
	NeedsIntervention int `json:"needsIntervention"`
}

type aiProcurementMonitorCount struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type aiProcurementMonitorItem struct {
	ID                 int64      `json:"id"`
	ProcurementNo      string     `json:"procurementNo"`
	Prompt             string     `json:"prompt,omitempty"`
	ProductName        string     `json:"productName,omitempty"`
	Category           string     `json:"category,omitempty"`
	Quantity           float64    `json:"quantity"`
	Unit               string     `json:"unit,omitempty"`
	FinancingRequired  bool       `json:"financingRequired"`
	CurrentStage       string     `json:"currentStage"`
	CurrentStageLabel  string     `json:"currentStageLabel"`
	Status             string     `json:"status"`
	StatusLabel        string     `json:"statusLabel"`
	ParseConfidence    float64    `json:"parseConfidence,omitempty"`
	MatchCount         int        `json:"matchCount"`
	OrderCount         int        `json:"orderCount"`
	FinancePlanCount   int        `json:"financePlanCount"`
	AuditItemCount     int        `json:"auditItemCount"`
	HistorySaved       bool       `json:"historySaved"`
	EnterpriseID       int64      `json:"enterpriseId"`
	CreatedBy          int64      `json:"createdBy"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	ReviewSubmittedAt  *time.Time `json:"reviewSubmittedAt,omitempty"`
	NeedsIntervention  bool       `json:"needsIntervention"`
	InterventionReason string     `json:"interventionReason,omitempty"`
}

func GetAIProcurementMonitor(c *gin.Context) {
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	config := defaultAIProcurementConfig()
	_ = loadAIProcurementConfig(ctx, &config)
	response := aiProcurementMonitorResponse{
		TableReady:  false,
		GeneratedAt: time.Now(),
		Config:      buildAIProcurementMonitorConfigSummary(config),
		Records:     pageResult[aiProcurementMonitorItem]{List: []aiProcurementMonitorItem{}, Total: 0},
	}

	if !tableExists(ctx, db, "ai_procurements") {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": response})
		return
	}
	response.TableReady = true

	columns := columnsExist(ctx, db, "ai_procurements")
	baseWhere, baseArgs := aiProcurementBaseWhere(columns)
	staleBefore := time.Now().Add(-aiProcurementMonitorStaleAfter)

	stats, err := buildAIProcurementMonitorStats(ctx, db, columns, config, baseWhere, baseArgs, staleBefore)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	response.Stats = stats

	stageCounts, err := listAIProcurementGroupedCounts(ctx, db, "current_stage", baseWhere, baseArgs, aiProcurementStageLabel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	response.StageCounts = stageCounts

	statusCounts, err := listAIProcurementGroupedCounts(ctx, db, "status", baseWhere, baseArgs, aiProcurementStatusLabel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	response.StatusCounts = statusCounts

	records, total, err := listAIProcurementMonitorItems(ctx, db, columns, config, staleBefore, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	response.Records = pageResult[aiProcurementMonitorItem]{List: records, Total: total}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": response})
}

func aiProcurementBaseWhere(columns map[string]bool) ([]string, []any) {
	whereParts := []string{"1=1"}
	if columns["deleted_at"] {
		whereParts = append(whereParts, "deleted_at IS NULL")
	}
	return whereParts, []any{}
}

func buildAIProcurementMonitorStats(ctx context.Context, db *sql.DB, columns map[string]bool, config aiProcurementConfig, baseWhere []string, baseArgs []any, staleBefore time.Time) (aiProcurementMonitorStats, error) {
	whereClause := strings.Join(baseWhere, " AND ")
	stats := aiProcurementMonitorStats{}
	var err error
	if stats.Total, err = queryCount(ctx, db, fmt.Sprintf("SELECT COUNT(*) FROM ai_procurements WHERE %s", whereClause), baseArgs...); err != nil {
		return stats, err
	}
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if stats.Today, err = queryCount(ctx, db, fmt.Sprintf("SELECT COUNT(*) FROM ai_procurements WHERE %s AND created_at >= ?", whereClause), appendArgs(baseArgs, todayStart)...); err != nil {
		return stats, err
	}
	if stats.ReviewSubmitted, err = queryCount(ctx, db, fmt.Sprintf("SELECT COUNT(*) FROM ai_procurements WHERE %s AND (status = 'review_submitted' OR current_stage = 'review_submission')", whereClause), baseArgs...); err != nil {
		return stats, err
	}
	if stats.Failed, err = queryCount(ctx, db, fmt.Sprintf("SELECT COUNT(*) FROM ai_procurements WHERE %s AND LOWER(COALESCE(status, '')) IN ('failed', 'error', 'cancelled')", whereClause), baseArgs...); err != nil {
		return stats, err
	}
	if stats.Stale, err = queryCount(ctx, db, fmt.Sprintf("SELECT COUNT(*) FROM ai_procurements WHERE %s AND updated_at < ? AND COALESCE(status, '') <> 'review_submitted' AND LOWER(COALESCE(status, '')) NOT IN ('failed', 'error', 'cancelled')", whereClause), appendArgs(baseArgs, staleBefore)...); err != nil {
		return stats, err
	}
	interventionWhere, interventionArgs := aiProcurementInterventionWhere(columns, config, staleBefore)
	if stats.NeedsIntervention, err = queryCount(ctx, db, fmt.Sprintf("SELECT COUNT(*) FROM ai_procurements WHERE %s AND %s", whereClause, interventionWhere), appendArgs(baseArgs, interventionArgs...)...); err != nil {
		return stats, err
	}
	stats.Active = stats.Total - stats.ReviewSubmitted - stats.Failed
	if stats.Active < 0 {
		stats.Active = 0
	}
	return stats, nil
}

func listAIProcurementGroupedCounts(ctx context.Context, db *sql.DB, column string, baseWhere []string, baseArgs []any, labeler func(string) string) ([]aiProcurementMonitorCount, error) {
	whereClause := strings.Join(baseWhere, " AND ")
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT COALESCE(%s, ''), COUNT(*)
		FROM ai_procurements
		WHERE %s
		GROUP BY COALESCE(%s, '')
		ORDER BY COUNT(*) DESC`, column, whereClause, column), baseArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]aiProcurementMonitorCount, 0)
	for rows.Next() {
		var key string
		var count int
		if err := rows.Scan(&key, &count); err != nil {
			return nil, err
		}
		items = append(items, aiProcurementMonitorCount{Key: key, Label: labeler(key), Count: count})
	}
	return items, rows.Err()
}

func listAIProcurementMonitorItems(ctx context.Context, db *sql.DB, columns map[string]bool, config aiProcurementConfig, staleBefore time.Time, c *gin.Context) ([]aiProcurementMonitorItem, int, error) {
	limit, offset := parsePageParams(c.Query("page"), c.Query("pageSize"))
	whereParts, args := aiProcurementBaseWhere(columns)
	if keyword := strings.TrimSpace(c.Query("keyword")); keyword != "" {
		like := "%" + keyword + "%"
		whereParts = append(whereParts, "(procurement_no LIKE ? OR prompt LIKE ? OR product_name LIKE ? OR category LIKE ?)")
		args = append(args, like, like, like, like)
	}
	if status := strings.TrimSpace(c.Query("status")); status != "" {
		whereParts = append(whereParts, "status = ?")
		args = append(args, status)
	}
	if stage := strings.TrimSpace(c.Query("stage")); stage != "" {
		whereParts = append(whereParts, "current_stage = ?")
		args = append(args, stage)
	}
	if c.Query("intervention") == "1" {
		interventionWhere, interventionArgs := aiProcurementInterventionWhere(columns, config, staleBefore)
		whereParts = append(whereParts, interventionWhere)
		args = append(args, interventionArgs...)
	}

	whereClause := strings.Join(whereParts, " AND ")
	total, err := queryCount(ctx, db, fmt.Sprintf("SELECT COUNT(*) FROM ai_procurements WHERE %s", whereClause), args...)
	if err != nil {
		return nil, 0, err
	}

	queryArgs := appendArgs(args, limit, offset)
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, COALESCE(procurement_no, ''), COALESCE(prompt, ''), COALESCE(product_name, ''), COALESCE(category, ''),
		       COALESCE(quantity, 0), COALESCE(unit, ''), %s,
		       COALESCE(current_stage, ''), COALESCE(status, ''),
		       %s, %s, %s, %s, %s,
		       %s, COALESCE(enterprise_id, 0), COALESCE(created_by, 0),
		       created_at, updated_at, %s
		FROM ai_procurements
		WHERE %s
		ORDER BY updated_at DESC, id DESC
		LIMIT ? OFFSET ?`,
		aiProcurementIntColumnExpr(columns, "financing_required", "0"),
		aiProcurementJSONColumnExpr(columns, "parsed_instruction"),
		aiProcurementJSONColumnExpr(columns, "matched_products"),
		aiProcurementJSONColumnExpr(columns, "order_draft"),
		aiProcurementJSONColumnExpr(columns, "finance_plans"),
		aiProcurementJSONColumnExpr(columns, "audit_items"),
		aiProcurementIntColumnExpr(columns, "history_saved", "1"),
		aiProcurementNullableColumnExpr(columns, "review_submitted_at"),
		whereClause), queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]aiProcurementMonitorItem, 0)
	for rows.Next() {
		item, err := scanAIProcurementMonitorItem(rows, config, staleBefore)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func scanAIProcurementMonitorItem(rows *sql.Rows, config aiProcurementConfig, staleBefore time.Time) (aiProcurementMonitorItem, error) {
	var item aiProcurementMonitorItem
	var financingRequired int
	var parsedInstruction sql.NullString
	var matchedProducts sql.NullString
	var orderDraft sql.NullString
	var financePlans sql.NullString
	var auditItems sql.NullString
	var historySaved int
	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	var reviewSubmittedAt sql.NullTime
	if err := rows.Scan(
		&item.ID,
		&item.ProcurementNo,
		&item.Prompt,
		&item.ProductName,
		&item.Category,
		&item.Quantity,
		&item.Unit,
		&financingRequired,
		&item.CurrentStage,
		&item.Status,
		&parsedInstruction,
		&matchedProducts,
		&orderDraft,
		&financePlans,
		&auditItems,
		&historySaved,
		&item.EnterpriseID,
		&item.CreatedBy,
		&createdAt,
		&updatedAt,
		&reviewSubmittedAt,
	); err != nil {
		return item, err
	}
	item.FinancingRequired = financingRequired != 0
	item.HistorySaved = historySaved != 0
	if createdAt.Valid {
		item.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		item.UpdatedAt = updatedAt.Time
	}
	if reviewSubmittedAt.Valid {
		item.ReviewSubmittedAt = &reviewSubmittedAt.Time
	}
	item.CurrentStageLabel = aiProcurementStageLabel(item.CurrentStage)
	item.StatusLabel = aiProcurementStatusLabel(item.Status)
	item.ParseConfidence = aiProcurementJSONNumber(parsedInstruction.String, "confidence")
	item.MatchCount = aiProcurementJSONArrayLength(matchedProducts.String)
	item.OrderCount = aiProcurementOrderCount(orderDraft.String)
	item.FinancePlanCount = aiProcurementJSONArrayLength(financePlans.String)
	item.AuditItemCount = aiProcurementJSONArrayLength(auditItems.String)
	item.NeedsIntervention, item.InterventionReason = aiProcurementIntervention(item, config, staleBefore)
	return item, nil
}

func aiProcurementInterventionWhere(columns map[string]bool, config aiProcurementConfig, staleBefore time.Time) (string, []any) {
	parts := []string{
		"status = 'review_submitted'",
		"current_stage = 'review_submission'",
		"LOWER(COALESCE(status, '')) IN ('failed', 'error', 'cancelled')",
		"(updated_at < ? AND COALESCE(status, '') <> 'review_submitted')",
	}
	args := []any{staleBefore}

	if columns["parsed_instruction"] && config.Model.ConfidenceThreshold > 0 {
		parts = append(parts, "(JSON_EXTRACT(parsed_instruction, '$.confidence') IS NOT NULL AND CAST(JSON_UNQUOTE(JSON_EXTRACT(parsed_instruction, '$.confidence')) AS DECIMAL(10,4)) < ?)")
		args = append(args, config.Model.ConfidenceThreshold)
	}
	if columns["matched_products"] {
		parts = append(parts, "(current_stage IN ('product_matching', 'logistics_warehouse_matching', 'order_generation', 'financing_matching', 'review_submission') AND COALESCE(JSON_LENGTH(matched_products), 0) = 0)")
	}
	if columns["order_draft"] {
		parts = append(parts, "(current_stage IN ('order_generation', 'financing_matching', 'review_submission') AND COALESCE(JSON_LENGTH(JSON_EXTRACT(order_draft, '$.orders')), 0) = 0 AND COALESCE(CAST(JSON_UNQUOTE(JSON_EXTRACT(order_draft, '$.order_id')) AS UNSIGNED), 0) = 0)")
	}

	return "(" + strings.Join(parts, " OR ") + ")", args
}

func aiProcurementJSONColumnExpr(columns map[string]bool, column string) string {
	if !columns[column] {
		return "''"
	}
	return fmt.Sprintf("COALESCE(CAST(%s AS CHAR), '')", column)
}

func aiProcurementIntColumnExpr(columns map[string]bool, column string, fallback string) string {
	if !columns[column] {
		return fallback
	}
	return fmt.Sprintf("COALESCE(%s, %s)", column, fallback)
}

func aiProcurementNullableColumnExpr(columns map[string]bool, column string) string {
	if !columns[column] {
		return "NULL"
	}
	return column
}

func appendArgs(args []any, values ...any) []any {
	result := append([]any{}, args...)
	result = append(result, values...)
	return result
}

func buildAIProcurementMonitorConfigSummary(config aiProcurementConfig) aiProcurementMonitorConfigSummary {
	warnings := make([]string, 0)
	if !config.Enabled {
		warnings = append(warnings, "AI采购总开关未启用")
	}
	if config.Model.ParsingEnabled && !config.Model.APIKeyConfigured {
		warnings = append(warnings, "模型解析已启用但模型 API Key 未配置")
	}
	if config.Matching.VectorEnabled && strings.TrimSpace(config.Matching.VectorConnectionEnv) != "" && os.Getenv(config.Matching.VectorConnectionEnv) == "" {
		warnings = append(warnings, "向量召回已启用但 pgvector 连接环境变量未配置")
	}
	runtimeStatus := "healthy"
	if !config.Enabled {
		runtimeStatus = "disabled"
	} else if len(warnings) > 0 {
		runtimeStatus = "warning"
	}
	return aiProcurementMonitorConfigSummary{
		Enabled:                config.Enabled,
		RuntimeStatus:          runtimeStatus,
		Warnings:               warnings,
		UpdatedAt:              config.UpdatedAt,
		ParsingEnabled:         config.Model.ParsingEnabled,
		ModelProvider:          config.Model.Provider,
		Model:                  config.Model.Model,
		ModelAPIKeyConfigured:  config.Model.APIKeyConfigured,
		VisionAPIKeyConfigured: config.Model.VisionAPIKeyConfigured,
		VectorEnabled:          config.Matching.VectorEnabled,
		VectorProvider:         config.Matching.VectorProvider,
		VectorTopK:             config.Matching.VectorTopK,
		VectorConnectionReady:  strings.TrimSpace(config.Matching.VectorConnectionEnv) == "" || os.Getenv(config.Matching.VectorConnectionEnv) != "",
		FinanceEnabled:         config.Finance.Enabled,
		AutoSubmitReview:       config.Finance.AutoSubmitReview,
		MaxRecommendations:     config.Matching.MaxRecommendations,
	}
}

func aiProcurementIntervention(item aiProcurementMonitorItem, config aiProcurementConfig, staleBefore time.Time) (bool, string) {
	status := strings.ToLower(item.Status)
	if status == "failed" || status == "error" || status == "cancelled" {
		return true, "流程异常：" + aiProcurementStatusLabel(item.Status)
	}
	if item.Status == "review_submitted" || item.CurrentStage == "review_submission" {
		return true, "已提交审核，等待人工复核"
	}
	if item.ParseConfidence > 0 && item.ParseConfidence < config.Model.ConfidenceThreshold {
		return true, fmt.Sprintf("解析置信度 %.0f%% 低于阈值 %.0f%%", item.ParseConfidence*100, config.Model.ConfidenceThreshold*100)
	}
	if aiProcurementStageReached(item.CurrentStage, "product_matching") && item.MatchCount == 0 {
		return true, "已进入商品匹配阶段但暂无匹配商品"
	}
	if aiProcurementStageReached(item.CurrentStage, "order_generation") && item.OrderCount == 0 {
		return true, "已进入订单生成阶段但暂无订单草稿"
	}
	if !item.UpdatedAt.IsZero() && item.UpdatedAt.Before(staleBefore) && !isFinalAIProcurementStatus(item.Status) {
		return true, "超过1小时未推进"
	}
	return false, ""
}

func aiProcurementStageReached(current, target string) bool {
	positions := map[string]int{
		"instruction_parsing":          1,
		"product_matching":             2,
		"logistics_warehouse_matching": 3,
		"order_generation":             4,
		"financing_matching":           5,
		"review_submission":            6,
	}
	return positions[current] >= positions[target] && positions[target] > 0
}

func isFinalAIProcurementStatus(status string) bool {
	switch strings.ToLower(status) {
	case "review_submitted", "failed", "error", "cancelled":
		return true
	default:
		return false
	}
}

func aiProcurementStageLabel(stage string) string {
	switch stage {
	case "instruction_parsing":
		return "指令解析"
	case "product_matching":
		return "商品匹配"
	case "logistics_warehouse_matching":
		return "物流/仓储匹配"
	case "order_generation":
		return "订单生成"
	case "financing_matching":
		return "融资匹配"
	case "review_submission":
		return "审核提交"
	default:
		if strings.TrimSpace(stage) == "" {
			return "未开始"
		}
		return stage
	}
}

func aiProcurementStatusLabel(status string) string {
	switch status {
	case "draft":
		return "草稿"
	case "parsed":
		return "已解析"
	case "matched":
		return "已匹配"
	case "order_generated":
		return "已生成订单"
	case "financing_ready":
		return "融资方案就绪"
	case "review_submitted":
		return "已提交审核"
	case "failed":
		return "失败"
	case "error":
		return "异常"
	case "cancelled":
		return "已取消"
	default:
		if strings.TrimSpace(status) == "" {
			return "未知"
		}
		return status
	}
}

func aiProcurementJSONArrayLength(raw string) int {
	if strings.TrimSpace(raw) == "" {
		return 0
	}
	var items []any
	if err := json.Unmarshal([]byte(raw), &items); err == nil {
		return len(items)
	}
	return 0
}

func aiProcurementOrderCount(raw string) int {
	if strings.TrimSpace(raw) == "" {
		return 0
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return 0
	}
	if orders, ok := payload["orders"].([]any); ok {
		return len(orders)
	}
	if id, ok := numericJSONValue(payload["order_id"]); ok && id > 0 {
		return 1
	}
	return 0
}

func aiProcurementJSONNumber(raw string, key string) float64 {
	if strings.TrimSpace(raw) == "" {
		return 0
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return 0
	}
	value, _ := numericJSONValue(payload[key])
	return value
}

func numericJSONValue(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case json.Number:
		parsed, err := strconv.ParseFloat(string(typed), 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}
