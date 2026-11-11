package lend

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/url"
	"regexp"

	"github.com/ilkamo/jupiter-go/internal/address"
	"github.com/ilkamo/jupiter-go/internal/listparam"
	"github.com/ilkamo/jupiter-go/internal/transport"
	"github.com/ilkamo/jupiter-go/response"
)

const (
	MarketMain   = "main"
	MarketEthena = "ethena"

	// NewPosition as OperateRequest.PositionID creates a position.
	NewPosition = 0

	// WithdrawAllCollateral as ColAmount withdraws all collateral, and RepayAllDebt
	// as DebtAmount repays all debt including interest dust. Both are the API's
	// MIN_I128 sentinel. Repaying the exact borrow amount can leave dust below the
	// vault's minimum and fail.
	WithdrawAllCollateral = "-170141183460469231731687303715884105728"
	RepayAllDebt          = WithdrawAllCollateral

	signedAmountBits = 128
	decimalBase      = 10
)

var (
	// signedAmountPattern is the documented OperatePayload amount pattern.
	signedAmountPattern = regexp.MustCompile(`^-?[0-9]+$`)

	minSignedAmount = new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), signedAmountBits-1))
	maxSignedAmount = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), signedAmountBits-1), big.NewInt(1))
)

// BorrowClient reads borrow vaults and positions and builds operate transactions.
// Vault and position IDs are scoped to a market; use the same Market for reads and
// operations. An empty Market selects MarketMain. Other values are passed through so
// markets added upstream remain usable.
type BorrowClient struct {
	transport *transport.Client
}

func marketQuery(market string) (url.Values, error) {
	query := url.Values{}
	if market == "" {
		return query, nil
	}

	if err := listparam.ValidateItem("market", market); err != nil {
		return nil, err
	}

	query.Set("market", market)

	return query, nil
}

// VaultsRequest selects a market. RPCURL optionally asks Jupiter to resolve on-chain
// state through your RPC node; the URL, including any credentials in it, is sent
// to Jupiter.
type VaultsRequest struct {
	Market string
	RPCURL string
}

type BorrowVaults struct {
	Vaults   []BorrowVault
	Metadata response.Metadata
}

