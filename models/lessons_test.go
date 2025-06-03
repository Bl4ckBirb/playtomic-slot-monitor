package models

import (
	"net/url"
	"testing"
	"time"
)

func TestSearchLessonsParamsToURLValues(t *testing.T) {
	tests := []struct {
		name     string
		params   SearchLessonsParams
		expected url.Values
	}{
		{
			name:   "Empty params",
			params: SearchLessonsParams{},
			expected: url.Values{
				"page": []string{"0"},
			},
		},
		{
			name: "Complete params",
			params: SearchLessonsParams{
				Sort:                 "start_date,ASC",
				TenantID:             "tenant-123",
				TournamentVisibility: "PUBLIC",
				Status:               "REGISTRATION_OPEN",
				Size:                 50,
				Page:                 2,
				FromStartDate:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expected: url.Values{
				"sort":                  []string{"start_date,ASC"},
				"tenant_id":             []string{"tenant-123"},
				"tournament_visibility": []string{"PUBLIC"},
				"status":                []string{"REGISTRATION_OPEN"},
				"size":                  []string{"50"},
				"page":                  []string{"2"},
				"from_start_date":       []string{"2023-01-01T00:00:00"},
			},
		},
		{
			name: "Whitespace handling",
			params: SearchLessonsParams{
				Sort:                 "  start_date,ASC  ",
				TenantID:             "  tenant-123  ",
				TournamentVisibility: "  PUBLIC  ",
				Status:               "  REGISTRATION_OPEN  ",
			},
			expected: url.Values{
				"sort":                  []string{"start_date,ASC"},
				"tenant_id":             []string{"tenant-123"},
				"tournament_visibility": []string{"PUBLIC"},
				"status":                []string{"REGISTRATION_OPEN"},
				"page":                  []string{"0"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.params.ToURLValues()

			for k, expectedVals := range tt.expected {
				resultVals, ok := result[k]
				if !ok {
					t.Errorf("Expected key %q not found in result", k)
					continue
				}

				if len(resultVals) != len(expectedVals) {
					t.Errorf("Key %q: expected %d values, got %d", k, len(expectedVals), len(resultVals))
					continue
				}

				for i, expectedVal := range expectedVals {
					if resultVals[i] != expectedVal {
						t.Errorf("Key %q, index %d: expected %q, got %q", k, i, expectedVal, resultVals[i])
					}
				}
			}

			for k := range result {
				if _, ok := tt.expected[k]; !ok {
					t.Errorf("Unexpected key %q in result", k)
				}
			}
		})
	}
}

func TestSearchLessonsParamsSportLocationAndRange(t *testing.T) {
	v := (&SearchLessonsParams{
		ToStartDate: time.Date(2025, 6, 8, 22, 59, 59, 0, time.UTC),
		SportID:     "PADEL",
		UserID:      "me",
		Coordinate:  &Coordinate{Lat: 51.5, Lon: -0.1},
		Radius:      25000,
	}).ToURLValues()

	if v.Get("to_start_date") != "2025-06-08T22:59:59" || v.Get("sport_id") != "PADEL" || v.Get("user_id") != "me" {
		t.Errorf("lessons params = %v", v)
	}
	if v.Get("coordinate") == "" || v.Get("radius") != "25000" {
		t.Errorf("coordinate/radius = %q/%q", v.Get("coordinate"), v.Get("radius"))
	}
}
