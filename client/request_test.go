package client

import (
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func fast() Option { return WithBackoff(time.Millisecond, 5*time.Millisecond) }

func TestReplaysBodyOnRetry(t *testing.T) {
	var (
		mu     sync.Mutex
		bodies []string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)

		mu.Lock()
		bodies = append(bodies, string(b))
		first := len(bodies) == 1
		mu.Unlock()

		if first {
			conn, _, _ := w.(http.Hijacker).Hijack()
			conn.Close()
			return
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	var out struct {
		OK bool `json:"ok"`
	}
	body := []byte(`{"hello":"world"}`)
	c := NewClient(WithBaseURL(srv.URL), WithRetries(2), fast())

	if err := c.sendRequest(context.Background(), http.MethodPut, "/v1/thing", "", body, &out); err != nil {
		t.Fatalf("sendRequest: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(bodies) != 2 || bodies[0] != string(body) || bodies[1] != string(body) {
		t.Errorf("attempts sent %q, want the same body twice", bodies)
	}
	if !out.OK {
		t.Error("retried attempt did not decode")
	}
}

func TestNegativeRetriesDoesNotPanic(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	srv.Close()

	c := NewClient(WithBaseURL(url), WithRetries(-1))
	if err := c.sendRequest(context.Background(), http.MethodGet, "/v1/thing", "", nil, &struct{}{}); err == nil {
		t.Fatal("expected an error against a closed server")
	}
}

// A nil *SearchClassesParams arrives as a non-nil interface holding a nil
// pointer, so this exercises the nil receiver as well as the empty query.
func TestNilParamsSendNoQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("got query %q, want none", r.URL.RawQuery)
		}
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	if _, err := c.SearchClasses(context.Background(), nil); err != nil {
		t.Fatalf("SearchClasses(nil): %v", err)
	}
}

func TestRetryPolicy(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		status   int
		attempts int32
	}{
		{"503 on GET", http.MethodGet, http.StatusServiceUnavailable, 3},
		{"429 on GET", http.MethodGet, http.StatusTooManyRequests, 3},
		{"429 on POST", http.MethodPost, http.StatusTooManyRequests, 1},
		{"503 on POST", http.MethodPost, http.StatusServiceUnavailable, 1},
		{"501 on GET", http.MethodGet, http.StatusNotImplemented, 1},
		{"400 on GET", http.MethodGet, http.StatusBadRequest, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			c := NewClient(WithBaseURL(srv.URL), WithRetries(2), fast())
			if err := c.sendRequest(context.Background(), tt.method, "/v1/thing", "", nil, &struct{}{}); err == nil {
				t.Fatalf("expected an error for %d", tt.status)
			}
			if got := calls.Load(); got != tt.attempts {
				t.Errorf("server saw %d attempts, want %d", got, tt.attempts)
			}
		})
	}
}

// A Retry-After longer than we are willing to wait is a refusal, not a retry:
// coming back early would only deepen the throttle.
func TestLongRetryAfterGivesUp(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL), WithRetries(3), fast())

	start := time.Now()
	err := c.sendRequest(context.Background(), http.MethodGet, "/v1/thing", "", nil, &struct{}{})
	elapsed := time.Since(start)

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("got %T, want *Error", err)
	}
	if apiErr.RetryAfter != time.Minute {
		t.Errorf("RetryAfter = %v, want exactly what the server asked for", apiErr.RetryAfter)
	}
	if calls.Load() != 1 {
		t.Errorf("%d attempts, want 1", calls.Load())
	}
	if elapsed > time.Second {
		t.Errorf("waited %v, should not have waited at all", elapsed)
	}
}

// An empty 200 used to decode to a zero value and come back as success.
func TestEmptySuccessBodyIsAnError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	if _, err := c.GetTenant(context.Background(), "tenant-1"); err == nil {
		t.Fatal("an empty 200 should not read as a tenant")
	}
}

// A custom RoundTripper may leave resp.Request nil, which newError used to read.
func TestErrorWithoutRequestOnResponse(t *testing.T) {
	c := NewClient(
		WithBaseURL("https://example.test"),
		WithHTTPClient(&http.Client{Transport: bareTransport{}}),
	)

	_, err := c.SearchClasses(context.Background(), nil)

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("got %T, want *Error", err)
	}
	if apiErr.Method != http.MethodGet || apiErr.URL != "https://example.test/v1/classes" {
		t.Errorf("method = %q, url = %q", apiErr.Method, apiErr.URL)
	}
}

// bareTransport answers without setting Response.Request, which the stdlib
// transport does but a third-party one need not.
type bareTransport struct{}

func (bareTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusForbidden,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader("nope")),
	}, nil
}

func TestCancelledContextStopsRetrying(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := NewClient(WithBaseURL(srv.URL), WithRetries(5), fast())
	if err := c.sendRequest(ctx, http.MethodGet, "/v1/thing", "", nil, &struct{}{}); err == nil {
		t.Fatal("expected a context error")
	}
}

func TestRetryAfter(t *testing.T) {
	future := time.Now().Add(30 * time.Second).UTC().Format(http.TimeFormat)

	tests := []struct {
		header string
		want   time.Duration
	}{
		{"", 0},
		{"30", 30 * time.Second},
		{"0", 0},
		{"-5", 0},
		{"soon", 0},
		{"172800", 48 * time.Hour},
		{"99999999999999", math.MaxInt64},
		{"Mon, 02 Jan 2006 15:04:05 GMT", 0},
	}

	for _, tt := range tests {
		resp := &http.Response{Header: http.Header{"Retry-After": {tt.header}}}
		if got := retryAfter(resp); got != tt.want {
			t.Errorf("retryAfter(%q) = %v, want %v", tt.header, got, tt.want)
		}
	}

	resp := &http.Response{Header: http.Header{"Retry-After": {future}}}
	if got := retryAfter(resp); got < 25*time.Second || got > 30*time.Second {
		t.Errorf("retryAfter(%q) = %v, want roughly 30s", future, got)
	}
}

// A window that would overflow while doubling must land on the cap, not on
// whatever a wrapped value happens to be. The old shift could wrap to a small
// positive and pass a naive bounds check.
func TestBackoffDoesNotOverflow(t *testing.T) {
	c := NewClient(WithBackoff((1<<62)+1, math.MaxInt64))

	for range 50 {
		if got := c.backoff(2); got < c.maxRetryWait/2 {
			t.Fatalf("backoff %v below half the cap %v", got, c.maxRetryWait)
		}
	}
}

func TestBackoffStaysInsideItsWindow(t *testing.T) {
	c := NewClient(WithBackoff(100*time.Millisecond, 2*time.Second))

	for attempt := range 8 {
		window := c.retryWait
		for range attempt {
			window = min(window*2, c.maxRetryWait)
		}
		for range 50 {
			if got := c.backoff(attempt); got < window/2 || got > window {
				t.Fatalf("attempt %d: %v outside [%v, %v]", attempt, got, window/2, window)
			}
		}
	}
}
