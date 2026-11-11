// Package swap implements Jupiter's Swap V2 order and custom-build flows.
package swap

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ilkamo/jupiter-go/internal/listparam"
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

// Validate checks the referral account address and the documented managed-order fee range.
func (r Referral) Validate() error {
	if err := validateAddress("referral account", r.Account); err != nil {
		return err
	}

	if r.BasisPoints < MinReferralFeeBps || r.BasisPoints > MaxReferralFeeBps {
		return fmt.Errorf(
			"referral fee must be between %d and %d basis points",
			MinReferralFeeBps,
			MaxReferralFeeBps,
		)
	}

	return nil
}

func NewClient(t *transport.Client, referral Referral) *Client {
	return &Client{transport: t, referral: referral}
}

func (c *Client) Referral() Referral {
	return c.referral
}

// OrderRequest describes a managed order. Without Taker, Jupiter returns a quote without a
// transaction; Receiver and Payer only shape that transaction and therefore require Taker.
type OrderRequest struct {
	InputMint                      string
	OutputMint                     string
	Amount                         string
	Taker                          string
	Receiver                       string
	Payer                          string
	SwapMode                       string
	SlippageBps                    *uint16
	PriorityFeeLamports            *uint64
	JitoTipLamports                *uint64
	BroadcastFeeType               string
	ExcludeRouters                 []string
	ExcludeDexes                   []string
	MaxSupportedTransactionVersion *uint8
}

func (r OrderRequest) validate() error {
	if err := validateMintPair(r.InputMint, r.OutputMint); err != nil {
		return err
	}

	if err := validateOptionalAddress("taker", r.Taker); err != nil {
		return err
	}

	if err := validateOptionalAddress("receiver", r.Receiver); err != nil {
		return err
	}

	if err := validateOptionalAddress("payer", r.Payer); err != nil {
		return err
	}

	transactionAccountSet := r.Receiver != "" || r.Payer != ""
	if transactionAccountSet && r.Taker == "" {
		return fmt.Errorf("receiver and payer require a taker")
	}

	if r.Receiver != "" && r.Receiver == r.Taker {
		return fmt.Errorf("receiver must differ from taker")
	}

	if r.SwapMode != "" && r.SwapMode != SwapModeExactIn {
		return fmt.Errorf("unsupported swap mode %q", r.SwapMode)
	}

	if r.SlippageBps != nil && *r.SlippageBps > BasisPointsPerWhole {
		return fmt.Errorf("slippage exceeds %d bps", BasisPointsPerWhole)
	}

	versionSet := r.MaxSupportedTransactionVersion != nil
	if versionSet && *r.MaxSupportedTransactionVersion > MaxTransactionVersion {
		return fmt.Errorf("unsupported transaction version")
	}

	switch r.BroadcastFeeType {
	case "", BroadcastFeeMaxCap, BroadcastFeeExact:
	default:
		return fmt.Errorf("invalid broadcast fee type %q", r.BroadcastFeeType)
	}

	broadcastFeeSet := r.PriorityFeeLamports != nil || r.JitoTipLamports != nil
	if r.BroadcastFeeType != "" && !broadcastFeeSet {
		return fmt.Errorf("broadcast fee type requires a priority fee or Jito tip")
	}

	// Router names are format-checked only, so routers added upstream remain usable.
	if err := listparam.Validate("excluded routers", r.ExcludeRouters); err != nil {
		return err
	}

	return listparam.Validate("excluded DEXes", r.ExcludeDexes)
}

