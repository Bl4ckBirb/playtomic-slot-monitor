package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

const classesJSON = `[{
  "type": "COURSE",
  "academy_class_id": "class-123",
  "sport_id": "PADEL",
  "start_date": "2023-01-01T10:00:00",
  "end_date": "2023-01-01T12:00:00",
  "course_summary": {"course_id": "course-456", "name": "Test Class", "min_players": 2, "max_players": 4},
  "resource": {"id": "resource-789", "name": "Court 1"},
  "tenant": {"tenant_id": "test-tenant-id-1", "tenant_name": "Test Club"}
}]`

func TestSearchClasses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/classes" {
			t.Errorf("path = %s, want /v1/classes", r.URL.Path)
		}
		q := r.URL.Query()
		if got := q.Get("tenant_id"); got != "test-tenant-id-1,test-tenant-id-2" {
			t.Errorf("tenant_id = %q", got)
		}
		if got := q.Get("include_summary"); got != "true" {
			t.Errorf("include_summary = %q", got)
		}
		w.Write([]byte(classesJSON))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	classes, err := c.SearchClasses(context.Background(), &models.SearchClassesParams{
		TenantIDs:      []string{"test-tenant-id-1", "test-tenant-id-2"},
		IncludeSummary: true,
	})
	if err != nil {
		t.Fatalf("SearchClasses: %v", err)
	}
	if len(classes) != 1 {
		t.Fatalf("got %d classes, want 1", len(classes))
	}

	got := classes[0]
	if got.AcademyClassID != "class-123" || got.Type != "COURSE" {
		t.Errorf("class = %+v", got)
	}
	if want := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC); !got.StartDate.Equal(want) {
		t.Errorf("StartDate = %v, want %v", got.StartDate, want)
	}
	if got.CourseSummary == nil || got.CourseSummary.Name != "Test Class" {
		t.Errorf("CourseSummary = %+v", got.CourseSummary)
	}
	if got.Resource.Name != "Court 1" || got.Tenant.TenantName != "Test Club" {
		t.Errorf("resource = %+v, tenant = %+v", got.Resource, got.Tenant)
	}
}
