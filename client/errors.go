package client

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Sentinels for the statuses worth branching on. Match them with errors.Is
// rather than comparing StatusCode.
var (
	ErrBadRequest   = errors.New("bad request")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrNotFound     = errors.New("not found")
	ErrRateLimited  = errors.New("rate limited")
	ErrServer       = errors.New("server error")

	// ErrMissingID guards the by-ID calls, which would otherwise request the
	// collection and decode a list into a single value.
	ErrMissingID = errors.New("missing id")
)

// maxSnippet bounds how much of an unparseable body reaches the error string.
const maxSnippet = 200

// Error is a non-2xx response from the Playtomic API.
type Error struct {
	Method     string
	URL        string
	StatusCode int
	Message    string
	Body       string
	RequestID  string
	RetryAfter time.Duration
	Details    map[string]any
}

func (e *Error) Error() string {
	if e == nil {
		return "playtomic: <nil>"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "playtomic: %s %s: %d ", e.Method, e.URL, e.StatusCode)

	// Falling back to the body matters: an edge that blocks the request never
	// speaks the API's error envelope, and the status alone says nothing.
	b.WriteString(cmp.Or(e.Message, e.Body, http.StatusText(e.StatusCode)))

	if e.RequestID != "" {
		fmt.Fprintf(&b, " (request %s)", e.RequestID)
	}
	return b.String()
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	switch {
	case e.StatusCode == http.StatusBadRequest:
		return ErrBadRequest
	case e.StatusCode == http.StatusUnauthorized:
		return ErrUnauthorized
	case e.StatusCode == http.StatusForbidden:
		return ErrForbidden
	case e.StatusCode == http.StatusNotFound:
		return ErrNotFound
	case e.StatusCode == http.StatusTooManyRequests:
		return ErrRateLimited
	case e.StatusCode >= 500:
		return ErrServer
	default:
		return nil
	}
}

// newError takes the method and URL as arguments rather than reading
// resp.Request, which a custom RoundTripper is free to leave nil.
func newError(method, url string, resp *http.Response, body []byte) *Error {
	e := &Error{
		Method:     method,
		URL:        url,
		StatusCode: resp.StatusCode,
		Body:       snippet(body),
		RequestID:  requestID(resp),
		RetryAfter: retryAfter(resp),
	}

	var payload struct {
		Error   string         `json:"error"`
		Message string         `json:"message"`
		Details map[string]any `json:"details"`
	}
	if json.Unmarshal(body, &payload) == nil {
		e.Message = cmp.Or(payload.Error, payload.Message)
		e.Details = payload.Details
	}

	return e
}

func requestID(resp *http.Response) string {
	for _, h := range []string{"X-Request-Id", "X-Amz-Cf-Id", "X-Amzn-Requestid"} {
		if v := resp.Header.Get(h); v != "" {
			return v
		}
	}
	return ""
}

var htmlTag = regexp.MustCompile(`(?s)<[^>]*>`)

// snippet flattens the body to one line. A body that starts with a tag came
// from something upstream of the API, and only the prose inside it is worth
// keeping.
func snippet(b []byte) string {
	s := strings.TrimSpace(string(b))
	if strings.HasPrefix(s, "<") {
		s = htmlTag.ReplaceAllString(s, " ")
	}

	s = strings.Join(strings.Fields(s), " ")
	if len(s) > maxSnippet {
		s = strings.ToValidUTF8(s[:maxSnippet], "") + "..."
	}
	return s
}
