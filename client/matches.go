package client

import (
	"context"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

// SearchMatches returns matches matching params. A nil params applies no filters.
func (c *Client) SearchMatches(ctx context.Context, params *models.SearchMatchesParams) ([]models.Match, error) {
	return get[[]models.Match](ctx, c, "/v1/matches", params)
}
