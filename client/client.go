// Package client provides a Go client for accessing the Playtomic API.
package client

import (
	"log/slog"
	"net/http"
	"time"
)

const (
	// DefaultBaseURL is the API host. Version prefixes live in the paths.
	// The app talks to api.app.playtomic.io. The old api.playtomic.io sits
	// behind a CloudFront rule that 403s everything.
	DefaultBaseURL = "https://api.app.playtomic.io"

	// DefaultTimeout is the default client timeout
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the default number of request retries
	DefaultMaxRetries = 3

	// DefaultUserAgent is the default User-Agent sent with requests
	DefaultUserAgent = "PlaytomicGoClient/1.0"

	// Auth sits under /v3 while the rest of the API is /v1.
	DefaultLoginPath   = "/v3/auth/login"
	DefaultRefreshPath = "/v3/auth/token"

	// DefaultRetryWait is the first backoff window. It doubles per attempt.
	DefaultRetryWait = 500 * time.Millisecond

	// DefaultMaxRetryWait caps a single wait.
	DefaultMaxRetryWait = 10 * time.Second
)

// Client provides access to the Playtomic API
type Client struct {
	httpClient   *http.Client
	baseURL      string
	timeout      time.Duration
	maxRetries   int
	retryWait    time.Duration
	maxRetryWait time.Duration
	headers      http.Header
	loginPath    string
	refreshPath  string
	tokenSource  TokenSource
	logger       *slog.Logger
}

// NewClient creates a new Playtomic API client with the given options
func NewClient(opts ...Option) *Client {
	c := &Client{
		httpClient:   &http.Client{},
		baseURL:      DefaultBaseURL,
		timeout:      DefaultTimeout,
		maxRetries:   DefaultMaxRetries,
		retryWait:    DefaultRetryWait,
		maxRetryWait: DefaultMaxRetryWait,
		loginPath:    DefaultLoginPath,
		refreshPath:  DefaultRefreshPath,
		logger:       slog.New(slog.DiscardHandler),
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	// After all options, so WithTimeout and WithHTTPClient cannot fight, and
	// the caller's own client is left alone.
	httpClient := *c.httpClient
	httpClient.Timeout = c.timeout
	c.httpClient = &httpClient

	return c
}
