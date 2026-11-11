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
)

const (
	MinDCAOrderCount   = 2
	MinDCAInterval     = time.Minute
	MaxDCAInterval     = 365 * 24 * time.Hour
	MaxDCAStartDelay   = 30 * 24 * time.Hour
	dcaIntervalUnit    = time.Second
	dcaTimestampLayout = time.RFC3339Nano
)

// DCAOrderRequest splits a signed deposit, crafted with OrderTypeDCA, into
// OrderCount equal rounds every Interval. OrderType is DCATimeBased (the default
// when empty) or DCAPriceConditional, which fills only while TriggerMint's USD price
// is within [MinPriceUSD, MaxPriceUSD]. JLEnabled (Earn While You Wait) requires
// DCATimeBased, a supported stablecoin input, and a deposit crafted with JLMint.
//
// Jupiter enforces at least 10 USD per round and at most 10 active DCA orders.
// Rounds use automatic slippage; there is no slippage parameter.
type DCAOrderRequest struct {
	DepositRequestID         string
	DepositSignedTransaction string
	InputMint                string
	OutputMint               string
	InputAmount              string
	OrderCount               uint32
	Interval                 time.Duration
	OrderType                string
	MinPriceUSD              json.Number
	MaxPriceUSD              json.Number
	TriggerMint              string
	BeginFillAt              time.Time
	JLEnabled                bool
	JLMint                   string
}

type dcaOrderBody struct {
	DepositRequestID string      `json:"depositRequestId"`
	DepositSignedTx  string      `json:"depositSignedTx"`
	UserPubkey       string      `json:"userPubkey"`
	InputMint        string      `json:"inputMint"`
	OutputMint       string      `json:"outputMint"`
	InputAmount      string      `json:"inputAmount"`
	OrderCount       uint32      `json:"orderCount"`
	IntervalSeconds  int64       `json:"intervalSeconds"`
	OrderType        string      `json:"orderType,omitempty"`
	MinPriceUSD      json.Number `json:"minPriceUsd,omitempty"`
	MaxPriceUSD      json.Number `json:"maxPriceUsd,omitempty"`
	TriggerMint      string      `json:"triggerMint,omitempty"`
	BeginFillAt      string      `json:"beginFillAt,omitempty"`
	JLEnabled        bool        `json:"jlEnabled,omitempty"`
	JLMint           string      `json:"jlMint,omitempty"`
}

func (r DCAOrderRequest) body(wallet string, now time.Time) (dcaOrderBody, error) {
	if err := validateFunding(
		r.DepositRequestID,
		r.DepositSignedTransaction,
		r.InputMint,
		r.OutputMint,
	); err != nil {
		return dcaOrderBody{}, err
	}

	inputAmount, err := amount.ParseRaw(r.InputAmount)
	if err != nil {
		return dcaOrderBody{}, fmt.Errorf("input %w", err)
	}

	if r.OrderCount < MinDCAOrderCount {
		return dcaOrderBody{}, fmt.Errorf("DCA orders require at least %d rounds", MinDCAOrderCount)
	}

	if r.Interval < MinDCAInterval || r.Interval > MaxDCAInterval || r.Interval%dcaIntervalUnit != 0 {
		return dcaOrderBody{}, fmt.Errorf(
			"interval must be whole seconds between %s and %s",
			MinDCAInterval,
			MaxDCAInterval,
		)
	}

	if err := r.validateType(); err != nil {
		return dcaOrderBody{}, err
	}

	beginFillAt := ""
	if !r.BeginFillAt.IsZero() {
		if r.BeginFillAt.After(now.Add(MaxDCAStartDelay)) {
			return dcaOrderBody{}, fmt.Errorf("the first round can start at most %s ahead", MaxDCAStartDelay)
		}

		beginFillAt = r.BeginFillAt.UTC().Format(dcaTimestampLayout)
	}

	return dcaOrderBody{
		DepositRequestID: r.DepositRequestID,
		DepositSignedTx:  r.DepositSignedTransaction,
		UserPubkey:       wallet,
		InputMint:        r.InputMint,
		OutputMint:       r.OutputMint,
		InputAmount:      inputAmount,
		OrderCount:       r.OrderCount,
		IntervalSeconds:  int64(r.Interval / dcaIntervalUnit),
		OrderType:        r.OrderType,
		MinPriceUSD:      r.MinPriceUSD,
		MaxPriceUSD:      r.MaxPriceUSD,
		TriggerMint:      r.TriggerMint,
		BeginFillAt:      beginFillAt,
		JLEnabled:        r.JLEnabled,
		JLMint:           r.JLMint,
	}, nil
}

func (r DCAOrderRequest) validateType() error {
	priceBandSet := r.MinPriceUSD != "" || r.MaxPriceUSD != ""

	switch r.OrderType {
	case "", DCATimeBased:
		if priceBandSet || r.TriggerMint != "" {
			return fmt.Errorf("price bounds and trigger mint apply only to price-conditional DCA")
		}
	case DCAPriceConditional:
		if err := address.Validate(r.TriggerMint); err != nil {
			return fmt.Errorf("price-conditional DCA trigger mint: %w", err)
		}

		if r.JLEnabled {
			return fmt.Errorf("Earn While You Wait requires time-based DCA")
		}
	default:
		return fmt.Errorf("DCA order type must be %q or %q", DCATimeBased, DCAPriceConditional)
	}

	minPrice, err := parseOptionalUSD("minimum price", r.MinPriceUSD)
	if err != nil {
		return err
	}

	maxPrice, err := parseOptionalUSD("maximum price", r.MaxPriceUSD)
	if err != nil {
		return err
	}

	if minPrice != nil && maxPrice != nil && minPrice.Cmp(maxPrice) >= 0 {
		return fmt.Errorf("minimum price must be below the maximum price")
	}

	if r.JLEnabled != (r.JLMint != "") {
		return fmt.Errorf("Earn While You Wait requires both JLEnabled and JLMint")
	}

	if r.JLMint != "" {
		if err := address.Validate(r.JLMint); err != nil {
			return fmt.Errorf("JL mint: %w", err)
		}
	}

	return nil
}

// CreateDCAOrder submits the signed deposit and creates the DCA order. It makes
// one attempt; failures after which the deposit may have landed wrap
// ErrOutcomeUnknown. Check DCAOrderHistory before creating another order.
func (s *Session) CreateDCAOrder(ctx context.Context, r DCAOrderRequest) (*Submission, error) {
	body, err := r.body(s.wallet, s.now())
	if err != nil {
		return nil, err
	}

	// The documented path has no trailing slash; "/orders/dca/" returns 404.
	request := transport.Request{
		Method: http.MethodPost,
		Path:   basePath + "/orders/dca",
		Body:   body,
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

// CancelDCAOrder starts cancelling a DCA order and returns the withdrawal for the
// unfilled remainder. Filled rounds are not reversed and the order cannot return
// to active. Repeating the call while withdrawing is documented as idempotent.
func (s *Session) CancelDCAOrder(ctx context.Context, id string) (*Cancellation, error) {
	return s.cancel(ctx, "/orders/dca/cancel/", id)
}

// ConfirmCancelDCAOrder submits the signed withdrawal. It makes one attempt;
// failures after which the withdrawal may have landed wrap ErrOutcomeUnknown. A
// failed confirmation leaves the order withdrawing; call CancelDCAOrder again to
// re-craft the withdrawal.
func (s *Session) ConfirmCancelDCAOrder(
	ctx context.Context,
	id string,
	r ConfirmCancelRequest,
) (*Submission, error) {
	return s.confirmCancel(ctx, "/orders/dca/confirm-cancel/", id, r)
}
