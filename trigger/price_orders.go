package trigger

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ilkamo/jupiter-go/internal/address"
	"github.com/ilkamo/jupiter-go/internal/amount"
	"github.com/ilkamo/jupiter-go/internal/transport"
	"github.com/ilkamo/jupiter-go/response"
)

// PriceOrderRequest creates a price order funded by a signed deposit. OrderType is
// OrderSubTypeSingle, OrderSubTypeOCO, or OrderSubTypeOTOCO and must match the
// deposit's sub-type. Prices are USD JSON numbers, kept exact as json.Number.
//
//   - single: TriggerCondition plus exactly one of TriggerPriceUSD or TrailingBps.
//     A trailing order uses TriggerMint == InputMint for TriggerBelow and
//     TriggerMint == OutputMint for TriggerAbove.
//   - oco: TakeProfitPriceUSD above StopLossPriceUSD; no trigger fields.
//   - otoco: TriggerCondition and TriggerPriceUSD for the parent, plus the
//     take-profit and stop-loss pair.
//
// Jupiter enforces a 10 USD minimum and does not check whether a trigger price
// makes economic sense.
type PriceOrderRequest struct {
	OrderType                string
	DepositRequestID         string
	DepositSignedTransaction string
	InputMint                string
	InputAmount              string
	OutputMint               string
	TriggerMint              string
	TriggerCondition         string
	TriggerPriceUSD          json.Number
	TrailingBps              *uint16
	SlippageBps              *uint16
	TakeProfitPriceUSD       json.Number
	StopLossPriceUSD         json.Number
	TakeProfitSlippageBps    *uint16
	StopLossSlippageBps      *uint16
	ExpiresAt                time.Time
}

type priceOrderBody struct {
	OrderType        string      `json:"orderType"`
	DepositRequestID string      `json:"depositRequestId"`
	DepositSignedTx  string      `json:"depositSignedTx"`
	UserPubkey       string      `json:"userPubkey"`
	InputMint        string      `json:"inputMint"`
	InputAmount      string      `json:"inputAmount"`
	OutputMint       string      `json:"outputMint"`
	TriggerMint      string      `json:"triggerMint"`
	TriggerCondition string      `json:"triggerCondition,omitempty"`
	TriggerPriceUSD  json.Number `json:"triggerPriceUsd,omitempty"`
	TrailingBps      *uint16     `json:"trailingBps,omitempty"`
	SlippageBps      *uint16     `json:"slippageBps,omitempty"`
	TPPriceUSD       json.Number `json:"tpPriceUsd,omitempty"`
	SLPriceUSD       json.Number `json:"slPriceUsd,omitempty"`
	TPSlippageBps    *uint16     `json:"tpSlippageBps,omitempty"`
	SLSlippageBps    *uint16     `json:"slSlippageBps,omitempty"`
	ExpiresAt        int64       `json:"expiresAt"`
}

// PriceOrder is the result of creating a price order. The deposit lands during the
// call; DepositConfirmed reports whether it has landed and the order is active.
type PriceOrder struct {
	ID               string `json:"id"`
	TxSignature      string `json:"txSignature"`
	DepositConfirmed bool   `json:"depositConfirmed"`
	Message          string `json:"message"`

	Metadata response.Metadata `json:"-"`
}

func (r PriceOrderRequest) body(wallet string, now time.Time) (priceOrderBody, error) {
	if err := validateFunding(
		r.DepositRequestID,
		r.DepositSignedTransaction,
		r.InputMint,
		r.OutputMint,
	); err != nil {
		return priceOrderBody{}, err
	}

	inputAmount, err := amount.ParseRaw(r.InputAmount)
	if err != nil {
		return priceOrderBody{}, fmt.Errorf("input %w", err)
	}

	if err := address.Validate(r.TriggerMint); err != nil {
		return priceOrderBody{}, fmt.Errorf("trigger mint: %w", err)
	}

	if !r.ExpiresAt.After(now) {
		return priceOrderBody{}, fmt.Errorf("price orders require a future expiry")
	}

	if err := r.validateVariant(); err != nil {
		return priceOrderBody{}, err
	}

	return priceOrderBody{
		OrderType:        r.OrderType,
		DepositRequestID: r.DepositRequestID,
		DepositSignedTx:  r.DepositSignedTransaction,
		UserPubkey:       wallet,
		InputMint:        r.InputMint,
		InputAmount:      inputAmount,
		OutputMint:       r.OutputMint,
		TriggerMint:      r.TriggerMint,
		TriggerCondition: r.TriggerCondition,
		TriggerPriceUSD:  r.TriggerPriceUSD,
		TrailingBps:      r.TrailingBps,
		SlippageBps:      r.SlippageBps,
		TPPriceUSD:       r.TakeProfitPriceUSD,
		SLPriceUSD:       r.StopLossPriceUSD,
		TPSlippageBps:    r.TakeProfitSlippageBps,
		SLSlippageBps:    r.StopLossSlippageBps,
		ExpiresAt:        r.ExpiresAt.UnixMilli(),
	}, nil
}

