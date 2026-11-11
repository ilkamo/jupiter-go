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

	"github.com/ilkamo/jupiter-go/response"
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
	Headers    http.Header
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

	// Timeout bounds a whole operation, including retries and the waits between them.
	Timeout time.Duration

	// DisableRetries forces one attempt regardless of the operation's policy.
	DisableRetries bool

	// sleep waits between attempts; nil selects a context-aware timer.
	sleep func(context.Context, time.Duration) error
}

// Request describes one API operation. Retry must reflect the operation's replay
// semantics, not only its HTTP method.
type Request struct {
	Method string
	Path   string
	Query  url.Values
	Body   any
	Retry  RetryPolicy

	// Header adds request-specific headers, such as a wallet session's Authorization.
	Header http.Header
}

type attemptOutcome struct {
	metadata  response.Metadata
	data      []byte
	err       error
	retryable bool
}

func (c *Client) Do(
	ctx context.Context,
	request Request,
	result any,
) (response.Metadata, error) {
	var metadata response.Metadata

	policy := request.Retry
	if c.DisableRetries {
		policy = NoRetry
	}

	if err := policy.validate(); err != nil {
		return metadata, err
	}

	ctx, cancel := context.WithTimeout(ctx, c.Timeout)
	defer cancel()

	var payload []byte
	if request.Body != nil {
		encoded, err := json.Marshal(request.Body)
		if err != nil {
			return metadata, fmt.Errorf("encode request: %w", err)
		}

		payload = encoded
	}

	for attempt := 1; ; attempt++ {
		outcome := c.attempt(ctx, request, payload)
		outcome.metadata.Attempts = attempt

		if outcome.err == nil {
			if err := json.Unmarshal(outcome.data, result); err != nil {
				return outcome.metadata, fmt.Errorf("decode response: %w", err)
			}

			return outcome.metadata, nil
		}

		if !outcome.retryable || attempt >= policy.MaxAttempts {
			return outcome.metadata, outcome.err
		}

		delay, ok := retryDelay(ctx, policy, attempt, outcome.metadata.Headers)
		if !ok {
			return outcome.metadata, outcome.err
		}

		if err := c.wait(ctx, delay); err != nil {
			return outcome.metadata, errors.Join(outcome.err, fmt.Errorf("wait before retry: %w", err))
		}
	}
}

// retryDelay chooses the wait before the next attempt. It declines to retry when the
// server asks for a longer wait than the policy accepts or than the deadline allows.
func retryDelay(
	ctx context.Context,
	policy RetryPolicy,
	completedAttempts int,
	headers http.Header,
) (time.Duration, bool) {
	delay, requested := retryAfter(headers, time.Now())
	if requested && delay > policy.MaxRetryAfter {
		return 0, false
	}

	if !requested {
		delay = policy.backoff(completedAttempts - 1)
	}

	deadline, hasDeadline := ctx.Deadline()
	if hasDeadline && time.Until(deadline) <= delay {
		return 0, false
	}

	return delay, true
}

func (c *Client) wait(ctx context.Context, delay time.Duration) error {
	if c.sleep != nil {
		return c.sleep(ctx, delay)
	}

	return sleepContext(ctx, delay)
}

func (c *Client) attempt(
	ctx context.Context,
	request Request,
	payload []byte,
) (outcome attemptOutcome) {
	req, err := http.NewRequestWithContext(
		ctx,
		request.Method,
		c.BaseURL+request.Path,
		bytes.NewReader(payload),
	)
	if err != nil {
		outcome.err = fmt.Errorf("create request: %w", err)

		return outcome
	}

	req.URL.RawQuery = request.Query.Encode()

	for name, values := range request.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	req.Header.Set("Accept", "application/json")

	if request.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.APIKey != "" {
		req.Header.Set("x-api-key", c.APIKey)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		outcome.err = fmt.Errorf("jupiter request: %w", err)
		outcome.retryable = ctx.Err() == nil

		return outcome
	}

	outcome.metadata = response.Metadata{
		StatusCode: resp.StatusCode,
		RequestID:  resp.Header.Get("x-request-id"),
		Headers:    resp.Header.Clone(),
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			outcome.err = errors.Join(outcome.err, fmt.Errorf("close response body: %w", err))
		}
	}()

	data, err := readResponseBody(resp.Body, maxResponseBodyBytes)
	if err != nil {
		outcome.err = fmt.Errorf("read response: %w", err)

		return outcome
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		outcome.err = newAPIError(resp, data)
		outcome.retryable = retryableStatus(resp.StatusCode)

		return outcome
	}

	outcome.data = data

	return outcome
}

func newAPIError(resp *http.Response, data []byte) *APIError {
	apiError := &APIError{
		Headers:    resp.Header.Clone(),
		StatusCode: resp.StatusCode,
		RequestID:  resp.Header.Get("x-request-id"),
		Message:    http.StatusText(resp.StatusCode),
		Body:       append(json.RawMessage(nil), data...),
	}

	var detail struct {
		Error   string          `json:"error"`
		Message string          `json:"message"`
		Code    json.RawMessage `json:"code"`
	}

	if err := json.Unmarshal(data, &detail); err != nil {
		apiError.Cause = fmt.Errorf("decode error response: %w", err)

		return apiError
	}

	apiError.Code = detail.Code

	switch {
	case detail.Error != "":
		apiError.Message = detail.Error
	case detail.Message != "":
		apiError.Message = detail.Message
	}

	return apiError
}

// MayHaveBeenProcessed reports whether a failed request might still have taken
// effect upstream. Only an HTTP response below 500 proves the server rejected it;
// transport failures, timeouts, server errors, and unreadable success responses
// leave the outcome unknown.
func MayHaveBeenProcessed(err error) bool {
	if apiError, ok := errors.AsType[*APIError](err); ok {
		return apiError.StatusCode >= http.StatusInternalServerError
	}

	return true
}
