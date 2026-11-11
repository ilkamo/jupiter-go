package trigger

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ilkamo/jupiter-go/internal/address"
	"github.com/ilkamo/jupiter-go/internal/listparam"
	"github.com/ilkamo/jupiter-go/internal/transport"
	"github.com/ilkamo/jupiter-go/response"
)

// HistoryRequest filters and pages order history. State, Sort, and Direction take
// the documented constants; other values are passed through so upstream additions
// remain usable. Zero values use Jupiter's defaults: limit 20, offset 0, sort by
// update time, descending. Mint matches either side of the order.
type HistoryRequest struct {
	State     string
	Mint      string
	Limit     *uint8
	Offset    uint32
	Sort      string
	Direction string
}

func (r HistoryRequest) query() (url.Values, error) {
	query := url.Values{}

	for _, field := range []struct {
		name  string
		key   string
		value string
	}{
		{name: "state", key: "state", value: r.State},
		{name: "sort", key: "sort", value: r.Sort},
		{name: "direction", key: "dir", value: r.Direction},
	} {
		if field.value == "" {
			continue
		}

		if err := listparam.ValidateItem(field.name, field.value); err != nil {
			return nil, err
		}

		query.Set(field.key, field.value)
	}

	if r.Mint != "" {
		if err := address.Validate(r.Mint); err != nil {
			return nil, fmt.Errorf("mint: %w", err)
		}

		query.Set("mint", r.Mint)
	}

	if r.Limit != nil {
		if *r.Limit < MinHistoryLimit || *r.Limit > MaxHistoryLimit {
			return nil, fmt.Errorf("limit must be between %d and %d", MinHistoryLimit, MaxHistoryLimit)
		}

		query.Set("limit", strconv.Itoa(int(*r.Limit)))
	}

	if r.Offset > 0 {
		query.Set("offset", strconv.FormatUint(uint64(r.Offset), decimalBase))
	}

	return query, nil
}

type Pagination struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// Price-order timestamps are Unix milliseconds. Numbers are json.Number; an empty
// value means null or absent.

type PriceOrderHistory struct {
	Orders     []PriceOrderRecord `json:"orders"`
	Pagination Pagination         `json:"pagination"`

	Metadata response.Metadata `json:"-"`
}

type PriceOrderRecord struct {
	ID                   string            `json:"id"`
	OrderType            string            `json:"orderType"`
	OrderState           string            `json:"orderState"`
	RawState             string            `json:"rawState"`
	UserPubkey           string            `json:"userPubkey"`
	PrivyWalletPubkey    string            `json:"privyWalletPubkey"`
	InputMint            string            `json:"inputMint"`
	InitialInputAmount   string            `json:"initialInputAmount"`
	RemainingInputAmount string            `json:"remainingInputAmount"`
	OutputMint           string            `json:"outputMint"`
	TriggerMint          string            `json:"triggerMint"`
	TriggerCondition     string            `json:"triggerCondition"`
	TriggerPriceUSD      json.Number       `json:"triggerPriceUsd"`
	SlippageBps          json.Number       `json:"slippageBps"`
	ExpiresAt            *int64            `json:"expiresAt"`
	CreatedAt            int64             `json:"createdAt"`
	UpdatedAt            int64             `json:"updatedAt"`
	TriggeredAt          *int64            `json:"triggeredAt"`
	OutputAmount         *string           `json:"outputAmount"`
	InputUsed            *string           `json:"inputUsed"`
	FillPercent          json.Number       `json:"fillPercent"`
	TrailingBps          json.Number       `json:"trailingBps"`
	HighWatermark        json.Number       `json:"highWatermark"`
	LowWatermark         json.Number       `json:"lowWatermark"`
	Events               []PriceOrderEvent `json:"events"`
}

type PriceOrderEvent struct {
	Type         string `json:"type"`
	Timestamp    int64  `json:"timestamp"`
	TxSignature  string `json:"txSignature"`
	State        string `json:"state"`
	Mint         string `json:"mint"`
	Amount       string `json:"amount"`
	OutputMint   string `json:"outputMint"`
	OutputAmount string `json:"outputAmount"`
	OrderContext string `json:"orderContext"`
}

