package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

// tokenMargin renews slightly early so a token cannot expire mid-flight.
const tokenMargin = 30 * time.Second

// Token is the credential pair the auth endpoints return.
type Token struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresAt    models.Time `json:"access_token_expiration"`
	UserID       string      `json:"user_id"`
}

// Expired reports whether the token needs replacing. No stated expiry means
// trust it until the API says otherwise.
func (t *Token) Expired() bool {
	if t == nil || t.AccessToken == "" {
		return true
	}
	if t.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().UTC().After(t.ExpiresAt.Add(-tokenMargin))
}

// TokenSource supplies the bearer token. Implement it to keep tokens somewhere
// this library need not know about.
type TokenSource interface {
	Token(ctx context.Context) (string, error)
}

type staticToken string

func (s staticToken) Token(context.Context) (string, error) { return string(s), nil }

// credentials logs in on demand and refreshes on expiry.
type credentials struct {
	client   *Client
	email    string
	password string

	// gate serialises renewal. A channel, not a mutex, so a waiter can leave
	// when its own context dies.
	gate chan struct{}

	mu    sync.Mutex
	token *Token
}

func newCredentials(c *Client, email, password string) *credentials {
	return &credentials{
		client:   c,
		email:    email,
		password: password,
		gate:     make(chan struct{}, 1),
	}
}

func (s *credentials) cached() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token.Expired() {
		return ""
	}
	return s.token.AccessToken
}

// invalidate clears the token only when it is the one that was rejected.
func (s *credentials) invalidate(bearer string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token != nil && s.token.AccessToken == bearer {
		s.token = nil
	}
}

func (s *credentials) Token(ctx context.Context) (string, error) {
	if token := s.cached(); token != "" {
		return token, nil
	}

	select {
	case s.gate <- struct{}{}:
		defer func() { <-s.gate }()
	case <-ctx.Done():
		return "", ctx.Err()
	}

	// Someone may have renewed it while we were waiting for the gate.
	if token := s.cached(); token != "" {
		return token, nil
	}

	token, err := s.renew(ctx)
	if err != nil {
		return "", err
	}

	// Otherwise we come straight back here, forever.
	if token.Expired() {
		return "", fmt.Errorf("playtomic returned a token that is already expired")
	}

	s.mu.Lock()
	s.token = token
	s.mu.Unlock()

	return token.AccessToken, nil
}

func (s *credentials) renew(ctx context.Context) (*Token, error) {
	s.mu.Lock()
	previous := s.token
	s.mu.Unlock()

	if previous != nil && previous.RefreshToken != "" {
		// A refresh token the server has already retired is not fatal here.
		if token, err := s.client.Refresh(ctx, previous.RefreshToken); err == nil {
			return token, nil
		}
	}

	return s.client.Login(ctx, s.email, s.password)
}

// Login exchanges credentials for a token. It does not change the client's own
// authentication: for that use WithToken or WithCredentials.
func (c *Client) Login(ctx context.Context, email, password string) (*Token, error) {
	return c.authenticate(ctx, c.loginPath, map[string]string{
		"email":    email,
		"password": password,
	})
}

// Refresh trades a refresh token for a fresh pair.
func (c *Client) Refresh(ctx context.Context, refreshToken string) (*Token, error) {
	return c.authenticate(ctx, c.refreshPath, map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	})
}

func (c *Client) authenticate(ctx context.Context, path string, payload map[string]string) (*Token, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encoding credentials: %w", err)
	}

	var token Token
	if err := c.withoutAuth().sendRequest(ctx, http.MethodPost, path, "", body, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// withoutAuth returns a shallow copy that sends no bearer: no token source and
// no Authorization header. The auth endpoints produce authorization, they must
// not consume it, and neither must the pre-login methods check.
func (c *Client) withoutAuth() *Client {
	bare := *c
	bare.tokenSource = nil
	if bare.headers != nil {
		bare.headers = bare.headers.Clone()
		bare.headers.Del("Authorization")
	}
	return &bare
}
