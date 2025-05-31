package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

// authServer answers the auth endpoints and one data endpoint, recording what
// it was asked and what bearer it saw.
type authServer struct {
	logins, refreshes atomic.Int32
	bearer            atomic.Value
	loginExpiry       string
}

func (a *authServer) handler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case DefaultLoginPath, DefaultRefreshPath:
			if got := r.Header.Get("Authorization"); got != "" {
				t.Errorf("auth request carried Authorization %q, want none", got)
			}

			var payload map[string]string
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Errorf("auth body %s: %v", body, err)
			}

			expiry := a.loginExpiry
			if r.URL.Path == DefaultLoginPath {
				a.logins.Add(1)
				if payload["email"] != "player@example.com" || payload["password"] != "hunter2" {
					t.Errorf("login payload = %v", payload)
				}
			} else {
				a.refreshes.Add(1)
				if payload["grant_type"] != "refresh_token" {
					t.Errorf("refresh payload = %v", payload)
				}
				expiry = time.Now().UTC().Add(time.Hour).Format("2006-01-02T15:04:05")
			}

			fmt.Fprintf(w, `{"access_token":"token-%d","refresh_token":"refresh-1",`+
				`"access_token_expiration":%q,"user_id":"user-1"}`,
				a.logins.Load()+a.refreshes.Load(), expiry)

		default:
			a.bearer.Store(r.Header.Get("Authorization"))
			w.Write([]byte(`[]`))
		}
	})
}

func TestLogin(t *testing.T) {
	a := &authServer{loginExpiry: time.Now().UTC().Add(time.Hour).Format("2006-01-02T15:04:05")}
	srv := httptest.NewServer(a.handler(t))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	token, err := c.Login(context.Background(), "player@example.com", "hunter2")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if token.AccessToken == "" || token.RefreshToken != "refresh-1" || token.UserID != "user-1" {
		t.Errorf("token = %+v", token)
	}
	if token.Expired() {
		t.Error("a token valid for an hour should not read as expired")
	}
}

func TestWithTokenSetsBearer(t *testing.T) {
	a := &authServer{}
	srv := httptest.NewServer(a.handler(t))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL), WithToken("static-token"))
	if _, err := c.SearchClasses(context.Background(), nil); err != nil {
		t.Fatalf("SearchClasses: %v", err)
	}
	if got := a.bearer.Load(); got != "Bearer static-token" {
		t.Errorf("Authorization = %v", got)
	}
	if a.logins.Load() != 0 {
		t.Error("a static token should not trigger a login")
	}
}

func TestCredentialsLoginOnceThenReuse(t *testing.T) {
	a := &authServer{loginExpiry: time.Now().UTC().Add(time.Hour).Format("2006-01-02T15:04:05")}
	srv := httptest.NewServer(a.handler(t))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL), WithCredentials("player@example.com", "hunter2"))
	for range 3 {
		if _, err := c.SearchClasses(context.Background(), nil); err != nil {
			t.Fatalf("SearchClasses: %v", err)
		}
	}

	if got := a.logins.Load(); got != 1 {
		t.Errorf("%d logins for 3 calls, want 1", got)
	}
	if got := a.bearer.Load(); got != "Bearer token-1" {
		t.Errorf("Authorization = %v", got)
	}
}

func TestCredentialsRefreshWhenExpired(t *testing.T) {
	a := &authServer{loginExpiry: time.Now().UTC().Add(time.Hour).Format("2006-01-02T15:04:05")}
	srv := httptest.NewServer(a.handler(t))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL), WithCredentials("player@example.com", "hunter2"))
	if _, err := c.SearchClasses(context.Background(), nil); err != nil {
		t.Fatalf("SearchClasses: %v", err)
	}

	// Age the cached token rather than have the server issue a dead one, which
	// is a different failure and is rejected outright.
	source := c.tokenSource.(*credentials)
	source.mu.Lock()
	source.token.ExpiresAt = models.Time{Time: time.Now().UTC().Add(-time.Minute)}
	source.mu.Unlock()

	if _, err := c.SearchClasses(context.Background(), nil); err != nil {
		t.Fatalf("SearchClasses after expiry: %v", err)
	}

	if got := a.logins.Load(); got != 1 {
		t.Errorf("%d logins, want 1", got)
	}
	if got := a.refreshes.Load(); got != 1 {
		t.Errorf("%d refreshes, want 1", got)
	}
}