func (r OrderRequest) query(referral Referral) (url.Values, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}

	amount, err := parseRawAmount(r.Amount)
	if err != nil {
		return nil, err
	}

	q := url.Values{
		"inputMint":  {r.InputMint},
		"outputMint": {r.OutputMint},
		"amount":     {amount},
	}

	set(q, "swapMode", r.SwapMode)
	set(q, "taker", r.Taker)
	set(q, "receiver", r.Receiver)
	set(q, "payer", r.Payer)

	if r.SlippageBps != nil {
		q.Set("slippageBps", strconv.Itoa(int(*r.SlippageBps)))
	}

	if r.MaxSupportedTransactionVersion != nil {
		q.Set(
			"maxSupportedTransactionVersion",
			strconv.Itoa(int(*r.MaxSupportedTransactionVersion)),
		)
	}

	if r.PriorityFeeLamports != nil {
		q.Set("priorityFeeLamports", strconv.FormatUint(*r.PriorityFeeLamports, decimalBase))
	}

	if r.JitoTipLamports != nil {
		q.Set("jitoTipLamports", strconv.FormatUint(*r.JitoTipLamports, decimalBase))
	}

	set(q, "broadcastFeeType", r.BroadcastFeeType)
	set(q, "excludeRouters", listparam.Join(r.ExcludeRouters))
	set(q, "excludeDexes", listparam.Join(r.ExcludeDexes))

	if referral.Account != "" {
		q.Set("referralAccount", referral.Account)
		q.Set("referralFee", strconv.Itoa(int(referral.BasisPoints)))
	}

	return q, nil
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
	q, err := r.query(c.referral)
	if err != nil {
		return nil, err
	}

	var result OrderResponse
	request := transport.Request{
		Method: http.MethodGet,
		Path:   "/swap/v2/order",
		Query:  q,
		Retry:  quoteRetryPolicy,
	}

	metadata, err := c.transport.Do(ctx, request, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	if result.ErrorCode != nil {
		return &result, &ExecutionError{
			Code:      result.ErrorCode,
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

func (r ExecuteRequest) validate() error {
	if r.RequestID == "" || r.SignedTransaction == "" {
		return fmt.Errorf("request ID and signed transaction required")
	}

	if _, err := base64.StdEncoding.DecodeString(r.SignedTransaction); err != nil {
		return fmt.Errorf("signed transaction: %w", err)
	}

	if r.LastValidBlockHeight != "" {
		_, err := strconv.ParseUint(r.LastValidBlockHeight, decimalBase, blockHeightBits)
		if err != nil {
			return fmt.Errorf("parse last valid block height: %w", err)
		}
	}

	return nil
}

func (c *Client) Execute(
	ctx context.Context,
	r ExecuteRequest,
) (*ExecuteResponse, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}

	// A context that has already ended guarantees nothing was sent, so the outcome is known.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	request := transport.Request{
		Method: http.MethodPost,
		Path:   "/swap/v2/execute",
		Body:   r,
		Retry:  executeRetryPolicy,
	}

	var result ExecuteResponse
	metadata, err := c.transport.Do(ctx, request, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, classifyExecuteFailure(err)
	}

	confirmed := result.Status == ExecutionStatusSuccess &&
		result.Code != nil &&
		*result.Code == executionSuccessCode
	if !confirmed {
		message := result.Error
		if message == "" {
			message = fmt.Sprintf("execution did not confirm success (status %q)", result.Status)
		}

		return &result, &ExecutionError{
			Code:      result.Code,
			Message:   message,
			RequestID: r.RequestID,
			Signature: result.Signature,
		}
	}

	return &result, nil
}

func (c *Client) ProgramLabels(ctx context.Context) (*ProgramLabelsResponse, error) {
	var result ProgramLabelsResponse
	request := transport.Request{
		Method: http.MethodGet,
		Path:   "/swap/v2/program-id-to-label",
		Retry:  transport.ReadOnlyRetry,
	}

	metadata, err := c.transport.Do(ctx, request, &result.Labels)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

type ExecutionError struct {
	Code      *int
	Message   string
	RequestID string
	Signature string
}

func (e *ExecutionError) Error() string {
	if e.Code == nil {
		return fmt.Sprintf("jupiter execution: %s", e.Message)
	}

	return fmt.Sprintf("jupiter execution %d: %s", *e.Code, e.Message)
}
