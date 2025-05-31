// Command basic looks up a club, lists its upcoming matches and prints what
// courts are free today.
//
// Set PLAYTOMIC_TENANT_ID to a club ID. Everything else the client needs comes
// from the other PLAYTOMIC_* variables.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/rafa-garcia/go-playtomic-api/client"
	"github.com/rafa-garcia/go-playtomic-api/models"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	tenantID := os.Getenv("PLAYTOMIC_TENANT_ID")
	if tenantID == "" {
		return errors.New("set PLAYTOMIC_TENANT_ID to a club ID")
	}

	c, err := client.NewFromEnv(client.WithTimeout(15 * time.Second))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	club, err := c.GetTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("looking up the club: %w", err)
	}
	fmt.Printf("%s, %s\n\n", club.TenantName, club.Address.City)

	if err := printMatches(ctx, c, tenantID); err != nil {
		return err
	}
	return printFreeCourts(ctx, c, tenantID)
}

func printMatches(ctx context.Context, c *client.Client, tenantID string) error {
	fmt.Println("Upcoming matches")

	params := &models.SearchMatchesParams{
		SportID:       "PADEL",
		TenantIDs:     []string{tenantID},
		FromStartDate: time.Now(),
		Visibility:    "VISIBLE",
		HasPlayers:    true,
	}

	var seen int
	for match, err := range c.AllMatches(ctx, params) {
		if err != nil {
			return fmt.Errorf("listing matches: %w", err)
		}

		fmt.Printf("  %s  level %.1f to %.1f  %d players\n",
			match.StartDate, match.MinLevel, match.MaxLevel, players(match))

		if seen++; seen == 10 {
			break
		}
	}

	if seen == 0 {
		fmt.Println("  none")
	}
	fmt.Println()
	return nil
}

func printFreeCourts(ctx context.Context, c *client.Client, tenantID string) error {
	now := time.Now()
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	courts, err := c.GetAvailability(ctx, &models.AvailabilityParams{
		TenantID: tenantID,
		SportID:  "PADEL",
		From:     day,
		To:       day.Add(24 * time.Hour),
	})
	if err != nil {
		return fmt.Errorf("reading availability: %w", err)
	}

	fmt.Println("Free today")
	for _, court := range courts {
		for _, slot := range court.Slots {
			fmt.Printf("  %s  %s  %d min  %s\n", court.ResourceID, slot.StartTime, slot.Duration, slot.Price)
		}
	}
	return nil
}

func players(m models.Match) int {
	var n int
	for _, team := range m.Teams {
		n += len(team.Players)
	}
	return n
}
