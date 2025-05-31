package client

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment variables NewFromEnv reads. Deployment-specific values belong
// here, not in source, so a moved host costs a redeploy and not a release.
const (
	EnvBaseURL   = "PLAYTOMIC_BASE_URL"
	EnvUserAgent = "PLAYTOMIC_USER_AGENT"
	EnvTimeout   = "PLAYTOMIC_TIMEOUT"
	EnvRetries   = "PLAYTOMIC_MAX_RETRIES"
	EnvHeaders   = "PLAYTOMIC_HEADERS"

	// Auth endpoints, in case the API moves them.
	EnvLoginPath   = "PLAYTOMIC_LOGIN_PATH"
	EnvRefreshPath = "PLAYTOMIC_REFRESH_PATH"

	// Credentials. Never commit these.
	EnvAccessToken = "PLAYTOMIC_ACCESS_TOKEN"
	EnvEmail       = "PLAYTOMIC_EMAIL"
	EnvPassword    = "PLAYTOMIC_PASSWORD"
)

// NewFromEnv reads the PLAYTOMIC_* environment then applies opts, so an
// explicit option wins. A set but unusable variable is an error.
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

	if login, refresh := os.Getenv(EnvLoginPath), os.Getenv(EnvRefreshPath); login != "" || refresh != "" {
		env = append(env, WithAuthPaths(login, refresh))
	}

	// A token beats credentials. Half a pair is a mistake worth naming.
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

// parseHeaders reads one "Name: value" per line, so a comma in a value survives.
func parseHeaders(s string) ([]Option, error) {
	var opts []Option

	for line := range strings.SplitSeq(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		name, value, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("%q is not a Name: value pair", line)
		}

		name, value = strings.TrimSpace(name), strings.TrimSpace(value)
		if !validHeaderName(name) {
			return nil, fmt.Errorf("%q is not a valid header name", name)
		}
		if strings.ContainsFunc(value, func(r rune) bool { return (r < ' ' && r != '\t') || r == 0x7f }) {
			return nil, fmt.Errorf("header %s has a control character in its value", name)
		}

		opts = append(opts, WithHeader(name, value))
	}

	return opts, nil
}

// validHeaderName is RFC 9110's token rule, checked at construction.
func validHeaderName(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if r <= ' ' || r >= 0x7f || strings.ContainsRune(":()<>@,;\\\"/[]?={}", r) {
			return false
		}
	}
	return true
}
