package models

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Tournament is a competition, from /v2/tournaments.
type Tournament struct {
	TournamentID    string  `json:"tournament_id"`
	Name            string  `json:"name"`
	SportID         string  `json:"sport_id"`
	StartDate       Time    `json:"start_date"`
	EndDate         Time    `json:"end_date"`
	Type            string  `json:"type"`
	Visibility      string  `json:"visibility"`
	Status          string  `json:"status"`
	Tenant          Tenant  `json:"tenant"`
	AvailablePlaces int     `json:"available_places"`
	AccessCode      *string `json:"access_code"`
}

// SearchTournamentsParams defines parameters for searching tournaments.
type SearchTournamentsParams struct {
	Sort               string
	SportID            string
	TenantID           string
	Status             string
	RegistrationStatus string
	Type               string
	Visibility         string
	UserID             string
	FromStartDate      time.Time
	ToStartDate        time.Time
	AvailablePlaces    bool
	Coordinate         *Coordinate
	Radius             int
	Size               int
	Page               int
}

// ToURLValues converts SearchTournamentsParams to url.Values.
func (p *SearchTournamentsParams) ToURLValues() url.Values {
	values := url.Values{}
	if p == nil {
		return values
	}

	set := func(k, v string) {
		if s := strings.TrimSpace(v); s != "" {
			values.Set(k, s)
		}
	}
	set("sort", p.Sort)
	set("sport_id", p.SportID)
	set("tenant_id", p.TenantID)
	set("status", p.Status)
	set("registration_status", p.RegistrationStatus)
	set("type", p.Type)
	set("visibility", p.Visibility)
	set("user_id", p.UserID)

	if !p.FromStartDate.IsZero() {
		values.Set("from_start_date", FormatTime(p.FromStartDate))
	}
	if !p.ToStartDate.IsZero() {
		values.Set("to_start_date", FormatTime(p.ToStartDate))
	}

	if p.AvailablePlaces {
		values.Set("available_places", "true")
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
