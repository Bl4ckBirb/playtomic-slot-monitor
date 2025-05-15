package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
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

func TestOmitsEmptyQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("got query %q, want none", r.URL.RawQuery)
		}
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	if err := c.sendRequest(context.Background(), http.MethodGet, "/v1/thing", "", nil, &struct{}{}); err != nil {
		t.Fatalf("sendRequest: %v", err)
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
		{"429 on POST", http.MethodPost, http.StatusTooManyRequests, 3},
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

func TestRetryAfterIsCapped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL), WithRetries(1), fast())

	start := time.Now()
	if err := c.sendRequest(context.Background(), http.MethodGet, "/v1/thing", "", nil, &struct{}{}); err == nil {
		t.Fatal("expected an error")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Errorf("waited %v, expected the minute to be capped at maxRetryWait", elapsed)
	}
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

func TestBackoffStaysInsideItsWindow(t *testing.T) {
	c := NewClient(WithBackoff(100*time.Millisecond, 2*time.Second))

	for attempt := range 8 {
		window := min(c.retryWait<<attempt, c.maxRetryWait)
		for range 50 {
			if got := c.backoff(attempt, 0); got < window/2 || got > window {
				t.Fatalf("attempt %d: %v outside [%v, %v]", attempt, got, window/2, window)
			}
		}
	}
}
