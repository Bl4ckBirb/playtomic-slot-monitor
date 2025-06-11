package models

import (
	"testing"
	"time"
)

func TestSearchTournamentsParams(t *testing.T) {
	v := (&SearchTournamentsParams{
		SportID:            "PADEL",
		RegistrationStatus: "OPEN",
		AvailablePlaces:    true,
		FromStartDate:      time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		ToStartDate:        time.Date(2025, 6, 30, 23, 0, 0, 0, time.UTC),
	}).ToURLValues()

	if v.Get("sport_id") != "PADEL" || v.Get("registration_status") != "OPEN" {
		t.Errorf("tournaments params = %v", v)
	}
	if v.Get("available_places") != "true" {
		t.Errorf("available_places = %q", v.Get("available_places"))
	}
	if v.Get("from_start_date") != "2025-06-01T00:00:00" || v.Get("to_start_date") != "2025-06-30T23:00:00" {
		t.Errorf("date window = %q/%q", v.Get("from_start_date"), v.Get("to_start_date"))
	}

	empty := (&SearchTournamentsParams{}).ToURLValues()
	if empty.Has("from_start_date") || empty.Has("to_start_date") {
		t.Errorf("zero times leaked: %v", empty)
	}
}

func TestSearchTournamentsParamsNilReceiver(t *testing.T) {
	if (*SearchTournamentsParams)(nil).ToURLValues() == nil {
		t.Error("nil receiver returned nil")
	}
}