func (r PriceOrderRequest) validateVariant() error {
	fields := priceFields{
		orderType:             r.OrderType,
		triggerCondition:      r.TriggerCondition,
		triggerPriceUSD:       r.TriggerPriceUSD,
		trailingBps:           r.TrailingBps,
		slippageBps:           r.SlippageBps,
		takeProfitPriceUSD:    r.TakeProfitPriceUSD,
		stopLossPriceUSD:      r.StopLossPriceUSD,
		takeProfitSlippageBps: r.TakeProfitSlippageBps,
		stopLossSlippageBps:   r.StopLossSlippageBps,
	}
	if err := fields.validate(); err != nil {
		return err
	}

	switch r.OrderType {
	case OrderSubTypeSingle:
		if r.TriggerCondition == "" {
			return fmt.Errorf("single orders require a trigger condition")
		}

		if (r.TriggerPriceUSD == "") == (r.TrailingBps == nil) {
			return fmt.Errorf("single orders require exactly one of a trigger price or trailing distance")
		}

		if r.TrailingBps != nil {
			return r.validateTrailingMint()
		}
	case OrderSubTypeOCO:
		if r.TakeProfitPriceUSD == "" || r.StopLossPriceUSD == "" {
			return fmt.Errorf("oco orders require take-profit and stop-loss prices")
		}
	case OrderSubTypeOTOCO:
		if r.TriggerCondition == "" || r.TriggerPriceUSD == "" {
			return fmt.Errorf("otoco orders require a trigger condition and price")
		}

		if r.TakeProfitPriceUSD == "" || r.StopLossPriceUSD == "" {
			return fmt.Errorf("otoco orders require take-profit and stop-loss prices")
		}
	}

	return nil
}

func (r PriceOrderRequest) validateTrailingMint() error {
	switch r.TriggerCondition {
	case TriggerBelow:
		if r.TriggerMint != r.InputMint {
			return fmt.Errorf("a trailing sell-below order must trigger on the input mint")
		}
	case TriggerAbove:
		if r.TriggerMint != r.OutputMint {
			return fmt.Errorf("a trailing buy-above order must trigger on the output mint")
		}
	}

	return nil
}

// priceFields holds the variant-specific fields shared by create and update.
type priceFields struct {
	orderType             string
	triggerCondition      string
	triggerPriceUSD       json.Number
	trailingBps           *uint16
	slippageBps           *uint16
	takeProfitPriceUSD    json.Number
	stopLossPriceUSD      json.Number
	takeProfitSlippageBps *uint16
	stopLossSlippageBps   *uint16
}

// validate checks value ranges and rejects fields that do not apply to the
// variant. It does not decide which fields are required.
func (f priceFields) validate() error {
	switch f.orderType {
	case OrderSubTypeSingle, OrderSubTypeOCO, OrderSubTypeOTOCO:
	default:
		return fmt.Errorf(
			"order type must be %q, %q, or %q",
			OrderSubTypeSingle,
			OrderSubTypeOCO,
			OrderSubTypeOTOCO,
		)
	}

	switch f.triggerCondition {
	case "", TriggerAbove, TriggerBelow:
	default:
		return fmt.Errorf("trigger condition must be %q or %q", TriggerAbove, TriggerBelow)
	}

	hasTrigger := f.triggerCondition != "" || f.triggerPriceUSD != ""
	hasPair := f.takeProfitPriceUSD != "" ||
		f.stopLossPriceUSD != "" ||
		f.takeProfitSlippageBps != nil ||
		f.stopLossSlippageBps != nil

	if f.orderType == OrderSubTypeOCO && hasTrigger {
		return fmt.Errorf("oco orders take no trigger condition or price")
	}

	if f.orderType == OrderSubTypeSingle && hasPair {
		return fmt.Errorf("single orders take no take-profit or stop-loss fields")
	}

	if f.trailingBps != nil && f.orderType != OrderSubTypeSingle {
		return fmt.Errorf("only single orders can trail")
	}

	if f.trailingBps != nil && f.triggerPriceUSD != "" {
		return fmt.Errorf("set a trigger price or a trailing distance, not both")
	}

	if err := validateTrailing(f.trailingBps); err != nil {
		return err
	}

	if err := validateSlippage("slippage", f.slippageBps); err != nil {
		return err
	}

	if err := validateSlippage("take-profit slippage", f.takeProfitSlippageBps); err != nil {
		return err
	}

	if err := validateSlippage("stop-loss slippage", f.stopLossSlippageBps); err != nil {
		return err
	}

	if _, err := parseOptionalUSD("trigger price", f.triggerPriceUSD); err != nil {
		return err
	}

	return validatePricePair(f.takeProfitPriceUSD, f.stopLossPriceUSD)
}

