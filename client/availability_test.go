package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

const availabilityJSON = `[{
  "resource_id": "court-1",
  "start_date": "2025-05-24",
  "slots": [
    {"start_time": "08:00:00", "duration": 60, "price": "24 EUR"},
    {"start_time": "09:30:00", "duration": 90, "price": "36 EUR"}
  ]
}]`

func TestGetAvailability(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/availability" {
			t.Errorf("path = %s, want /v1/availability", r.URL.Path)
		}
		q := r.URL.Query()
		if got := q.Get("tenant_id"); got != "tenant-123" {
			t.Errorf("tenant_id = %q", got)
		}
		if got := q.Get("local_start_min"); got != "2025-05-24T00:00:00" {
			t.Errorf("local_start_min = %q", got)
		}
		if got := q.Get("local_start_max"); got != "2025-05-24T23:59:59" {
			t.Errorf("local_start_max = %q", got)
		}
		w.Write([]byte(availabilityJSON))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	got, err := c.GetAvailability(context.Background(), &models.AvailabilityParams{
		TenantIDs: []string{"tenant-123"},
		SportID:   "PADEL",
		From:      time.Date(2025, 5, 24, 0, 0, 0, 0, time.UTC),
		To:        time.Date(2025, 5, 24, 23, 59, 59, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("GetAvailability: %v", err)
	}
	if len(got) != 1 || got[0].ResourceID != "court-1" {
		t.Fatalf("availability = %+v", got)
	}

	// A date-only value has to parse, which is why Time carries DateFormat.
	if want := time.Date(2025, 5, 24, 0, 0, 0, 0, time.UTC); !got[0].StartDate.Equal(want) {
		t.Errorf("StartDate = %v, want %v", got[0].StartDate, want)
	}
	if len(got[0].Slots) != 2 {
		t.Fatalf("got %d slots, want 2", len(got[0].Slots))
	}
	if s := got[0].Slots[1]; s.StartTime != "09:30:00" || s.Duration != 90 || s.Price != "36 EUR" {
		t.Errorf("slot = %+v", s)
	}
}
