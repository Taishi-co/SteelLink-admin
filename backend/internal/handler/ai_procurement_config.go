package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
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
	ParsingEnabled         bool    `json:"parsingEnabled"`
	Provider               string  `json:"provider"`
	BaseURL                string  `json:"baseUrl"`
	APIKeyEnv              string  `json:"apiKeyEnv"`
	APIKeyConfigured       bool    `json:"apiKeyConfigured"`
	VisionProvider         string  `json:"visionProvider"`
	VisionBaseURL          string  `json:"visionBaseUrl"`
	VisionAPIKeyEnv        string  `json:"visionApiKeyEnv"`
	VisionAPIKeyConfigured bool    `json:"visionApiKeyConfigured"`
	Model                  string  `json:"model"`
	Temperature            float64 `json:"temperature"`
	TopP                   float64 `json:"topP"`
	MaxTokens              int     `json:"maxTokens"`
	TimeoutSec             int     `json:"timeoutSec"`
	RetryCount             int     `json:"retryCount"`
	ConfidenceThreshold    float64 `json:"confidenceThreshold"`
}

type aiProcurementMatchingConfig struct {
	Mode                     string  `json:"mode"`
	VectorEnabled            bool    `json:"vectorEnabled"`
	VectorProvider           string  `json:"vectorProvider"`
	VectorEndpoint           string  `json:"vectorEndpoint"`
	VectorDatabase           string  `json:"vectorDatabase"`
	VectorUsername           string  `json:"vectorUsername"`
	VectorPasswordEnv        string  `json:"vectorPasswordEnv"`
	VectorConnectionEnv      string  `json:"vectorConnectionEnv"`
	ProductCollection        string  `json:"productCollection"`
	SupplierCollection       string  `json:"supplierCollection"`
	FinanceProductCollection string  `json:"financeProductCollection"`
	EmbeddingProvider        string  `json:"embeddingProvider"`
	EmbeddingBaseURL         string  `json:"embeddingBaseUrl"`
	EmbeddingAPIKeyEnv       string  `json:"embeddingApiKeyEnv"`
	EmbeddingModel           string  `json:"embeddingModel"`
	EmbeddingDimension       int     `json:"embeddingDimension"`
	VectorTopK               int     `json:"vectorTopK"`
	VectorMinScore           float64 `json:"vectorMinScore"`
	HybridKeywordWeight      float64 `json:"hybridKeywordWeight"`
	HybridVectorWeight       float64 `json:"hybridVectorWeight"`
	CategoryBoost            int     `json:"categoryBoost"`
	SupplierPriority         string  `json:"supplierPriority"`
	PriceTolerancePct        float64 `json:"priceTolerancePct"`
	MaxRecommendations       int     `json:"maxRecommendations"`
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
	if !loadAIProcurementConfig(c.Request.Context(), &config) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": config})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": config})
}

func GetAIProcurementRuntimeConfig(c *gin.Context) {
	expectedToken := strings.TrimSpace(os.Getenv("AI_PROCUREMENT_CONFIG_SYNC_TOKEN"))
	if expectedToken != "" && c.GetHeader("X-Internal-Token") != expectedToken {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "invalid internal token"})
		return
	}

	config := defaultAIProcurementConfig()
	_ = loadAIProcurementConfig(c.Request.Context(), &config)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": config})
}

func loadAIProcurementConfig(parent context.Context, config *aiProcurementConfig) bool {
	db := database.GetDB()
	if db == nil {
		normalizeAIProcurementConfig(config)
		return false
	}

	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	if err := ensureAdminConfigTable(ctx, db); err != nil {
		normalizeAIProcurementConfig(config)
		return false
	}

	var raw string
	var updatedAt sql.NullTime
	err := db.QueryRowContext(ctx, `SELECT config_value, updated_at FROM admin_system_configs WHERE config_key = ?`, aiProcurementConfigKey).Scan(&raw, &updatedAt)
	if err == nil && raw != "" {
		_ = json.Unmarshal([]byte(raw), config)
		if updatedAt.Valid {
			config.UpdatedAt = &updatedAt.Time
		}
	}
	normalizeAIProcurementConfig(config)
	return true
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
	if err := syncAIProcurementConfigToRedis(ctx, payload); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "message": "redis unavailable: " + err.Error()})
		return
	}
	req.UpdatedAt = &now
	logOpFromContext(c, db, "ai_procurement", "update_config", 0, string(payload))
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": req})
}

