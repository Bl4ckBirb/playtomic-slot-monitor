package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

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

		resp, err := c.httpClient.Do(req)
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
