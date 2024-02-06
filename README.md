## Go library to interact with [Jupiter](https://jup.ag) and the solana blockchain

This library is a Go client for [Jupiter](https://jup.ag). It provides a simple way to interact with the Jupiter API to get quotes and perform swaps.

## Installation

```bash
go get github.com/jupiter-dev/jupiter-go
```

## Usage

Here's a simple example to get a quote and the related swap instructions from Jupiter:

```go
package main

import (
	"context"

	"github.com/ilkamo/go-jup"
	"github.com/ilkamo/go-jup/openapi"
)

func main() {
	jupClient, err := gojup.NewJupClient("https://quote-api.jup.ag/v6")
	if err != nil {
		// handle me
	}

	ctx := context.TODO()

	slippageBps := 250

	// Get the current quote for a swap
	quote, err := jupClient.GetQuote(ctx, openapi.GetQuoteParams{
		InputMint:   "So11111111111111111111111111111111111111112",
		OutputMint:  "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		Amount:      "10000000",
		SlippageBps: &slippageBps,
	})
	if err != nil {
		// handle me
	}

	// More info: https://station.jup.ag/docs/apis/troubleshooting
	prioritizationFeeLamports := "auto"
	dynamicComputeUnitLimit := true
	// Get instructions for a swap
	swap, err := jupClient.PostSwap(ctx, openapi.PostSwapJSONRequestBody{
		PrioritizationFeeLamports: &prioritizationFeeLamports,
		QuoteResponse:             quote,
		UserPublicKey:             "the public key of your wallet",
		DynamicComputeUnitLimit:   &dynamicComputeUnitLimit,
	})
	if err != nil {
		// handle me
	}
}
```

Once you have the swap instructions, you can use the [Solana engine](engine.go) to sign and send the transaction.

```go
package main

import (
	"github.com/ilkamo/go-jup"
)

func main() {
	// ...
	// swap, err := jupClient.PostSwap(ctx, openapi.PostSwapJSONRequestBody{...})
	// ...

	// Create a wallet from private key
	walletPrivateKey := "your private key"
	wallet, err := gojup.NewWalletFromPrivateKeyBase58(walletPrivateKey)
	if err != nil {
		// handle me
	}

	// Sign and send the transaction
	eng, err := gojup.NewSolanaEngine(wallet, "https://api.mainnet-beta.solana.com")
	if err != nil {
		// handle me
	}

	// Sign the transaction
	signedTx, err := eng.SendSwapOnChain(ctx, swap)
	if err != nil {
		// handle me
	}
}
```

## License

This library is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
```
