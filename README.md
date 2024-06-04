# Jupiter-go

### Go library to interact with [Jupiter](https://jup.ag) to get quotes, perform swaps and send them on-chain
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GoDoc](https://pkg.go.dev/badge/github.com/ilkamo/jupiter-go?status.svg)](https://pkg.go.dev/github.com/ilkamo/jupiter-go?tab=doc)
[![Go Report Card](https://goreportcard.com/badge/github.com/ilkamo/jupiter-go)](https://goreportcard.com/report/ilkamo/jupiter-go)

This library provides a simple way to interact with the [Jupiter](https://jup.ag) API to get quotes and perform swaps. 

It also provides: 
- A [solana client](solana/client.go) to send the swap transaction on-chain and check its status.
- A [solana monitor](solana/monitor.go) to wait for a transaction to reach a specific commitment status.

<img align="right" width="200" src="assets/jup-gopher.png">

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

	"github.com/ilkamo/jupiter-go/jupiter"
)

func main() {
	jupClient, err := jupiter.NewClientWithResponses(jupiter.DefaultAPIURL)
	// handle the error

	ctx := context.TODO()

	slippageBps := 250

	// Get the current quote for a swap.
	// Ensure that the input and output mints are valid.
	// The amount is the smallest unit of the input token.
	quoteResponse, err := jupClient.GetQuoteWithResponse(ctx, &jupiter.GetQuoteParams{
		InputMint:   "So11111111111111111111111111111111111111112",
		OutputMint:  "WENWENvqqNya429ubCdR81ZmD69brwQaaBYY6p3LCpk",
		Amount:      100000,
		SlippageBps: &slippageBps,
	})
	// handle the error
	
	quote := quoteResponse.JSON200

	// More info: https://station.jup.ag/docs/apis/troubleshooting
	prioritizationFeeLamports := jupiter.SwapRequest_PrioritizationFeeLamports{}
	if err = prioritizationFeeLamports.UnmarshalJSON([]byte(`"auto"`)); err != nil {
		// handle the error
	}

	dynamicComputeUnitLimit := true
	// Get instructions for a swap.
	// Ensure your public key is valid.
	swapResponse, err := jupClient.PostSwapWithResponse(ctx, jupiter.PostSwapJSONRequestBody{
		PrioritizationFeeLamports: &prioritizationFeeLamports,
		QuoteResponse:             *quote,
		UserPublicKey:             "{YOUR_PUBLIC_KEY}",
		DynamicComputeUnitLimit:   &dynamicComputeUnitLimit,
	})
	// handle the error
	
	swap := swapResponse.JSON200
}
```

Once you have the swap instructions, you can use the [solana client](solana/client.go) to sign and send the transaction on-chain.
Please remember, when a transaction is sent on-chain it doesn't mean that the swap is completed. The instruction could error, that's why you [should monitor the transaction status](_examples/txmonitor/main.go) and confirm the transaction is finalized without errors.

```go
package main

import (
	"time"

	"github.com/ilkamo/jupiter-go/jupiter"
	"github.com/ilkamo/jupiter-go/solana"
)

func main() {
	// ... previous code
	// swap := swapResponse.JSON200

	// Create a wallet from private key
	walletPrivateKey := "{YOUR_PRIVATE_KEY}"
	wallet, err := solana.NewWalletFromPrivateKeyBase58(walletPrivateKey)
	// handle the error

	// Create a Solana client. Change the URL to the desired Solana node.
	solanaClient, err := solana.NewClient(wallet, "https://api.mainnet-beta.solana.com")
	// handle the error

	// Sign and send the transaction.
	signedTx, err := solanaClient.SendTransactionOnChain(ctx, swap.SwapTransaction)
	// handle the error

	// Wait a bit to let the transaction propagate to the network.
	// This is just an example and not a best practice.
	// You could use a ticker or initialize a monitor to wait for the transaction to be confirmed.
	time.Sleep(20 * time.Second)

	// Get the status of the transaction (pull the status from the blockchain at intervals 
	// until the transaction is confirmed).
	confirmed, err := solanaClient.CheckSignature(ctx, signedTx)
	// handle the error
}
```

A full swap example is available in the [examples/swap](_examples/swap) folder.

A transaction monitoring example using websocket is available in the [examples/txmonitor](_examples/txmonitor) folder.

## Jupiter client

The Jupiter client is generated from the [official Jupiter openapi definition](https://github.com/jup-ag/jupiter-quote-api-node/blob/main/swagger.yaml) and provides the following methods to interact with the Jupiter API:

```go
// GetIndexedRouteMapWithResponse request
GetIndexedRouteMapWithResponse(
	ctx context.Context, 
	params *GetIndexedRouteMapParams, 
	reqEditors ...RequestEditorFn, 
) (*GetIndexedRouteMapResponse, error)

// GetProgramIdToLabelWithResponse request
GetProgramIdToLabelWithResponse(
	ctx context.Context, 
	reqEditors ...RequestEditorFn,
) (*GetProgramIdToLabelResponse, error)

// GetQuoteWithResponse request
GetQuoteWithResponse(
	ctx context.Context, 
	params *GetQuoteParams, 
	reqEditors ...RequestEditorFn, 
) (*GetQuoteResponse, error)

// PostSwapWithBodyWithResponse request with any body
PostSwapWithBodyWithResponse(
	ctx context.Context, 
	contentType string, 
	body io.Reader, 
	reqEditors ...RequestEditorFn, 
) (*PostSwapResponse, error)

PostSwapWithResponse(
	ctx context.Context, 
	body PostSwapJSONRequestBody, 
	reqEditors ...RequestEditorFn,
) (*PostSwapResponse, error)

// PostSwapInstructionsWithBodyWithResponse request with any body
PostSwapInstructionsWithBodyWithResponse(
	ctx context.Context, 
	contentType string, 
	body io.Reader, 
	reqEditors ...RequestEditorFn,
) (*PostSwapInstructionsResponse, error)

PostSwapInstructionsWithResponse(
	ctx context.Context, 
	body PostSwapInstructionsJSONRequestBody, 
	reqEditors ...RequestEditorFn, 
) (*PostSwapInstructionsResponse, error)

GetTokensWithResponse(
	ctx context.Context, 
	reqEditors ...RequestEditorFn, 
) (*GetTokensResponse, error)
```

## Solana client

The Solana client provides the following methods to interact with the Solana blockchain:

```go
// SendTransactionOnChain signs and sends a transaction on-chain.
SendTransactionOnChain(
	ctx context.Context, 
	txBase64 string,
) (TxID, error)

// CheckSignature checks the status of a transaction on-chain.
CheckSignature(
	ctx context.Context, 
	tx TxID,
) (bool, error)

// GetTokenAccountBalance returns the balance of an SPL token account.
GetTokenAccountBalance(
	ctx context.Context, 
	tokenAccount string, 
) (TokenAccount, error)

// Close closes the client.
Close() error
```

## Solana monitor

The Solana monitor provides the following methods to monitor the Solana blockchain:

```go
// WaitForCommitmentStatus waits for a transaction to reach a specific commitment status.
WaitForCommitmentStatus(
    context.Context, 
    TxID, 
    CommitmentStatus,
) (MonitorResponse, error)
```

## API Urls

This library provides two options to leverage for Jupiter APIs. 

The default Jupiter API provided by the official Jupiter team can be used via `jupiter.DefaultAPIURL`.

The second Jupiter API option leverages [jupiterapi.com](https://www.jupiterapi.com/) (Community Project). This endpoint provides higher rate limits, but *includes a small 0.2% platform fee*. This API can be used via `jupiter.JupiterAPIURL`.



## License

This library is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Donate

If you find this library useful and want to support its development, consider donating some JUP/Solana to the following address:

`BXzmfHxfEMcMj8hDccUNdrwXVNeybyfb2iV2nktE1VnJ`
