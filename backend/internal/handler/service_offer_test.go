package handler

import "testing"

func TestNormalizeServiceOfferAuditStatusUsesExplicitAction(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		auditAction string
		want        int
	}{
		{name: "approve action writes approved offer status", status: 3, auditAction: "approve", want: 2},
		{name: "reject action writes rejected offer status", status: 2, auditAction: "reject", want: 3},
		{name: "action ignores whitespace and case", status: 3, auditAction: " Approve ", want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := normalizeServiceOfferAuditStatus(tt.status, tt.auditAction)
			if !ok {
				t.Fatalf("expected valid status")
			}
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNormalizeServiceOfferAuditStatusAcceptsCanonicalStatusValues(t *testing.T) {
	tests := []struct {
		status int
		want   int
	}{
		{status: 1, want: 2},
		{status: 2, want: 2},
		{status: 3, want: 3},
	}

	for _, tt := range tests {
		got, ok := normalizeServiceOfferAuditStatus(tt.status, "")
		if !ok {
			t.Fatalf("expected status %d to be valid", tt.status)
		}
		if got != tt.want {
			t.Fatalf("status %d got %d, want %d", tt.status, got, tt.want)
		}
	}
}

func TestNormalizeServiceOfferAuditStatusRejectsUnknownStatus(t *testing.T) {
	if got, ok := normalizeServiceOfferAuditStatus(4, ""); ok || got != 0 {
		t.Fatalf("got (%d, %v), want (0, false)", got, ok)
	}
}
