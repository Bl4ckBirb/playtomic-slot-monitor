package models

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Class represents a class from the Playtomic API
type Class struct {
	Type             string           `json:"type"`
	AcademyClassID   string           `json:"academy_class_id"`
	SportID          string           `json:"sport_id"`
	Tenant           Tenant           `json:"tenant"`
	Resource         Resource         `json:"resource"`
	StartDate        Time             `json:"start_date"`
	EndDate          Time             `json:"end_date"`
	Coaches          []Coach          `json:"coaches"`
	RegistrationInfo RegistrationInfo `json:"registration_info"`
	CourseSummary    *CourseSummary   `json:"course_summary,omitempty"`
	AccessCode       *string          `json:"access_code"`
	Origin           string           `json:"origin"`
	IsCanceled       bool             `json:"is_canceled"`
	PrivateNotes     *string          `json:"private_notes"`
	PublicNotes      string           `json:"public_notes"`
	Status           string           `json:"status"`
	PaymentStatus    string           `json:"payment_status"`
}

// CourseSummary represents summary information about a course
type CourseSummary struct {
	CourseID   string `json:"course_id"`
	Name       string `json:"name"`
	Gender     string `json:"gender"`
	Visibility string `json:"visibility"`
	MinPlayers int    `json:"min_players"`
	MaxPlayers int    `json:"max_players"`
}

// SearchClassesParams defines parameters for searching classes
type SearchClassesParams struct {
	Sort             string
	Status           string
	Type             string
	TenantIDs        []string
	IncludeSummary   bool
	Size             int
	Page             int
	CourseVisibility string
	FromStartDate    time.Time
	Coordinate       *Coordinate
	Radius           int
}

// ToURLValues converts SearchClassesParams to url.Values
func (p *SearchClassesParams) ToURLValues() url.Values {
	values := url.Values{}
	if p == nil {
		return values
	}

	if s := strings.TrimSpace(p.Sort); s != "" {
		values.Set("sort", s)
	}

	if s := strings.TrimSpace(p.Status); s != "" {
		values.Set("status", s)
	}

	if t := strings.TrimSpace(p.Type); t != "" {
		values.Set("type", t)
	}

	if len(p.TenantIDs) > 0 {
		values.Set("tenant_id", strings.Join(p.TenantIDs, ","))
	}

	if p.IncludeSummary {
		values.Set("include_summary", "true")
	}

	if p.Size > 0 {
		values.Set("size", strconv.Itoa(p.Size))
	}

	values.Set("page", strconv.Itoa(p.Page))

	if cv := strings.TrimSpace(p.CourseVisibility); cv != "" {
		values.Set("course_visibility", cv)
	}

	if !p.FromStartDate.IsZero() {
		values.Set("from_start_date", FormatTime(p.FromStartDate))
	}

	if p.Coordinate != nil {
		values.Set("coordinate", p.Coordinate.String())

		if p.Radius > 0 {
			values.Set("radius", strconv.Itoa(p.Radius))
		}
	}

	return values
}
