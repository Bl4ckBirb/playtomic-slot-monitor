package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// TimeFormat is the standard time format used by Playtomic API
const TimeFormat = "2006-01-02T15:04:05"

// Time is an API timestamp. The API sends wall-clock times with no zone, so a
// parsed value reads as UTC and only the tenant's Address.Timezone turns it
// into a real instant.
type Time struct {
	time.Time
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

	for _, layout := range []string{TimeFormat, time.RFC3339} {
		if parsed, err := time.Parse(layout, s); err == nil {
			t.Time = parsed
			return nil
		}
	}
	return fmt.Errorf("time %q matches neither %q nor RFC 3339", s, TimeFormat)
}

func (t Time) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + t.Format(TimeFormat) + `"`), nil
}

func (t Time) String() string {
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
