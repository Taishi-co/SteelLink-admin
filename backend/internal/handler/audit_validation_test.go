package handler

import "testing"

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
