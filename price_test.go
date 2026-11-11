package jupiter_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/price"
)

func TestPriceContract(t *testing.T) {
	const body = `{
  "So11111111111111111111111111111111111111112": {
    "createdAt": "2024-06-05T08:55:25.527Z",
    "liquidity": 621679197.670000000000001,
    "usdPrice": 147.48,
    "blockId": null,
    "decimals": 9,
    "priceChange24h": null
  },
  "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v": {
    "liquidity": 0,
    "usdPrice": 0.9999,
    "blockId": 348004025,
    "decimals": 6,
    "priceChange24h": 0
  }
}`

	client, recorded := recordingServer(t, body)
	unpriced := takerAddress
	mints := []string{wrappedSOLMint, unpriced, usdcMint}

	result, err := client.Price.Get(context.Background(), mints)
	if err != nil {
		t.Fatal(err)
	}

	requests := recorded()
	if len(requests) != 1 || requests[0].path != "/price/v3/" {
		t.Fatalf("unexpected requests: %+v", requests)
	}

	if ids := requests[0].query["ids"]; len(ids) != 1 || ids[0] != strings.Join(mints, ",") {
		t.Fatalf("ids = %v", ids)
	}

	if len(result.Missing) != 1 || result.Missing[0] != unpriced {
		t.Fatalf("missing = %v", result.Missing)
	}

	sol := result.Prices[wrappedSOLMint]
	if sol.Liquidity.String() != "621679197.670000000000001" || sol.Decimals != 9 {
		t.Fatalf("precision lost: %+v", sol)
	}

	if sol.BlockID != nil || sol.PriceChange24h != "" {
		t.Fatalf("null fields acquired values: %+v", sol)
	}

	usdc := result.Prices[usdcMint]
	if usdc.BlockID == nil || *usdc.BlockID != 348004025 || usdc.PriceChange24h.String() != "0" {
		t.Fatalf("present fields lost: %+v", usdc)
	}
}

func TestPriceAcceptsMaximumMints(t *testing.T) {
	client, recorded := recordingServer(t, `{}`)

	mints := distinctMints(price.MaxMints)
	result, err := client.Price.Get(context.Background(), mints)
	if err != nil {
		t.Fatal(err)
	}

	if len(recorded()) != 1 || len(result.Missing) != price.MaxMints {
		t.Fatalf("missing = %d", len(result.Missing))
	}
}

func TestInvalidPriceRequestsDoNotReachTransport(t *testing.T) {
	client, err := jupiter.NewClient(jupiter.WithHTTPClient(unreachableHTTPClient{t: t}))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		mints        []string
		addressError bool
	}{
		{name: "no mints"},
		{name: "above maximum", mints: distinctMints(price.MaxMints + 1)},
		{name: "malformed mint", mints: []string{wrappedSOLMint, "SOL"}, addressError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := client.Price.Get(context.Background(), test.mints)
			if err == nil {
				t.Fatal("invalid request accepted")
			}

			if errors.Is(err, price.ErrInvalidAddress) != test.addressError {
				t.Fatalf("unexpected address error classification: %v", err)
			}
		})
	}
}
