package handler

import (
	"strings"
	"testing"
	"time"
)

func TestAIProcurementJSONHelpers(t *testing.T) {
	if got := aiProcurementJSONArrayLength(`[{"id":1},{"id":2}]`); got != 2 {
		t.Fatalf("array length=%d, want 2", got)
	}
	if got := aiProcurementOrderCount(`{"order_id":12}`); got != 1 {
		t.Fatalf("single order count=%d, want 1", got)
	}
	if got := aiProcurementOrderCount(`{"orders":[{"order_id":1},{"order_id":2}]}`); got != 2 {
		t.Fatalf("order list count=%d, want 2", got)
	}
	if got := aiProcurementJSONNumber(`{"confidence":0.62}`, "confidence"); got != 0.62 {
		t.Fatalf("confidence=%v, want 0.62", got)
	}
}

func TestAIProcurementInterventionDetectsReviewAndLowConfidence(t *testing.T) {
	cfg := defaultAIProcurementConfig()
	cfg.Model.ConfidenceThreshold = 0.8

	needs, reason := aiProcurementIntervention(aiProcurementMonitorItem{
		Status:       "review_submitted",
		CurrentStage: "review_submission",
		UpdatedAt:    time.Now(),
	}, cfg, time.Now().Add(-time.Hour))
	if !needs || !strings.Contains(reason, "人工复核") {
		t.Fatalf("review intervention=(%v, %q), want manual review", needs, reason)
	}

	needs, reason = aiProcurementIntervention(aiProcurementMonitorItem{
		Status:          "parsed",
		CurrentStage:    "instruction_parsing",
		ParseConfidence: 0.5,
		UpdatedAt:       time.Now(),
	}, cfg, time.Now().Add(-time.Hour))
	if !needs || !strings.Contains(reason, "置信度") {
		t.Fatalf("low confidence intervention=(%v, %q), want confidence reason", needs, reason)
	}
}

func TestAIProcurementInterventionDetectsStaleAndMissingMatches(t *testing.T) {
	cfg := defaultAIProcurementConfig()
	staleBefore := time.Now().Add(-time.Hour)

	needs, reason := aiProcurementIntervention(aiProcurementMonitorItem{
		Status:       "matched",
		CurrentStage: "product_matching",
		MatchCount:   0,
		UpdatedAt:    time.Now(),
	}, cfg, staleBefore)
	if !needs || !strings.Contains(reason, "匹配商品") {
		t.Fatalf("missing match intervention=(%v, %q), want match reason", needs, reason)
	}

	needs, reason = aiProcurementIntervention(aiProcurementMonitorItem{
		Status:       "matched",
		CurrentStage: "product_matching",
		MatchCount:   2,
		UpdatedAt:    staleBefore.Add(-time.Minute),
	}, cfg, staleBefore)
	if !needs || !strings.Contains(reason, "未推进") {
		t.Fatalf("stale intervention=(%v, %q), want stale reason", needs, reason)
	}
}

func TestAIProcurementLabels(t *testing.T) {
	if aiProcurementStageLabel("logistics_warehouse_matching") != "物流/仓储匹配" {
		t.Fatalf("unexpected stage label")
	}
	if aiProcurementStatusLabel("order_generated") != "已生成订单" {
		t.Fatalf("unexpected status label")
	}
}