func syncAIProcurementConfigToRedis(ctx context.Context, payload []byte) error {
	client := database.GetRedis()
	if client == nil {
		return errors.New("redis client is not initialized")
	}
	return client.Set(ctx, aiProcurementConfigKey, string(payload), 0).Err()
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
			ParsingEnabled:      false,
			Provider:            "minimax",
			BaseURL:             "https://api.minimax.io/v1/chat/completions",
			APIKeyEnv:           "MINIMAX_API_KEY",
			VisionProvider:      "minimax_token_plan",
			VisionBaseURL:       "https://api.minimax.io/v1/coding_plan/vlm",
			VisionAPIKeyEnv:     "MINIMAX_TOKEN_PLAN_API_KEY",
			Model:               "MiniMax-M2.7",
			Temperature:         0.2,
			MaxTokens:           1200,
			TopP:                0.9,
			TimeoutSec:          20,
			RetryCount:          1,
			ConfidenceThreshold: 0.78,
		},
		Matching: aiProcurementMatchingConfig{
			Mode:                     "hybrid",
			VectorEnabled:            false,
			VectorProvider:           "pgvector",
			VectorEndpoint:           "",
			VectorDatabase:           "steellink_vectors",
			VectorPasswordEnv:        "PGVECTOR_PASSWORD",
			VectorConnectionEnv:      "PGVECTOR_DSN",
			ProductCollection:        "ai_product_embeddings",
			SupplierCollection:       "ai_supplier_embeddings",
			FinanceProductCollection: "ai_finance_product_embeddings",
			EmbeddingProvider:        "minimax",
			EmbeddingBaseURL:         "https://api.minimax.io/v1/embeddings",
			EmbeddingAPIKeyEnv:       "MINIMAX_API_KEY",
			EmbeddingModel:           "embo-01",
			EmbeddingDimension:       1536,
			VectorTopK:               10,
			VectorMinScore:           0.65,
			HybridKeywordWeight:      0.35,
			HybridVectorWeight:       0.65,
			CategoryBoost:            30,
			SupplierPriority:         "approved_first",
			PriceTolerancePct:        8,
			MaxRecommendations:       5,
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
	defaults := defaultAIProcurementConfig()
	if len(config.Stages) == 0 {
		config.Stages = defaults.Stages
	}
	config.Model.Provider = strings.TrimSpace(config.Model.Provider)
	if config.Model.Provider == "" {
		config.Model.Provider = defaults.Model.Provider
	}
	config.Model.BaseURL = strings.TrimSpace(config.Model.BaseURL)
	if config.Model.BaseURL == "" {
		config.Model.BaseURL = defaults.Model.BaseURL
	}
	config.Model.APIKeyEnv = strings.TrimSpace(config.Model.APIKeyEnv)
	if config.Model.APIKeyEnv == "" {
		config.Model.APIKeyEnv = defaults.Model.APIKeyEnv
	}
	config.Model.VisionProvider = strings.TrimSpace(config.Model.VisionProvider)
	if config.Model.VisionProvider == "" {
		config.Model.VisionProvider = defaults.Model.VisionProvider
	}
	config.Model.VisionBaseURL = strings.TrimRight(strings.TrimSpace(config.Model.VisionBaseURL), "/")
	if config.Model.VisionBaseURL == "" {
		config.Model.VisionBaseURL = defaults.Model.VisionBaseURL
	}
	config.Model.VisionAPIKeyEnv = strings.TrimSpace(config.Model.VisionAPIKeyEnv)
	if config.Model.VisionAPIKeyEnv == "" {
		config.Model.VisionAPIKeyEnv = defaults.Model.VisionAPIKeyEnv
	}
	config.Model.Model = strings.TrimSpace(config.Model.Model)
	if config.Model.Model == "" {
		config.Model.Model = defaults.Model.Model
	}
	if config.Model.MaxTokens <= 0 {
		config.Model.MaxTokens = defaults.Model.MaxTokens
	}
	if config.Model.Temperature <= 0 || config.Model.Temperature > 1 {
		config.Model.Temperature = defaults.Model.Temperature
	}
	if config.Model.TopP <= 0 || config.Model.TopP > 1 {
		config.Model.TopP = defaults.Model.TopP
	}
	if config.Model.TimeoutSec <= 0 {
		config.Model.TimeoutSec = defaults.Model.TimeoutSec
	}
	if config.Model.RetryCount < 0 {
		config.Model.RetryCount = 0
	}
	if config.Model.ConfidenceThreshold <= 0 {
		config.Model.ConfidenceThreshold = defaults.Model.ConfidenceThreshold
	}
	config.Model.APIKeyConfigured = os.Getenv(config.Model.APIKeyEnv) != ""
	config.Model.VisionAPIKeyConfigured = os.Getenv(config.Model.VisionAPIKeyEnv) != ""
	if config.Matching.MaxRecommendations <= 0 {
		config.Matching.MaxRecommendations = 5
	}
	if config.Matching.Mode == "" {
		config.Matching.Mode = defaults.Matching.Mode
	}
	if config.Matching.VectorProvider == "" {
		config.Matching.VectorProvider = defaults.Matching.VectorProvider
	}
	if config.Matching.VectorEndpoint == "" {
		config.Matching.VectorEndpoint = defaults.Matching.VectorEndpoint
	}
	if config.Matching.VectorDatabase == "" {
		config.Matching.VectorDatabase = defaults.Matching.VectorDatabase
	}
	if config.Matching.VectorPasswordEnv == "" {
		config.Matching.VectorPasswordEnv = defaults.Matching.VectorPasswordEnv
	}
	if config.Matching.VectorConnectionEnv == "" {
		config.Matching.VectorConnectionEnv = defaults.Matching.VectorConnectionEnv
	}
	if config.Matching.ProductCollection == "" {
		config.Matching.ProductCollection = defaults.Matching.ProductCollection
	}
	if config.Matching.SupplierCollection == "" {
		config.Matching.SupplierCollection = defaults.Matching.SupplierCollection
	}
	if config.Matching.FinanceProductCollection == "" {
		config.Matching.FinanceProductCollection = defaults.Matching.FinanceProductCollection
	}
	if config.Matching.EmbeddingProvider == "" {
		config.Matching.EmbeddingProvider = defaults.Matching.EmbeddingProvider
	}
	if config.Matching.EmbeddingBaseURL == "" {
		config.Matching.EmbeddingBaseURL = defaults.Matching.EmbeddingBaseURL
	}
	if config.Matching.EmbeddingAPIKeyEnv == "" {
		config.Matching.EmbeddingAPIKeyEnv = defaults.Matching.EmbeddingAPIKeyEnv
	}
	if config.Matching.EmbeddingModel == "" {
		config.Matching.EmbeddingModel = defaults.Matching.EmbeddingModel
	}
	if config.Matching.EmbeddingDimension <= 0 {
		config.Matching.EmbeddingDimension = defaults.Matching.EmbeddingDimension
	}
	if config.Matching.VectorTopK <= 0 {
		config.Matching.VectorTopK = defaults.Matching.VectorTopK
	}
	if config.Matching.VectorMinScore <= 0 || config.Matching.VectorMinScore > 1 {
		config.Matching.VectorMinScore = defaults.Matching.VectorMinScore
	}
	if config.Matching.HybridKeywordWeight <= 0 {
		config.Matching.HybridKeywordWeight = defaults.Matching.HybridKeywordWeight
	}
	if config.Matching.HybridVectorWeight <= 0 {
		config.Matching.HybridVectorWeight = defaults.Matching.HybridVectorWeight
	}
	if config.Finance.RequiredDocs == nil {
		config.Finance.RequiredDocs = []string{}
	}
	if config.Review.RiskKeywords == nil {
		config.Review.RiskKeywords = []string{}
	}
}
