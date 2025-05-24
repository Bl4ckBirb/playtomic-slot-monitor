package client

import (
	"context"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

// GetAvailability returns open slots per resource for a club and window.
func (c *Client) GetAvailability(ctx context.Context, params *models.AvailabilityParams) ([]models.Availability, error) {
	return get[[]models.Availability](ctx, c, "/v1/availability", params)
}
