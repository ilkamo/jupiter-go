package tokens

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ilkamo/jupiter-go/internal/address"
	"github.com/ilkamo/jupiter-go/internal/amount"
	"github.com/ilkamo/jupiter-go/internal/transport"
	"github.com/ilkamo/jupiter-go/response"
)

// Payment currencies are token symbols, not mints. JUP transfers directly; the
// others are swapped to 1000 JUP.
const (
	PaymentJUP    = "JUP"
	PaymentSOL    = "SOL"
	PaymentUSDC   = "USDC"
	PaymentJUPUSD = "JUPUSD"

	ExpressStatusSuccess = "Success"
	ExpressStatusFailed  = "Failed"

	expressPath = "/verify/express"
)

// ErrOutcomeUnknown marks Express Execute failures after which the 1000 JUP
// payment may have landed. Check the submission at verified.jup.ag before paying
// again.
var ErrOutcomeUnknown = errors.New("express verification outcome unknown")

// ExpressError reports a craft or execute response that did not succeed. Code is
// nil when the API sent none.
type ExpressError struct {
	Code      *int
	Message   string
	Signature string
}

func (e *ExpressError) Error() string {
	if e.Code == nil {
		return fmt.Sprintf("express verification: %s", e.Message)
	}

	return fmt.Sprintf("express verification %d: %s", *e.Code, e.Message)
}

// ExpressClient submits tokens to Jupiter VRFD Express Verification, which costs
// 1000 JUP per submission. Standard submissions on verified.jup.ag are free.
type ExpressClient struct {
	transport *transport.Client
}

// Eligibility reports whether verification and metadata updates can be submitted.
// If neither can, Execute rejects the submission before charging.
type Eligibility struct {
	TokenExists       bool   `json:"tokenExists"`
	IsVerified        bool   `json:"isVerified"`
	CanVerify         bool   `json:"canVerify"`
	CanMetadata       bool   `json:"canMetadata"`
	VerificationError string `json:"verificationError"`
	MetadataError     string `json:"metadataError"`

	Metadata response.Metadata `json:"-"`
}

func (c *ExpressClient) CheckEligibility(ctx context.Context, tokenID string) (*Eligibility, error) {
	if err := address.Validate(tokenID); err != nil {
		return nil, fmt.Errorf("token: %w", err)
	}

	request := transport.Request{
		Method: http.MethodGet,
		Path:   basePath + expressPath + "/check-eligibility",
		Query:  url.Values{"tokenId": {tokenID}},
		Retry:  transport.ReadOnlyRetry,
	}

	var eligibility Eligibility
	metadata, err := c.transport.Do(ctx, request, &eligibility)
	eligibility.Metadata = metadata
	if err != nil {
		return &eligibility, err
	}

	return &eligibility, nil
}

// Payment is an unsigned 1000 JUP payment transaction. For non-JUP currencies the
// Input* and *InputAmount fields describe the swap; pass QuotedInputAmount to
// ExpressRequest.PaymentAmount and Amount to JUPOutputAmount.
type Payment struct {
	Transaction       string      `json:"transaction"`
	RequestID         string      `json:"requestId"`
	ReceiverAddress   string      `json:"receiverAddress"`
	Mint              string      `json:"mint"`
	Amount            string      `json:"amount"`
	TokenDecimals     int         `json:"tokenDecimals"`
	TokenUSDRate      json.Number `json:"tokenUsdRate"`
	FeeLamports       json.Number `json:"feeLamports"`
	FeeUSDAmount      json.Number `json:"feeUsdAmount"`
	FeeMint           string      `json:"feeMint"`
	FeeTokenDecimals  int         `json:"feeTokenDecimals"`
	FeeAmount         json.Number `json:"feeAmount"`
	TotalTime         json.Number `json:"totalTime"`
	ExpireAt          string      `json:"expireAt"`
	Error             string      `json:"error"`
	Code              *int        `json:"code"`
	Gasless           bool        `json:"gasless"`
	InputMint         string      `json:"inputMint"`
	InputDecimals     int         `json:"inputDecimals"`
	QuotedInputAmount string      `json:"quotedInputAmount"`
	MaxInputAmount    string      `json:"maxInputAmount"`

	Metadata response.Metadata `json:"-"`
}

// CraftPayment builds the payment transaction for sender, paying in currency
// (PaymentJUP when empty). Crafting submits nothing and returns a fresh request
// ID, so transient failures are retried.
func (c *ExpressClient) CraftPayment(ctx context.Context, sender string, currency string) (*Payment, error) {
	if err := address.Validate(sender); err != nil {
		return nil, fmt.Errorf("sender: %w", err)
	}

	query := url.Values{"senderAddress": {sender}}
	if currency != "" {
		query.Set("paymentCurrency", currency)
	}

	request := transport.Request{
		Method: http.MethodGet,
		Path:   basePath + expressPath + "/craft-txn",
		Query:  query,
		Retry:  transport.RepeatableRetry,
	}

	var payment Payment
	metadata, err := c.transport.Do(ctx, request, &payment)
	payment.Metadata = metadata
	if err != nil {
		return &payment, err
	}

	failed := payment.Error != "" || (payment.Code != nil && *payment.Code != 0)
	if failed {
		return &payment, &ExpressError{Code: payment.Code, Message: payment.Error}
	}

	return &payment, nil
}

