package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

const resourcesJSON = `[
  {"resource_id": "court-1", "name": "Padel 1", "properties": {"resource_type": "indoor", "resource_size": "double", "resource_feature": "panoramic"}},
  {"resource_id": "court-2", "name": "Padel 2", "properties": {"resource_type": "outdoor", "resource_size": "single", "resource_feature": ""}}
]`

func TestGetResources(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/tenants/tenant-123/resources" {
			t.Errorf("path = %s, want /v1/tenants/tenant-123/resources", r.URL.Path)
		}
		if _, err := w.Write([]byte(resourcesJSON)); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	resources, err := c.GetResources(context.Background(), "tenant-123")
	if err != nil {
		t.Fatalf("GetResources: %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("resources = %+v", resources)
	}
	if resources[0].ResourceID != "court-1" || resources[0].Name != "Padel 1" {
		t.Errorf("resource[0] = %+v", resources[0])
	}
	if !resources[0].IsIndoor() {
		t.Errorf("resource[0] should be indoor: %+v", resources[0].Properties)
	}
	if resources[1].IsIndoor() {
		t.Errorf("resource[1] should be outdoor: %+v", resources[1].Properties)
	}
	if resources[1].Properties.ResourceSize != "single" {
		t.Errorf("resource[1] size = %q", resources[1].Properties.ResourceSize)
	}
}

func TestGetResourcesMissingID(t *testing.T) {
	c := NewClient()
	if _, err := c.GetResources(context.Background(), ""); !errors.Is(err, ErrMissingID) {
		t.Fatalf("err = %v, want ErrMissingID", err)
	}
}