func validatePricePair(takeProfit json.Number, stopLoss json.Number) error {
	tp, err := parseOptionalUSD("take-profit price", takeProfit)
	if err != nil {
		return err
	}

	sl, err := parseOptionalUSD("stop-loss price", stopLoss)
	if err != nil {
		return err
	}

	if tp != nil && sl != nil && tp.Cmp(sl) <= 0 {
		return fmt.Errorf("take-profit price must be above the stop-loss price")
	}

	return nil
}

func validateFunding(
	depositRequestID string,
	depositSignedTransaction string,
	inputMint string,
	outputMint string,
) error {
	if depositRequestID == "" {
		return fmt.Errorf("deposit request ID required")
	}

	if err := validateSignedTransaction("signed deposit transaction", depositSignedTransaction); err != nil {
		return err
	}

	if err := address.Validate(inputMint); err != nil {
		return fmt.Errorf("input mint: %w", err)
	}

	if err := address.Validate(outputMint); err != nil {
		return fmt.Errorf("output mint: %w", err)
	}

	if inputMint == outputMint {
		return fmt.Errorf("input and output mints must differ")
	}

	return nil
}

// CreatePriceOrder submits the signed deposit and creates the order. It makes one
// attempt; failures after which the deposit may have landed wrap
// ErrOutcomeUnknown. Check PriceOrderHistory before creating another order.
func (s *Session) CreatePriceOrder(ctx context.Context, r PriceOrderRequest) (*PriceOrder, error) {
	body, err := r.body(s.wallet, s.now())
	if err != nil {
		return nil, err
	}

	request := transport.Request{
		Method: http.MethodPost,
		Path:   basePath + "/orders/price",
		Body:   body,
		Retry:  transport.NoRetry,
	}

	var order PriceOrder
	metadata, err := s.submit(ctx, request, &order)
	order.Metadata = metadata
	if err != nil {
		return &order, err
	}

	return &order, nil
}

// PriceOrderUpdate changes an open price order in place. OrderType is required and
// must be the order's type; set only the fields to change.
type PriceOrderUpdate struct {
	OrderType             string
	TriggerPriceUSD       json.Number
	TrailingBps           *uint16
	SlippageBps           *uint16
	TakeProfitPriceUSD    json.Number
	StopLossPriceUSD      json.Number
	TakeProfitSlippageBps *uint16
	StopLossSlippageBps   *uint16
}

type priceOrderUpdateBody struct {
	OrderType       string      `json:"orderType"`
	TriggerPriceUSD json.Number `json:"triggerPriceUsd,omitempty"`
	TrailingBps     *uint16     `json:"trailingBps,omitempty"`
	SlippageBps     *uint16     `json:"slippageBps,omitempty"`
	TPPriceUSD      json.Number `json:"tpPriceUsd,omitempty"`
	SLPriceUSD      json.Number `json:"slPriceUsd,omitempty"`
	TPSlippageBps   *uint16     `json:"tpSlippageBps,omitempty"`
	SLSlippageBps   *uint16     `json:"slSlippageBps,omitempty"`
}

// OrderUpdate is the result of UpdatePriceOrder.
type OrderUpdate struct {
	ID string `json:"id"`

	Metadata response.Metadata `json:"-"`
}

func (u PriceOrderUpdate) body() (priceOrderUpdateBody, error) {
	fields := priceFields{
		orderType:             u.OrderType,
		triggerPriceUSD:       u.TriggerPriceUSD,
		trailingBps:           u.TrailingBps,
		slippageBps:           u.SlippageBps,
		takeProfitPriceUSD:    u.TakeProfitPriceUSD,
		stopLossPriceUSD:      u.StopLossPriceUSD,
		takeProfitSlippageBps: u.TakeProfitSlippageBps,
		stopLossSlippageBps:   u.StopLossSlippageBps,
	}
	if err := fields.validate(); err != nil {
		return priceOrderUpdateBody{}, err
	}

	body := priceOrderUpdateBody{
		OrderType:       u.OrderType,
		TriggerPriceUSD: u.TriggerPriceUSD,
		TrailingBps:     u.TrailingBps,
		SlippageBps:     u.SlippageBps,
		TPPriceUSD:      u.TakeProfitPriceUSD,
		SLPriceUSD:      u.StopLossPriceUSD,
		TPSlippageBps:   u.TakeProfitSlippageBps,
		SLSlippageBps:   u.StopLossSlippageBps,
	}

	if body == (priceOrderUpdateBody{OrderType: u.OrderType}) {
		return priceOrderUpdateBody{}, fmt.Errorf("update changes no fields")
	}

	return body, nil
}

