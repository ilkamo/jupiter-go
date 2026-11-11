package transport

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	testTimeout        = 10 * time.Second
	testInitialBackoff = 10 * time.Millisecond
	testMaxBackoff     = 80 * time.Millisecond
	testMaxRetryAfter  = 5 * time.Second
	testMaxAttempts    = 3
	backoffSamples     = 200
)

var testPolicy = RetryPolicy{
	MaxAttempts:    testMaxAttempts,
	InitialBackoff: testInitialBackoff,
	MaxBackoff:     testMaxBackoff,
	MaxRetryAfter:  testMaxRetryAfter,
}

type scriptedResponse struct {
	status     int
	body       string
	retryAfter string
	err        error
}

// scriptedHTTP replays responses in order and records each request body.
type scriptedHTTP struct {
	t         *testing.T
	responses []scriptedResponse
	bodies    []string
}

func (s *scriptedHTTP) Do(r *http.Request) (*http.Response, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	s.bodies = append(s.bodies, string(data))

	if len(s.bodies) > len(s.responses) {
		s.t.Fatalf("unexpected attempt %d", len(s.bodies))
	}

	next := s.responses[len(s.bodies)-1]
	if next.err != nil {
		return nil, next.err
	}

	header := make(http.Header)
	if next.retryAfter != "" {
		header.Set(retryAfterHeader, next.retryAfter)
	}

	return &http.Response{
		StatusCode: next.status,
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(next.body)),
	}, nil
}

type recordedWaits struct {
	delays []time.Duration
}

func (w *recordedWaits) sleep(_ context.Context, delay time.Duration) error {
	w.delays = append(w.delays, delay)

	return nil
}

func newScriptedClient(t *testing.T, responses ...scriptedResponse) (*Client, *scriptedHTTP, *recordedWaits) {
	t.Helper()

	httpClient := &scriptedHTTP{t: t, responses: responses}
	waits := &recordedWaits{}
	client := &Client{
		BaseURL: "https://example.com",
		Timeout: testTimeout,
		HTTP:    httpClient,
		sleep:   waits.sleep,
	}

	return client, httpClient, waits
}

func TestRetryReplaysBodyUntilSuccess(t *testing.T) {
	client, httpClient, waits := newScriptedClient(
		t,
		scriptedResponse{status: http.StatusServiceUnavailable, body: `{"error":"busy"}`},
		scriptedResponse{status: http.StatusTooManyRequests, body: `{"error":"slow down"}`},
		scriptedResponse{status: http.StatusOK, body: `{"value":"done"}`},
	)

	request := Request{
		Method: http.MethodPost,
		Path:   "/",
		Body:   map[string]string{"key": "value"},
		Retry:  testPolicy,
	}

	var result map[string]string
	metadata, err := client.Do(context.Background(), request, &result)
	if err != nil {
		t.Fatal(err)
	}

	if result["value"] != "done" || metadata.Attempts != testMaxAttempts {
		t.Fatalf("result=%v attempts=%d", result, metadata.Attempts)
	}

	for i, body := range httpClient.bodies {
		if body != httpClient.bodies[0] || body == "" {
			t.Fatalf("attempt %d body %q differs from %q", i+1, body, httpClient.bodies[0])
		}
	}

	if len(waits.delays) != testMaxAttempts-1 {
		t.Fatalf("waits = %v", waits.delays)
	}
}

func TestRetryStopsAtMaxAttempts(t *testing.T) {
	failure := scriptedResponse{status: http.StatusBadGateway, body: `{"error":"upstream"}`}
	client, _, _ := newScriptedClient(t, failure, failure, failure)

	var result any
	request := Request{Method: http.MethodGet, Path: "/", Retry: testPolicy}
	metadata, err := client.Do(context.Background(), request, &result)

	apiError, ok := errors.AsType[*APIError](err)
	if !ok || apiError.StatusCode != http.StatusBadGateway {
		t.Fatalf("expected final HTTP error, got %v", err)
	}

	if metadata.Attempts != testMaxAttempts || metadata.StatusCode != http.StatusBadGateway {
		t.Fatalf("unexpected metadata: %+v", metadata)
	}
}

