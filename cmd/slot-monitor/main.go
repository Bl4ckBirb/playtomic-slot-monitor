// Command slot-monitor polls Playtomic court availability for configured clubs
// and time windows and sends a Telegram message when new matching slots appear.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rafa-garcia/go-playtomic-api/client"
	"github.com/rafa-garcia/go-playtomic-api/models"
)

const requestedWith = "com.playtomic.web"

// browserUserAgent presents the client as a normal browser rather than the
// library default ("PlaytomicGoClient/1.0"), which reads as an obvious bot.
const browserUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
	"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

func main() {
	configPath := flag.String("config", "config.json", "path to the JSON config file")
	statePath := flag.String("state", "state/seen.json", "path to the seen-slots state file")
	validate := flag.Bool("validate", false, "validate the config file and exit, without contacting the API")
	flag.Parse()

	if err := run(*configPath, *statePath, *validate); err != nil {
		log.Fatalf("slot-monitor: %v", err)
	}
}

func run(configPath, statePath string, validateOnly bool) error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	loc, err := cfg.location()
	if err != nil {
		return err
	}
	windows, err := cfg.parseWindows()
	if err != nil {
		return err
	}

	if validateOnly {
		log.Printf("config OK: %d clubs, %d watch windows, timezone %s", len(cfg.Clubs), len(windows), cfg.Timezone)
		return nil
	}

	email := os.Getenv("PLAYTOMIC_EMAIL")
	password := os.Getenv("PLAYTOMIC_PASSWORD")
	if email == "" || password == "" {
		return fmt.Errorf("PLAYTOMIC_EMAIL and PLAYTOMIC_PASSWORD must be set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Always log in fresh (no token caching), then use the access token.
	login := client.NewClient(
		client.WithUserAgent(browserUserAgent),
		client.WithHeader("X-Requested-With", requestedWith),
		client.WithTimeout(20*time.Second),
	)
	token, err := login.Login(ctx, email, password)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	api := client.NewClient(
		client.WithToken(token.AccessToken),
		client.WithUserAgent(browserUserAgent),
		client.WithHeader("X-Requested-With", requestedWith),
		client.WithTimeout(20*time.Second),
	)

	m := &matcher{filters: cfg.Filters, windows: windows, loc: loc}
	now := time.Now().In(loc)

	var allMatches []matchedSlot
	clubsOK := 0
	for _, club := range cfg.Clubs {
		matches, err := collectClub(ctx, api, m, club, cfg, windows, now)
		if err != nil {
			log.Printf("club %s: %v (skipping)", clubLabel(club), err)
			continue
		}
		clubsOK++
		allMatches = append(allMatches, matches...)
	}
	if clubsOK == 0 {
		return fmt.Errorf("all %d clubs failed", len(cfg.Clubs))
	}

	sortSlots(allMatches)

	st, err := loadState(statePath)
	if err != nil {
		return err
	}
	fresh := st.newSlots(allMatches)
	st.prune(now.UTC())

	log.Printf("clubs ok: %d/%d | matching slots: %d | new: %d", clubsOK, len(cfg.Clubs), len(allMatches), len(fresh))

	if len(fresh) > 0 {
		if err := notify(ctx, fresh); err != nil {
			return err
		}
	}

	if err := st.save(statePath); err != nil {
		return err
	}
	return nil
}

// collectClub fetches one club's courts and availability and returns the slots
// that pass every filter.
func collectClub(ctx context.Context, api *client.Client, m *matcher, club Club, cfg *Config, windows []parsedWindow, now time.Time) ([]matchedSlot, error) {
	resList, err := api.GetResources(ctx, club.TenantID)
	if err != nil {
		return nil, fmt.Errorf("get resources: %w", err)
	}
	resources := make(map[string]models.TenantResource, len(resList))
	for _, r := range resList {
		resources[r.ResourceID] = r
	}

	var matches []matchedSlot
	for _, day := range candidateDays(now, cfg.LookAheadDays, windows, m.loc) {
		start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, m.loc)
		end := start.Add(24 * time.Hour)
		avails, err := api.GetAvailability(ctx, &models.AvailabilityParams{
			TenantIDs: []string{club.TenantID},
			SportID:   sportID,
			StartMin:  start.UTC(),
			StartMax:  end.UTC(),
		})
		if err != nil {
			log.Printf("club %s %s: availability: %v (skipping day)", clubLabel(club), day.Format("2006-01-02"), err)
			continue
		}
		dayMatches, err := m.match(club, resources, avails)
		if err != nil {
			return nil, err
		}
		matches = append(matches, dayMatches...)
	}
	return matches, nil
}

// candidateDays returns the local dates within look-ahead that have at least one
// watch window on their weekday.
func candidateDays(now time.Time, lookAhead int, windows []parsedWindow, loc *time.Location) []time.Time {
	var days []time.Time
	for offset := 0; offset < lookAhead; offset++ {
		day := now.AddDate(0, 0, offset).In(loc)
		for _, w := range windows {
			if w.days[day.Weekday()] {
				days = append(days, day)
				break
			}
		}
	}
	return days
}

// notify sends the new slots to Telegram, or just logs them when Telegram is
// not configured.
func notify(ctx context.Context, slots []matchedSlot) error {
	msg := formatMessage(slots)
	tg := newTelegram(os.Getenv("TELEGRAM_BOT_TOKEN"), os.Getenv("TELEGRAM_CHAT_ID"))
	if !tg.configured() {
		log.Printf("telegram not configured; would send:\n%s", msg)
		return nil
	}
	if err := tg.send(ctx, msg); err != nil {
		return fmt.Errorf("telegram: %w", err)
	}
	log.Printf("sent telegram message with %d new slots", len(slots))
	return nil
}

func clubLabel(club Club) string {
	if club.Name != "" {
		return club.Name
	}
	return club.TenantID
}
