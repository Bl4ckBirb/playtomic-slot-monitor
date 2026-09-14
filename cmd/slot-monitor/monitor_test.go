package main

import (
	"testing"
	"time"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

func mustDate(s string) models.Time {
	t, err := time.Parse(models.DateFormat, s)
	if err != nil {
		panic(err)
	}
	return models.Time{Time: t}
}

func testMatcher(t *testing.T, filters Filters) *matcher {
	t.Helper()
	cfg := &Config{
		WatchWindows: []Window{
			{Days: []string{"mon", "tue", "wed", "thu"}, Start: "17:30", End: "21:00"},
			{Days: []string{"sat", "sun"}, Start: "09:00", End: "22:00"},
		},
	}
	windows, err := cfg.parseWindows()
	if err != nil {
		t.Fatalf("parseWindows: %v", err)
	}
	// Fixed +02:00 zone so the test does not depend on system tzdata.
	return &matcher{filters: filters, windows: windows, loc: time.FixedZone("TEST", 2*3600)}
}

func indoorDouble(name string) models.TenantResource {
	return models.TenantResource{ResourceID: name, Name: name, Properties: models.ResourceProperties{ResourceType: "indoor", ResourceSize: "double"}}
}

func outdoorSingle(name string) models.TenantResource {
	return models.TenantResource{ResourceID: name, Name: name, Properties: models.ResourceProperties{ResourceType: "outdoor", ResourceSize: "single"}}
}

func TestResolveSlotStartUTC(t *testing.T) {
	got, err := resolveSlotStartUTC("2025-05-26", "08:00:00")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	want := time.Date(2025, 5, 26, 8, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %s, want %s", got, want)
	}
	// Date with a trailing T-part and a short HH:MM time still parse.
	got2, err := resolveSlotStartUTC("2025-05-26T00:00:00", "17:30")
	if err != nil {
		t.Fatalf("resolve2: %v", err)
	}
	if !got2.Equal(time.Date(2025, 5, 26, 17, 30, 0, 0, time.UTC)) {
		t.Errorf("got2 %s", got2)
	}
}

func TestMatchFiltersAndWindow(t *testing.T) {
	m := testMatcher(t, Filters{Durations: []int{90}, CourtType: "indoor"})
	resources := map[string]models.TenantResource{
		"c1": indoorDouble("c1"),
		"c2": outdoorSingle("c2"),
	}
	// 2025-05-26 is a Monday. UTC times; +02:00 local.
	avails := []models.Availability{
		{ResourceID: "c1", StartDate: mustDate("2025-05-26"), Slots: []models.Slot{
			{StartTime: "16:00:00", Duration: 90, Price: "24 EUR"}, // 18:00–19:30 local -> match
			{StartTime: "06:00:00", Duration: 90, Price: "20 EUR"}, // 08:00 local -> outside window
			{StartTime: "16:00:00", Duration: 60, Price: "18 EUR"}, // wrong duration
		}},
		{ResourceID: "c2", StartDate: mustDate("2025-05-26"), Slots: []models.Slot{
			{StartTime: "16:00:00", Duration: 90, Price: "24 EUR"}, // outdoor -> filtered
		}},
		{ResourceID: "unknown", StartDate: mustDate("2025-05-26"), Slots: []models.Slot{
			{StartTime: "16:00:00", Duration: 90, Price: "24 EUR"}, // no metadata -> skipped
		}},
	}

	got, err := m.match(Club{TenantID: "club-1"}, resources, avails)
	if err != nil {
		t.Fatalf("match: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d matches, want 1: %+v", len(got), got)
	}
	if got[0].courtName != "c1" || got[0].duration != 90 {
		t.Errorf("match = %+v", got[0])
	}
	if h := got[0].started.Hour(); h != 18 {
		t.Errorf("local start hour = %d, want 18", h)
	}
	if !got[0].indoor {
		t.Errorf("expected indoor")
	}
}

func TestWindowBoundaryFit(t *testing.T) {
	m := testMatcher(t, Filters{Durations: []int{90}})
	resources := map[string]models.TenantResource{"c1": indoorDouble("c1")}
	// 19:30 local start (17:30 UTC) + 90m = 21:00 local -> fits exactly.
	// 19:31 local start would end 21:01 -> must not fit.
	avails := []models.Availability{
		{ResourceID: "c1", StartDate: mustDate("2025-05-26"), Slots: []models.Slot{
			{StartTime: "17:30:00", Duration: 90}, // 19:30–21:00 -> match
			{StartTime: "17:31:00", Duration: 90}, // 19:31–21:01 -> no
		}},
	}
	got, err := m.match(Club{TenantID: "c"}, resources, avails)
	if err != nil {
		t.Fatalf("match: %v", err)
	}
	if len(got) != 1 || got[0].started.Format("15:04") != "19:30" {
		t.Fatalf("got %+v", got)
	}
}

func TestStateNewSlotsAndPrune(t *testing.T) {
	st := &state{Seen: map[string]string{}}
	slot := matchedSlot{
		club:     Club{TenantID: "club-1"},
		startUTC: time.Date(2025, 5, 26, 16, 0, 0, 0, time.UTC),
		duration: 90,
	}
	fresh := st.newSlots([]matchedSlot{slot})
	if len(fresh) != 1 {
		t.Fatalf("first pass: got %d, want 1", len(fresh))
	}
	// Same slot again -> not new.
	if again := st.newSlots([]matchedSlot{slot}); len(again) != 0 {
		t.Fatalf("second pass: got %d, want 0", len(again))
	}
	wantSig := "club-1|2025-05-26T16:00:00Z|90"
	if sigs := st.signatures(); len(sigs) != 1 || sigs[0] != wantSig {
		t.Fatalf("signatures = %v", sigs)
	}
	// Pruning after the slot's start drops it.
	st.prune(time.Date(2025, 5, 27, 0, 0, 0, 0, time.UTC))
	if len(st.Seen) != 0 {
		t.Fatalf("prune left %d entries", len(st.Seen))
	}
}
