package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type APIError struct {
	StatusCode int
	RequestID  string
	Code       json.RawMessage
	Message    string
	Body       json.RawMessage
	Cause      error
}

func (e *APIError) Error() string {
	return fmt.Sprintf("jupiter: HTTP %d: %s", e.StatusCode, e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Cause
}

type Client struct {
	BaseURL string
	APIKey  string
	HTTP    HTTPClient
	Timeout time.Duration
}

// Do never retries: an ambiguous execution response must not create another order.
func (c *Client) Do(
	ctx context.Context,
	method string,
	path string,
	query url.Values,
	body any,
	result any,
) (resultErr error) {
	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	var payload bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&payload).Encode(body); err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, &payload)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.URL.RawQuery = query.Encode()
	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.APIKey != "" {
		req.Header.Set("x-api-key", c.APIKey)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("jupiter request: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close response body: %w", err))
		}
	}()

	data, err := readResponseBody(resp.Body, maxResponseBodyBytes)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		var detail struct {
			Error   string          `json:"error"`
			Message string          `json:"message"`
			Code    json.RawMessage `json:"code"`
		}

		decodeErr := json.Unmarshal(data, &detail)
		if decodeErr != nil {
			return &APIError{
				StatusCode: resp.StatusCode,
				RequestID:  resp.Header.Get("x-request-id"),
				Message:    http.StatusText(resp.StatusCode),
				Body:       append(json.RawMessage(nil), data...),
				Cause:      fmt.Errorf("decode error response: %w", decodeErr),
			}
		}

		if detail.Error == "" {
			detail.Error = detail.Message
		}

		if detail.Error == "" {
			detail.Error = http.StatusText(resp.StatusCode)
		}

		return &APIError{
			StatusCode: resp.StatusCode,
			RequestID:  resp.Header.Get("x-request-id"),
			Code:       detail.Code,
			Message:    detail.Error,
			Body:       append(json.RawMessage(nil), data...),
		}
	}

	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}
