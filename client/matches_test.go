package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

const matchesJSON = `[{
  "match_id": "match-123",
  "sport_id": "PADEL",
  "start_date": "2023-01-01T10:00:00",
  "end_date": "2023-01-01T12:00:00",
  "created_at": "2022-12-20T09:30:00",
  "match_type": "COMPETITIVE",
  "min_players_per_team": 2,
  "max_players_per_team": 2,
  "min_level": 2.5,
  "max_level": 4.0,
  "gender": "MIXED",
  "price": "30 EUR",
  "teams": [
    {"team_id": "0", "min_players": 2, "max_players": 2,
     "players": [{"user_id": "user-123", "name": "Player One", "level_value": 3.5}]},
    {"team_id": "1", "min_players": 2, "max_players": 2,
     "players": [{"user_id": "user-456", "name": "Player Two", "level_value": 3.0}]}
  ],
  "tenant": {"tenant_id": "test-tenant-id-1", "tenant_name": "Test Club"}
}]`

func TestSearchMatches(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/matches" {
			t.Errorf("path = %s, want /v1/matches", r.URL.Path)
		}
		q := r.URL.Query()
		if got := q.Get("tenant_id"); got != "test-tenant-id-1,test-tenant-id-2" {
			t.Errorf("tenant_id = %q", got)
		}
		if got := q.Get("has_players"); got != "true" {
			t.Errorf("has_players = %q", got)
		}
		if got := q.Get("sport_id"); got != "PADEL" {
			t.Errorf("sport_id = %q", got)
		}
		w.Write([]byte(matchesJSON))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	matches, err := c.SearchMatches(context.Background(), &models.SearchMatchesParams{
		TenantIDs:  []string{"test-tenant-id-1", "test-tenant-id-2"},
		HasPlayers: true,
		SportID:    "PADEL",
	})
	if err != nil {
		t.Fatalf("SearchMatches: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("got %d matches, want 1", len(matches))
	}

	got := matches[0]
	if got.MatchID != "match-123" || got.MatchType != "COMPETITIVE" {
		t.Errorf("match = %+v", got)
	}
	if got.StartDate.String() != "2023-01-01T10:00:00" || got.CreatedAt.String() != "2022-12-20T09:30:00" {
		t.Errorf("StartDate = %q, CreatedAt = %q", got.StartDate, got.CreatedAt)
	}
	if len(got.Teams) != 2 {
		t.Fatalf("got %d teams, want 2", len(got.Teams))
	}
	if got.Teams[0].Players[0].Name != "Player One" || got.Teams[1].Players[0].UserID != "user-456" {
		t.Errorf("teams = %+v", got.Teams)
	}
	if got.MinLevel != 2.5 || got.MaxLevel != 4.0 {
		t.Errorf("levels = %v..%v", got.MinLevel, got.MaxLevel)
	}
}
