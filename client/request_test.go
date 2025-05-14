package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestSendRequestReplaysBodyOnRetry(t *testing.T) {
	var (
		mu     sync.Mutex
		bodies []string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading request body: %v", err)
			return
		}

		mu.Lock()
		bodies = append(bodies, string(b))
		attempt := len(bodies)
		mu.Unlock()

		if attempt == 1 {
			conn, _, err := w.(http.Hijacker).Hijack()
			if err != nil {
				t.Errorf("hijacking connection: %v", err)
				return
			}
			conn.Close()
			return
		}

		if _, err := w.Write([]byte(`{"ok":true}`)); err != nil {
			t.Errorf("writing response: %v", err)
		}
	}))
	defer server.Close()

	c := NewClient(WithBaseURL(server.URL), WithRetries(2))

	var out struct {
		OK bool `json:"ok"`
	}

	body := []byte(`{"hello":"world"}`)
	if err := c.sendRequest(context.Background(), http.MethodPost, "/v1/thing", "", body, &out); err != nil {
		t.Fatalf("sendRequest: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(bodies) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(bodies))
	}
	for i, got := range bodies {
		if got != string(body) {
			t.Errorf("attempt %d sent %q, want %q", i+1, got, body)
		}
	}
	if !out.OK {
		t.Error("expected the retried attempt to decode")
	}
}

func TestSendRequestNegativeRetries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := server.URL
	server.Close()

	c := NewClient(WithBaseURL(url), WithRetries(-1))

	err := c.sendRequest(context.Background(), http.MethodGet, "/v1/thing", "", nil, &struct{}{})
	if err == nil {
		t.Fatal("expected an error against a closed server")
	}
}

func TestSendRequestOmitsEmptyQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("expected no query string, got %q", r.URL.RawQuery)
		}
		if _, err := w.Write([]byte(`{}`)); err != nil {
			t.Errorf("writing response: %v", err)
		}
	}))
	defer server.Close()

	c := NewClient(WithBaseURL(server.URL))

	if err := c.sendRequest(context.Background(), http.MethodGet, "/v1/thing", "", nil, &struct{}{}); err != nil {
		t.Fatalf("sendRequest: %v", err)
	}
}
