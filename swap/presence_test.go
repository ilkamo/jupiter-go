package swap

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ilkamo/jupiter-go/internal/transport"
)

type executionFixtureClient struct {
	body string
}

func (c executionFixtureClient) Do(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(c.body)),
	}, nil
}

func TestExecutionRequiresExplicitSuccess(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		success     bool
		codePresent bool
	}{
		{name: "confirmed", body: `{"status":"Success","code":0}`, success: true, codePresent: true},
		{name: "missing code", body: `{"status":"Success"}`},
		{name: "null code", body: `{"status":"Success","code":null}`},
		{name: "missing status", body: `{"code":0}`, codePresent: true},
		{name: "failed with zero", body: `{"status":"Failed","code":0}`, codePresent: true},
		{name: "unknown status", body: `{"status":"Pending","code":0}`, codePresent: true},
		{name: "contradictory code", body: `{"status":"Success","code":-1001}`, codePresent: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := NewClient(&transport.Client{
				BaseURL: "https://example.com",
				Timeout: time.Second,
				HTTP:    executionFixtureClient{body: test.body},
			}, Referral{})

			result, err := client.Execute(context.Background(), ExecuteRequest{
				RequestID:         "order",
				SignedTransaction: "AQ==",
			})
			if test.success {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				var executionError *ExecutionError
				if !errors.As(err, &executionError) {
					t.Fatalf("expected execution error, got %v", err)
				}

				if (executionError.Code != nil) != test.codePresent {
					t.Fatal("error lost code presence")
				}

				if executionError.Message == "" {
					t.Fatal("error has no explanation")
				}
			}

			if result == nil || (result.Code != nil) != test.codePresent {
				t.Fatalf("response lost code presence: %+v", result)
			}
		})
	}
}

func TestOptionalZeroValues(t *testing.T) {
	var order OrderResponse
	if err := json.Unmarshal([]byte(`{}`), &order); err != nil {
		t.Fatal(err)
	}

	if order.Gasless != nil {
		t.Fatal("omitted gasless field acquired a value")
	}

	if err := json.Unmarshal([]byte(`{"gasless":false}`), &order); err != nil {
		t.Fatal(err)
	}

	if order.Gasless == nil || *order.Gasless {
		t.Fatal("explicit false was lost")
	}

	var build BuildResponse
	if err := json.Unmarshal([]byte(`{}`), &build); err != nil {
		t.Fatal(err)
	}

	if build.TransactionVersion != nil || build.BlockhashWithMetadata != nil {
		t.Fatal("omitted build fields acquired values")
	}

	const body = `{
  "transactionVersion": 0,
  "blockhashWithMetadata": {
   "lastValidBlockHeight": 0,
   "fetchedAt": {"secs_since_epoch": 0, "nanos_since_epoch": 0}
  }
 }`
	if err := json.Unmarshal([]byte(body), &build); err != nil {
		t.Fatal(err)
	}

	if build.TransactionVersion == nil || *build.TransactionVersion != 0 {
		t.Fatal("explicit transaction version zero was lost")
	}

	blockhash := build.BlockhashWithMetadata
	if blockhash == nil || blockhash.LastValidBlockHeight == nil || *blockhash.LastValidBlockHeight != 0 {
		t.Fatal("explicit block height zero was lost")
	}

	timestamp := blockhash.FetchedAt
	if timestamp == nil || timestamp.Seconds == nil || timestamp.Nanoseconds == nil {
		t.Fatal("timestamp component presence was lost")
	}
}
