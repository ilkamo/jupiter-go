// Package swap implements Jupiter's Swap V2 order and custom-build flows.
package swap

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"

	"github.com/ilkamo/jupiter-go/internal/transport"
)

type Client struct {
	transport *transport.Client
	referral  Referral
}

type Referral struct {
	Account     string
	BasisPoints uint16
}

func NewClient(t *transport.Client, referral Referral) *Client {
	return &Client{transport: t, referral: referral}
}

func (c *Client) Referral() Referral {
	return c.referral
}

type OrderRequest struct {
	InputMint, OutputMint, Amount, Taker string
	Receiver, Payer                      string
	SwapMode                             string
	SlippageBps                          *uint16
	PriorityFeeLamports, JitoTipLamports *uint64
	BroadcastFeeType                     string
	ExcludeRouters, ExcludeDexes         []string
	MaxSupportedTransactionVersion       *uint8
}

func swapQuery(input, output, amount string) (url.Values, error) {
	if input == "" || output == "" || amount == "" {
		return nil, fmt.Errorf("input mint, output mint and amount required")
	}

	rawAmount, ok := new(big.Int).SetString(amount, decimalBase)
	if !ok {
		return nil, fmt.Errorf("amount must be a base-10 raw integer")
	}

	if rawAmount.Sign() <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	return url.Values{
		"inputMint":  {input},
		"outputMint": {output},
		"amount":     {rawAmount.String()},
	}, nil
}

func set(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}

func (c *Client) Order(
	ctx context.Context,
	r OrderRequest,
) (*OrderResponse, error) {
	q, err := swapQuery(r.InputMint, r.OutputMint, r.Amount)
	if err != nil {
		return nil, err
	}

	if r.SwapMode != "" && r.SwapMode != SwapModeExactIn {
		return nil, fmt.Errorf("unsupported swap mode")
	}

	set(q, "swapMode", r.SwapMode)
	set(q, "taker", r.Taker)
	set(q, "receiver", r.Receiver)
	set(q, "payer", r.Payer)

	if r.SlippageBps != nil {
		if *r.SlippageBps > BasisPointsPerWhole {
			return nil, fmt.Errorf("slippage exceeds %d bps", BasisPointsPerWhole)
		}

		q.Set("slippageBps", strconv.Itoa(int(*r.SlippageBps)))
	}

	if r.MaxSupportedTransactionVersion != nil {
		if *r.MaxSupportedTransactionVersion > MaxTransactionVersion {
			return nil, fmt.Errorf("unsupported transaction version")
		}

		q.Set("maxSupportedTransactionVersion", strconv.Itoa(int(*r.MaxSupportedTransactionVersion)))
	}

	if r.PriorityFeeLamports != nil {
		q.Set("priorityFeeLamports", strconv.FormatUint(*r.PriorityFeeLamports, decimalBase))
	}

	if r.JitoTipLamports != nil {
		q.Set("jitoTipLamports", strconv.FormatUint(*r.JitoTipLamports, decimalBase))
	}

	switch r.BroadcastFeeType {
	case "", BroadcastFeeMaxCap, BroadcastFeeExact:
	default:
		return nil, fmt.Errorf("invalid broadcast fee type")
	}

	set(q, "broadcastFeeType", r.BroadcastFeeType)
	set(q, "excludeRouters", strings.Join(r.ExcludeRouters, ","))
	set(q, "excludeDexes", strings.Join(r.ExcludeDexes, ","))

	if c.referral.Account != "" {
		q.Set("referralAccount", c.referral.Account)
		q.Set("referralFee", strconv.Itoa(int(c.referral.BasisPoints)))
	}

	var result OrderResponse
	metadata, err := c.transport.Do(ctx, "GET", "/swap/v2/order", q, nil, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	if result.ErrorCode != nil {
		return &result, &ExecutionError{
			Code:      *result.ErrorCode,
			Message:   result.ErrorMessage,
			RequestID: result.RequestID,
		}
	}

	return &result, nil
}

type ExecuteRequest struct {
	SignedTransaction    string `json:"signedTransaction"`
	RequestID            string `json:"requestId"`
	LastValidBlockHeight string `json:"lastValidBlockHeight,omitempty"`
}

func (c *Client) Execute(
	ctx context.Context,
	r ExecuteRequest,
) (*ExecuteResponse, error) {
	if r.RequestID == "" || r.SignedTransaction == "" {
		return nil, fmt.Errorf("request ID and signed transaction required")
	}

	if _, err := base64.StdEncoding.DecodeString(r.SignedTransaction); err != nil {
		return nil, fmt.Errorf("signed transaction: %w", err)
	}

	var result ExecuteResponse
	metadata, err := c.transport.Do(ctx, "POST", "/swap/v2/execute", nil, r, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	if result.Status != ExecutionStatusSuccess || result.Code != executionSuccessCode {
		return &result, &ExecutionError{
			Code:      result.Code,
			Message:   result.Error,
			RequestID: r.RequestID,
			Signature: result.Signature,
		}
	}

	return &result, nil
}

func (c *Client) ProgramLabels(ctx context.Context) (*ProgramLabelsResponse, error) {
	var result ProgramLabelsResponse
	metadata, err := c.transport.Do(ctx, "GET", "/swap/v2/program-id-to-label", nil, nil, &result.Labels)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

type ExecutionError struct {
	Code                          int
	Message, RequestID, Signature string
}

func (e *ExecutionError) Error() string {
	return fmt.Sprintf("jupiter execution %d: %s", e.Code, e.Message)
}
