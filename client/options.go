package client

import (
	"net/http"
	"time"
)

// Option defines a function that configures the client
type Option func(*Client)

// WithBaseURL sets a custom base URL for the client
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithRetries sets the maximum number of retries for failed requests
func WithRetries(retries int) Option {
	return func(c *Client) {
		c.maxRetries = retries
	}
}

// WithBackoff sets the first retry window and the cap on any single wait.
// Non-positive values leave the default in place.
func WithBackoff(first, maxWait time.Duration) Option {
	return func(c *Client) {
		if first > 0 {
			c.retryWait = first
		}
		if maxWait > 0 {
			c.maxRetryWait = maxWait
		}
	}
}

// WithDebug enables debug logging
func WithDebug(enabled bool) Option {
	return func(c *Client) {
		c.debug = enabled
	}
}

// WithHeader sets a request header, overriding a default of the same name.
// A header the API starts demanding can be supplied without a release.
func WithHeader(name, value string) Option {
	return func(c *Client) {
		if c.headers == nil {
			c.headers = http.Header{}
		}
		c.headers.Set(name, value)
	}
}

// WithUserAgent sets a custom User-Agent header
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}
