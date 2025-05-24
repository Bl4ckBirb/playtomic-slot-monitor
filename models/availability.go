package models

import (
	"net/url"
	"strings"
	"time"
)

// Availability is one resource's open slots for a day.
type Availability struct {
	ResourceID string `json:"resource_id"`
	StartDate  Time   `json:"start_date"`
	Slots      []Slot `json:"slots"`
}

// Slot is a bookable interval. StartTime is a wall-clock time of day on the
// parent's StartDate, and Duration is in minutes.
type Slot struct {
	StartTime string `json:"start_time"`
	Duration  int    `json:"duration"`
	Price     string `json:"price"`
}

// AvailabilityParams defines parameters for querying court availability
type AvailabilityParams struct {
	TenantID string
	SportID  string
	From     time.Time
	To       time.Time
}

// ToURLValues converts AvailabilityParams to url.Values
func (p *AvailabilityParams) ToURLValues() url.Values {
	values := url.Values{}
	if p == nil {
		return values
	}

	if s := strings.TrimSpace(p.TenantID); s != "" {
		values.Set("tenant_id", s)
	}

	if s := strings.TrimSpace(p.SportID); s != "" {
		values.Set("sport_id", s)
	}

	if !p.From.IsZero() {
		values.Set("local_start_min", FormatTime(p.From))
	}

	if !p.To.IsZero() {
		values.Set("local_start_max", FormatTime(p.To))
	}

	return values
}
