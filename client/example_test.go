package client_test

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

func ExampleClient_SearchMatches() {
	c := client.NewClient(client.WithTimeout(10 * time.Second))

	matches, err := c.SearchMatches(context.Background(), &models.SearchMatchesParams{
		SportID:       "PADEL",
		TenantIDs:     []string{os.Getenv("PLAYTOMIC_TENANT_ID")},
		FromStartDate: time.Now(),
		HasPlayers:    true,
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range matches {
		fmt.Printf("%s %s %.1f-%.1f\n", m.StartDate, m.Tenant.TenantName, m.MinLevel, m.MaxLevel)
	}
}

// Searching by area rather than by club, which is how you find tenant IDs in
// the first place.
func ExampleClient_SearchTenants() {
	c := client.NewClient()

	clubs, err := c.SearchTenants(context.Background(), &models.SearchTenantsParams{
		Coordinate:      &models.Coordinate{Lat: 51.5074, Lon: -0.1278},
		Radius:          25000,
		SportID:         "PADEL",
		PlaytomicStatus: "ACTIVE",
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, club := range clubs {
		fmt.Println(club.TenantID, club.TenantName, club.Address.City)
	}
}

func ExampleClient_AllMatches() {
	c := client.NewClient()

	for match, err := range c.AllMatches(context.Background(), &models.SearchMatchesParams{SportID: "PADEL"}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(match.MatchID, match.StartDate)
	}
}

func ExampleClient_GetAvailability() {
	c := client.NewClient()
	day := time.Now().Truncate(24 * time.Hour)

	courts, err := c.GetAvailability(context.Background(), &models.AvailabilityParams{
		TenantIDs: []string{os.Getenv("PLAYTOMIC_TENANT_ID")},
		SportID:   "PADEL",
		From:      day,
		To:        day.Add(24 * time.Hour),
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, court := range courts {
		for _, slot := range court.Slots {
			fmt.Printf("%s %s %dmin %s\n", court.ResourceID, slot.StartTime, slot.Duration, slot.Price)
		}
	}
}

// Credentials belong in the environment, never in source.
func ExampleWithCredentials() {
	c := client.NewClient(client.WithCredentials(
		os.Getenv("PLAYTOMIC_EMAIL"),
		os.Getenv("PLAYTOMIC_PASSWORD"),
	))

	if _, err := c.SearchClasses(context.Background(), nil); err != nil {
		log.Fatal(err)
	}
}

func ExampleError() {
	c := client.NewClient()

	_, err := c.SearchMatches(context.Background(), nil)

	switch {
	case errors.Is(err, client.ErrRateLimited):
		var apiErr *client.Error
		errors.As(err, &apiErr)
		fmt.Println("back off for", apiErr.RetryAfter)
	case errors.Is(err, client.ErrUnauthorized):
		fmt.Println("token expired")
	case err != nil:
		fmt.Println(err)
	}
}
