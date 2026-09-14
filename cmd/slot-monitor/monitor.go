package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

// matchedSlot is a slot that passed every filter, ready to report.
type matchedSlot struct {
	club      Club
	courtName string
	indoor    bool
	startUTC  time.Time
	started   time.Time // start in the configured timezone
	duration  int
	price     string
}

// signature identifies a slot across runs, independent of timezone.
func (m matchedSlot) signature() string {
	return fmt.Sprintf("%s|%s|%d", m.club.TenantID, m.startUTC.UTC().Format(time.RFC3339), m.duration)
}

// matcher holds the resolved config needed to filter slots.
type matcher struct {
	filters Filters
	windows []parsedWindow
	loc     *time.Location
}

// resolveSlotStartUTC combines a resource's start_date with a slot's start_time
// and reads the pair as UTC, matching how the app API reports times (verified
// against the padel-cli client).
func resolveSlotStartUTC(startDate, startTime string) (time.Time, error) {
	date := startDate
	if len(date) >= 10 {
		date = date[:10]
	}
	if len(startTime) == 5 {
		startTime += ":00"
	}
	return time.ParseInLocation(models.TimeFormat, date+"T"+startTime, time.UTC)
}

// match returns every slot of one club that passes the filters and windows.
func (m *matcher) match(club Club, resources map[string]models.TenantResource, avails []models.Availability) ([]matchedSlot, error) {
	var out []matchedSlot
	for _, avail := range avails {
		res, known := resources[avail.ResourceID]
		if !known {
			continue // a court we have no metadata for
		}
		if !m.courtPasses(res) {
			continue
		}
		for _, slot := range avail.Slots {
			if !durationAllowed(slot.Duration, m.filters.Durations) {
				continue
			}
			startUTC, err := resolveSlotStartUTC(avail.StartDate.String(), slot.StartTime)
			if err != nil {
				return nil, fmt.Errorf("club %s resource %s: %w", club.TenantID, avail.ResourceID, err)
			}
			startLocal := startUTC.In(m.loc)
			if !m.withinWindow(startLocal, slot.Duration) {
				continue
			}
			out = append(out, matchedSlot{
				club:      club,
				courtName: courtLabel(res, avail.ResourceID),
				indoor:    res.IsIndoor(),
				startUTC:  startUTC,
				started:   startLocal,
				duration:  slot.Duration,
				price:     slot.Price,
			})
		}
	}
	return out, nil
}

// courtPasses applies the court-type/size/name filters.
func (m *matcher) courtPasses(res models.TenantResource) bool {
	f := m.filters
	if f.CourtType != "" && !strings.EqualFold(res.Properties.ResourceType, f.CourtType) {
		return false
	}
	if f.CourtSize != "" && !strings.EqualFold(res.Properties.ResourceSize, f.CourtSize) {
		return false
	}
	name := strings.ToLower(res.Name)
	if len(f.IncludeCourtNames) > 0 && !containsAnyFold(name, f.IncludeCourtNames) {
		return false
	}
	if containsAnyFold(name, f.ExcludeCourtNames) {
		return false
	}
	return true
}

// withinWindow reports whether a slot fits fully inside any watch window on its
// own weekday (evaluated in the configured timezone).
func (m *matcher) withinWindow(startLocal time.Time, duration int) bool {
	startMin := startLocal.Hour()*60 + startLocal.Minute()
	endMin := startMin + duration
	for _, w := range m.windows {
		if !w.days[startLocal.Weekday()] {
			continue
		}
		if startMin >= w.start.minutes() && endMin <= w.end.minutes() {
			return true
		}
	}
	return false
}

func durationAllowed(duration int, allowed []int) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, d := range allowed {
		if d == duration {
			return true
		}
	}
	return false
}

func containsAnyFold(haystackLower string, needles []string) bool {
	for _, n := range needles {
		if n == "" {
			continue
		}
		if strings.Contains(haystackLower, strings.ToLower(n)) {
			return true
		}
	}
	return false
}

func courtLabel(res models.TenantResource, fallback string) string {
	if res.Name != "" {
		return res.Name
	}
	return fallback
}

// sortSlots orders matches by start time, then club, then court.
func sortSlots(slots []matchedSlot) {
	sort.Slice(slots, func(i, j int) bool {
		if !slots[i].startUTC.Equal(slots[j].startUTC) {
			return slots[i].startUTC.Before(slots[j].startUTC)
		}
		if slots[i].club.TenantID != slots[j].club.TenantID {
			return slots[i].club.TenantID < slots[j].club.TenantID
		}
		return slots[i].courtName < slots[j].courtName
	})
}
