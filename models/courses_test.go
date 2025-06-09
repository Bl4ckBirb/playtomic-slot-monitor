package models

import (
	"testing"
	"time"
)

func TestSearchCoursesParams(t *testing.T) {
	v := (&SearchCoursesParams{
		TenantID:            "t-1",
		Visibility:          "PUBLIC",
		AvailablePlacesFrom: time.Date(2025, 6, 2, 9, 0, 0, 0, time.UTC),
		CourseEndsAfter:     time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC),
		Coordinate:          &Coordinate{Lat: 51.5, Lon: -0.1},
		Radius:              10000,
		Size:                10,
	}).ToURLValues()

	if v.Get("tenant_id") != "t-1" || v.Get("visibility") != "PUBLIC" {
		t.Errorf("courses params = %v", v)
	}
	if v.Get("available_places_from") != "2025-06-02T09:00:00" || v.Get("course_ends_after") != "2025-06-30T00:00:00" {
		t.Errorf("dates = %q/%q", v.Get("available_places_from"), v.Get("course_ends_after"))
	}
	if v.Get("radius") != "10000" || v.Get("size") != "10" || v.Get("page") != "0" {
		t.Errorf("radius/size/page = %q/%q/%q", v.Get("radius"), v.Get("size"), v.Get("page"))
	}

	// Zero times must be omitted, not sent as an empty or epoch value.
	empty := (&SearchCoursesParams{}).ToURLValues()
	if empty.Has("available_places_from") || empty.Has("course_ends_after") {
		t.Errorf("zero times leaked: %v", empty)
	}
}

func TestSearchCoursesParamsNilReceiver(t *testing.T) {
	if (*SearchCoursesParams)(nil).ToURLValues() == nil {
		t.Error("nil receiver returned nil")
	}
}
