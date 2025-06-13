package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

// GetAuthMethods reports how a given email can authenticate. It needs no token,
// so it is the one auth call safe to make before logging in.
func (c *Client) GetAuthMethods(ctx context.Context, email string) (*models.AuthMethods, error) {
	if email == "" {
		return nil, ErrMissingID
	}
	q := url.Values{"email": {email}}
	methods, err := get[*models.AuthMethods](ctx, c.withoutAuth(), "/v3/auth/methods", rawValues(q))
	if err != nil {
		return nil, err
	}
	if methods == nil {
		return nil, fmt.Errorf("no auth methods in the response")
	}
	return methods, nil
}
