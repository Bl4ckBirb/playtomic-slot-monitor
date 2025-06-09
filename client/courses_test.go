package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

const courseJSON = `{
  "course_id": "course-1",
  "sport_id": "PADEL",
  "name": "Improvers",
  "description": "Tuesday nights",
  "start_date": "2025-06-03T19:00:00",
  "end_date": "2025-07-29T20:00:00",
  "visibility": "PUBLIC",
  "status": "PENDING",
  "is_canceled": false,
  "tenant": {"tenant_id": "t-1", "tenant_name": "Club"}
}`

func TestSearchCourses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/courses" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("tenant_id"); got != "t-1" {
			t.Errorf("tenant_id = %q", got)
		}
		if got := r.URL.Query().Get("visibility"); got != "PUBLIC" {
			t.Errorf("visibility = %q", got)
		}
		w.Write([]byte("[" + courseJSON + "]"))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	courses, err := c.SearchCourses(context.Background(), &models.SearchCoursesParams{
		TenantID:   "t-1",
		Visibility: "PUBLIC",
	})
	if err != nil {
		t.Fatalf("SearchCourses: %v", err)
	}
	if len(courses) != 1 || courses[0].Name != "Improvers" {
		t.Fatalf("courses = %+v", courses)
	}
	if courses[0].Tenant.TenantName != "Club" {
		t.Errorf("tenant = %+v", courses[0].Tenant)
	}
}

func TestGetCourse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/courses/course-1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write([]byte(courseJSON))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	course, err := c.GetCourse(context.Background(), "course-1")
	if err != nil {
		t.Fatalf("GetCourse: %v", err)
	}
	if course.CourseID != "course-1" {
		t.Errorf("course = %+v", course)
	}

	if _, err := c.GetCourse(context.Background(), ""); !errors.Is(err, ErrMissingID) {
		t.Errorf("empty id gave %v", err)
	}
}

func TestAllCoursesWalksPages(t *testing.T) {
	var reqs atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqs.Add(1)
		// Two full pages of one, then an empty page to stop on.
		if p := r.URL.Query().Get("page"); p == "0" || p == "1" {
			w.Write([]byte("[" + courseJSON + "]"))
		} else {
			w.Write([]byte("[]"))
		}
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	var n int
	for course, err := range c.AllCourses(context.Background(), &models.SearchCoursesParams{Size: 1}) {
		if err != nil {
			t.Fatalf("AllCourses: %v", err)
		}
		if course.CourseID != "course-1" {
			t.Errorf("course = %+v", course)
		}
		n++
	}
	if n != 2 {
		t.Errorf("collected %d courses, want 2", n)
	}
	if reqs.Load() != 3 {
		t.Errorf("%d requests, want 3 (stops on the empty page)", reqs.Load())
	}
}
