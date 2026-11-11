package jupiter_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/swap"
)

func TestOrderMetadataAndRFQFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("swapMode") != swap.SwapModeExactIn {
			t.Error("swap mode was not encoded")
		}

		w.Header().Set("x-request-id", "http-request")
		w.Header().Set("x-ratelimit-remaining", "12")
		const body = `{
   "requestId": "order-request",
   "swapUsdValue": 123.1234567890123456789,
   "priceImpactPct": "0.001",
   "swapType": "rfq",
   "totalTime": 1.25,
   "taker": null,
   "quoteId": "quote",
   "maker": "maker",
   "expireAt": "2026-11-11T12:00:00Z",
   "routePlan": [{"usdValue": 123.1234567890123456789}]
  }`
		if _, err := w.Write([]byte(body)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer server.Close()

	client, err := jupiter.NewClient(jupiter.WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Swap.Order(context.Background(), swap.OrderRequest{
		InputMint:  wrappedSOLMint,
		OutputMint: usdcMint,
		Amount:     "1",
		SwapMode:   swap.SwapModeExactIn,
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.Metadata.StatusCode != http.StatusOK || result.Metadata.RequestID != "http-request" {
		t.Fatalf("missing metadata: %+v", result.Metadata)
	}

	if result.RequestID != "order-request" || result.Metadata.Headers.Get("x-ratelimit-remaining") != "12" {
		t.Fatal("order identity or rate limit lost")
	}

	const usdValue = "123.1234567890123456789"
	if result.SwapUSDValue.String() != usdValue || result.RoutePlan[0].USDValue.String() != usdValue {
		t.Fatal("route or swap USD precision lost")
	}

	if result.QuoteID != "quote" || result.Maker != "maker" || result.Taker != nil {
		t.Fatalf("RFQ fields lost: %+v", result)
	}
}

func TestProgramLabelsMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-request-id", "labels")

		if _, err := w.Write([]byte(`{"program":"Example DEX"}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer server.Close()

	client, err := jupiter.NewClient(jupiter.WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Swap.ProgramLabels(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if result.Labels["program"] != "Example DEX" || result.Metadata.RequestID != "labels" {
		t.Fatalf("unexpected labels response: %+v", result)
	}
}
