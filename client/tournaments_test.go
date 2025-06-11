package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

const tournamentJSON = `{
  "tournament_id": "tour-1",
  "name": "Summer Open",
  "sport_id": "PADEL",
  "start_date": "2025-06-14T09:00:00",
  "end_date": "2025-06-15T20:00:00",
  "type": "QUICK",
  "visibility": "PUBLIC",
  "status": "PENDING",
  "available_places": 8,
  "tenant": {"tenant_id": "t-1", "tenant_name": "Club"}
}`

func TestSearchTournaments(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/tournaments" {
			t.Errorf("path = %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("sport_id") != "PADEL" || q.Get("registration_status") != "OPEN" {
			t.Errorf("query = %s", r.URL.RawQuery)
		}
		if q.Get("available_places") != "true" {
			t.Errorf("available_places = %q", q.Get("available_places"))
		}
		w.Write([]byte("[" + tournamentJSON + "]"))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	tours, err := c.SearchTournaments(context.Background(), &models.SearchTournamentsParams{
		SportID:            "PADEL",
		RegistrationStatus: "OPEN",
		AvailablePlaces:    true,
	})
	if err != nil {
		t.Fatalf("SearchTournaments: %v", err)
	}
	if len(tours) != 1 || tours[0].Name != "Summer Open" || tours[0].AvailablePlaces != 8 {
		t.Errorf("tournaments = %+v", tours)
	}
}

func TestAllTournamentsWalksPages(t *testing.T) {
	var reqs atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqs.Add(1)
		if r.URL.Query().Get("page") == "0" {
			w.Write([]byte("[" + tournamentJSON + "]"))
		} else {
			w.Write([]byte("[]"))
		}
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	var names []string
	for tour, err := range c.AllTournaments(context.Background(), &models.SearchTournamentsParams{Size: 1}) {
		if err != nil {
			t.Fatalf("AllTournaments: %v", err)
		}
		names = append(names, tour.Name)
	}
	if strings.Join(names, ",") != "Summer Open" {
		t.Errorf("names = %v", names)
	}
	if reqs.Load() != 2 {
		t.Errorf("%d requests, want 2", reqs.Load())
	}
}

func TestGetTournament(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// EscapedPath, not Path: a slash in the ID must arrive percent-encoded,
		// which is what pins url.PathEscape against a raw concatenation.
		if got := r.URL.EscapedPath(); got != "/v2/tournaments/tour%2F1" {
			t.Errorf("escaped path = %s, want the id percent-encoded", got)
		}
		w.Write([]byte(tournamentJSON))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	tour, err := c.GetTournament(context.Background(), "tour/1")
	if err != nil {
		t.Fatalf("GetTournament: %v", err)
	}
	if tour.TournamentID != "tour-1" {
		t.Errorf("tournament = %+v", tour)
	}

	if _, err := c.GetTournament(context.Background(), ""); !errors.Is(err, ErrMissingID) {
		t.Errorf("empty id gave %v, want ErrMissingID", err)
	}
}

func TestGetTournamentNullResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`null`))
	}))
	defer srv.Close()

	if _, err := NewClient(WithBaseURL(srv.URL)).GetTournament(context.Background(), "tour-1"); err == nil {
		t.Fatal("a null body should be an error, not a nil tournament")
	}
}
