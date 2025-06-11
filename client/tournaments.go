package client

import (
	"context"
	"fmt"
	"iter"
	"net/url"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

// SearchTournaments returns tournaments matching params. A nil params applies
// no filters.
func (c *Client) SearchTournaments(ctx context.Context, params *models.SearchTournamentsParams) ([]models.Tournament, error) {
	return get[[]models.Tournament](ctx, c, "/v2/tournaments", params)
}

// GetTournament returns one tournament by ID.
func (c *Client) GetTournament(ctx context.Context, id string) (*models.Tournament, error) {
	if id == "" {
		return nil, ErrMissingID
	}
	t, err := get[*models.Tournament](ctx, c, "/v2/tournaments/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, fmt.Errorf("no tournament in the response for %q", id)
	}
	return t, nil
}

// AllTournaments iterates every page of a tournament search.
func (c *Client) AllTournaments(ctx context.Context, params *models.SearchTournamentsParams) iter.Seq2[models.Tournament, error] {
	var q models.SearchTournamentsParams
	if params != nil {
		q = *params
	}
	if q.Size <= 0 {
		q.Size = DefaultPageSize
	}

	return paginate(ctx, func(ctx context.Context, page int) ([]models.Tournament, error) {
		q.Page = page
		return c.SearchTournaments(ctx, &q)
	})
}
