package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

// SearchSocialUsers returns players from the social graph. A nil params applies
// no filters.
func (c *Client) SearchSocialUsers(ctx context.Context, params *models.SearchSocialUsersParams) ([]models.SocialUser, error) {
	return get[[]models.SocialUser](ctx, c, "/v1/social/users", params)
}

// GetUserStats returns follower counts for one user, or for the authenticated
// user when id is "me".
func (c *Client) GetUserStats(ctx context.Context, id string) (*models.UserStats, error) {
	if id == "" {
		return nil, ErrMissingID
	}
	stats, err := get[*models.UserStats](ctx, c, "/v1/social/users/"+url.PathEscape(id)+"/stats", nil)
	if err != nil {
		return nil, err
	}
	if stats == nil {
		return nil, fmt.Errorf("no stats in the response for %q", id)
	}
	return stats, nil
}
