package jupiter_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/swap"
)

func TestOrderContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/swap/v2/order" || r.Header.Get("x-api-key") != "secret" {
			t.Error("incorrect path or authentication")
		}

		q := r.URL.Query()
		if q.Get("amount") != "9007199254740993" ||
			q.Get("referralFee") != "50" {
			t.Error("amount or referral lost")
		}

		if _, err := w.Write([]byte(`{
  "outAmount": "9007199254740994",
  "transaction": null,
  "feeBps": 50,
  "inUsdValue": 0.1234567890123456789
}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))

	defer server.Close()

	c, err := jupiter.NewClient(
		jupiter.WithBaseURL(server.URL),
		jupiter.WithAPIKey("secret"),
		jupiter.WithReferral("account", 50),
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := c.Swap.Order(
		context.Background(),
		swap.OrderRequest{
			InputMint:  "in",
			OutputMint: "out",
			Amount:     "9007199254740993",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.Transaction != nil || result.InUSDValue.String() != "0.1234567890123456789" {
		t.Fatalf("precision or null lost: %+v", result)
	}
}

func TestErrorsAndNoRetry(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{
			name:       "execution failure",
			statusCode: http.StatusOK,
			body: `{
  "status": "Failed",
  "code": -1001,
  "error": "unknown",
  "signature": "sig"
}`,
		},
		{
			name:       "HTTP failure",
			statusCode: http.StatusForbidden,
			body:       `{"error":"denied","code":42}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.WriteHeader(test.statusCode)

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
				RequestID:         "id",
				SignedTransaction: "AQ==",
			})
			if err == nil {
				t.Fatal("expected execution to fail")
			}

			const expectedAttempts = 1
			if calls != expectedAttempts {
				t.Fatalf("expected %d attempt, got %d", expectedAttempts, calls)
			}

			if test.statusCode == http.StatusOK {
				var executionError *swap.ExecutionError
				if !errors.As(err, &executionError) {
					t.Fatalf("expected execution error, got %v", err)
				}

				return
			}

			var apiError *jupiter.APIError
			if !errors.As(err, &apiError) || apiError.StatusCode != test.statusCode {
				t.Fatalf("expected HTTP error with status %d, got %v", test.statusCode, err)
			}
		})
	}
}

func TestConfiguration(t *testing.T) {
	tests := [][]jupiter.Option{
		{jupiter.WithTimeout(0)},
		{jupiter.WithBaseURL("ftp://host")},
		{jupiter.WithReferral("a", swap.MinReferralFeeBps-1)},
		{jupiter.WithReferral("a", swap.MinReferralFeeBps), jupiter.WithoutMaintainerFee()},
		{nil},
	}

	for _, opts := range tests {
		if _, err := jupiter.NewClient(opts...); err == nil {
			t.Fatal("invalid configuration accepted")
		}
	}
}

func TestCancellation(t *testing.T) {
	c, err := jupiter.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = c.Swap.Order(ctx, swap.OrderRequest{
		InputMint:  "in",
		OutputMint: "out",
		Amount:     "1",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
}

func TestRedirectDoesNotLeakKey(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("redirect followed")
	}))
	defer target.Close()

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer origin.Close()

	c, err := jupiter.NewClient(
		jupiter.WithBaseURL(origin.URL),
		jupiter.WithAPIKey("secret"),
		jupiter.WithTimeout(time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Swap.ProgramLabels(context.Background())
	var api *jupiter.APIError
	if !errors.As(err, &api) || api.StatusCode != http.StatusFound {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildPreservesInstructionsAndSeparatesFees(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/swap/v2/build" ||
			q.Get("referralAccount") != "" ||
			q.Get("platformFeeBps") != "25" ||
			q.Get("feeAccount") != "token-account" ||
			q.Get("slippageBps") != swap.SlippageRTSE {
			t.Errorf("unexpected build: %s", r.URL)
		}

		if _, err := w.Write([]byte(`{
  "transactionVersion": 1,
 "swapMode": "ExactIn",
 "slippageBps": 0,
  "computeUnitPrice": "12345678901234567890",
  "swapInstruction": {
    "programId": "program",
    "accounts": [
      {
        "pubkey": "owner",
        "isSigner": true,
        "isWritable": false
      }
    ],
    "data": "AQ=="
  },
  "cleanupInstruction": null,
  "addressesByLookupTableAddress": {}
}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))

	defer server.Close()

	c, err := jupiter.NewClient(jupiter.WithBaseURL(server.URL), jupiter.WithReferral("referral", 50))
	if err != nil {
		t.Fatal(err)
	}

	result, err := c.Swap.Build(
		context.Background(),
		swap.BuildRequest{
			InputMint:      "in",
			OutputMint:     "out",
			Amount:         "1",
			Taker:          "owner",
			SlippageBps:    swap.SlippageRTSE,
			PlatformFeeBps: 25,
			FeeAccount:     "token-account",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.SwapMode != swap.SwapModeExactIn || result.SlippageBps.String() != "0" {
		t.Fatal("build mode or explicit zero slippage lost")
	}

	if result.ComputeUnitPrice != "12345678901234567890" ||
		result.SwapInstruction.Data != "AQ==" ||
		!result.SwapInstruction.Accounts[0].IsSigner ||
		result.CleanupInstruction != nil {
		t.Fatalf("instruction changed: %+v", result)
	}
}

func TestExecutePreservesSignedPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request swap.ExecuteRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}

		if request.RequestID != "original-id" || request.SignedTransaction != "AAEC/w==" {
			t.Errorf("payload changed: %+v", request)
		}

		if _, err := w.Write([]byte(`{
  "status": "Success",
  "code": 0,
  "signature": "signature",
  "totalOutputAmount": "9007199254740993"
}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))

	defer server.Close()

	c, err := jupiter.NewClient(jupiter.WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	result, err := c.Swap.Execute(
		context.Background(),
		swap.ExecuteRequest{
			RequestID:         "original-id",
			SignedTransaction: "AAEC/w==",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if result.TotalOutputAmount != "9007199254740993" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestBuildValidation(t *testing.T) {
	client, err := jupiter.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		configure  func(*swap.BuildRequest)
		parseError bool
	}{
		{
			name: "malformed slippage",
			configure: func(r *swap.BuildRequest) {
				r.SlippageBps = "invalid"
			},
			parseError: true,
		},
		{
			name: "slippage above maximum",
			configure: func(r *swap.BuildRequest) {
				r.SlippageBps = strconv.Itoa(swap.BasisPointsPerWhole + 1)
			},
		},
		{
			name: "negative percentile",
			configure: func(r *swap.BuildRequest) {
				r.ComputeUnitPricePercentile = "-1"
			},
		},
		{
			name: "malformed tip",
			configure: func(r *swap.BuildRequest) {
				r.TipAmount = "invalid"
			},
			parseError: true,
		},
		{
			name: "platform fee above maximum",
			configure: func(r *swap.BuildRequest) {
				r.PlatformFeeBps = swap.BasisPointsPerWhole + 1
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := swap.BuildRequest{
				InputMint:  "in",
				OutputMint: "out",
				Amount:     "1",
				Taker:      "owner",
			}
			test.configure(&request)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err := client.Swap.Build(ctx, request)
			if err == nil {
				t.Fatal("invalid request accepted")
			}

			if errors.Is(err, context.Canceled) {
				t.Fatalf("invalid request reached transport: %v", err)
			}

			var parseError *strconv.NumError
			if errors.As(err, &parseError) != test.parseError {
				t.Fatalf("unexpected parse cause: %v", err)
			}
		})
	}
}
