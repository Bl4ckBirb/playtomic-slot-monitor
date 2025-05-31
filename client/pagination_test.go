package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

// pagedTenants serves size-2 pages: two full, one short, then an empty one,
// which is what actually ends the walk.
func pagedTenants(t *testing.T, pages *atomic.Int32) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pages.Add(1)
		page := r.URL.Query().Get("page")

		var ids []string
		switch page {
		case "0":
			ids = []string{"a", "b"}
		case "1":
			ids = []string{"c", "d"}
		case "2":
			ids = []string{"e"}
		case "3":
			ids = nil
		default:
			t.Errorf("unexpected page %q", page)
		}

		items := make([]string, 0, len(ids))
		for _, id := range ids {
			items = append(items, fmt.Sprintf(`{"tenant_id":%q}`, id))
		}
		fmt.Fprintf(w, "[%s]", strings.Join(items, ","))
	})
}

func TestAllTenantsWalksPages(t *testing.T) {
	var requests atomic.Int32
	srv := httptest.NewServer(pagedTenants(t, &requests))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))

	var got []string
	for tenant, err := range c.AllTenants(context.Background(), &models.SearchTenantsParams{Size: 2}) {
		if err != nil {
			t.Fatalf("AllTenants: %v", err)
		}
		got = append(got, tenant.TenantID)
	}

	if want := "a b c d e"; strings.Join(got, " ") != want {
		t.Errorf("collected %v, want %s", got, want)
	}
	// Four, not three: the walk ends on an empty page rather than a short one.
	if requests.Load() != 4 {
		t.Errorf("%d requests, want 4", requests.Load())
	}
}

func TestAllTenantsStopsOnBreak(t *testing.T) {
	var requests atomic.Int32
	srv := httptest.NewServer(pagedTenants(t, &requests))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))

	var seen int
	for range c.AllTenants(context.Background(), &models.SearchTenantsParams{Size: 2}) {
		seen++
		break
	}

	if seen != 1 || requests.Load() != 1 {
		t.Errorf("saw %d items over %d requests, want 1 and 1", seen, requests.Load())
	}
}

func TestAllTenantsYieldsErrorOnce(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"nope"}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))

	var errs int
	for _, err := range c.AllTenants(context.Background(), nil) {
		if !errors.Is(err, ErrBadRequest) {
			t.Errorf("got %v, want a 400", err)
		}
		errs++
	}

	if errs != 1 {
		t.Errorf("yielded %d errors, want exactly 1", errs)
	}
}

func TestAllLeavesCallerParamsAlone(t *testing.T) {
	var requests atomic.Int32
	srv := httptest.NewServer(pagedTenants(t, &requests))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	params := &models.SearchTenantsParams{Size: 2}

	var seen int
	for range c.AllTenants(context.Background(), params) {
		seen++
	}
	if seen != 5 {
		t.Errorf("saw %d tenants, want 5", seen)
	}

	if params.Page != 0 {
		t.Errorf("caller's Page moved to %d", params.Page)
	}
}
