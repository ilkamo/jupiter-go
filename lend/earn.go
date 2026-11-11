package lend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ilkamo/jupiter-go/internal/address"
	"github.com/ilkamo/jupiter-go/internal/amount"
	"github.com/ilkamo/jupiter-go/internal/listparam"
	"github.com/ilkamo/jupiter-go/internal/transport"
	"github.com/ilkamo/jupiter-go/response"
)

// EarnClient deposits into and withdraws from Jupiter Earn vaults. Deposit and
// Withdraw move an amount of the underlying asset; Mint and Redeem move a number
// of vault shares (jlTokens).
type EarnClient struct {
	transport *transport.Client
}

// AmountRequest deposits or withdraws Amount raw units of Asset for Signer.
type AmountRequest struct {
	Asset  string `json:"asset"`
	Signer string `json:"signer"`
	Amount string `json:"amount"`
}

func (r AmountRequest) normalized() (AmountRequest, error) {
	if err := validateParties(r.Asset, r.Signer); err != nil {
		return AmountRequest{}, err
	}

	rawAmount, err := amount.ParseRaw(r.Amount)
	if err != nil {
		return AmountRequest{}, err
	}

	r.Amount = rawAmount

	return r, nil
}

// SharesRequest mints or redeems Shares raw vault-share units of Asset for Signer.
type SharesRequest struct {
	Asset  string `json:"asset"`
	Signer string `json:"signer"`
	Shares string `json:"shares"`
}

func (r SharesRequest) normalized() (SharesRequest, error) {
	if err := validateParties(r.Asset, r.Signer); err != nil {
		return SharesRequest{}, err
	}

	shares, err := amount.ParseRaw(r.Shares)
	if err != nil {
		return SharesRequest{}, fmt.Errorf("shares: %w", err)
	}

	r.Shares = shares

	return r, nil
}

func validateParties(asset string, signer string) error {
	if err := address.Validate(asset); err != nil {
		return fmt.Errorf("asset: %w", err)
	}

	if err := address.Validate(signer); err != nil {
		return fmt.Errorf("signer: %w", err)
	}

	return nil
}

// Deposit builds an unsigned deposit transaction.
func (c *EarnClient) Deposit(ctx context.Context, r AmountRequest) (*Transaction, error) {
	body, err := r.normalized()
	if err != nil {
		return nil, err
	}

	return c.transaction(ctx, "/earn/deposit", body)
}

// Withdraw builds an unsigned withdrawal transaction.
func (c *EarnClient) Withdraw(ctx context.Context, r AmountRequest) (*Transaction, error) {
	body, err := r.normalized()
	if err != nil {
		return nil, err
	}

	return c.transaction(ctx, "/earn/withdraw", body)
}

// Mint builds an unsigned transaction that deposits enough assets to mint the shares.
func (c *EarnClient) Mint(ctx context.Context, r SharesRequest) (*Transaction, error) {
	body, err := r.normalized()
	if err != nil {
		return nil, err
	}

	return c.transaction(ctx, "/earn/mint", body)
}

// Redeem builds an unsigned transaction that burns the shares for assets.
func (c *EarnClient) Redeem(ctx context.Context, r SharesRequest) (*Transaction, error) {
	body, err := r.normalized()
	if err != nil {
		return nil, err
	}

	return c.transaction(ctx, "/earn/redeem", body)
}

// InstructionResult is a single instruction to compose into your own transaction.
type InstructionResult struct {
	Instruction Instruction
	Metadata    response.Metadata
}

func (c *EarnClient) DepositInstructions(ctx context.Context, r AmountRequest) (*InstructionResult, error) {
	body, err := r.normalized()
	if err != nil {
		return nil, err
	}

	return c.instruction(ctx, "/earn/deposit-instructions", body)
}

func (c *EarnClient) WithdrawInstructions(ctx context.Context, r AmountRequest) (*InstructionResult, error) {
	body, err := r.normalized()
	if err != nil {
		return nil, err
	}

	return c.instruction(ctx, "/earn/withdraw-instructions", body)
}

func (c *EarnClient) MintInstructions(ctx context.Context, r SharesRequest) (*InstructionResult, error) {
	body, err := r.normalized()
	if err != nil {
		return nil, err
	}

	return c.instruction(ctx, "/earn/mint-instructions", body)
}

func (c *EarnClient) RedeemInstructions(ctx context.Context, r SharesRequest) (*InstructionResult, error) {
	body, err := r.normalized()
	if err != nil {
		return nil, err
	}

	return c.instruction(ctx, "/earn/redeem-instructions", body)
}

// buildRequest describes a POST that builds an unsigned transaction or
// instruction. Building submits nothing and each call returns fresh data, so
// transient failures are retried.
func buildRequest(path string, body any, query url.Values) transport.Request {
	return transport.Request{
		Method: http.MethodPost,
		Path:   basePath + path,
		Query:  query,
		Body:   body,
		Retry:  transport.RepeatableRetry,
	}
}

