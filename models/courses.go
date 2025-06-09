package models

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Course is an academy course, a series of scheduled classes, from /v1/courses.
type Course struct {
	CourseID    string  `json:"course_id"`
	SportID     string  `json:"sport_id"`
	Tenant      Tenant  `json:"tenant"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	StartDate   Time    `json:"start_date"`
	EndDate     Time    `json:"end_date"`
	Visibility  string  `json:"visibility"`
	Status      string  `json:"status"`
	IsCanceled  bool    `json:"is_canceled"`
	AccessCode  *string `json:"access_code"`
}

// SearchCoursesParams defines parameters for searching courses.
type SearchCoursesParams struct {
	Sort                string
	TenantID            string
	Visibility          string
	AvailablePlacesFrom time.Time
	CourseEndsAfter     time.Time
	Coordinate          *Coordinate
	Radius              int
	Size                int
	Page                int
}

// ToURLValues converts SearchCoursesParams to url.Values.
func (p *SearchCoursesParams) ToURLValues() url.Values {
	values := url.Values{}
	if p == nil {
		return values
	}

	if s := strings.TrimSpace(p.Sort); s != "" {
		values.Set("sort", s)
	}

	if s := strings.TrimSpace(p.TenantID); s != "" {
		values.Set("tenant_id", s)
	}

	if s := strings.TrimSpace(p.Visibility); s != "" {
		values.Set("visibility", s)
	}

	if !p.AvailablePlacesFrom.IsZero() {
		values.Set("available_places_from", FormatTime(p.AvailablePlacesFrom))
	}

	if !p.CourseEndsAfter.IsZero() {
		values.Set("course_ends_after", FormatTime(p.CourseEndsAfter))
	}

	if p.Coordinate != nil {
		values.Set("coordinate", p.Coordinate.String())

		if p.Radius > 0 {
			values.Set("radius", strconv.Itoa(p.Radius))
		}
	}

	if p.Size > 0 {
		values.Set("size", strconv.Itoa(p.Size))
	}

	values.Set("page", strconv.Itoa(p.Page))

	return values
}
