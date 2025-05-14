package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		if err == nil {
			return decode(resp, result)
		}

		if attempt >= c.maxRetries {
			return fmt.Errorf("sending request after %d attempts: %w", attempt+1, err)
		}

		if err := sleep(ctx, time.Duration(attempt+1)*500*time.Millisecond); err != nil {
			return err
		}
	}
}

// bodyReader hands each attempt a fresh reader over the same bytes. A nil body
// has to stay a nil reader, not a reader over nothing.
func bodyReader(body []byte) io.Reader {
	if body == nil {
		return nil
	}
	return bytes.NewReader(body)
}

func decode(resp *http.Response, result any) error {
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr struct {
			Error   string         `json:"error"`
			Details map[string]any `json:"details"`
		}

		if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Error != "" {
			return &APIError{
				StatusCode: resp.StatusCode,
				Message:    apiErr.Error,
				Details:    apiErr.Details,
			}
		}

		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Unexpected response from API",
		}
	}

	if err := json.Unmarshal(respBody, result); err != nil {
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
