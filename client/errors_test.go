package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// cloudfront is what the edge returns when it blocks a request outright. It is
// not the API's error envelope, so nothing structured can be pulled from it.
const cloudfront = `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN">
<HTML><HEAD><TITLE>ERROR: The request could not be satisfied</TITLE></HEAD><BODY>
<H1>403 ERROR</H1><H2>The request could not be satisfied.</H2>
Request blocked.
</BODY></HTML>`

func TestErrorSentinels(t *testing.T) {
	tests := []struct {
		status int
		want   error
	}{
		{http.StatusUnauthorized, ErrUnauthorized},
		{http.StatusForbidden, ErrForbidden},
		{http.StatusNotFound, ErrNotFound},
		{http.StatusTooManyRequests, ErrRateLimited},
		{http.StatusInternalServerError, ErrServer},
		{http.StatusBadGateway, ErrServer},
	}

	for _, tt := range tests {
		err := &Error{StatusCode: tt.status}
		if !errors.Is(err, tt.want) {
			t.Errorf("status %d: errors.Is(%v) = false", tt.status, tt.want)
		}
	}

	if got := (&Error{StatusCode: http.StatusBadRequest}).Unwrap(); got != nil {
		t.Errorf("400 unwrapped to %v, want nil", got)
	}
}

func TestErrorCarriesBlockedBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Amz-Cf-Id", "GLRveArDyxQX1uyNVrl1bgs8T0bHcBsP")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(cloudfront))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	err := c.sendRequest(context.Background(), http.MethodGet, "/v1/matches", "size=1", nil, &struct{}{})

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("got %T, want *Error", err)
	}
	if !errors.Is(err, ErrForbidden) {
		t.Error("a 403 should match ErrForbidden")
	}
	if apiErr.RequestID != "GLRveArDyxQX1uyNVrl1bgs8T0bHcBsP" {
		t.Errorf("RequestID = %q", apiErr.RequestID)
	}

	// The whole point: the message names the failure instead of swallowing it.
	msg := err.Error()
	for _, want := range []string{"403", "Request blocked.", "/v1/matches?size=1", "request GLRve"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error string missing %q:\n%s", want, msg)
		}
	}
	if strings.Contains(msg, "\n") {
		t.Errorf("error string should stay one line:\n%s", msg)
	}
}

func TestErrorUsesAPIEnvelopeWhenPresent(t *testing.T) {
	srv := httptest.NewServer(func() http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"invalid tenant_id","details":{"field":"tenant_id"}}`))
		}
	}())
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	err := c.sendRequest(context.Background(), http.MethodGet, "/v1/classes", "", nil, &struct{}{})

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("got %T, want *Error", err)
	}
	if apiErr.Message != "invalid tenant_id" {
		t.Errorf("Message = %q", apiErr.Message)
	}
	if apiErr.Details["field"] != "tenant_id" {
		t.Errorf("Details = %v", apiErr.Details)
	}
}

func TestSnippetTruncates(t *testing.T) {
	got := snippet([]byte(strings.Repeat("x", maxSnippet+50)))
	if len(got) != maxSnippet+3 {
		t.Errorf("len = %d, want %d", len(got), maxSnippet+3)
	}
	if !strings.HasSuffix(got, "...") {
		t.Error("want an elision marker")
	}
}

func TestDecodeAcceptsEmptyBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	if err := c.sendRequest(context.Background(), http.MethodGet, "/v1/thing", "", nil, &struct{}{}); err != nil {
		t.Errorf("204 should not be an error, got %v", err)
	}
}
