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

// WithTimeout sets the HTTP client timeout. It is applied after every other
// option, so it holds whether or not WithHTTPClient is also given.
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

// WithLogger logs one record per request attempt at debug level. Headers stay
// out of it, since they carry the bearer token.
func WithLogger(l *slog.Logger) Option {
	return func(c *Client) {
		if l != nil {
			c.logger = l
		}
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

// WithCredentials logs in on the first call that needs a token and refreshes it
// on expiry. The credentials stay in memory for the client's lifetime.
func WithCredentials(email, password string) Option {
	return func(c *Client) {
		c.tokenSource = newCredentials(c, email, password)
	}
}

// WithUserAgent sets a custom User-Agent header. It is a header like any
// other, so the last option to name it wins.
func WithUserAgent(userAgent string) Option {
	return WithHeader("User-Agent", userAgent)
}

// WithAuthPaths overrides the login and refresh endpoints. Empty values keep
// the defaults. Here so a moved endpoint needs no release.
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

// WithHTTPClient sets a custom HTTP client. A nil client is ignored rather
// than left to panic on the first request.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}
