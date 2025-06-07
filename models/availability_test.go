package models

import (
	"testing"
	"time"
)

func TestAvailabilityParamsMultiTenantAndWindow(t *testing.T) {
	v := (&AvailabilityParams{
		TenantIDs: []string{"a", "b"},
		From:      time.Date(2025, 6, 8, 0, 0, 0, 0, time.UTC),
		To:        time.Date(2025, 6, 8, 23, 59, 59, 0, time.UTC),
		StartMin:  time.Date(2025, 6, 8, 9, 0, 0, 0, time.UTC),
		StartMax:  time.Date(2025, 6, 8, 22, 0, 0, 0, time.UTC),
	}).ToURLValues()

	if got := v.Get("tenant_id"); got != "a,b" {
		t.Errorf("tenant_id = %q, want the joined list", got)
	}
	if v.Get("local_start_min") != "2025-06-08T00:00:00" || v.Get("local_start_max") != "2025-06-08T23:59:59" {
		t.Errorf("local window = %q/%q", v.Get("local_start_min"), v.Get("local_start_max"))
	}
	if v.Get("start_min") != "2025-06-08T09:00:00" || v.Get("start_max") != "2025-06-08T22:00:00" {
		t.Errorf("absolute window = %q/%q", v.Get("start_min"), v.Get("start_max"))
	}
}

func TestAvailabilityParamsNilReceiver(t *testing.T) {
	if (*AvailabilityParams)(nil).ToURLValues() == nil {
		t.Error("nil receiver returned nil")
	}
}
