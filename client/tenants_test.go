package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

const tenantJSON = `{
  "tenant_id": "tenant-123",
  "tenant_name": "Padel Club London",
  "playtomic_status": "ACTIVE",
  "address": {
    "street": "1 Court Road", "city": "London", "country_code": "GB",
    "timezone": "Europe/London",
    "coordinate": {"lat": 51.5074, "lon": -0.1278}
  }
}`

func TestSearchTenants(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/tenants" {
			t.Errorf("path = %s, want /v1/tenants", r.URL.Path)
		}
		q := r.URL.Query()
		if got := q.Get("coordinate"); got != "51.507400,-0.127800" {
			t.Errorf("coordinate = %q", got)
		}
		if got := q.Get("radius"); got != "50000" {
			t.Errorf("radius = %q", got)
		}
		if got := q.Get("playtomic_status"); got != "ACTIVE" {
			t.Errorf("playtomic_status = %q", got)
		}
		w.Write([]byte("[" + tenantJSON + "]"))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	tenants, err := c.SearchTenants(context.Background(), &models.SearchTenantsParams{
		Coordinate:      &models.Coordinate{Lat: 51.5074, Lon: -0.1278},
		Radius:          50000,
		PlaytomicStatus: "ACTIVE",
	})
	if err != nil {
		t.Fatalf("SearchTenants: %v", err)
	}
	if len(tenants) != 1 || tenants[0].TenantName != "Padel Club London" {
		t.Fatalf("tenants = %+v", tenants)
	}
	if tenants[0].Address.Timezone != "Europe/London" {
		t.Errorf("timezone = %q", tenants[0].Address.Timezone)
	}
}

func TestGetTenant(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/tenants/tenant-123" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write([]byte(tenantJSON))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	tenant, err := c.GetTenant(context.Background(), "tenant-123")
	if err != nil {
		t.Fatalf("GetTenant: %v", err)
	}
	if tenant.TenantID != "tenant-123" || tenant.Address.Coordinate.Lat != 51.5074 {
		t.Errorf("tenant = %+v", tenant)
	}

	if _, err := c.GetTenant(context.Background(), ""); !errors.Is(err, ErrMissingID) {
		t.Errorf("empty id gave %v, want ErrMissingID", err)
	}
}
