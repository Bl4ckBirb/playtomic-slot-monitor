package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.EscapedPath(); got != "/v1/matches/m%2F1" {
			t.Errorf("escaped path = %s, want the id percent-encoded", got)
		}
		w.Write([]byte(`{"match_id":"m-1","sport_id":"PADEL","start_date":"2025-06-08T10:00:00"}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	match, err := c.GetMatch(context.Background(), "m/1")
	if err != nil {
		t.Fatalf("GetMatch: %v", err)
	}
	if match.MatchID != "m-1" || match.StartDate.String() != "2025-06-08T10:00:00" {
		t.Errorf("match = %+v", match)
	}

	if _, err := c.GetMatch(context.Background(), ""); !errors.Is(err, ErrMissingID) {
		t.Errorf("empty id gave %v, want ErrMissingID", err)
	}
}

func TestGetMatchNullResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`null`))
	}))
	defer srv.Close()

	if _, err := NewClient(WithBaseURL(srv.URL)).GetMatch(context.Background(), "m-1"); err == nil {
		t.Fatal("a null body should be an error, not a nil match")
	}
}