func TestRetryEligibility(t *testing.T) {
	transportFailure := errors.New("connection reset")

	tests := []struct {
		name     string
		first    scriptedResponse
		attempts int
	}{
		{name: "rate limited", first: scriptedResponse{status: http.StatusTooManyRequests}, attempts: 2},
		{
			name:     "internal error",
			first:    scriptedResponse{status: http.StatusInternalServerError},
			attempts: 2,
		},
		{name: "gateway timeout", first: scriptedResponse{status: http.StatusGatewayTimeout}, attempts: 2},
		{name: "transport failure", first: scriptedResponse{err: transportFailure}, attempts: 2},
		{name: "bad request", first: scriptedResponse{status: http.StatusBadRequest}, attempts: 1},
		{name: "unauthorized", first: scriptedResponse{status: http.StatusUnauthorized}, attempts: 1},
		{name: "not implemented", first: scriptedResponse{status: http.StatusNotImplemented}, attempts: 1},
		{name: "redirect", first: scriptedResponse{status: http.StatusFound}, attempts: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, _, _ := newScriptedClient(
				t,
				test.first,
				scriptedResponse{status: http.StatusOK, body: `{}`},
			)

			var result map[string]string
			request := Request{Method: http.MethodGet, Path: "/", Retry: testPolicy}
			metadata, err := client.Do(context.Background(), request, &result)

			if metadata.Attempts != test.attempts {
				t.Fatalf("attempts = %d, want %d (err %v)", metadata.Attempts, test.attempts, err)
			}

			if (err == nil) != (test.attempts > 1) {
				t.Fatalf("unexpected result: %v", err)
			}
		})
	}
}

func TestRetryAfterControlsWait(t *testing.T) {
	const requestedSeconds = 2

	client, _, waits := newScriptedClient(
		t,
		scriptedResponse{
			status:     http.StatusTooManyRequests,
			retryAfter: strconv.Itoa(requestedSeconds),
		},
		scriptedResponse{status: http.StatusOK, body: `{}`},
	)

	var result map[string]string
	request := Request{Method: http.MethodGet, Path: "/", Retry: testPolicy}
	if _, err := client.Do(context.Background(), request, &result); err != nil {
		t.Fatal(err)
	}

	want := requestedSeconds * time.Second
	if len(waits.delays) != 1 || waits.delays[0] != want {
		t.Fatalf("waits = %v, want [%v]", waits.delays, want)
	}
}

func TestRetryAfterBeyondLimitsStopsRetrying(t *testing.T) {
	tests := []struct {
		name       string
		retryAfter time.Duration
		timeout    time.Duration
	}{
		{name: "above policy maximum", retryAfter: testMaxRetryAfter + time.Second, timeout: time.Minute},
		{name: "beyond operation deadline", retryAfter: testMaxRetryAfter, timeout: time.Second},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, _, waits := newScriptedClient(t, scriptedResponse{
				status:     http.StatusServiceUnavailable,
				retryAfter: strconv.Itoa(int(test.retryAfter / time.Second)),
			})
			client.Timeout = test.timeout

			var result any
			request := Request{Method: http.MethodGet, Path: "/", Retry: testPolicy}
			metadata, err := client.Do(context.Background(), request, &result)

			apiError, ok := errors.AsType[*APIError](err)
			if !ok || apiError.StatusCode != http.StatusServiceUnavailable {
				t.Fatalf("expected the server's error, got %v", err)
			}

			if metadata.Attempts != 1 || len(waits.delays) != 0 {
				t.Fatalf("attempts=%d waits=%v", metadata.Attempts, waits.delays)
			}
		})
	}
}

func TestCancellationDuringRetryWait(t *testing.T) {
	client, _, _ := newScriptedClient(t, scriptedResponse{status: http.StatusServiceUnavailable})

	ctx, cancel := context.WithCancel(context.Background())
	client.sleep = func(waitCtx context.Context, delay time.Duration) error {
		cancel()

		return sleepContext(waitCtx, delay)
	}

	var result any
	request := Request{Method: http.MethodGet, Path: "/", Retry: testPolicy}
	_, err := client.Do(ctx, request, &result)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}

	if _, ok := errors.AsType[*APIError](err); !ok {
		t.Fatalf("last response error lost: %v", err)
	}
}

