package models

import (
	"net/url"
	"testing"
	"time"
)

func TestSearchMatchesParamsToURLValues(t *testing.T) {
	tests := []struct {
		name     string
		params   SearchMatchesParams
		expected url.Values
	}{
		{
			name:   "Empty params",
			params: SearchMatchesParams{},
			expected: url.Values{
				"page": []string{"0"},
			},
		},
		{
			name: "Complete params",
			params: SearchMatchesParams{
				Sort:          "start_date,DESC",
				HasPlayers:    true,
				SportID:       "PADEL",
				TenantIDs:     []string{"tenant-123", "tenant-456"},
				Visibility:    "VISIBLE",
				FromStartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				Size:          50,
				Page:          2,
			},
			expected: url.Values{
				"sort":            []string{"start_date,DESC"},
				"has_players":     []string{"true"},
				"sport_id":        []string{"PADEL"},
				"tenant_id":       []string{"tenant-123,tenant-456"},
				"visibility":      []string{"VISIBLE"},
				"from_start_date": []string{"2023-01-01T00:00:00"},
				"size":            []string{"50"},
				"page":            []string{"2"},
			},
		},
		{
			name: "Has players only",
			params: SearchMatchesParams{
				HasPlayers: true,
			},
			expected: url.Values{
				"page":        []string{"0"},
				"has_players": []string{"true"},
			},
		},
		{
			name: "Whitespace handling",
			params: SearchMatchesParams{
				Sort:       "  start_date,DESC  ",
				SportID:    "  PADEL  ",
				Visibility: "  VISIBLE  ",
			},
			expected: url.Values{
				"page":       []string{"0"},
				"sort":       []string{"start_date,DESC"},
				"sport_id":   []string{"PADEL"},
				"visibility": []string{"VISIBLE"},
			},
		},
		{
			name: "Single tenant ID",
			params: SearchMatchesParams{
				TenantIDs: []string{"tenant-123"},
			},
			expected: url.Values{
				"page":      []string{"0"},
				"tenant_id": []string{"tenant-123"},
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

func TestSearchMatchesParamsDateAndPlayerFilters(t *testing.T) {
	v := (&SearchMatchesParams{
		ToStartDate:  time.Date(2025, 6, 8, 22, 59, 59, 0, time.UTC),
		ToCreatedAt:  time.Date(2025, 6, 1, 9, 0, 0, 0, time.UTC),
		MatchStatus:  "PENDING,PLAYED",
		UserID:       "me",
		PlayerUserID: "123",
	}).ToURLValues()

	want := map[string]string{
		"to_start_date":  "2025-06-08T22:59:59",
		"to_created_at":  "2025-06-01T09:00:00",
		"match_status":   "PENDING,PLAYED",
		"user_id":        "me",
		"player_user_id": "123",
	}
	for k, w := range want {
		if got := v.Get(k); got != w {
			t.Errorf("%s = %q, want %q", k, got, w)
		}
	}
}
