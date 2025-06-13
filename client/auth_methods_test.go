package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAuthMethods(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/auth/methods" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("email"); got != "player@example.com" {
			t.Errorf("email = %q", got)
		}
		// No token needed for this one.
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("auth-methods carried Authorization %q", got)
		}
		w.Write([]byte(`{"available_methods":[{"type":"PASSWORD"}],"action_required":false}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	m, err := c.GetAuthMethods(context.Background(), "player@example.com")
	if err != nil {
		t.Fatalf("GetAuthMethods: %v", err)
	}
	if len(m.AvailableMethods) != 1 || m.AvailableMethods[0].Type != "PASSWORD" {
		t.Errorf("methods = %+v", m)
	}
	if m.ActionRequired {
		t.Errorf("action_required = %v, want false", m.ActionRequired)
	}

	if _, err := c.GetAuthMethods(context.Background(), ""); !errors.Is(err, ErrMissingID) {
		t.Errorf("empty email gave %v", err)
	}
}

// A JSON null body must not come back as a nil methods with no error.
func TestGetAuthMethodsNullResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`null`))
	}))
	defer srv.Close()

	if _, err := NewClient(WithBaseURL(srv.URL)).GetAuthMethods(context.Background(), "x@example.com"); err == nil {
		t.Fatal("a null body should be an error, not a nil result")
	}
}

// The auth-method check must not carry a bearer, even under WithCredentials or a
// configured Authorization header.
func TestGetAuthMethodsSendsNoBearer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Presence, not value: the bug this guards leaves the header set but
		// empty, which Header.Get cannot tell from absent.
		if _, ok := r.Header["Authorization"]; ok {
			t.Errorf("Authorization header present, want it dropped: %q", r.Header.Get("Authorization"))
		}
		w.Write([]byte(`{"available_methods":[{"type":"PASSWORD"}]}`))
	}))
	defer srv.Close()

	c := NewClient(
		WithBaseURL(srv.URL),
		WithToken("should-not-be-sent"),
		WithHeader("Authorization", "Bearer nor-this"),
	)
	if _, err := c.GetAuthMethods(context.Background(), "x@example.com"); err != nil {
		t.Fatalf("GetAuthMethods: %v", err)
	}

	// Even an empty-valued Authorization header must be dropped, not shared
	// through from the client's header map.
	c2 := NewClient(WithBaseURL(srv.URL), WithHeader("Authorization", ""))
	if _, err := c2.GetAuthMethods(context.Background(), "x@example.com"); err != nil {
		t.Fatalf("GetAuthMethods (empty auth): %v", err)
	}
}
