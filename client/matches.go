package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

// SearchMatches returns matches matching params. A nil params applies no filters.
func (c *Client) SearchMatches(ctx context.Context, params *models.SearchMatchesParams) ([]models.Match, error) {
	return get[[]models.Match](ctx, c, "/v1/matches", params)
}

// GetMatch returns one match by ID.
func (c *Client) GetMatch(ctx context.Context, id string) (*models.Match, error) {
	if id == "" {
		return nil, ErrMissingID
	}
	match, err := get[*models.Match](ctx, c, "/v1/matches/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}
	if match == nil {
		return nil, fmt.Errorf("no match in the response for %q", id)
	}
	return match, nil
}
