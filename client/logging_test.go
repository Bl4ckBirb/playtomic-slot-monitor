package client

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// capture collects log output so a test can read it back.
type capture struct {
	mu sync.Mutex
	b  strings.Builder
}

func (c *capture) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.b.Write(p)
}

func (c *capture) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.b.String()
}

func TestWithLogger(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	out := &capture{}
	logger := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}))

	c := NewClient(WithBaseURL(srv.URL), WithLogger(logger), WithToken("secret-token"))
	if _, err := c.SearchClasses(context.Background(), nil); err != nil {
		t.Fatalf("SearchClasses: %v", err)
	}

	logged := out.String()
	for _, want := range []string{"playtomic request", "method=GET", "status=200", "attempt=1"} {
		if !strings.Contains(logged, want) {
			t.Errorf("log missing %q:\n%s", want, logged)
		}
	}
	// The bearer must never reach the log.
	if strings.Contains(logged, "secret-token") {
		t.Errorf("log leaked the token:\n%s", logged)
	}
}

func TestDefaultLoggerIsSilent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	// The zero configuration must not need a nil check at any call site.
	c := NewClient(WithBaseURL(srv.URL))
	if _, err := c.SearchClasses(context.Background(), nil); err != nil {
		t.Fatalf("SearchClasses: %v", err)
	}
}
