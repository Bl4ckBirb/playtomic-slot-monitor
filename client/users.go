package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

// GetMe returns the authenticated user. It needs a token.
func (c *Client) GetMe(ctx context.Context) (*models.User, error) {
	user, err := get[*models.User](ctx, c, "/v2/users/me", nil)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("no user in the response")
	}
	return user, nil
}

// GetUser returns one user by ID.
func (c *Client) GetUser(ctx context.Context, id string) (*models.User, error) {
	if id == "" {
		return nil, ErrMissingID
	}
	user, err := get[*models.User](ctx, c, "/v2/users/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("no user in the response for %q", id)
	}
	return user, nil
}
