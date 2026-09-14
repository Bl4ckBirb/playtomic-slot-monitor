package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// sportID is fixed: this monitor only ever looks for padel.
const sportID = "PADEL"

// defaultTimezone is used when the config leaves timezone empty.
const defaultTimezone = "Europe/Berlin"

// weekdays maps the config's short day names to time.Weekday.
var weekdays = map[string]time.Weekday{
	"sun": time.Sunday,
	"mon": time.Monday,
	"tue": time.Tuesday,
	"wed": time.Wednesday,
	"thu": time.Thursday,
	"fri": time.Friday,
	"sat": time.Saturday,
}

// Config is the whole monitor configuration, supplied as JSON.
type Config struct {
	Timezone      string   `json:"timezone"`
	LookAheadDays int      `json:"look_ahead_days"`
	Clubs         []Club   `json:"clubs"`
	WatchWindows  []Window `json:"watch_windows"`
	Filters       Filters  `json:"filters"`
}

// Club is one venue to watch. Only TenantID is required.
type Club struct {
	TenantID string `json:"tenant_id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
}

// Window is a weekday/time span slots must fall inside.
type Window struct {
	Days  []string `json:"days"`
	Start string   `json:"start"`
	End   string   `json:"end"`
}

// Filters narrows which slots qualify. Empty fields disable that filter.
type Filters struct {
	Durations         []int    `json:"durations"`
	CourtType         string   `json:"court_type"` // "" | "indoor" | "outdoor"
	CourtSize         string   `json:"court_size"` // "" | "single" | "double"
	IncludeCourtNames []string `json:"include_court_names"`
	ExcludeCourtNames []string `json:"exclude_court_names"`
}

// parsedWindow is a Window with its days and times resolved once.
type parsedWindow struct {
	days  map[time.Weekday]bool
	start timeOfDay
	end   timeOfDay
}

type timeOfDay struct {
	hour, minute int
}

func (t timeOfDay) minutes() int { return t.hour*60 + t.minute }

// loadConfig reads and validates the JSON config file.
func loadConfig(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse config JSON: %w", err)
	}
	if cfg.Timezone == "" {
		cfg.Timezone = defaultTimezone
	}
	if cfg.LookAheadDays <= 0 {
		cfg.LookAheadDays = 7
	}
	if len(cfg.Clubs) == 0 {
		return nil, fmt.Errorf("config has no clubs")
	}
	for i, club := range cfg.Clubs {
		if strings.TrimSpace(club.TenantID) == "" {
			return nil, fmt.Errorf("club %d has no tenant_id", i)
		}
	}
	if len(cfg.WatchWindows) == 0 {
		return nil, fmt.Errorf("config has no watch_windows")
	}
	return &cfg, nil
}

// location resolves the config timezone to a *time.Location.
func (c *Config) location() (*time.Location, error) {
	loc, err := time.LoadLocation(c.Timezone)
	if err != nil {
		return nil, fmt.Errorf("load timezone %q: %w", c.Timezone, err)
	}
	return loc, nil
}

// parseWindows resolves every watch window's days and start/end times.
func (c *Config) parseWindows() ([]parsedWindow, error) {
	windows := make([]parsedWindow, 0, len(c.WatchWindows))
	for i, w := range c.WatchWindows {
		days := make(map[time.Weekday]bool, len(w.Days))
		for _, d := range w.Days {
			wd, ok := weekdays[strings.ToLower(strings.TrimSpace(d))]
			if !ok {
				return nil, fmt.Errorf("watch_windows[%d]: invalid day %q", i, d)
			}
			days[wd] = true
		}
		start, err := parseTimeOfDay(w.Start)
		if err != nil {
			return nil, fmt.Errorf("watch_windows[%d].start: %w", i, err)
		}
		end, err := parseTimeOfDay(w.End)
		if err != nil {
			return nil, fmt.Errorf("watch_windows[%d].end: %w", i, err)
		}
		windows = append(windows, parsedWindow{days: days, start: start, end: end})
	}
	return windows, nil
}

// parseTimeOfDay parses "HH:MM" into a timeOfDay.
func parseTimeOfDay(s string) (timeOfDay, error) {
	var t timeOfDay
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) != 2 {
		return t, fmt.Errorf("invalid time %q, want HH:MM", s)
	}
	if _, err := fmt.Sscanf(s, "%d:%d", &t.hour, &t.minute); err != nil {
		return t, fmt.Errorf("invalid time %q: %w", s, err)
	}
	if t.hour < 0 || t.hour > 23 || t.minute < 0 || t.minute > 59 {
		return t, fmt.Errorf("time %q out of range", s)
	}
	return t, nil
}
