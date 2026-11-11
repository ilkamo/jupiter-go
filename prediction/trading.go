package prediction

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ilkamo/jupiter-go/internal/amount"
	"github.com/ilkamo/jupiter-go/internal/transport"
	"github.com/ilkamo/jupiter-go/response"
)

const (
	// Deposit mints documented for orders; both use 6 decimals.
	MintJupUSD = "JuprjznTrTSp2UFa3ZBUFgwdAmtZCq4MQCwysN55USD"
	MintUSDC   = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"

	ExecutionStatusSuccess = "Success"
	ExecutionStatusFailed  = "Failed"

	BasisPointsPerWhole = 10000
)

var (
	// ErrOutcomeUnknown marks Execute failures after which the transaction may still
	// land. Check the order status or positions before building another order.
	ErrOutcomeUnknown = errors.New("prediction execution outcome unknown")
)

// ExecutionError reports an Execute response that did not confirm success.
type ExecutionError struct {
	Status    string
	Message   string
	Signature string
	RequestID string
}

func (e *ExecutionError) Error() string {
	return fmt.Sprintf("prediction execution %s: %s", e.Status, e.Message)
}

// OrderRequest builds a buy or sell order for Owner.
//
//   - Buy (IsBuy): MarketID, IsYes, DepositAmount in raw units of DepositMint
//     (MintJupUSD or MintUSDC; the API defaults to USDC when empty). Jupiter's
//     minimum order is 5 USD.
//   - Sell: PositionPubkey and exactly one quantity: ContractsMicro (1000000 = 1
//     contract), ContractsDecimal, or the legacy whole-number Contracts.
//
// The built transaction only opens an order; Jupiter's keepers fill it later.
type OrderRequest struct {
	Owner            string
	MarketID         string
	PositionPubkey   string
	IsYes            bool
	IsBuy            bool
	Contracts        string
	ContractsMicro   string
	ContractsDecimal string
	DepositAmount    string
	DepositMint      string
}

type orderBody struct {
	OwnerPubkey      string `json:"ownerPubkey"`
	MarketID         string `json:"marketId,omitempty"`
	PositionPubkey   string `json:"positionPubkey,omitempty"`
	IsYes            bool   `json:"isYes"`
	IsBuy            bool   `json:"isBuy"`
	Contracts        string `json:"contracts,omitempty"`
	ContractsMicro   string `json:"contractsMicro,omitempty"`
	ContractsDecimal string `json:"contractsDecimal,omitempty"`
	DepositAmount    string `json:"depositAmount,omitempty"`
	DepositMint      string `json:"depositMint,omitempty"`
}

func (r OrderRequest) body() (orderBody, error) {
	if err := validatePubkey("owner", r.Owner); err != nil {
		return orderBody{}, err
	}

	body := orderBody{
		OwnerPubkey: r.Owner,
		MarketID:    r.MarketID,
		IsYes:       r.IsYes,
		IsBuy:       r.IsBuy,
	}

	quantitySet := r.Contracts != "" || r.ContractsMicro != "" || r.ContractsDecimal != ""
	if r.IsBuy {
		if r.MarketID == "" {
			return orderBody{}, fmt.Errorf("buy orders require a market ID")
		}

		if quantitySet || r.PositionPubkey != "" {
			return orderBody{}, fmt.Errorf("buy orders take a deposit, not a position or contract quantity")
		}

		deposit, err := amount.ParseRaw(r.DepositAmount)
		if err != nil {
			return orderBody{}, fmt.Errorf("deposit %w", err)
		}

		if r.DepositMint != "" {
			if err := validatePubkey("deposit mint", r.DepositMint); err != nil {
				return orderBody{}, err
			}
		}

		body.DepositAmount = deposit
		body.DepositMint = r.DepositMint

		return body, nil
	}

	if err := validatePubkey("position", r.PositionPubkey); err != nil {
		return orderBody{}, err
	}

	if r.DepositAmount != "" || r.DepositMint != "" {
		return orderBody{}, fmt.Errorf("sell orders take a contract quantity, not a deposit")
	}

	quantity, err := r.sellQuantity()
	if err != nil {
		return orderBody{}, err
	}

	body.PositionPubkey = r.PositionPubkey
	body.Contracts = quantity.whole
	body.ContractsMicro = quantity.micro
	body.ContractsDecimal = quantity.decimal

	return body, nil
}

// contractQuantity holds the one sell quantity form a request uses.
type contractQuantity struct {
	whole   string
	micro   string
	decimal string
}

// sellQuantity requires exactly one quantity field and returns it canonicalized.
func (r OrderRequest) sellQuantity() (contractQuantity, error) {
	set := 0
	for _, value := range []string{r.Contracts, r.ContractsMicro, r.ContractsDecimal} {
		if value != "" {
			set++
		}
	}

	if set != 1 {
		return contractQuantity{}, fmt.Errorf("sell orders require exactly one contract quantity")
	}

	var quantity contractQuantity
	var err error
	switch {
	case r.Contracts != "":
		quantity.whole, err = amount.ParseRaw(r.Contracts)
	case r.ContractsMicro != "":
		quantity.micro, err = amount.ParseRaw(r.ContractsMicro)
	default:
		quantity.decimal, err = parsePositiveDecimal(r.ContractsDecimal)
	}

	if err != nil {
		return contractQuantity{}, fmt.Errorf("contracts: %w", err)
	}

	return quantity, nil
}

// parsePositiveDecimal accepts a positive decimal JSON number such as 5.777773.
// json.Valid rejects fractions, hexadecimal, and special values big.Rat accepts.
func parsePositiveDecimal(value string) (string, error) {
	if !json.Valid([]byte(value)) {
		return "", fmt.Errorf("%q is not a decimal number", value)
	}

	parsed, ok := new(big.Rat).SetString(value)
	if !ok || parsed.Sign() <= 0 {
		return "", fmt.Errorf("%q must be a positive decimal number", value)
	}

	return value, nil
}

