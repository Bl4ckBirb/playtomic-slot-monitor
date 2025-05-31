package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// maxErrorBody bounds a failed response: it is diagnostic, not payload.
const maxErrorBody = 32 << 10

// maxRetryAfterSeconds is the largest whole-second delay a Duration holds.
const maxRetryAfterSeconds = int64(math.MaxInt64) / int64(time.Second)

// queryParams lives here, not in models, to keep transport out of models.
type queryParams interface {
	ToURLValues() url.Values
}

// get fetches path into a T. A nil params arrives non-nil here, hence the
// nil-safe ToURLValues.
func get[T any](ctx context.Context, c *Client, path string, params queryParams) (T, error) {
	var out T
	var query string
	if params != nil {
		query = params.ToURLValues().Encode()
	}

	err := c.sendRequest(ctx, http.MethodGet, path, query, nil, &out)
	return out, err
}

// sendRequest sends a request to the Playtomic API and decodes the response
func (c *Client) sendRequest(ctx context.Context, method, endpoint, query string, body []byte, result any) error {
	reqURL := c.baseURL + endpoint
	if query != "" {
		reqURL += "?" + query
	}

	// Once per call. Per attempt would hit a user's TokenSource maxRetries+1 times.
	var bearer string
	if c.tokenSource != nil {
		token, err := c.tokenSource.Token(ctx)
		if err != nil {
			return fmt.Errorf("acquiring token: %w", err)
		}
		bearer = token
	}

	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader(body))
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}
		c.setHeaders(req, bearer)

		start := time.Now()
		resp, err := c.httpClient.Do(req)
		c.log(ctx, method, reqURL, attempt, start, resp, err)

		final := attempt >= c.maxRetries

		if err != nil {
			if final || !idempotent(method) {
				return fmt.Errorf("sending request after %d attempts: %w", attempt+1, err)
			}
			if err := sleep(ctx, c.backoff(attempt)); err != nil {
				return err
			}
			continue
		}

		if final || !retriable(method, resp.StatusCode) {
			return c.finish(method, reqURL, bearer, resp, result)
		}

		// Honour it in full. Coming back early only deepens the throttle, and
		// a wait we will not sit through goes to the caller to decide.
		wait := retryAfter(resp)
		if wait > c.maxRetryWait {
			return c.finish(method, reqURL, bearer, resp, result)
		}
		if wait == 0 {
			wait = c.backoff(attempt)
		}

		drain(resp)
		if err := sleep(ctx, wait); err != nil {
			return err
		}
	}
}

func (c *Client) setHeaders(req *http.Request, bearer string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", DefaultUserAgent)

	// Last, so precedence is option order.
	for name, values := range c.headers {
		req.Header[name] = values
	}

	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
}

// finish names the bearer so a late 401 for an old token cannot discard the
// newer one another caller just fetched.
func (c *Client) finish(method, url, bearer string, resp *http.Response, result any) error {
	err := decode(method, url, resp, result)

	if errors.Is(err, ErrUnauthorized) {
		if source, ok := c.tokenSource.(interface{ invalidate(string) }); ok {
			source.invalidate(bearer)
		}
	}
	return err
}

// bodyReader gives each attempt a fresh reader. A nil body must stay nil.
func bodyReader(body []byte) io.Reader {
	if body == nil {
		return nil
	}
	return bytes.NewReader(body)
}

// idempotent reports whether a method survives replay.
func idempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut,
		http.MethodDelete, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

// retriable never replays a non-idempotent request. 501 is a permanent refusal
// wearing a 5xx.
func retriable(method string, status int) bool {
	if !idempotent(method) {
		return false
	}
	return status == http.StatusTooManyRequests ||
		(status >= 500 && status != http.StatusNotImplemented)
}

// backoff doubles per attempt and lands in the window's upper half, so a fleet
// does not come back in step. Doubling, not shifting: a shift wraps.
func (c *Client) backoff(attempt int) time.Duration {
	window := c.retryWait
	for i := 0; i < attempt && window > 0 && window < c.maxRetryWait; i++ {
		window *= 2
	}

	if window <= 0 || window > c.maxRetryWait {
		window = c.maxRetryWait
	}
	if window <= 0 {
		return 0
	}
	return window/2 + rand.N(window/2+1)
}

// retryAfter reads the header, which RFC 9110 allows as seconds or an HTTP date.
func retryAfter(resp *http.Response) time.Duration {
	v := resp.Header.Get("Retry-After")
	if v == "" {
		return 0
	}

	if secs, err := strconv.Atoi(v); err == nil {
		switch {
		case secs <= 0:
			return 0
		case int64(secs) > maxRetryAfterSeconds:
			return math.MaxInt64
		default:
			return time.Duration(secs) * time.Second
		}
	}

	if when, err := http.ParseTime(v); err == nil {
		if d := time.Until(when); d > 0 {
			return d
		}
	}

	return 0
}

// drain bounds the read. A larger body never reaches EOF and its connection is
// dropped rather than reused.
func drain(resp *http.Response) {
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	_ = resp.Body.Close()
}

func decode(method, url string, resp *http.Response, result any) error {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		return newError(method, url, resp, body)
	}

	if result == nil {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	// Otherwise a by-ID call hands back a nil and no error.
	if len(bytes.TrimSpace(body)) == 0 {
		if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusResetContent {
			return nil
		}
		return fmt.Errorf("empty body from %s %s (status %d)", method, url, resp.StatusCode)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	return nil
}

func (c *Client) log(ctx context.Context, method, url string, attempt int, start time.Time, resp *http.Response, err error) {
	attrs := []any{
		slog.String("method", method),
		slog.String("url", url),
		slog.Int("attempt", attempt+1),
		slog.Duration("took", time.Since(start)),
	}
	if resp != nil {
		attrs = append(attrs, slog.Int("status", resp.StatusCode))
	}
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
	}

	c.logger.DebugContext(ctx, "playtomic request", attrs...)
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