func (c *EarnClient) transaction(ctx context.Context, path string, body any) (*Transaction, error) {
	var result Transaction
	metadata, err := c.transport.Do(ctx, buildRequest(path, body, nil), &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

func (c *EarnClient) instruction(ctx context.Context, path string, body any) (*InstructionResult, error) {
	var result InstructionResult
	metadata, err := c.transport.Do(ctx, buildRequest(path, body, nil), &result.Instruction)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

// Numbers are json.Number: they keep upstream precision and accept both the JSON
// number and numeric string forms, which the Lend schemas and examples mix.

// EarnToken is a vault's share token (jlToken) and its underlying asset.
type EarnToken struct {
	ID                  int                 `json:"id"`
	Address             string              `json:"address"`
	Name                string              `json:"name"`
	Symbol              string              `json:"symbol"`
	Decimals            int                 `json:"decimals"`
	AssetAddress        string              `json:"assetAddress"`
	Asset               EarnAsset           `json:"asset"`
	TotalAssets         json.Number         `json:"totalAssets"`
	TotalSupply         json.Number         `json:"totalSupply"`
	ConvertToShares     json.Number         `json:"convertToShares"`
	ConvertToAssets     json.Number         `json:"convertToAssets"`
	RewardsRate         json.Number         `json:"rewardsRate"`
	SupplyRate          json.Number         `json:"supplyRate"`
	TotalRate           json.Number         `json:"totalRate"`
	RebalanceDifference json.Number         `json:"rebalanceDifference"`
	LiquiditySupplyData LiquiditySupplyData `json:"liquiditySupplyData"`
}

type EarnAsset struct {
	Address     string      `json:"address"`
	ChainID     ChainID     `json:"chain_id"`
	Name        string      `json:"name"`
	Symbol      string      `json:"symbol"`
	Decimals    int         `json:"decimals"`
	LogoURL     string      `json:"logo_url"`
	Price       json.Number `json:"price"`
	CoingeckoID string      `json:"coingecko_id"`
}

type LiquiditySupplyData struct {
	ModeWithInterest       bool        `json:"modeWithInterest"`
	Supply                 json.Number `json:"supply"`
	WithdrawalLimit        json.Number `json:"withdrawalLimit"`
	LastUpdateTimestamp    json.Number `json:"lastUpdateTimestamp"`
	ExpandPercent          json.Number `json:"expandPercent"`
	ExpandDuration         json.Number `json:"expandDuration"`
	BaseWithdrawalLimit    json.Number `json:"baseWithdrawalLimit"`
	WithdrawableUntilLimit json.Number `json:"withdrawableUntilLimit"`
	Withdrawable           json.Number `json:"withdrawable"`
}

type EarnTokens struct {
	Tokens   []EarnToken
	Metadata response.Metadata
}

// Tokens lists the assets that can be deposited.
func (c *EarnClient) Tokens(ctx context.Context) (*EarnTokens, error) {
	var result EarnTokens
	metadata, err := c.transport.Do(ctx, readRequest("/earn/tokens", nil), &result.Tokens)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

type EarnPosition struct {
	Token             EarnToken   `json:"token"`
	OwnerAddress      string      `json:"ownerAddress"`
	Shares            json.Number `json:"shares"`
	UnderlyingAssets  json.Number `json:"underlyingAssets"`
	UnderlyingBalance json.Number `json:"underlyingBalance"`
	Allowance         json.Number `json:"allowance"`
}

type EarnPositions struct {
	Positions []EarnPosition
	Metadata  response.Metadata
}

// Positions returns Earn positions for one or more wallets.
func (c *EarnClient) Positions(ctx context.Context, users []string) (*EarnPositions, error) {
	if err := validateAddresses("user", users); err != nil {
		return nil, err
	}

	query := url.Values{"users": {listparam.Join(users)}}

	var result EarnPositions
	metadata, err := c.transport.Do(ctx, readRequest("/earn/positions", query), &result.Positions)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

type Earning struct {
	Address      string      `json:"address"`
	OwnerAddress string      `json:"ownerAddress"`
	Earnings     json.Number `json:"earnings"`
	Slot         json.Number `json:"slot"`
}

// earningList accepts the single object the schema declares and the array the
// documented example returns.
type earningList []Earning

func (l *earningList) UnmarshalJSON(data []byte) error {
	var list []Earning
	if err := json.Unmarshal(data, &list); err == nil {
		*l = list

		return nil
	}

	var single Earning
	if err := json.Unmarshal(data, &single); err != nil {
		return fmt.Errorf("earnings must be an object or array: %w", err)
	}

	*l = earningList{single}

	return nil
}

type Earnings struct {
	Earnings []Earning
	Metadata response.Metadata
}

// Earnings returns a wallet's earnings for the given position (jlToken) addresses.
func (c *EarnClient) Earnings(ctx context.Context, user string, positions []string) (*Earnings, error) {
	if err := address.Validate(user); err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	if err := validateAddresses("position", positions); err != nil {
		return nil, err
	}

	query := url.Values{
		"user":      {user},
		"positions": {listparam.Join(positions)},
	}

	var list earningList
	metadata, err := c.transport.Do(ctx, readRequest("/earn/earnings", query), &list)
	result := &Earnings{Earnings: list, Metadata: metadata}
	if err != nil {
		return result, err
	}

	return result, nil
}

func readRequest(path string, query url.Values) transport.Request {
	return transport.Request{
		Method: http.MethodGet,
		Path:   basePath + path,
		Query:  query,
		Retry:  transport.ReadOnlyRetry,
	}
}