// TokenMetadataUpdate submits metadata alongside verification. Only TokenID is
// required; unset fields are omitted. Use* flags select the provided supply or
// price sources.
type TokenMetadataUpdate struct {
	TokenID                 string `json:"tokenId"`
	Name                    string `json:"name,omitempty"`
	Symbol                  string `json:"symbol,omitempty"`
	Icon                    string `json:"icon,omitempty"`
	Website                 string `json:"website,omitempty"`
	Twitter                 string `json:"twitter,omitempty"`
	TwitterCommunity        string `json:"twitterCommunity,omitempty"`
	Telegram                string `json:"telegram,omitempty"`
	Discord                 string `json:"discord,omitempty"`
	Instagram               string `json:"instagram,omitempty"`
	TikTok                  string `json:"tiktok,omitempty"`
	TokenDescription        string `json:"tokenDescription,omitempty"`
	CirculatingSupply       string `json:"circulatingSupply,omitempty"`
	UseCirculatingSupply    *bool  `json:"useCirculatingSupply,omitempty"`
	CirculatingSupplyURL    string `json:"circulatingSupplyUrl,omitempty"`
	UseCirculatingSupplyURL *bool  `json:"useCirculatingSupplyUrl,omitempty"`
	CoingeckoCoinID         string `json:"coingeckoCoinId,omitempty"`
	UseCoingeckoCoinID      *bool  `json:"useCoingeckoCoinId,omitempty"`
	OtherURL                string `json:"otherUrl,omitempty"`
}

// ExpressRequest submits the signed payment with the verification request.
// PaymentCurrency must match CraftPayment; for currencies other than PaymentJUP,
// PaymentAmount is required.
type ExpressRequest struct {
	Transaction         string               `json:"transaction"`
	RequestID           string               `json:"requestId"`
	SenderAddress       string               `json:"senderAddress"`
	TokenID             string               `json:"tokenId"`
	TwitterHandle       string               `json:"twitterHandle"`
	SenderTwitterHandle string               `json:"senderTwitterHandle,omitempty"`
	Description         string               `json:"description"`
	TokenMetadata       *TokenMetadataUpdate `json:"tokenMetadata,omitempty"`
	PaymentCurrency     string               `json:"paymentCurrency,omitempty"`
	PaymentAmount       string               `json:"paymentAmount,omitempty"`
	JUPOutputAmount     string               `json:"jupOutputAmount,omitempty"`
}

func (r ExpressRequest) normalized() (ExpressRequest, error) {
	if r.Transaction == "" || r.RequestID == "" {
		return ExpressRequest{}, fmt.Errorf("signed transaction and request ID required")
	}

	if _, err := base64.StdEncoding.DecodeString(r.Transaction); err != nil {
		return ExpressRequest{}, fmt.Errorf("signed transaction: %w", err)
	}

	if err := address.Validate(r.SenderAddress); err != nil {
		return ExpressRequest{}, fmt.Errorf("sender: %w", err)
	}

	if err := address.Validate(r.TokenID); err != nil {
		return ExpressRequest{}, fmt.Errorf("token: %w", err)
	}

	if r.TwitterHandle == "" || r.Description == "" {
		return ExpressRequest{}, fmt.Errorf("twitter handle and description required")
	}

	if r.TokenMetadata != nil {
		if err := address.Validate(r.TokenMetadata.TokenID); err != nil {
			return ExpressRequest{}, fmt.Errorf("metadata token: %w", err)
		}
	}

	paidInJUP := r.PaymentCurrency == "" || r.PaymentCurrency == PaymentJUP
	if !paidInJUP && r.PaymentAmount == "" {
		return ExpressRequest{}, fmt.Errorf("payment amount required for %s payments", r.PaymentCurrency)
	}

	if r.PaymentAmount != "" {
		paid, err := amount.ParseRaw(r.PaymentAmount)
		if err != nil {
			return ExpressRequest{}, fmt.Errorf("payment %w", err)
		}

		r.PaymentAmount = paid
	}

	if r.JUPOutputAmount != "" {
		output, err := amount.ParseRaw(r.JUPOutputAmount)
		if err != nil {
			return ExpressRequest{}, fmt.Errorf("JUP output %w", err)
		}

		r.JUPOutputAmount = output
	}

	return r, nil
}

type ExpressResult struct {
	Status              string      `json:"status"`
	Signature           string      `json:"signature"`
	Error               string      `json:"error"`
	Code                *int        `json:"code"`
	TotalTime           json.Number `json:"totalTime"`
	VerificationCreated bool        `json:"verificationCreated"`
	MetadataCreated     bool        `json:"metadataCreated"`

	Metadata response.Metadata `json:"-"`
}

// Execute submits the signed payment and creates the verification and metadata
// requests, which are reviewed independently. It makes one attempt because the
// payment is spent on success; a response without an explicit Success status
// returns an *ExpressError, and failures after which the payment may have landed
// wrap ErrOutcomeUnknown.
func (c *ExpressClient) Execute(ctx context.Context, r ExpressRequest) (*ExpressResult, error) {
	body, err := r.normalized()
	if err != nil {
		return nil, err
	}

	// An ended context guarantees nothing is sent, so the outcome is known.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	request := transport.Request{
		Method: http.MethodPost,
		Path:   basePath + expressPath + "/execute",
		Body:   body,
		Retry:  transport.NoRetry,
	}

	var result ExpressResult
	metadata, err := c.transport.Do(ctx, request, &result)
	result.Metadata = metadata
	if err != nil {
		if transport.MayHaveBeenProcessed(err) {
			return &result, fmt.Errorf("%w: %w", ErrOutcomeUnknown, err)
		}

		return &result, err
	}

	if result.Status != ExpressStatusSuccess {
		message := result.Error
		if message == "" {
			message = fmt.Sprintf("execution did not confirm success (status %q)", result.Status)
		}

		return &result, &ExpressError{Code: result.Code, Message: message, Signature: result.Signature}
	}

	return &result, nil
}