func TestDisableRetriesForcesSingleAttempt(t *testing.T) {
	client, _, _ := newScriptedClient(t, scriptedResponse{status: http.StatusServiceUnavailable})
	client.DisableRetries = true

	var result any
	request := Request{Method: http.MethodGet, Path: "/", Retry: testPolicy}
	metadata, err := client.Do(context.Background(), request, &result)
	if err == nil || metadata.Attempts != 1 {
		t.Fatalf("attempts=%d err=%v", metadata.Attempts, err)
	}
}

func TestRetryPolicyValidation(t *testing.T) {
	tests := []struct {
		name   string
		policy RetryPolicy
		valid  bool
	}{
		{name: "no retry", policy: NoRetry, valid: true},
		{name: "bounded", policy: testPolicy, valid: true},
		{name: "unset", policy: RetryPolicy{}},
		{
			name:   "missing backoff",
			policy: RetryPolicy{MaxAttempts: testMaxAttempts, MaxBackoff: testMaxBackoff},
		},
		{
			name: "maximum below initial",
			policy: RetryPolicy{
				MaxAttempts:    testMaxAttempts,
				InitialBackoff: testMaxBackoff,
				MaxBackoff:     testInitialBackoff,
			},
		},
		{
			name: "negative Retry-After limit",
			policy: RetryPolicy{
				MaxAttempts:    testMaxAttempts,
				InitialBackoff: testInitialBackoff,
				MaxBackoff:     testMaxBackoff,
				MaxRetryAfter:  -time.Second,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.policy.validate(); (err == nil) != test.valid {
				t.Fatalf("validate() = %v", err)
			}
		})
	}
}

func TestInvalidPolicyMakesNoRequest(t *testing.T) {
	client, httpClient, _ := newScriptedClient(t)

	var result any
	request := Request{Method: http.MethodGet, Path: "/"}
	if _, err := client.Do(context.Background(), request, &result); err == nil {
		t.Fatal("unset retry policy accepted")
	}

	if len(httpClient.bodies) != 0 {
		t.Fatal("request sent with an invalid policy")
	}
}

func TestBackoffStaysWithinJitterBounds(t *testing.T) {
	const retriesBeyondCap = 100

	for retry := range retriesBeyondCap {
		ceiling := testInitialBackoff
		for range retry {
			ceiling = min(ceiling*backoffGrowthFactor, testMaxBackoff)
		}

		for range backoffSamples {
			delay := testPolicy.backoff(retry)
			if delay < 0 || delay > ceiling {
				t.Fatalf("retry %d delay %v outside [0, %v]", retry, delay, ceiling)
			}
		}
	}
}

func TestRetryAfterParsing(t *testing.T) {
	now := time.Date(2026, time.November, 11, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		value string
		want  time.Duration
		ok    bool
	}{
		{name: "absent", value: ""},
		{name: "zero seconds", value: "0", want: 0, ok: true},
		{name: "seconds", value: "3", want: 3 * time.Second, ok: true},
		{name: "negative", value: "-1"},
		{name: "fractional", value: "1.5"},
		{name: "malformed", value: "soon"},
		{name: "overflowing seconds", value: "9223372036854775807"},
		{
			name:  "future date",
			value: now.Add(4 * time.Second).Format(http.TimeFormat),
			want:  4 * time.Second,
			ok:    true,
		},
		{name: "past date", value: now.Add(-time.Minute).Format(http.TimeFormat), want: 0, ok: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			header := make(http.Header)
			if test.value != "" {
				header.Set(retryAfterHeader, test.value)
			}

			got, ok := retryAfter(header, now)
			if ok != test.ok || got != test.want {
				t.Fatalf("retryAfter(%q) = %v, %v; want %v, %v", test.value, got, ok, test.want, test.ok)
			}
		})
	}
}
