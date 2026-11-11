package jupiter_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/response"
	"github.com/ilkamo/jupiter-go/swap"
)

const (
	recoveredAttempts = 2
	singleAttempt     = 1
)

// flakyServer fails the first request with the given status, then serves body.
func flakyServer(t *testing.T, failureStatus int, body string) (*httptest.Server, *atomic.Int32) {
	t.Helper()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK
		payload := body
		if calls.Add(1) == 1 {
			status = failureStatus
			payload = `{"error":"transient"}`
		}

		w.WriteHeader(status)

		if _, err := w.Write([]byte(payload)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	return server, &calls
}

func TestReadOperationsRetryTransientFailures(t *testing.T) {
	tests := []struct {
		name string
		body string
		call func(context.Context, *jupiter.Client) (response.Metadata, error)
	}{
		{
			name: "order",
			body: `{"outAmount":"1"}`,
			call: func(ctx context.Context, c *jupiter.Client) (response.Metadata, error) {
				result, err := c.Swap.Order(ctx, swap.OrderRequest{
					InputMint:  wrappedSOLMint,
					OutputMint: usdcMint,
					Amount:     "1",
				})
				if result == nil {
					return response.Metadata{}, err
				}

				return result.Metadata, err
			},
		},
		{
			name: "build",
			body: `{"swapMode":"ExactIn"}`,
			call: func(ctx context.Context, c *jupiter.Client) (response.Metadata, error) {
				result, err := c.Swap.Build(ctx, swap.BuildRequest{
					InputMint:  wrappedSOLMint,
					OutputMint: usdcMint,
					Amount:     "1",
					Taker:      takerAddress,
				})
				if result == nil {
					return response.Metadata{}, err
				}

				return result.Metadata, err
			},
		},
		{
			name: "program labels",
			body: `{}`,
			call: func(ctx context.Context, c *jupiter.Client) (response.Metadata, error) {
				result, err := c.Swap.ProgramLabels(ctx)
				if result == nil {
					return response.Metadata{}, err
				}

				return result.Metadata, err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, calls := flakyServer(t, http.StatusServiceUnavailable, test.body)

			client, err := jupiter.NewClient(jupiter.WithBaseURL(server.URL))
			if err != nil {
				t.Fatal(err)
			}

			metadata, err := test.call(context.Background(), client)
			if err != nil {
				t.Fatal(err)
			}

			if calls.Load() != recoveredAttempts || metadata.Attempts != recoveredAttempts {
				t.Fatalf("calls=%d attempts=%d", calls.Load(), metadata.Attempts)
			}
		})
	}
}

func TestWithoutRetriesMakesOneAttempt(t *testing.T) {
	server, calls := flakyServer(t, http.StatusServiceUnavailable, `{}`)

	client, err := jupiter.NewClient(jupiter.WithBaseURL(server.URL), jupiter.WithoutRetries())
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Swap.ProgramLabels(context.Background())

	apiError, ok := errors.AsType[*jupiter.APIError](err)
	if !ok || apiError.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected unavailable error, got %v", err)
	}

	if calls.Load() != singleAttempt || result.Metadata.Attempts != singleAttempt {
		t.Fatalf("calls=%d attempts=%d", calls.Load(), result.Metadata.Attempts)
	}
}

func TestExecuteFailureClassification(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		unknown bool
	}{
		{
			name:    "server error",
			status:  http.StatusInternalServerError,
			body:    `{"signature":"sig"}`,
			unknown: true,
		},
		{name: "unavailable", status: http.StatusServiceUnavailable, body: `{}`, unknown: true},
		{name: "unreadable success", status: http.StatusOK, body: `not json`, unknown: true},
		{name: "rejected", status: http.StatusBadRequest, body: `{"error":"invalid","code":-2}`},
		{name: "rate limited", status: http.StatusTooManyRequests, body: `{}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(test.status)

				if _, err := w.Write([]byte(test.body)); err != nil {
					t.Errorf("write response: %v", err)
				}
			}))
			defer server.Close()

			client, err := jupiter.NewClient(jupiter.WithBaseURL(server.URL))
			if err != nil {
				t.Fatal(err)
			}

			_, err = client.Swap.Execute(context.Background(), swap.ExecuteRequest{
				RequestID:         "order",
				SignedTransaction: "AQ==",
			})
			if err == nil {
				t.Fatal("expected failure")
			}

			if calls.Load() != singleAttempt {
				t.Fatalf("execute retried: %d calls", calls.Load())
			}

			if errors.Is(err, swap.ErrExecutionOutcomeUnknown) != test.unknown {
				t.Fatalf("unexpected outcome classification: %v", err)
			}

			_, hasAPIError := errors.AsType[*jupiter.APIError](err)
			isHTTPFailure := test.status != http.StatusOK
			if hasAPIError != isHTTPFailure {
				t.Fatalf("HTTP error not preserved: %v", err)
			}
		})
	}
}

func TestExecuteTransportFailureIsUnknown(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	server.Close()

	client, err := jupiter.NewClient(jupiter.WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.Swap.Execute(context.Background(), swap.ExecuteRequest{
		RequestID:         "order",
		SignedTransaction: "AQ==",
	})
	if !errors.Is(err, swap.ErrExecutionOutcomeUnknown) {
		t.Fatalf("expected unknown outcome, got %v", err)
	}
}

func TestExecuteWithEndedContextSendsNothing(t *testing.T) {
	client, err := jupiter.NewClient(jupiter.WithHTTPClient(unreachableHTTPClient{t: t}))
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = client.Swap.Execute(ctx, swap.ExecuteRequest{
		RequestID:         "order",
		SignedTransaction: "AQ==",
	})
	if !errors.Is(err, context.Canceled) || errors.Is(err, swap.ErrExecutionOutcomeUnknown) {
		t.Fatalf("expected known cancellation, got %v", err)
	}
}
