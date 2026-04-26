package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Taishi-co/SteelLink-admin/backend/internal/database"
	"github.com/gin-gonic/gin"
)

const aiProcurementConfigKey = "ai_procurement_config"

type aiProcurementStageConfig struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	Enabled    bool   `json:"enabled"`
	Required   bool   `json:"required"`
	Sort       int    `json:"sort"`
	TimeoutSec int    `json:"timeoutSec"`
}

type aiProcurementModelConfig struct {
	Provider            string  `json:"provider"`
	Model               string  `json:"model"`
	Temperature         float64 `json:"temperature"`
	MaxTokens           int     `json:"maxTokens"`
	ConfidenceThreshold float64 `json:"confidenceThreshold"`
}

type aiProcurementMatchingConfig struct {
	CategoryBoost      int     `json:"categoryBoost"`
	SupplierPriority   string  `json:"supplierPriority"`
	PriceTolerancePct  float64 `json:"priceTolerancePct"`
	MaxRecommendations int     `json:"maxRecommendations"`
}

type aiProcurementFinanceConfig struct {
	Enabled          bool     `json:"enabled"`
	MinAmount        float64  `json:"minAmount"`
	MaxAmount        float64  `json:"maxAmount"`
	RequiredDocs     []string `json:"requiredDocs"`
	AutoSubmitReview bool     `json:"autoSubmitReview"`
}

type aiProcurementReviewConfig struct {
	ReviewerRole          string   `json:"reviewerRole"`
	ManualReviewThreshold float64  `json:"manualReviewThreshold"`
	RiskKeywords          []string `json:"riskKeywords"`
}

type aiProcurementConfig struct {
	Enabled   bool                        `json:"enabled"`
	Stages    []aiProcurementStageConfig  `json:"stages"`
	Model     aiProcurementModelConfig    `json:"model"`
	Matching  aiProcurementMatchingConfig `json:"matching"`
	Finance   aiProcurementFinanceConfig  `json:"finance"`
	Review    aiProcurementReviewConfig   `json:"review"`
	UpdatedAt *time.Time                  `json:"updatedAt,omitempty"`
}

func GetAIProcurementConfig(c *gin.Context) {
	config := defaultAIProcurementConfig()
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": config})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := ensureAdminConfigTable(ctx, db); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": config})
		return
	}

	var raw string
	var updatedAt sql.NullTime
	err := db.QueryRowContext(ctx, `SELECT config_value, updated_at FROM admin_system_configs WHERE config_key = ?`, aiProcurementConfigKey).Scan(&raw, &updatedAt)
	if err == nil && raw != "" {
		_ = json.Unmarshal([]byte(raw), &config)
		if updatedAt.Valid {
			config.UpdatedAt = &updatedAt.Time
		}
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": config})
}

func UpdateAIProcurementConfig(c *gin.Context) {
	var req aiProcurementConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid config payload"})
		return
	}
	normalizeAIProcurementConfig(&req)

	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "database unavailable"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := ensureAdminConfigTable(ctx, db); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	payload, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid config payload"})
		return
	}
	now := time.Now()
	_, err = db.ExecContext(ctx, `
		INSERT INTO admin_system_configs (config_key, config_value, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE config_value = VALUES(config_value), updated_at = VALUES(updated_at)`,
		aiProcurementConfigKey, string(payload), now, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	req.UpdatedAt = &now
	logOpFromContext(c, db, "ai_procurement", "update_config", 0, string(payload))
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": req})
}

func ensureAdminConfigTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS admin_system_configs (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
			config_key VARCHAR(100) NOT NULL UNIQUE,
			config_value JSON NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`)
	return err
}

func defaultAIProcurementConfig() aiProcurementConfig {
	return aiProcurementConfig{
		Enabled: true,
		Stages: []aiProcurementStageConfig{
			{Code: "instruction_parsing", Name: "指令解析", Enabled: true, Required: true, Sort: 10, TimeoutSec: 20},
			{Code: "product_matching", Name: "商品匹配", Enabled: true, Required: true, Sort: 20, TimeoutSec: 30},
			{Code: "order_generation", Name: "订单生成", Enabled: true, Required: true, Sort: 30, TimeoutSec: 20},
			{Code: "financing_matching", Name: "融资匹配", Enabled: true, Required: false, Sort: 40, TimeoutSec: 30},
			{Code: "review_submission", Name: "审核提交", Enabled: true, Required: true, Sort: 50, TimeoutSec: 15},
		},
		Model: aiProcurementModelConfig{
			Provider:            "platform",
			Model:               "steel-procurement-router",
			Temperature:         0.2,
			MaxTokens:           1200,
			ConfidenceThreshold: 0.78,
		},
		Matching: aiProcurementMatchingConfig{
			CategoryBoost:      30,
			SupplierPriority:   "approved_first",
			PriceTolerancePct:  8,
			MaxRecommendations: 5,
		},
		Finance: aiProcurementFinanceConfig{
			Enabled:          true,
			MinAmount:        10000,
			MaxAmount:        5000000,
			RequiredDocs:     []string{"营业执照", "采购订单", "历史交易流水"},
			AutoSubmitReview: false,
		},
		Review: aiProcurementReviewConfig{
			ReviewerRole:          "finance_reviewer",
			ManualReviewThreshold: 1000000,
			RiskKeywords:          []string{"超预算", "高风险供应商", "授信不足"},
		},
	}
}

func normalizeAIProcurementConfig(config *aiProcurementConfig) {
	if len(config.Stages) == 0 {
		config.Stages = defaultAIProcurementConfig().Stages
	}
	if config.Model.MaxTokens <= 0 {
		config.Model.MaxTokens = 1200
	}
	if config.Model.ConfidenceThreshold <= 0 {
		config.Model.ConfidenceThreshold = 0.78
	}
	if config.Matching.MaxRecommendations <= 0 {
		config.Matching.MaxRecommendations = 5
	}
	if config.Finance.RequiredDocs == nil {
		config.Finance.RequiredDocs = []string{}
	}
	if config.Review.RiskKeywords == nil {
		config.Review.RiskKeywords = []string{}
	}
}
