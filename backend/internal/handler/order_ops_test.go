package handler

import (
	"database/sql"
	"testing"
)

func TestRestoreStatusFromFreezePayloadRejectsMissingPayload(t *testing.T) {
	if got, ok := restoreStatusFromFreezePayload(sql.NullString{}); ok || got != 0 {
		t.Fatalf("got (%d, %v), want (0, false)", got, ok)
	}
}

func TestRestoreStatusFromFreezePayloadAcceptsDraftStatus(t *testing.T) {
	got, ok := restoreStatusFromFreezePayload(sql.NullString{String: `{"from_status":0}`, Valid: true})
	if !ok || got != 0 {
		t.Fatalf("got (%d, %v), want (0, true)", got, ok)
	}
}

func TestRestoreStatusFromFreezePayloadRejectsFrozenStatus(t *testing.T) {
	if got, ok := restoreStatusFromFreezePayload(sql.NullString{String: `{"from_status":10}`, Valid: true}); ok || got != 0 {
		t.Fatalf("got (%d, %v), want (0, false)", got, ok)
	}
}

func TestBuildOrderTimelineUsesNewFulfillmentFlow(t *testing.T) {
	timeline := buildOrderTimeline(orderItem{Status: 13})
	assertTimelineStatus(t, timeline, "warehouse_select", "completed")
	assertTimelineStatus(t, timeline, "warehouse_check", "current")
	assertTimelineStatus(t, timeline, "payment_finance", "pending")
}

func TestBuildOrderTimelineDoesNotMarkShippingCompleteForWarehouseReselect(t *testing.T) {
	timeline := buildOrderTimeline(orderItem{Status: 24})
	assertTimelineStatus(t, timeline, "warehouse_check", "current")
	assertTimelineStatus(t, timeline, "shipping", "pending")
	assertTimelineStatus(t, timeline, "receipt", "pending")
}

func TestCanFreezeOrderStatusRejectsTerminalStatuses(t *testing.T) {
	for _, status := range []int{8, 9, 21} {
		if canFreezeOrderStatus(status) {
			t.Fatalf("status %d should not be freezable", status)
		}
	}
}

func TestCanFreezeOrderStatusAllowsActiveStatuses(t *testing.T) {
	for _, status := range []int{1, 13, 22, 24} {
		if !canFreezeOrderStatus(status) {
			t.Fatalf("status %d should be freezable", status)
		}
	}
}

func TestBuildOrderTimelineMarksLogisticsNoCapacityAsTerminal(t *testing.T) {
	timeline := buildOrderTimeline(orderItem{Status: 25})
	assertTimelineStatus(t, timeline, "payment_finance", "completed")
	assertTimelineStatus(t, timeline, "logistics_bidding", "failed")
	assertTimelineStatus(t, timeline, "logistics_confirm", "skipped")
	assertTimelineStatus(t, timeline, "shipping", "skipped")
}

func TestBuildOrderTerminationReasonForLogisticsNoCapacity(t *testing.T) {
	reason := buildOrderTerminationReason(orderItem{
		Status: 25,
		Remark: map[string]any{"logistics_no_capacity_reason": "暂无可承接车队"},
	})
	if reason == nil || reason.Code != "logistics_no_capacity" || reason.Party != "logistics" || reason.Reason != "暂无可承接车队" {
		t.Fatalf("unexpected termination reason: %#v", reason)
	}
}

func TestBuildOrderTerminationReasonForFinancingReject(t *testing.T) {
	reason := buildOrderTerminationReason(orderItem{
		Status:        22,
		FinancingInfo: map[string]any{"risk_result": "rejected", "riskRemark": "额度不足"},
	})
	if reason == nil || reason.Code != "financing_risk_rejected" || reason.Party != "funding" || reason.Reason != "额度不足" {
		t.Fatalf("unexpected termination reason: %#v", reason)
	}
}

func TestBuildOrderTerminationReasonForBuyerCancelEvent(t *testing.T) {
	reason := buildOrderTerminationReason(orderItem{
		Status: 9,
		Events: []orderEventItem{{
			EventType: "buyer_cancelled",
			Payload:   map[string]any{"reason": "采购计划取消"},
		}},
	})
	if reason == nil || reason.Code != "buyer_cancelled" || reason.Party != "buyer" || reason.Reason != "采购计划取消" {
		t.Fatalf("unexpected termination reason: %#v", reason)
	}
}

func TestBuildOrderTerminationReasonForGenericCancel(t *testing.T) {
	reason := buildOrderTerminationReason(orderItem{Status: 9})
	if reason == nil || reason.Code != "cancelled" || reason.Label != "订单取消-已终止" {
		t.Fatalf("unexpected termination reason: %#v", reason)
	}
}

func assertTimelineStatus(t *testing.T, timeline []orderTimelineItem, key string, want string) {
	t.Helper()
	for _, item := range timeline {
		if item.Key == key {
			if item.Status != want {
				t.Fatalf("%s status=%q, want %q", key, item.Status, want)
			}
			return
		}
	}
	t.Fatalf("timeline key %q not found", key)
}
