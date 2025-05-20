package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

const lessonsJSON = `[{
  "tournament_id": "lesson-123",
  "tournament_name": "Test Lesson",
  "start_date": "2023-01-01T10:00:00",
  "end_date": "2023-01-01T12:00:00",
  "type": "CLASS",
  "min_players": 2,
  "max_players": 4,
  "reservation_ids": null,
  "registered_players": [
    {"user_id": "user-123", "full_name": "John Doe", "level_value": 3.5, "registration_price": "30.00"}
  ],
  "tournament_visibility": "PUBLIC",
  "tournament_status": "REGISTRATION_OPEN",
  "available_places": 2,
  "tenant": {"tenant_id": "test-tenant-id", "tenant_name": "Test Club"}
}]`

func TestSearchLessons(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/lessons" {
			t.Errorf("path = %s, want /v1/lessons", r.URL.Path)
		}
		q := r.URL.Query()
		if got := q.Get("tenant_id"); got != "test-tenant-id" {
			t.Errorf("tenant_id = %q", got)
		}
		if got := q.Get("tournament_visibility"); got != "PUBLIC" {
			t.Errorf("tournament_visibility = %q", got)
		}
		w.Write([]byte(lessonsJSON))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	lessons, err := c.SearchLessons(context.Background(), &models.SearchLessonsParams{
		TenantID:             "test-tenant-id",
		TournamentVisibility: "PUBLIC",
	})
	if err != nil {
		t.Fatalf("SearchLessons: %v", err)
	}
	if len(lessons) != 1 {
		t.Fatalf("got %d lessons, want 1", len(lessons))
	}

	got := lessons[0]
	if got.TournamentID != "lesson-123" || got.TournamentName != "Test Lesson" {
		t.Errorf("lesson = %+v", got)
	}
	if got.StartDate.String() != "2023-01-01T10:00:00" {
		t.Errorf("StartDate = %q", got.StartDate)
	}
	// The API sends null here as often as it sends a list.
	if got.ReservationIDs != nil {
		t.Errorf("ReservationIDs = %v, want nil", got.ReservationIDs)
	}
	if len(got.RegisteredPlayers) != 1 || got.RegisteredPlayers[0].FullName != "John Doe" {
		t.Errorf("RegisteredPlayers = %+v", got.RegisteredPlayers)
	}
	if got.Tenant.TenantName != "Test Club" || got.AvailablePlaces != 2 {
		t.Errorf("tenant = %+v, places = %d", got.Tenant, got.AvailablePlaces)
	}
}
