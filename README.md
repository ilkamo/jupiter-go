## Go library to interact with [Jupiter](https://jup.ag) and the solana blockchain

This library is a Go client for [Jupiter](https://jup.ag). It provides a simple way to interact with the Jupiter API to get quotes and perform swaps.

## Installation

```bash
go get github.com/ilkamo/jupiter-go
```

## Usage

Here's a simple example to get a quote and the related swap instructions from Jupiter:

```go
package main

import (
	"context"

	"github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/openapi"
)

func main() {
	jupClient, err := jupitergo.NewJupClient("https://quote-api.jup.ag/v6")
	if err != nil {
		// handle me
	}

	ctx := context.TODO()

	slippageBps := 250

	// Get the current quote for a swap
	quote, err := jupClient.GetQuote(ctx, openapi.GetQuoteParams{
		InputMint:   "So11111111111111111111111111111111111111112",
		OutputMint:  "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
		Amount:      10000000,
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
	"github.com/ilkamo/jupiter-go"
)

func main() {
	// ...
	// swap, err := jupClient.PostSwap(ctx, openapi.PostSwapJSONRequestBody{...})
	// ...

	// Create a wallet from private key
	walletPrivateKey := "your private key"
	wallet, err := jupitergo.NewWalletFromPrivateKeyBase58(walletPrivateKey)
	if err != nil {
		// handle me
	}

	// Sign and send the transaction
	eng, err := jupitergo.NewSolanaEngine(wallet, "https://api.mainnet-beta.solana.com")
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

## TODO

Once a transaction is sent on-chain it doesn't mean that the swap is completed. You should monitor the transaction status and confirm the swap is completed. This library doesn't provide a way to monitor the transaction status yet but it's on the roadmap.

## License

This library is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Donate

If you find this library useful, consider donating some JUP or Solana to the following addresses:

`BXzmfHxfEMcMj8hDccUNdrwXVNeybyfb2iV2nktE1VnJ`
