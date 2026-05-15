package handler

import "testing"

func TestNormalizeAIProcurementConfigFillsVisionDefaults(t *testing.T) {
	config := aiProcurementConfig{}

	normalizeAIProcurementConfig(&config)

	if config.Model.VisionProvider != "minimax_token_plan" {
		t.Fatalf("VisionProvider = %q", config.Model.VisionProvider)
	}
	if config.Model.VisionBaseURL != "https://api.minimaxi.com/v1/coding_plan/vlm" {
		t.Fatalf("VisionBaseURL = %q", config.Model.VisionBaseURL)
	}
	if config.Model.VisionAPIKeyEnv != "MINIMAX_TOKEN_PLAN_API_KEY" {
		t.Fatalf("VisionAPIKeyEnv = %q", config.Model.VisionAPIKeyEnv)
	}
}

func TestNormalizeAIProcurementConfigReportsVisionKeyEnv(t *testing.T) {
	t.Setenv("MINIMAX_TOKEN_PLAN_API_KEY", "test-token-plan-key")
	config := aiProcurementConfig{}
	config.Model.VisionAPIKeyEnv = "MINIMAX_TOKEN_PLAN_API_KEY"

	normalizeAIProcurementConfig(&config)

	if !config.Model.VisionAPIKeyConfigured {
		t.Fatalf("VisionAPIKeyConfigured = false")
	}
}
