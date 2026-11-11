package jupiter_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
		jupiter.WithReferral(referralAccount, swap.MinReferralFeeBps),
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := c.Swap.Order(
		context.Background(),
		swap.OrderRequest{
			InputMint:  wrappedSOLMint,
			OutputMint: usdcMint,
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
				if _, ok := errors.AsType[*swap.ExecutionError](err); !ok {
					t.Fatalf("expected execution error, got %v", err)
				}

				return
			}

			apiError, ok := errors.AsType[*jupiter.APIError](err)
			if !ok || apiError.StatusCode != test.statusCode {
				t.Fatalf("expected HTTP error with status %d, got %v", test.statusCode, err)
			}
		})
	}
}

func TestConfiguration(t *testing.T) {
	tests := [][]jupiter.Option{
		{jupiter.WithTimeout(0)},
		{jupiter.WithBaseURL("ftp://host")},
		{jupiter.WithReferral(referralAccount, swap.MinReferralFeeBps-1)},
		{
			jupiter.WithReferral(referralAccount, swap.MinReferralFeeBps),
			jupiter.WithReferral(referralAccount, swap.MaxReferralFeeBps),
		},
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
		InputMint:  wrappedSOLMint,
		OutputMint: usdcMint,
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
	api, ok := errors.AsType[*jupiter.APIError](err)
	if !ok || api.StatusCode != http.StatusFound {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildPreservesInstructionsAndSeparatesFees(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.URL.Path != "/swap/v2/build" ||
			q.Get("referralAccount") != "" ||
			q.Get("platformFeeBps") != "25" ||
			q.Get("feeAccount") != tokenAccount ||
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

	c, err := jupiter.NewClient(
		jupiter.WithBaseURL(server.URL),
		jupiter.WithReferral(referralAccount, swap.MinReferralFeeBps),
	)
	if err != nil {
		t.Fatal(err)
	}

	result, err := c.Swap.Build(
		context.Background(),
		swap.BuildRequest{
			InputMint:      wrappedSOLMint,
			OutputMint:     usdcMint,
			Amount:         "1",
			Taker:          takerAddress,
			SlippageBps:    swap.SlippageRTSE,
			PlatformFeeBps: 25,
			FeeAccount:     tokenAccount,
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

type unreachableHTTPClient struct {
	t *testing.T
}

func (c unreachableHTTPClient) Do(r *http.Request) (*http.Response, error) {
	c.t.Errorf("invalid request reached transport: %s", r.URL)

	return nil, errors.New("unexpected request")
}

func TestInvalidRequestsDoNotReachTransport(t *testing.T) {
	client, err := jupiter.NewClient(jupiter.WithHTTPClient(unreachableHTTPClient{t: t}))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		call         func(context.Context) error
		addressError bool
	}{
		{
			name: "order with malformed taker",
			call: func(ctx context.Context) error {
				_, err := client.Swap.Order(ctx, swap.OrderRequest{
					InputMint:  wrappedSOLMint,
					OutputMint: usdcMint,
					Amount:     "1",
					Taker:      "owner",
				})

				return err
			},
			addressError: true,
		},
		{
			name: "order receiver without taker",
			call: func(ctx context.Context) error {
				_, err := client.Swap.Order(ctx, swap.OrderRequest{
					InputMint:  wrappedSOLMint,
					OutputMint: usdcMint,
					Amount:     "1",
					Receiver:   takerAddress,
				})

				return err
			},
		},
		{
			name: "build with malformed fee account",
			call: func(ctx context.Context) error {
				_, err := client.Swap.Build(ctx, swap.BuildRequest{
					InputMint:      wrappedSOLMint,
					OutputMint:     usdcMint,
					Amount:         "1",
					Taker:          takerAddress,
					PlatformFeeBps: 1,
					FeeAccount:     "token-account",
				})

				return err
			},
			addressError: true,
		},
		{
			name: "execute with malformed block height",
			call: func(ctx context.Context) error {
				_, err := client.Swap.Execute(ctx, swap.ExecuteRequest{
					RequestID:            "order",
					SignedTransaction:    "AQ==",
					LastValidBlockHeight: "-1",
				})

				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.call(context.Background())
			if err == nil {
				t.Fatal("invalid request accepted")
			}

			if errors.Is(err, swap.ErrInvalidAddress) != test.addressError {
				t.Fatalf("unexpected address error classification: %v", err)
			}
		})
	}
}

func TestReferralRequiresAddress(t *testing.T) {
	_, err := jupiter.NewClient(jupiter.WithReferral("account", swap.MinReferralFeeBps))
	if !errors.Is(err, swap.ErrInvalidAddress) {
		t.Fatalf("expected invalid address error, got %v", err)
	}
}