// A token that arrives already expired would loop forever, so it is refused.
func TestCredentialsRejectDeadToken(t *testing.T) {
	a := &authServer{loginExpiry: time.Now().UTC().Add(-time.Hour).Format("2006-01-02T15:04:05")}
	srv := httptest.NewServer(a.handler(t))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL), WithCredentials("player@example.com", "hunter2"))
	if _, err := c.SearchClasses(context.Background(), nil); err == nil {
		t.Fatal("expected a token that is already expired to be refused")
	}
}

func TestTokenExpired(t *testing.T) {
	if !(*Token)(nil).Expired() {
		t.Error("a nil token is expired")
	}
	if !(&Token{}).Expired() {
		t.Error("a token with no access token is expired")
	}

	// No stated expiry means trust it until the API says otherwise.
	if (&Token{AccessToken: "t"}).Expired() {
		t.Error("a token with no expiry should not read as expired")
	}
}

func TestCredentialsInvalidateOn401(t *testing.T) {
	var logins atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == DefaultLoginPath {
			logins.Add(1)
			fmt.Fprintf(w, `{"access_token":"token-%d","access_token_expiration":%q}`,
				logins.Load(), time.Now().UTC().Add(time.Hour).Format("2006-01-02T15:04:05"))
			return
		}
		// The API has revoked it, so the cached token must not be reused.
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL), WithCredentials("player@example.com", "hunter2"))
	for range 2 {
		if _, err := c.SearchClasses(context.Background(), nil); !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("got %v, want ErrUnauthorized", err)
		}
	}

	if got := logins.Load(); got != 2 {
		t.Errorf("%d logins, want 2: a rejected token should be dropped", got)
	}
}

func TestAuthRequestsDropConfiguredAuthorization(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("login carried Authorization %q, want none", got)
		}
		fmt.Fprint(w, `{"access_token":"t"}`)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL), WithHeader("Authorization", "Bearer stale"))
	if _, err := c.Login(context.Background(), "player@example.com", "hunter2"); err != nil {
		t.Fatalf("Login: %v", err)
	}
}

func TestTokenSourceRespectsContextWhileAnotherRenews(t *testing.T) {
	inLogin, release := make(chan struct{}), make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(inLogin)
		<-release
		fmt.Fprint(w, `{"access_token":"t"}`)
	}))
	defer srv.Close()
	defer close(release)

	c := NewClient(WithBaseURL(srv.URL), WithCredentials("player@example.com", "hunter2"))
	source := c.tokenSource

	go func() { _, _ = source.Token(context.Background()) }()

	// Wait until the first caller is inside the login and holding the gate.
	// Signalling before the goroutine starts would let the timed caller take
	// the gate itself and pass without testing anything.
	<-inLogin

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if _, err := source.Token(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("got %v, want the waiter to give up on its own deadline", err)
	}
}

// A 401 arriving late for a superseded token must not discard the replacement.
func TestInvalidateOnlyClearsTheRejectedToken(t *testing.T) {
	s := newCredentials(nil, "player@example.com", "hunter2")
	s.token = &Token{AccessToken: "token-B"}

	s.invalidate("token-A")
	if s.token == nil {
		t.Fatal("a 401 for an old bearer cleared the current token")
	}

	s.invalidate("token-B")
	if s.token != nil {
		t.Error("a 401 for the current bearer should have cleared it")
	}
}