// UpdatePriceOrder changes an open order. Updates set absolute values, so
// transient failures are retried.
func (s *Session) UpdatePriceOrder(
	ctx context.Context,
	id string,
	u PriceOrderUpdate,
) (*OrderUpdate, error) {
	path, err := orderPath("/orders/price/", id)
	if err != nil {
		return nil, err
	}

	body, err := u.body()
	if err != nil {
		return nil, err
	}

	request := transport.Request{
		Method: http.MethodPatch,
		Path:   path,
		Body:   body,
		Retry:  repeatableRetry,
	}

	var result OrderUpdate
	metadata, err := s.do(ctx, request, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

// Cancellation is an unsigned withdrawal transaction from step one of a
// two-step cancellation. RoundsRemaining and RefundAmount are set for DCA orders.
type Cancellation struct {
	ID              string `json:"id"`
	Transaction     string `json:"transaction"`
	RequestID       string `json:"requestId"`
	RoundsRemaining int    `json:"roundsRemaining"`
	RefundAmount    string `json:"refundAmount"`

	Metadata response.Metadata `json:"-"`
}

// ConfirmCancelRequest completes a cancellation with the signed withdrawal
// transaction and the Cancellation's RequestID.
type ConfirmCancelRequest struct {
	SignedTransaction string `json:"signedTransaction"`
	CancelRequestID   string `json:"cancelRequestId"`
}

func (r ConfirmCancelRequest) validate() error {
	if r.CancelRequestID == "" {
		return fmt.Errorf("cancel request ID required")
	}

	return validateSignedTransaction("signed withdrawal transaction", r.SignedTransaction)
}

// Submission identifies an order and the on-chain transaction submitted for it.
type Submission struct {
	ID          string `json:"id"`
	TxSignature string `json:"txSignature"`

	Metadata response.Metadata `json:"-"`
}

// CancelPriceOrder starts cancelling an order, including an expired one whose funds
// remain in the vault. It only crafts a withdrawal transaction, so transient
// failures are retried. HTTP 400 can mean the order is not yet ready to cancel.
func (s *Session) CancelPriceOrder(ctx context.Context, id string) (*Cancellation, error) {
	return s.cancel(ctx, "/orders/price/cancel/", id)
}

// ConfirmCancelPriceOrder submits the signed withdrawal. It makes one attempt;
// failures after which the withdrawal may have landed wrap ErrOutcomeUnknown. If
// the withdrawal did not land, the documented recovery is to confirm again with
// the same cancel request ID.
func (s *Session) ConfirmCancelPriceOrder(
	ctx context.Context,
	id string,
	r ConfirmCancelRequest,
) (*Submission, error) {
	return s.confirmCancel(ctx, "/orders/price/confirm-cancel/", id, r)
}

func (s *Session) cancel(ctx context.Context, prefix string, id string) (*Cancellation, error) {
	path, err := orderPath(prefix, id)
	if err != nil {
		return nil, err
	}

	request := transport.Request{
		Method: http.MethodPost,
		Path:   path,
		Body:   struct{}{},
		Retry:  repeatableRetry,
	}

	var cancellation Cancellation
	metadata, err := s.do(ctx, request, &cancellation)
	cancellation.Metadata = metadata
	if err != nil {
		return &cancellation, err
	}

	return &cancellation, nil
}

func (s *Session) confirmCancel(
	ctx context.Context,
	prefix string,
	id string,
	r ConfirmCancelRequest,
) (*Submission, error) {
	path, err := orderPath(prefix, id)
	if err != nil {
		return nil, err
	}

	if err := r.validate(); err != nil {
		return nil, err
	}

	request := transport.Request{
		Method: http.MethodPost,
		Path:   path,
		Body:   r,
		Retry:  transport.NoRetry,
	}

	var submission Submission
	metadata, err := s.submit(ctx, request, &submission)
	submission.Metadata = metadata
	if err != nil {
		return &submission, err
	}

	return &submission, nil
}