// CreateOrder builds an unsigned buy or sell order transaction. Building submits
// nothing and returns fresh data, so transient failures are retried.
func (c *Client) CreateOrder(ctx context.Context, r OrderRequest) (*TransactionBuild, error) {
	body, err := r.body()
	if err != nil {
		return nil, err
	}

	return c.build(ctx, http.MethodPost, "/orders", body)
}

type ownerBody struct {
	OwnerPubkey string `json:"ownerPubkey"`
}

// ClosePosition builds a transaction selling every contract of a position before
// the market resolves.
func (c *Client) ClosePosition(
	ctx context.Context,
	owner string,
	positionPubkey string,
) (*TransactionBuild, error) {
	if err := validatePubkey("owner", owner); err != nil {
		return nil, err
	}

	if err := validatePubkey("position", positionPubkey); err != nil {
		return nil, err
	}

	return c.build(ctx, http.MethodDelete, "/positions/"+positionPubkey, ownerBody{OwnerPubkey: owner})
}

type closeAllBody struct {
	OwnerPubkey             string `json:"ownerPubkey"`
	MinSellPriceSlippageBps uint16 `json:"minSellPriceSlippageBps"`
}

type TransactionBuilds struct {
	Builds   []TransactionBuild `json:"data"`
	Metadata response.Metadata  `json:"-"`
}

// CloseAllPositions builds one sell transaction per open position of owner.
// minSellPriceSlippageBps bounds the accepted sell price for each.
func (c *Client) CloseAllPositions(
	ctx context.Context,
	owner string,
	minSellPriceSlippageBps uint16,
) (*TransactionBuilds, error) {
	if err := validatePubkey("owner", owner); err != nil {
		return nil, err
	}

	if minSellPriceSlippageBps > BasisPointsPerWhole {
		return nil, fmt.Errorf("sell slippage exceeds %d bps", BasisPointsPerWhole)
	}

	body := closeAllBody{OwnerPubkey: owner, MinSellPriceSlippageBps: minSellPriceSlippageBps}
	request := buildRequest(http.MethodDelete, "/positions", body)

	var result TransactionBuilds
	metadata, err := c.transport.Do(ctx, request, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

// ClaimPosition builds a transaction paying out a winning position ($1 per
// contract, no claim fee) after resolution. Check Position.Claimable first; Jupiter
// Forecast rounds settle automatically and need no claim.
func (c *Client) ClaimPosition(
	ctx context.Context,
	owner string,
	positionPubkey string,
) (*TransactionBuild, error) {
	if err := validatePubkey("owner", owner); err != nil {
		return nil, err
	}

	if err := validatePubkey("position", positionPubkey); err != nil {
		return nil, err
	}

	path := "/positions/" + positionPubkey + "/claim"

	return c.build(ctx, http.MethodPost, path, ownerBody{OwnerPubkey: owner})
}

func buildRequest(method string, path string, body any) transport.Request {
	return transport.Request{
		Method: method,
		Path:   basePath + path,
		Body:   body,
		Retry:  transport.RepeatableRetry,
	}
}

func (c *Client) build(ctx context.Context, method string, path string, body any) (*TransactionBuild, error) {
	var result TransactionBuild
	metadata, err := c.transport.Do(ctx, buildRequest(method, path, body), &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

// ExecuteRequest submits a signed build. Pass the build's Execution.Context
// unchanged; RequestID is an optional correlation ID echoed back.
type ExecuteRequest struct {
	SignedTransaction string          `json:"signedTransaction"`
	Context           json.RawMessage `json:"context,omitempty"`
	RequestID         string          `json:"requestId,omitempty"`
}

type ExecuteResult struct {
	Status    string `json:"status"`
	Signature string `json:"signature"`
	Error     string `json:"error"`
	RequestID string `json:"requestId"`

	Metadata response.Metadata `json:"-"`
}

// Execute submits a signed order, sell, or Forecast swap transaction. It makes one
// attempt because no replay guarantee is documented. A response without an
// explicit Success status returns an *ExecutionError; failures after which the
// transaction may have landed wrap ErrOutcomeUnknown. For keeper-filled orders,
// success means the order was opened, not filled: poll OrderStatus.
func (c *Client) Execute(ctx context.Context, r ExecuteRequest) (*ExecuteResult, error) {
	if r.SignedTransaction == "" {
		return nil, fmt.Errorf("signed transaction required")
	}

	if _, err := base64.StdEncoding.DecodeString(r.SignedTransaction); err != nil {
		return nil, fmt.Errorf("signed transaction: %w", err)
	}

	if len(r.Context) > 0 && !json.Valid(r.Context) {
		return nil, fmt.Errorf("execution context is not valid JSON")
	}

	// An ended context guarantees nothing is sent, so the outcome is known.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	request := transport.Request{
		Method: http.MethodPost,
		Path:   basePath + "/execute",
		Body:   r,
		Retry:  transport.NoRetry,
	}

	var result ExecuteResult
	metadata, err := c.transport.Do(ctx, request, &result)
	result.Metadata = metadata
	if err != nil {
		if transport.MayHaveBeenProcessed(err) {
			return &result, fmt.Errorf("%w: %w", ErrOutcomeUnknown, err)
		}

		return &result, err
	}

	if result.Status != ExecutionStatusSuccess {
		message := result.Error
		if message == "" {
			message = "execution did not confirm success"
		}

		return &result, &ExecutionError{
			Status:    result.Status,
			Message:   message,
			Signature: result.Signature,
			RequestID: result.RequestID,
		}
	}

	return &result, nil
}
