package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// TimeFormat is the standard time format used by Playtomic API
const TimeFormat = "2006-01-02T15:04:05"

// DateFormat is the layout for the endpoints that carry a day rather than an
// instant, such as availability.
const DateFormat = "2006-01-02"

// Time is an API timestamp. The API sends wall-clock times with no zone, so a
// parsed value reads as UTC and only the tenant's Address.Timezone turns it
// into a real instant. A date-only value parses to midnight.
type Time struct {
	time.Time

	// raw is exactly what arrived. Marshalling hands the same bytes back, so a
	// date does not become a timestamp, an offset is not dropped, and
	// fractional seconds survive a round trip.
	raw string
}

func (t *Time) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	// null and "" both arrive here as the empty string.
	if s == "" {
		*t = Time{}
		return nil
	}

	for _, layout := range []string{TimeFormat, time.RFC3339, DateFormat} {
		if parsed, err := time.Parse(layout, s); err == nil {
			t.Time, t.raw = parsed, s
			return nil
		}
	}
	return fmt.Errorf("time %q matches none of %q, RFC 3339 or %q", s, TimeFormat, DateFormat)
}

func (t Time) MarshalJSON() ([]byte, error) {
	if t.raw != "" {
		return json.Marshal(t.raw)
	}
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Format(TimeFormat))
}

func (t Time) String() string {
	if t.raw != "" {
		return t.raw
	}
	if t.IsZero() {
		return ""
	}
	return t.Format(TimeFormat)
}

// ParseTime parses a time string in Playtomic's format
func ParseTime(timeStr string) (time.Time, error) {
	return time.Parse(TimeFormat, timeStr)
}

// FormatTime formats a time.Time into Playtomic's expected format
func FormatTime(t time.Time) string {
	return t.Format(TimeFormat)
}