// PriceOrderHistory lists the wallet's price orders.
func (s *Session) PriceOrderHistory(ctx context.Context, r HistoryRequest) (*PriceOrderHistory, error) {
	var history PriceOrderHistory
	metadata, err := s.history(ctx, "/orders/history", r, &history)
	history.Metadata = metadata
	if err != nil {
		return &history, err
	}

	return &history, nil
}

// DCA timestamps are RFC 3339 strings.

type DCAOrderHistory struct {
	Orders     []DCAOrderRecord `json:"orders"`
	Pagination Pagination       `json:"pagination"`

	Metadata response.Metadata `json:"-"`
}

type DCAOrderRecord struct {
	ID                   string      `json:"id"`
	RequestID            string      `json:"requestId"`
	UserPubkey           string      `json:"userPubkey"`
	VaultPubkey          string      `json:"vaultPubkey"`
	InputMint            string      `json:"inputMint"`
	OutputMint           string      `json:"outputMint"`
	InputAmountInitial   string      `json:"inputAmountInitial"`
	InputAmountRemaining string      `json:"inputAmountRemaining"`
	AmountPerRound       string      `json:"amountPerRound"`
	OutputAmountTotal    string      `json:"outputAmountTotal"`
	InputAmountUsed      string      `json:"inputAmountUsed"`
	OrderType            string      `json:"orderType"`
	MinPriceUSD          json.Number `json:"minPriceUsd"`
	MaxPriceUSD          json.Number `json:"maxPriceUsd"`
	TriggerMint          *string     `json:"triggerMint"`
	RetryWindowSeconds   json.Number `json:"retryWindowSeconds"`
	JLEnabled            bool        `json:"jlEnabled"`
	JLYieldUSD           json.Number `json:"jlYieldUsd"`
	NumberOfRounds       int         `json:"numberOfRounds"`
	IntervalSeconds      int64       `json:"intervalSeconds"`
	BeginFillAt          string      `json:"beginFillAt"`
	NextFillAt           string      `json:"nextFillAt"`
	LastFillAt           *string     `json:"lastFillAt"`
	RoundsFilled         int         `json:"roundsFilled"`

	// FillPercent is progress from 0 to 1, unlike the price-order field.
	FillPercent  json.Number `json:"fillPercent"`
	State        string      `json:"state"`
	DisplayState string      `json:"displayState"`
	CreatedAt    string      `json:"createdAt"`
	UpdatedAt    string      `json:"updatedAt"`
	Events       []DCAEvent  `json:"events"`

	Metadata response.Metadata `json:"-"`
}

type DCAEvent struct {
	Type         string `json:"type"`
	Timestamp    string `json:"timestamp"`
	TxSignature  string `json:"txSignature"`
	RoundNumber  int    `json:"roundNumber"`
	InputAmount  string `json:"inputAmount"`
	OutputAmount string `json:"outputAmount"`
	State        string `json:"state"`
	AttemptCount int    `json:"attemptCount"`
}

// DCAOrderHistory lists the wallet's DCA orders.
func (s *Session) DCAOrderHistory(ctx context.Context, r HistoryRequest) (*DCAOrderHistory, error) {
	var history DCAOrderHistory
	metadata, err := s.history(ctx, "/orders/history/dca", r, &history)
	history.Metadata = metadata
	if err != nil {
		return &history, err
	}

	return &history, nil
}

// DCAOrder returns one DCA order with its events.
func (s *Session) DCAOrder(ctx context.Context, id string) (*DCAOrderRecord, error) {
	path, err := orderPath("/orders/history/dca/", id)
	if err != nil {
		return nil, err
	}

	request := transport.Request{
		Method: http.MethodGet,
		Path:   path,
		Retry:  transport.ReadOnlyRetry,
	}

	var order DCAOrderRecord
	metadata, err := s.do(ctx, request, &order)
	order.Metadata = metadata
	if err != nil {
		return &order, err
	}

	return &order, nil
}

func (s *Session) history(
	ctx context.Context,
	path string,
	r HistoryRequest,
	result any,
) (response.Metadata, error) {
	query, err := r.query()
	if err != nil {
		return response.Metadata{}, err
	}

	request := transport.Request{
		Method: http.MethodGet,
		Path:   basePath + path,
		Query:  query,
		Retry:  transport.ReadOnlyRetry,
	}

	return s.do(ctx, request, result)
}
