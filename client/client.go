// Package client provides a Go client for accessing the Playtomic API.
package client

import (
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
	userAgent    string
	maxRetries   int
	retryWait    time.Duration
	maxRetryWait time.Duration
	headers      http.Header
	tokenSource  TokenSource
	debug        bool
}

// NewClient creates a new Playtomic API client with the given options
func NewClient(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		baseURL:      DefaultBaseURL,
		userAgent:    DefaultUserAgent,
		maxRetries:   DefaultMaxRetries,
		retryWait:    DefaultRetryWait,
		maxRetryWait: DefaultMaxRetryWait,
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	return c
}
