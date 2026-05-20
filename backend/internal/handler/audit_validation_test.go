package handler

import (
	"database/sql"
	"testing"
	"time"
)

func TestEnterpriseAuditDecisionOnlyAllowsApproveOrReject(t *testing.T) {
	for _, status := range []int{1, 2} {
		if !isEnterpriseAuditDecision(status) {
			t.Fatalf("status %d should be valid", status)
		}
	}
	for _, status := range []int{0, 3, 4} {
		if isEnterpriseAuditDecision(status) {
			t.Fatalf("status %d should be invalid", status)
		}
	}
}

func TestNormalizeProductAuditDecision(t *testing.T) {
	status, logStatus, ok := normalizeProductAuditDecision(1)
	if !ok || status != 2 || logStatus != "approved" {
		t.Fatalf("approve got (%d, %q, %v)", status, logStatus, ok)
	}

	status, logStatus, ok = normalizeProductAuditDecision(2)
	if !ok || status != 4 || logStatus != "rejected" {
		t.Fatalf("reject got (%d, %q, %v)", status, logStatus, ok)
	}

	status, logStatus, ok = normalizeProductAuditDecision(0)
	if ok || status != 0 || logStatus != "" {
		t.Fatalf("pending got (%d, %q, %v), want invalid", status, logStatus, ok)
	}
}

func TestRejectReasonProvided(t *testing.T) {
	if rejectReasonProvided(2, 2, "  ") {
		t.Fatalf("blank reject reason should be invalid")
	}
	if !rejectReasonProvided(2, 2, "资料不完整") {
		t.Fatalf("non-blank reject reason should be valid")
	}
	if !rejectReasonProvided(1, 2, "") {
		t.Fatalf("approve should not require reject reason")
	}
}

func TestBuildProductAuditSnapshotCapturesCoreFields(t *testing.T) {
	now := time.Date(2026, 5, 20, 12, 30, 0, 0, time.UTC)
	item := productItem{
		ID:                 11,
		Name:               "热轧板卷",
		EnterpriseID:       22,
		Category:           "板材",
		SubCategory:        "热轧",
		SteelMill:          "宝钢",
		GoodsType:          "spot",
		WarehouseServiceID: 33,
		Spec:               "5.5*1500*C",
		Material:           "Q235B",
		Province:           "上海",
		City:               "上海",
		District:           "宝山",
		BundleNo:           "B-001",
		ProductNo:          "P-001",
		Unit:               "吨",
		Price:              3900.5,
		Stock:              80,
		MinOrderQty:        5,
		DeliveryDays:       2,
		QuoteValidity:      24,
		Description:        "现货库存",
		Images:             []string{"https://example.com/p.png"},
		Status:             1,
		AuditRemark:        "待核对",
		AuditorID:          9,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	snapshot := buildProductAuditSnapshot(item)

	if snapshot["name"] != "热轧板卷" || snapshot["productNo"] != "P-001" {
		t.Fatalf("snapshot missed product identity: %#v", snapshot)
	}
	if snapshot["warehouseRegion"] != "上海 / 上海 / 宝山" {
		t.Fatalf("warehouseRegion = %#v", snapshot["warehouseRegion"])
	}
	if snapshot["status"] != 1 || snapshot["auditorId"] != int64(9) {
		t.Fatalf("snapshot missed audit fields: %#v", snapshot)
	}
}

func TestProductAuditSnapshotJSONRoundTrip(t *testing.T) {
	snapshot := buildProductAuditSnapshot(productItem{
		ID:           7,
		Name:         "螺纹钢",
		EnterpriseID: 3,
		ProductNo:    "P-007",
		Status:       2,
	})
	raw := productAuditSnapshotJSON(snapshot)
	if raw == "" {
		t.Fatal("snapshot json should not be empty")
	}

	parsed := parseProductAuditSnapshot(sql.NullString{String: raw, Valid: true})
	if parsed["name"] != "螺纹钢" || parsed["productNo"] != "P-007" {
		t.Fatalf("parsed snapshot = %#v", parsed)
	}
}
