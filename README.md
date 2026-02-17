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
	"net/http"

	"github.com/ilkamo/jupiter-go/jupiter"
)

func main() {
	// Initialize client with API key (automatically added to all requests)
	apiKey := "{YOUR_JUPITER_API_KEY}"
	jupClient, err := jupiter.NewClientWithResponses(
		jupiter.DefaultAPIURL,
		jupiter.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("x-api-key", apiKey)
			return nil
		}),
	)
	// handle the error

	ctx := context.TODO()

	slippageBps := uint64(250)

	// Get the current quote for a swap.
	// Ensure that the input and output mints are valid.
	// The amount is the smallest unit of the input token.
	quoteResponse, err := jupClient.QuoteGetWithResponse(ctx, &jupiter.QuoteGetParams{
		InputMint:   "So11111111111111111111111111111111111111112",
		OutputMint:  "JUPyiwrYJFskUPiHa7hkeR8VUtAeFoSYbKedZNsDvCN",
		Amount:      100000,
		SlippageBps: &slippageBps,
	})
	// handle the error
	
	quote := quoteResponse.JSON200

	// Define the prioritization fee in lamports.
	prioritizationFeeLamports := &struct {
		JitoTipLamports              *uint64 `json:"jitoTipLamports,omitempty"`
		PriorityLevelWithMaxLamports *struct {
			MaxLamports   *uint64                                                                                   `json:"maxLamports,omitempty"`
			PriorityLevel *jupiter.SwapRequestPrioritizationFeeLamportsPriorityLevelWithMaxLamportsPriorityLevel `json:"priorityLevel,omitempty"`
		} `json:"priorityLevelWithMaxLamports,omitempty"`
	}{
		PriorityLevelWithMaxLamports: &struct {
			MaxLamports   *uint64                                                                                   `json:"maxLamports,omitempty"`
			PriorityLevel *jupiter.SwapRequestPrioritizationFeeLamportsPriorityLevelWithMaxLamportsPriorityLevel `json:"priorityLevel,omitempty"`
		}{
			MaxLamports:   new(uint64),
			PriorityLevel: new(jupiter.SwapRequestPrioritizationFeeLamportsPriorityLevelWithMaxLamportsPriorityLevel),
		},
	}

	*prioritizationFeeLamports.PriorityLevelWithMaxLamports.MaxLamports = 1000
	*prioritizationFeeLamports.PriorityLevelWithMaxLamports.PriorityLevel = jupiter.High
	
	// If you prefer to set a Jito tip, you can use the following line instead of the above block.
	// Look at _examples/jitoswap/main.go for more details.
	// *prioritizationFeeLamports.JitoTipLamports = 1000

	dynamicComputeUnitLimit := true
	// Get instructions for a swap.
	// Ensure your public key is valid.
	swapResponse, err := jupClient.SwapPostWithResponse(ctx, jupiter.SwapPostJSONRequestBody{
		PrioritizationFeeLamports: prioritizationFeeLamports,
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

A full swap example is available in the [examples/swap](_examples/swap) folder. For a swap with Jito tips, check the [examples/jitoswap](_examples/jitoswap) folder.

A transaction monitoring example using websocket is available in the [examples/txmonitor](_examples/txmonitor) folder.

## API Key Usage

To use Jupiter's API, you need an API key. The recommended way is to configure the client once with `WithRequestEditorFn`, which automatically injects the API key into every request:

```go
apiKey := "{YOUR_JUPITER_API_KEY}"
jupClient, err := jupiter.NewClientWithResponses(
	jupiter.DefaultAPIURL,
	jupiter.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("x-api-key", apiKey)
		return nil
	}),
)
```

With this approach, you don't need to pass the API key to individual method calls—it's handled automatically.

## Jupiter client

The Jupiter client is generated from the [official Jupiter openapi definition](https://github.com/jup-ag/jupiter-quote-api-node/blob/main/swagger.yaml) and provides the following methods to interact with the Jupiter API:

```go
// ProgramIdToLabelGetWithResponse request
ProgramIdToLabelGetWithResponse(
	ctx context.Context, 
	reqEditors ...RequestEditorFn,
) (*ProgramIdToLabelGetResponse, error)

// QuoteGetWithResponse request
QuoteGetWithResponse(
	ctx context.Context, 
	params *QuoteGetParams, 
	reqEditors ...RequestEditorFn, 
) (*QuoteGetResponse, error)

// SwapPostWithBodyWithResponse request with any body
SwapPostWithBodyWithResponse(
	ctx context.Context, 
	contentType string, 
	body io.Reader, 
	reqEditors ...RequestEditorFn, 
) (*SwapPostResponse, error)

SwapPostWithResponse(
	ctx context.Context, 
	body SwapPostJSONRequestBody, 
	reqEditors ...RequestEditorFn,
) (*SwapPostResponse, error)

// SwapInstructionsPostWithBodyWithResponse request with any body
SwapInstructionsPostWithBodyWithResponse(
	ctx context.Context, 
	contentType string, 
	body io.Reader, 
	reqEditors ...RequestEditorFn,
) (*SwapInstructionsPostResponse, error)

SwapInstructionsPostWithResponse(
	ctx context.Context, 
	body SwapInstructionsPostJSONRequestBody, 
	reqEditors ...RequestEditorFn, 
) (*SwapInstructionsPostResponse, error)
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

## Notes
- Starting with **v0.2.0**, methods and parameters were renamed to align with the Jupiter OpenAPI definition.
- Starting with **v0.1.0**, _jupiter-go_ supports the new Jupiter API as documented at [station.jup.ag/docs](https://station.jup.ag/docs/).
  - It supports both **prioritization fee** and **Jito tips**.
- For those who need to use the legacy API, **v0.0.24** is the final version supporting it. Note that legacy Jupiter API hostnames will be fully deprecated on **June 1, 2025**.

## Contribute

Contributions are welcome! Feel free to open an issue or submit a pull request if you find a bug or want to add a new feature.

## License

This library is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Donate

If you find this library useful and want to support its development, consider donating some JUP/Solana to the following address:

`BXzmfHxfEMcMj8hDccUNdrwXVNeybyfb2iV2nktE1VnJ`

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=ilkamo/jupiter-go&type=Date)](https://www.star-history.com/#ilkamo/jupiter-go&Date)
