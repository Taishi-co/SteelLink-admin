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
