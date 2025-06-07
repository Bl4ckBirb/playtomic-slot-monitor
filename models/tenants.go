package models

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Tenant represents a club/venue in the Playtomic API
type Tenant struct {
	TenantID        string         `json:"tenant_id"`
	TenantName      string         `json:"tenant_name"`
	Address         Address        `json:"address"`
	Images          []string       `json:"images"`
	Properties      map[string]any `json:"properties"`
	PlaytomicStatus string         `json:"playtomic_status"`
}

// Address represents a physical address
type Address struct {
	Street                string     `json:"street"`
	PostalCode            string     `json:"postal_code"`
	City                  string     `json:"city"`
	SubAdministrativeArea string     `json:"sub_administrative_area"`
	AdministrativeArea    string     `json:"administrative_area"`
	Country               string     `json:"country"`
	CountryCode           string     `json:"country_code"`
	Coordinate            Coordinate `json:"coordinate"`
	Timezone              string     `json:"timezone"`
}

// Coordinate represents a geographical coordinate
type Coordinate struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// String renders a coordinate the way the API expects it in a query.
func (c Coordinate) String() string {
	return fmt.Sprintf("%f,%f", c.Lat, c.Lon)
}

// SearchTenantsParams defines parameters for searching clubs
type SearchTenantsParams struct {
	TenantIDs       []string
	Name            string
	SportID         string
	PlaytomicStatus string
	Coordinate      *Coordinate
	Radius          int
	WithProperties  []string
	Size            int
	Page            int
}

// ToURLValues converts SearchTenantsParams to url.Values
func (p *SearchTenantsParams) ToURLValues() url.Values {
	values := url.Values{}
	if p == nil {
		return values
	}

	if len(p.TenantIDs) > 0 {
		values.Set("tenant_id", strings.Join(p.TenantIDs, ","))
	}

	if s := strings.TrimSpace(p.Name); s != "" {
		values.Set("tenant_name", s)
	}

	if s := strings.TrimSpace(p.SportID); s != "" {
		values.Set("sport_id", s)
	}

	if s := strings.TrimSpace(p.PlaytomicStatus); s != "" {
		values.Set("playtomic_status", s)
	}

	if p.Coordinate != nil {
		values.Set("coordinate", p.Coordinate.String())

		if p.Radius > 0 {
			values.Set("radius", strconv.Itoa(p.Radius))
		}
	}

	if len(p.WithProperties) > 0 {
		values.Set("with_properties", strings.Join(p.WithProperties, ","))
	}

	if p.Size > 0 {
		values.Set("size", strconv.Itoa(p.Size))
	}

	values.Set("page", strconv.Itoa(p.Page))

	return values
}
