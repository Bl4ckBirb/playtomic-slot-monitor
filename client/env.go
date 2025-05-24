package client

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment variables NewFromEnv reads. Anything deployment-specific belongs
// here rather than in the source, so a moved host or a newly required header
// costs a redeploy and not a release.
const (
	EnvBaseURL   = "PLAYTOMIC_BASE_URL"
	EnvUserAgent = "PLAYTOMIC_USER_AGENT"
	EnvTimeout   = "PLAYTOMIC_TIMEOUT"
	EnvRetries   = "PLAYTOMIC_MAX_RETRIES"
	EnvHeaders   = "PLAYTOMIC_HEADERS"

	// Credentials. Never commit these.
	EnvAccessToken = "PLAYTOMIC_ACCESS_TOKEN"
	EnvEmail       = "PLAYTOMIC_EMAIL"
	EnvPassword    = "PLAYTOMIC_PASSWORD"
)

// NewFromEnv builds a client from the PLAYTOMIC_* environment and applies opts
// on top, so an explicit option still wins. An unset variable keeps its
// default. A set but unusable one is an error, not a silent fallback.
func NewFromEnv(opts ...Option) (*Client, error) {
	var env []Option

	if v := os.Getenv(EnvBaseURL); v != "" {
		env = append(env, WithBaseURL(v))
	}
	if v := os.Getenv(EnvUserAgent); v != "" {
		env = append(env, WithUserAgent(v))
	}

	if v := os.Getenv(EnvTimeout); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", EnvTimeout, err)
		}
		env = append(env, WithTimeout(d))
	}

	if v := os.Getenv(EnvRetries); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", EnvRetries, err)
		}
		env = append(env, WithRetries(n))
	}

	if v := os.Getenv(EnvHeaders); v != "" {
		headers, err := parseHeaders(v)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", EnvHeaders, err)
		}
		env = append(env, headers...)
	}

	// A token beats credentials, since it costs no round trip. Half a
	// credential pair is a mistake worth naming rather than ignoring.
	switch token, email, password := os.Getenv(EnvAccessToken), os.Getenv(EnvEmail), os.Getenv(EnvPassword); {
	case token != "":
		env = append(env, WithToken(token))
	case email != "" && password != "":
		env = append(env, WithCredentials(email, password))
	case email != "" || password != "":
		return nil, fmt.Errorf("%s and %s must be set together", EnvEmail, EnvPassword)
	}

	return NewClient(append(env, opts...)...), nil
}

// parseHeaders reads "Name: value" pairs separated by newlines or commas, which
// is close enough to how a proxy prints them to paste straight in.
func parseHeaders(s string) ([]Option, error) {
	var opts []Option

	for _, field := range strings.FieldsFunc(s, func(r rune) bool { return r == '\n' || r == ',' }) {
		name, value, ok := strings.Cut(field, ":")
		if !ok {
			return nil, fmt.Errorf("%q is not a Name: value pair", strings.TrimSpace(field))
		}

		name = strings.TrimSpace(name)
		if name == "" {
			return nil, fmt.Errorf("empty header name in %q", strings.TrimSpace(field))
		}
		opts = append(opts, WithHeader(name, strings.TrimSpace(value)))
	}

	return opts, nil
}