func (c *BorrowClient) Vaults(ctx context.Context, r VaultsRequest) (*BorrowVaults, error) {
	query, err := marketQuery(r.Market)
	if err != nil {
		return nil, err
	}

	if r.RPCURL != "" {
		parsed, err := url.Parse(r.RPCURL)
		if err != nil {
			return nil, fmt.Errorf("parse RPC URL: %w", err)
		}

		if !parsed.IsAbs() || parsed.Host == "" {
			return nil, fmt.Errorf("RPC URL must be absolute")
		}

		query.Set("rpcUrl", r.RPCURL)
	}

	var result BorrowVaults
	metadata, err := c.transport.Do(ctx, readRequest("/borrow/vaults", query), &result.Vaults)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

type BorrowPositions struct {
	Positions []BorrowPosition
	Metadata  response.Metadata
}

// Positions returns borrow positions (NFTs) for one or more wallets in a market.
func (c *BorrowClient) Positions(
	ctx context.Context,
	market string,
	users []string,
) (*BorrowPositions, error) {
	query, err := marketQuery(market)
	if err != nil {
		return nil, err
	}

	if err := validateAddresses("user", users); err != nil {
		return nil, err
	}

	query.Set("users", listparam.Join(users))

	var result BorrowPositions
	metadata, err := c.transport.Do(ctx, readRequest("/borrow/positions", query), &result.Positions)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

// OperateRequest changes a position by signed deltas in base units:
//
//   - ColAmount > 0 deposits collateral, < 0 withdraws it.
//   - DebtAmount > 0 borrows (at least the vault's minimum), < 0 repays.
//   - PositionID NewPosition creates a position, typically with a deposit.
//
// Use WithdrawAllCollateral and RepayAllDebt to close a position fully.
type OperateRequest struct {
	Market        string `json:"-"`
	VaultID       uint32 `json:"vaultId"`
	PositionID    uint32 `json:"positionId"`
	PositionOwner string `json:"positionOwner,omitempty"`
	Signer        string `json:"signer"`
	ColAmount     string `json:"colAmount"`
	DebtAmount    string `json:"debtAmount"`
}

func (r OperateRequest) normalized() (OperateRequest, url.Values, error) {
	query, err := marketQuery(r.Market)
	if err != nil {
		return OperateRequest{}, nil, err
	}

	if err := address.Validate(r.Signer); err != nil {
		return OperateRequest{}, nil, fmt.Errorf("signer: %w", err)
	}

	if r.PositionOwner != "" {
		if err := address.Validate(r.PositionOwner); err != nil {
			return OperateRequest{}, nil, fmt.Errorf("position owner: %w", err)
		}
	}

	collateral, err := parseSignedAmount("collateral amount", r.ColAmount)
	if err != nil {
		return OperateRequest{}, nil, err
	}

	debt, err := parseSignedAmount("debt amount", r.DebtAmount)
	if err != nil {
		return OperateRequest{}, nil, err
	}

	if collateral.Sign() == 0 && debt.Sign() == 0 {
		return OperateRequest{}, nil, fmt.Errorf("operate requires a non-zero collateral or debt amount")
	}

	r.ColAmount = collateral.String()
	r.DebtAmount = debt.String()

	return r, query, nil
}

// parseSignedAmount accepts a base-10 integer in the signed 128-bit range, the
// documented operate amount format.
func parseSignedAmount(field string, value string) (*big.Int, error) {
	if !signedAmountPattern.MatchString(value) {
		return nil, fmt.Errorf("%s must be a base-10 integer", field)
	}

	parsed, ok := new(big.Int).SetString(value, decimalBase)
	if !ok {
		return nil, fmt.Errorf("%s must be a base-10 integer", field)
	}

	if parsed.Cmp(minSignedAmount) < 0 || parsed.Cmp(maxSignedAmount) > 0 {
		return nil, fmt.Errorf("%s is outside the signed %d-bit range", field, signedAmountBits)
	}

	return parsed, nil
}

// OperateTransaction is an unsigned transaction for the operation. NFTID is the
// affected position, newly assigned when the request used NewPosition.
type OperateTransaction struct {
	NFTID       int    `json:"nftId"`
	Transaction string `json:"transaction"`

	Metadata response.Metadata `json:"-"`
}

// Operate builds an unsigned operate transaction. Building submits nothing, so
// transient failures are retried.
func (c *BorrowClient) Operate(ctx context.Context, r OperateRequest) (*OperateTransaction, error) {
	body, query, err := r.normalized()
	if err != nil {
		return nil, err
	}

	var result OperateTransaction
	metadata, err := c.transport.Do(ctx, buildRequest("/borrow/operate", body, query), &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

// OperateInstructions holds instructions for a version 0 transaction. Only the
// lookup table addresses are returned; fetch their contents from an RPC node before
// compiling the message.
type OperateInstructions struct {
	NFTID                       int           `json:"nftId"`
	Instructions                []Instruction `json:"instructions"`
	AddressLookupTableAddresses []string      `json:"addressLookupTableAddresses"`

	Metadata response.Metadata `json:"-"`
}

func (c *BorrowClient) OperateInstructions(
	ctx context.Context,
	r OperateRequest,
) (*OperateInstructions, error) {
	body, query, err := r.normalized()
	if err != nil {
		return nil, err
	}

	var result OperateInstructions
	request := buildRequest("/borrow/operate-instructions", body, query)
	metadata, err := c.transport.Do(ctx, request, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

// Numeric vault and position fields are json.Number; ratios such as
// CollateralFactor use a 1e3 scale (800 means 80%).

type BorrowVault struct {
	ID                     int                 `json:"id"`
	Address                string              `json:"address"`
	SupplyToken            BorrowToken         `json:"supplyToken"`
	BorrowToken            BorrowToken         `json:"borrowToken"`
	TotalSupply            json.Number         `json:"totalSupply"`
	TotalSupplyLiquidity   json.Number         `json:"totalSupplyLiquidity"`
	TotalBorrow            json.Number         `json:"totalBorrow"`
	TotalBorrowLiquidity   json.Number         `json:"totalBorrowLiquidity"`
	AbsorbedSupply         json.Number         `json:"absorbedSupply"`
	AbsorbedBorrow         json.Number         `json:"absorbedBorrow"`
	SupplyRateMagnifier    json.Number         `json:"supplyRateMagnifier"`
	BorrowRateMagnifier    json.Number         `json:"borrowRateMagnifier"`
	BorrowFee              json.Number         `json:"borrowFee"`
	CollateralFactor       json.Number         `json:"collateralFactor"`
	LiquidationThreshold   json.Number         `json:"liquidationThreshold"`
	LiquidationMaxLimit    json.Number         `json:"liquidationMaxLimit"`
	LiquidationPenalty     json.Number         `json:"liquidationPenalty"`
	WithdrawalGap          json.Number         `json:"withdrawalGap"`
	SupplyRate             json.Number         `json:"supplyRate"`
	SupplyRateLiquidity    json.Number         `json:"supplyRateLiquidity"`
	BorrowRate             json.Number         `json:"borrowRate"`
	BorrowRateLiquidity    json.Number         `json:"borrowRateLiquidity"`
	WithdrawLimit          json.Number         `json:"withdrawLimit"`
	WithdrawableUntilLimit json.Number         `json:"withdrawableUntilLimit"`
	Withdrawable           json.Number         `json:"withdrawable"`
	BorrowLimit            json.Number         `json:"borrowLimit"`
	BorrowableUntilLimit   json.Number         `json:"borrowableUntilLimit"`
	Borrowable             json.Number         `json:"borrowable"`
	BorrowLimitUtilization json.Number         `json:"borrowLimitUtilization"`
	MinimumBorrowing       json.Number         `json:"minimumBorrowing"`
	TotalPositions         int                 `json:"totalPositions"`
	Oracle                 string              `json:"oracle"`
	OraclePrice            json.Number         `json:"oraclePrice"`
	OraclePriceOperate     json.Number         `json:"oraclePriceOperate"`
	OraclePriceLiquidate   json.Number         `json:"oraclePriceLiquidate"`
	OracleSources          []OracleSource      `json:"oracleSources"`
	OracleTimestamp        int64               `json:"oracleTimestamp"`
	LiquiditySupplyData    LiquiditySupplyData `json:"liquiditySupplyData"`
	LiquidityBorrowData    LiquidityBorrowData `json:"liquidityBorrowData"`
	Rewards                []MarketReward      `json:"rewards"`
	TopTick                int64               `json:"topTick"`
	Metadata               VaultMetadata       `json:"metadata"`
}

type BorrowToken struct {
	Address                string          `json:"address"`
	ChainID                string          `json:"chainId"`
	Name                   string          `json:"name"`
	Symbol                 string          `json:"symbol"`
	UISymbol               string          `json:"uiSymbol"`
	Decimals               int             `json:"decimals"`
	LogoURL                string          `json:"logoUrl"`
	Price                  json.Number     `json:"price"`
	CoingeckoID            string          `json:"coingeckoId"`
	UpdatedAt              string          `json:"updatedAt"`
	StakingAPR             json.Number     `json:"stakingApr"`
	NativeStakeVoteAccount string          `json:"nativeStakeVoteAccount"`
	ScaledUIConfig         *ScaledUIConfig `json:"scaledUiConfig"`
}

// ScaledUIConfig is the Token-2022 scaled UI amount configuration, present on
// tokens such as tokenised stocks.
type ScaledUIConfig struct {
	Authority                       string      `json:"authority"`
	Multiplier                      json.Number `json:"multiplier"`
	NewMultiplier                   json.Number `json:"newMultiplier"`
	NewMultiplierEffectiveTimestamp string      `json:"newMultiplierEffectiveTimestamp"`
}

type OracleSource struct {
	Source     string          `json:"source"`
	Invert     bool            `json:"invert"`
	Multiplier json.Number     `json:"multiplier"`
	Divisor    json.Number     `json:"divisor"`
	SourceType json.RawMessage `json:"sourceType"`
}

type LiquidityBorrowData struct {
	ModeWithInterest       bool        `json:"modeWithInterest"`
	Borrow                 json.Number `json:"borrow"`
	BorrowLimit            json.Number `json:"borrowLimit"`
	LastUpdateTimestamp    json.Number `json:"lastUpdateTimestamp"`
	ExpandPercent          json.Number `json:"expandPercent"`
	ExpandDuration         json.Number `json:"expandDuration"`
	BaseBorrowLimit        json.Number `json:"baseBorrowLimit"`
	MaxBorrowLimit         json.Number `json:"maxBorrowLimit"`
	BorrowLimitUtilization json.Number `json:"borrowLimitUtilization"`
	BorrowableUntilLimit   json.Number `json:"borrowableUntilLimit"`
	Borrowable             json.Number `json:"borrowable"`
}

type MarketReward struct {
	RewardToken BorrowToken     `json:"rewardToken"`
	Type        string          `json:"type"`
	Side        string          `json:"side"`
	APR         json.Number     `json:"apr"`
	Metadata    json.RawMessage `json:"metadata"`
}

type VaultMetadata struct {
	MaxLeverage bool             `json:"maxLeverage"`
	NativeStake bool             `json:"nativeStake"`
	Pegged      bool             `json:"pegged"`
	Multiply    MultiplySettings `json:"multiply"`
}

type MultiplySettings struct {
	Enabled      bool `json:"enabled"`
	ReduceFactor int  `json:"reduceFactor"`
}

type BorrowPosition struct {
	ID                int               `json:"id"`
	VaultID           int               `json:"vaultId"`
	Address           string            `json:"address"`
	Supply            json.Number       `json:"supply"`
	BeforeSupply      json.Number       `json:"beforeSupply"`
	Borrow            json.Number       `json:"borrow"`
	BeforeBorrow      json.Number       `json:"beforeBorrow"`
	IsLiquidated      bool              `json:"isLiquidated"`
	SupplyLiquidation json.Number       `json:"supplyLiquidation"`
	BorrowLiquidation json.Number       `json:"borrowLiquidation"`
	IsSupplyPosition  bool              `json:"isSupplyPosition"`
	Tick              int64             `json:"tick"`
	TickID            int64             `json:"tickId"`
	DustBorrow        json.Number       `json:"dustBorrow"`
	OwnerAddress      string            `json:"ownerAddress"`
	Vault             BorrowVault       `json:"vault"`
	Config            map[string]string `json:"config"`
}
