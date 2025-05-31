// Package client provides a Go client for accessing the Playtomic API.
package client

import (
	"log/slog"
	"net/http"
	"time"
)

const (
	// DefaultBaseURL is the Playtomic API host. Version prefixes belong to
	// the endpoint paths, since the API mixes v1 and v3.
	DefaultBaseURL = "https://api.playtomic.io"

	// DefaultTimeout is the default client timeout
	DefaultTimeout = 30 * time.Second

	// DefaultMaxRetries is the default number of request retries
	DefaultMaxRetries = 3

	// DefaultUserAgent is the default User-Agent sent with requests
	DefaultUserAgent = "PlaytomicGoClient/1.0"

	// DefaultLoginPath and DefaultRefreshPath are the auth endpoints. They sit
	// under /v3 while the rest of the API is /v1, which is why the client is
	// based on the host rather than a version prefix.
	DefaultLoginPath   = "/v3/auth/login"
	DefaultRefreshPath = "/v3/auth/token"

	// DefaultRetryWait is the first backoff window. It doubles per attempt.
	DefaultRetryWait = 500 * time.Millisecond

	// DefaultMaxRetryWait caps a single wait, including one the server asked
	// for through Retry-After.
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

	// The timeout lands on a copy after every option has run, so WithTimeout
	// and WithHTTPClient no longer depend on which came last, and the caller's
	// own client is left alone.
	httpClient := *c.httpClient
	httpClient.Timeout = c.timeout
	c.httpClient = &httpClient

	return c
}
