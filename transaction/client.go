// Package transaction submits signed Solana transactions through Jupiter's landing
// infrastructure at tx.jup.ag, a send-only, Solana RPC-compatible JSON-RPC
// endpoint.
//
// Every transaction must transfer at least MinTipLamports to one of TipReceivers
// (solanatx.TipInstruction builds that instruction). A successful Send means the
// transaction was accepted and forwarded, not that it landed: confirm it on your
// own RPC node, which also serves blockhashes and simulation.
package transaction

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"

	"github.com/ilkamo/jupiter-go/internal/transport"
	"github.com/ilkamo/jupiter-go/response"
)

const (
	// DefaultURL is the transaction landing service, separate from the API gateway.
	DefaultURL = "https://tx.jup.ag"

	// MinTipLamports is the minimum Jupiter tip (0.001 SOL). The tip is flat:
	// tipping more does not improve landing.
	MinTipLamports = 1_000_000

	// Documented JSON-RPC error codes.
	CodeTipBelowMinimum  = -1013
	CodeMissingTip       = -1014
	CodeInvalidParameter = -1015
	CodeInvalidEncoding  = -1016

	jsonRPCVersion = "2.0"
	sendMethod     = "sendTransaction"
	encodingBase64 = "base64"
	sendRequestID  = 1
	rootPath       = "/"
)

// TipReceivers are the documented tip accounts. Spread tips across them, for
// example with RandomTipReceiver, to reduce write-lock contention.
var TipReceivers = []string{
	"GGztQqQ6pCPaJQnNpXBgELr5cs3WwDakRbh1iEMzjgSJ",
	"2MFoS3MPtvyQ4Wh4M9pdfPjz6UhVoNbFbGJAskCPCj3h",
	"BQ72nSv9f3PRyRKCBnHLVrerrv37CYTHm5h3s9VSGQDV",
	"6U91aKa8pmMxkJwBCfPTmUEfZi6dHe7DcFq2ALvB2tbB",
	"4xDsmeTWPNjgSVSS1VTfzFq3iHZhp77ffPkAmkZkdu71",
	"CapuXNQoDviLvU1PxFiizLgPNQCxrsag1uMeyk6zLVps",
	"9nnLbotNTcUhvbrsA6Mdkx45Sm82G35zo28AqUvjExn8",
	"6LXutJvKUw8Q5ue2gCgKHQdAN4suWW8awzFVC6XCguFx",
	"HFqp6ErWHY6Uzhj8rFyjYuDya2mXUpYEk8VW75K9PSiY",
	"DSN3j1ykL3obAVNv7ZX49VsFCPe4LqzxHnmtLiPwY6xg",
	"69yhtoJR4JYPPABZcSNkzuqbaFbwHsCkja1sP1Q2aVT5",
	"HU23r7UoZbqTUuh3vA7emAGztFtqwTeVips789vqxxBw",
	"3LoAYHuSd7Gh8d7RTFnhvYtiTiefdZ5ByamU42vkzd76",
	"3CgvbiM3op4vjrrjH2zcrQUwsqh5veNVRjFCB9N6sRoD",
	"GP8StUXNYSZjPikyRsvkTbvRV1GBxMErb59cpeCJnDf1",
	"7iWnBRRhBCiNXXPhqiGzvvBkKrvFSWqqmxRyu9VyYBxE",
}

// RandomTipReceiver picks a tip account. The choice only spreads load, so it uses
// a non-cryptographic generator.
func RandomTipReceiver() string {
	return TipReceivers[rand.N(len(TipReceivers))]
}

// ErrOutcomeUnknown marks Send failures after which the transaction may have been
// forwarded. Resubmitting the same signed transaction cannot execute it twice,
// because Solana deduplicates by signature, until its blockhash expires.
var ErrOutcomeUnknown = errors.New("transaction submission outcome unknown")

// RPCError is a JSON-RPC error returned with HTTP 200, such as CodeMissingTip. It
// is a definitive rejection.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string {
	return fmt.Sprintf("tx.jup.ag sendTransaction error %d: %s", e.Code, e.Message)
}

type Client struct {
	transport *transport.Client
}

// NewClient takes a transport configured with the transaction service URL.
func NewClient(t *transport.Client) *Client {
	return &Client{transport: t}
}

// SendOptions configures routing. SWQoSOnly selects SWQoS only; the default uses
// SWQoS and Jito, in which case Jupiter adds a small Jito tip.
type SendOptions struct {
	SWQoSOnly bool
}

type sendConfig struct {
	Encoding  string `json:"encoding"`
	SWQoSOnly bool   `json:"swqosOnly,omitempty"`
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
}

type rpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      *int      `json:"id"`
	Result  string    `json:"result"`
	Error   *RPCError `json:"error"`
}

type SendResult struct {
	// Signature identifies the forwarded transaction; confirm it on your RPC node.
	Signature string
	Metadata  response.Metadata
}

// Send forwards a base64-encoded signed transaction. Transport failures and
// server errors are retried with the same bytes, which is safe because a
// transaction executes at most once; JSON-RPC errors are returned as *RPCError.
func (c *Client) Send(
	ctx context.Context,
	signedTransaction string,
	options SendOptions,
) (*SendResult, error) {
	if signedTransaction == "" {
		return nil, fmt.Errorf("signed transaction required")
	}

	if _, err := base64.StdEncoding.DecodeString(signedTransaction); err != nil {
		return nil, fmt.Errorf("signed transaction: %w", err)
	}

	// An ended context guarantees nothing is sent, so the outcome is known.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	body := rpcRequest{
		JSONRPC: jsonRPCVersion,
		ID:      sendRequestID,
		Method:  sendMethod,
		Params: []any{
			signedTransaction,
			sendConfig{Encoding: encodingBase64, SWQoSOnly: options.SWQoSOnly},
		},
	}

	request := transport.Request{
		Method: http.MethodPost,
		Path:   rootPath,
		Body:   body,
		Retry:  transport.RepeatableRetry,
	}

	var decoded rpcResponse
	metadata, err := c.transport.Do(ctx, request, &decoded)
	result := &SendResult{Signature: decoded.Result, Metadata: metadata}
	if err != nil {
		if transport.MayHaveBeenProcessed(err) {
			return result, fmt.Errorf("%w: %w", ErrOutcomeUnknown, err)
		}

		return result, err
	}

	if decoded.Error != nil {
		return result, decoded.Error
	}

	if decoded.JSONRPC != jsonRPCVersion || decoded.ID == nil || *decoded.ID != sendRequestID {
		return result, fmt.Errorf("%w: unexpected JSON-RPC response envelope", ErrOutcomeUnknown)
	}

	if decoded.Result == "" {
		return result, fmt.Errorf("%w: response has no signature", ErrOutcomeUnknown)
	}

	return result, nil
}
