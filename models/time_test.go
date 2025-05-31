package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	timeStr := "2023-05-15T14:30:00"
	parsed, err := ParseTime(timeStr)
	if err != nil {
		t.Errorf("Expected successful parsing, got error: %v", err)
	}

	expected := time.Date(2023, 5, 15, 14, 30, 0, 0, time.UTC)
	if !parsed.Equal(expected) {
		t.Errorf("Expected %v, got %v", expected, parsed)
	}

	_, err = ParseTime("invalid-date")
	if err == nil {
		t.Errorf("Expected error for invalid date format, got no error")
	}
}

func TestFormatTime(t *testing.T) {
	testTime := time.Date(2023, 5, 15, 14, 30, 0, 0, time.UTC)
	formatted := FormatTime(testTime)

	expected := "2023-05-15T14:30:00"
	if formatted != expected {
		t.Errorf("Expected %q, got %q", expected, formatted)
	}
}

func TestTimeUnmarshal(t *testing.T) {
	tests := []struct {
		json string
		want string
	}{
		{`"2023-05-15T14:30:00"`, "2023-05-15T14:30:00"},
		{`"2023-05-15T14:30:00Z"`, "2023-05-15T14:30:00Z"},
		{`"2023-05-15T16:30:00+02:00"`, "2023-05-15T16:30:00+02:00"},
		{`"2025-05-24"`, "2025-05-24"},
		{`null`, ""},
		{`""`, ""},
	}

	for _, tt := range tests {
		var got Time
		if err := json.Unmarshal([]byte(tt.json), &got); err != nil {
			t.Errorf("Unmarshal(%s): %v", tt.json, err)
			continue
		}
		if got.String() != tt.want {
			t.Errorf("Unmarshal(%s) = %q, want %q", tt.json, got, tt.want)
		}
	}

	var bad Time
	if err := json.Unmarshal([]byte(`"15/05/2023"`), &bad); err == nil {
		t.Error("expected an error for an unrecognised layout")
	}
}

func TestTimeRoundTrip(t *testing.T) {
	type wrapper struct {
		At Time `json:"at"`
	}

	var v wrapper
	if err := json.Unmarshal([]byte(`{"at":"2023-05-15T14:30:00"}`), &v); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(b) != `{"at":"2023-05-15T14:30:00"}` {
		t.Errorf("round trip = %s", b)
	}

	if b, _ := json.Marshal(wrapper{}); string(b) != `{"at":null}` {
		t.Errorf("zero marshalled as %s, want null", b)
	}
}

// A date must not come back as a timestamp, and an offset must survive.
func TestTimeMarshalsInTheLayoutItArrivedIn(t *testing.T) {
	for _, want := range []string{
		`"2025-05-24"`,
		`"2025-05-24T08:00:00"`,
		`"2025-05-24T08:00:00+02:00"`,
		`"2025-05-24T08:00:00.123+02:00"`,
		`"2025-05-24T08:00:00.5Z"`,
	} {
		var got Time
		if err := json.Unmarshal([]byte(want), &got); err != nil {
			t.Errorf("Unmarshal(%s): %v", want, err)
			continue
		}

		b, err := json.Marshal(got)
		if err != nil {
			t.Errorf("Marshal(%s): %v", want, err)
			continue
		}
		if string(b) != want {
			t.Errorf("round trip of %s gave %s", want, b)
		}
	}
}
