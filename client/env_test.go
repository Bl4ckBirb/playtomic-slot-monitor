package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewFromEnv(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://example.test")
	t.Setenv(EnvUserAgent, "padel-bot/2.0")
	t.Setenv(EnvTimeout, "5s")
	t.Setenv(EnvRetries, "7")
	t.Setenv(EnvHeaders, "X-App-Token: abc123\nX-Client: ios")

	c, err := NewFromEnv()
	if err != nil {
		t.Fatalf("NewFromEnv: %v", err)
	}

	if c.baseURL != "https://example.test" || c.headers.Get("User-Agent") != "padel-bot/2.0" {
		t.Errorf("baseURL = %q, userAgent = %q", c.baseURL, c.headers.Get("User-Agent"))
	}
	if c.httpClient.Timeout != 5*time.Second || c.maxRetries != 7 {
		t.Errorf("timeout = %v, retries = %d", c.httpClient.Timeout, c.maxRetries)
	}
	if got := c.headers.Get("X-App-Token"); got != "abc123" {
		t.Errorf("X-App-Token = %q", got)
	}
	if got := c.headers.Get("X-Client"); got != "ios" {
		t.Errorf("X-Client = %q", got)
	}
}

func TestNewFromEnvExplicitOptionWins(t *testing.T) {
	t.Setenv(EnvBaseURL, "https://from-env.test")

	c, err := NewFromEnv(WithBaseURL("https://explicit.test"))
	if err != nil {
		t.Fatalf("NewFromEnv: %v", err)
	}
	if c.baseURL != "https://explicit.test" {
		t.Errorf("baseURL = %q, want the explicit option to win", c.baseURL)
	}
}

func TestNewFromEnvRejectsBadValues(t *testing.T) {
	tests := []struct{ name, key, value string }{
		{"timeout", EnvTimeout, "ages"},
		{"retries", EnvRetries, "lots"},
		{"headers", EnvHeaders, "no colon here"},
		{"header name", EnvHeaders, ": orphaned"},
		{"header name token", EnvHeaders, "Bad Name: value"},
		{"header value control", EnvHeaders, "X-Ok: va\x7flue"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.key, tt.value)
			if _, err := NewFromEnv(); err == nil {
				t.Errorf("%s=%q should not be accepted", tt.key, tt.value)
			}
		})
	}
}

func TestHeadersReachTheRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-App-Token"); got != "abc123" {
			t.Errorf("X-App-Token = %q", got)
		}
		// A configured header has to beat the built-in default.
		if got := r.Header.Get("User-Agent"); got != "override/1.0" {
			t.Errorf("User-Agent = %q", got)
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := NewClient(
		WithBaseURL(srv.URL),
		WithHeader("X-App-Token", "abc123"),
		WithHeader("User-Agent", "override/1.0"),
	)
	if _, err := c.SearchClasses(context.Background(), nil); err != nil {
		t.Fatalf("SearchClasses: %v", err)
	}
}

func TestUserAgentPrecedenceIsOptionOrder(t *testing.T) {
	agent := func(opts ...Option) string {
		return NewClient(opts...).headers.Get("User-Agent")
	}

	if got := agent(WithUserAgent("first"), WithHeader("User-Agent", "second")); got != "second" {
		t.Errorf("User-Agent = %q, want the later option", got)
	}
	if got := agent(WithHeader("User-Agent", "first"), WithUserAgent("second")); got != "second" {
		t.Errorf("User-Agent = %q, want the later option", got)
	}
}

func TestExplicitOptionBeatsEnvHeader(t *testing.T) {
	t.Setenv(EnvHeaders, "User-Agent: from-env")

	c, err := NewFromEnv(WithUserAgent("explicit"))
	if err != nil {
		t.Fatalf("NewFromEnv: %v", err)
	}
	if got := c.headers.Get("User-Agent"); got != "explicit" {
		t.Errorf("User-Agent = %q, want the explicit option to win", got)
	}
}

// One header per line is what lets a value carry a comma.
func TestHeaderValueKeepsCommas(t *testing.T) {
	t.Setenv(EnvHeaders, "X-List: a,b\nX-Other: c")

	c, err := NewFromEnv()
	if err != nil {
		t.Fatalf("NewFromEnv: %v", err)
	}
	if got := c.headers.Get("X-List"); got != "a,b" {
		t.Errorf("X-List = %q, want it kept whole", got)
	}
	if got := c.headers.Get("X-Other"); got != "c" {
		t.Errorf("X-Other = %q", got)
	}
}

// HTAB is legal inside a field value, unlike the other control characters.
func TestHeaderValueAllowsTab(t *testing.T) {
	t.Setenv(EnvHeaders, "X-Spaced: a\tb")

	if _, err := NewFromEnv(); err != nil {
		t.Errorf("an internal tab should be accepted: %v", err)
	}
}
