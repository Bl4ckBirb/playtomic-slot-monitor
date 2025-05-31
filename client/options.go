package client

import (
	"log/slog"
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

// WithTimeout sets the HTTP client timeout, applied after every other option.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
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

// WithLogger logs one record per attempt at debug level. Headers stay out:
// one of them is the bearer.
func WithLogger(l *slog.Logger) Option {
	return func(c *Client) {
		if l != nil {
			c.logger = l
		}
	}
}

// WithHeader sets a request header, overriding a default of the same name.
func WithHeader(name, value string) Option {
	return func(c *Client) {
		if c.headers == nil {
			c.headers = http.Header{}
		}
		c.headers.Set(name, value)
	}
}

// WithToken authenticates with an access token you already hold.
func WithToken(accessToken string) Option {
	return func(c *Client) {
		c.tokenSource = staticToken(accessToken)
	}
}

// WithTokenSource authenticates through a source you control.
func WithTokenSource(src TokenSource) Option {
	return func(c *Client) {
		c.tokenSource = src
	}
}

// WithCredentials logs in on first need and refreshes on expiry.
func WithCredentials(email, password string) Option {
	return func(c *Client) {
		c.tokenSource = newCredentials(c, email, password)
	}
}

// WithUserAgent sets the User-Agent. A header like any other, so the last
// option naming it wins.
func WithUserAgent(userAgent string) Option {
	return WithHeader("User-Agent", userAgent)
}

// WithAuthPaths overrides the auth endpoints. Empty values keep the defaults.
func WithAuthPaths(login, refresh string) Option {
	return func(c *Client) {
		if login != "" {
			c.loginPath = login
		}
		if refresh != "" {
			c.refreshPath = refresh
		}
	}
}

// WithHTTPClient sets a custom HTTP client. A nil client is ignored.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}
