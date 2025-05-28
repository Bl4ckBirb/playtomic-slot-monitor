package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// queryParams is satisfied by the search parameter types in models. Declared
// here rather than there so models stays free of transport concerns.
type queryParams interface {
	ToURLValues() url.Values
}

// get fetches path into a T. params may be a nil pointer inside a non-nil
// interface, which is why every ToURLValues tolerates a nil receiver.
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

	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader(body))
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.userAgent)
		for name, values := range c.headers {
			req.Header[name] = values
		}

		if c.tokenSource != nil {
			token, err := c.tokenSource.Token(ctx)
			if err != nil {
				return fmt.Errorf("acquiring token: %w", err)
			}
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
		}

		start := time.Now()
		resp, err := c.httpClient.Do(req)
		c.log(ctx, method, reqURL, attempt, start, resp, err)

		final := attempt >= c.maxRetries

		if err != nil {
			if final || !idempotent(method) {
				return fmt.Errorf("sending request after %d attempts: %w", attempt+1, err)
			}
			if err := sleep(ctx, c.backoff(attempt, 0)); err != nil {
				return err
			}
			continue
		}

		if final || !retriable(method, resp.StatusCode) {
			return decode(resp, result)
		}

		wait := c.backoff(attempt, retryAfter(resp))
		drain(resp)
		if err := sleep(ctx, wait); err != nil {
			return err
		}
	}
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

// bodyReader gives each attempt a fresh reader. A nil body must stay nil.
func bodyReader(body []byte) io.Reader {
	if body == nil {
		return nil
	}
	return bytes.NewReader(body)
}

// idempotent reports whether a method survives replay after a failure that may
// already have reached the server.
func idempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut,
		http.MethodDelete, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

// retriable: a 429 was refused rather than processed, so any method can go
// again. 501 is a permanent refusal wearing a 5xx.
func retriable(method string, status int) bool {
	if status == http.StatusTooManyRequests {
		return true
	}
	return status >= 500 && status != http.StatusNotImplemented && idempotent(method)
}

// backoff doubles the window per attempt and lands in its upper half, so
// clients that all got the same 503 do not come back in step. Retry-After wins.
func (c *Client) backoff(attempt int, after time.Duration) time.Duration {
	if after > 0 {
		return min(after, c.maxRetryWait)
	}

	window := c.retryWait << attempt
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
		if secs <= 0 {
			return 0
		}
		return time.Duration(secs) * time.Second
	}

	if when, err := http.ParseTime(v); err == nil {
		if d := time.Until(when); d > 0 {
			return d
		}
	}

	return 0
}

// drain returns the connection to the pool instead of dropping it.
func drain(resp *http.Response) {
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	_ = resp.Body.Close()
}

func decode(resp *http.Response, result any) error {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return newError(resp, body)
	}

	if result == nil || len(bytes.TrimSpace(body)) == 0 {
		return nil
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	return nil
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
