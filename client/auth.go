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

// Auth endpoints. They sit under /v3 while everything else is /v1, which is why
// the client is based on the host rather than a version prefix.
const (
	loginPath   = "/v3/auth/login"
	refreshPath = "/v3/auth/token"
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

// Expired reports whether the token needs replacing. A token with no stated
// expiry is taken at face value until the API rejects it.
func (t *Token) Expired() bool {
	if t == nil || t.AccessToken == "" {
		return true
	}
	if t.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().UTC().After(t.ExpiresAt.Add(-tokenMargin))
}

// TokenSource supplies the bearer token for a request. Implement it to hold
// tokens somewhere this library does not need to know about.
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

	mu    sync.Mutex
	token *Token
}

// Token holds the lock across the network call on purpose: concurrent callers
// finding an expired token should produce one login, not one each.
func (s *credentials) Token(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.token.Expired() {
		return s.token.AccessToken, nil
	}

	if s.token != nil && s.token.RefreshToken != "" {
		// A refresh token the server has already retired is not fatal here.
		if token, err := s.client.Refresh(ctx, s.token.RefreshToken); err == nil {
			s.token = token
			return token.AccessToken, nil
		}
	}

	token, err := s.client.Login(ctx, s.email, s.password)
	if err != nil {
		return "", err
	}

	s.token = token
	return token.AccessToken, nil
}

// Login exchanges credentials for a token. It leaves the client's own
// authentication alone: pass the result to WithToken, or use WithCredentials
// and let the client manage the lifecycle.
func (c *Client) Login(ctx context.Context, email, password string) (*Token, error) {
	return c.authenticate(ctx, loginPath, map[string]string{
		"email":    email,
		"password": password,
	})
}

// Refresh trades a refresh token for a fresh pair.
func (c *Client) Refresh(ctx context.Context, refreshToken string) (*Token, error) {
	return c.authenticate(ctx, refreshPath, map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	})
}

func (c *Client) authenticate(ctx context.Context, path string, payload map[string]string) (*Token, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encoding credentials: %w", err)
	}

	// A copy with no token source, so acquiring a token cannot ask for one.
	bare := *c
	bare.tokenSource = nil

	var token Token
	if err := bare.sendRequest(ctx, http.MethodPost, path, "", body, &token); err != nil {
		return nil, err
	}
	return &token, nil
}
