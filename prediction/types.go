package prediction

import (
	"encoding/json"
	"fmt"

	"github.com/ilkamo/jupiter-go/response"
)

// USD amounts are micro USD (1000000 = $1) and contract quantities in *Micro
// fields are micro contracts (1000000 = 1 contract). Fields the API declares as
// strings stay strings to keep exact values; declared numbers are json.Number.
// Timestamps are Unix seconds unless noted.

// StringOrNumber holds a value the API sends as either a JSON string or number,
// such as Market.ResolveAt. Numbers keep their text; null leaves it empty.
type StringOrNumber string

// UnmarshalJSON implements json.Unmarshaler for string, number, and null.
func (v *StringOrNumber) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*v = StringOrNumber(text)

		return nil
	}

	var number json.Number
	if err := json.Unmarshal(data, &number); err != nil {
		return fmt.Errorf("value must be a string or number: %w", err)
	}

	*v = StringOrNumber(number.String())

	return nil
}

type Pagination struct {
	Start   int  `json:"start"`
	End     int  `json:"end"`
	Total   int  `json:"total"`
	HasNext bool `json:"hasNext"`
}

type EventMetadata struct {
	EventID   string `json:"eventId"`
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle"`
	Slug      string `json:"slug"`
	Series    string `json:"series"`
	CloseTime string `json:"closeTime"`
	ImageURL  string `json:"imageUrl"`
	IsLive    bool   `json:"isLive"`
}

type MarketOption struct {
	Label  string `json:"label"`
	BuyYes bool   `json:"buyYes"`
}

type MarketMetadata struct {
	MarketID       string         `json:"marketId"`
	EventID        string         `json:"eventId"`
	Title          string         `json:"title"`
	Subtitle       string         `json:"subtitle"`
	Description    string         `json:"description"`
	Provider       string         `json:"provider"`
	Status         string         `json:"status"`
	Result         string         `json:"result"`
	CloseTime      json.Number    `json:"closeTime"`
	OpenTime       json.Number    `json:"openTime"`
	IsTeamMarket   bool           `json:"isTeamMarket"`
	MarketOptions  []MarketOption `json:"marketOptions"`
	RulesPrimary   string         `json:"rulesPrimary"`
	RulesSecondary string         `json:"rulesSecondary"`
}

// MarketPricing prices are micro USD per contract and imply the outcome's
// probability; an empty value means the API reported null.
type MarketPricing struct {
	BuyYesPriceUSD  json.Number `json:"buyYesPriceUsd"`
	BuyNoPriceUSD   json.Number `json:"buyNoPriceUsd"`
	SellYesPriceUSD json.Number `json:"sellYesPriceUsd"`
	SellNoPriceUSD  json.Number `json:"sellNoPriceUsd"`
	Volume          json.Number `json:"volume"`
}

// Market is one binary YES/NO market. The Outcome*, MarketPDA, OutcomeSide,
// Tradable, and LifecycleStatus fields apply to Jupiter Forecast markets.
type Market struct {
	MarketID            string          `json:"marketId"`
	EventID             string          `json:"eventId"`
	Provider            string          `json:"provider"`
	Title               string          `json:"title"`
	Status              string          `json:"status"`
	Result              string          `json:"result"`
	OpenTime            json.Number     `json:"openTime"`
	CloseTime           json.Number     `json:"closeTime"`
	ResolveAt           StringOrNumber  `json:"resolveAt"`
	MarketResultPubkey  string          `json:"marketResultPubkey"`
	ImageURL            string          `json:"imageUrl"`
	RulesPrimary        string          `json:"rulesPrimary"`
	RulesSecondary      string          `json:"rulesSecondary"`
	Outcomes            []string        `json:"outcomes"`
	MarketOptions       []MarketOption  `json:"marketOptions"`
	ClobTokenIDs        []string        `json:"clobTokenIds"`
	IsTeamMarket        bool            `json:"isTeamMarket"`
	Team                json.RawMessage `json:"team"`
	SportsLine          json.Number     `json:"sportsLine"`
	SportsMarketType    string          `json:"sportsMarketType"`
	OutcomeMint         string          `json:"outcomeMint"`
	OutcomeTokenProgram string          `json:"outcomeTokenProgram"`
	MarketPDA           string          `json:"marketPda"`
	OutcomeSide         string          `json:"outcomeSide"`
	Tradable            bool            `json:"tradable"`
	LifecycleStatus     string          `json:"lifecycleStatus"`
	Pricing             *MarketPricing  `json:"pricing"`

	Metadata response.Metadata `json:"-"`
}

type Event struct {
	EventID        string          `json:"eventId"`
	IsActive       bool            `json:"isActive"`
	IsLive         bool            `json:"isLive"`
	Category       string          `json:"category"`
	Subcategory    string          `json:"subcategory"`
	Tags           []string        `json:"tags"`
	EventMetadata  EventMetadata   `json:"metadata"`
	Markets        []Market        `json:"markets"`
	VolumeUSD      string          `json:"volumeUsd"`
	Volume24h      string          `json:"volume24hr"`
	LiveScore      json.RawMessage `json:"liveScore"`
	CloseCondition string          `json:"closeCondition"`
	BeginAt        string          `json:"beginAt"`
	RulesPDF       string          `json:"rulesPdf"`

	Metadata response.Metadata `json:"-"`
}

type GameScore struct {
	EventID            string `json:"eventId"`
	GameID             string `json:"gameId"`
	LeagueAbbreviation string `json:"leagueAbbreviation"`
	HomeTeam           string `json:"homeTeam"`
	AwayTeam           string `json:"awayTeam"`
	Score              string `json:"score"`
	Period             string `json:"period"`
	Elapsed            string `json:"elapsed"`
	Status             string `json:"status"`
	Live               bool   `json:"live"`
	Ended              bool   `json:"ended"`
	FinishedTimestamp  string `json:"finishedTimestamp"`
	UpdatedAt          string `json:"updatedAt"`
}

// Orderbook levels are [price, quantity] pairs for each side.
type Orderbook struct {
	Yes        [][]json.Number `json:"yes"`
	No         [][]json.Number `json:"no"`
	YesDollars [][]json.Number `json:"yes_dollars"`
	NoDollars  [][]json.Number `json:"no_dollars"`
}

type Order struct {
	Pubkey          string         `json:"pubkey"`
	Owner           string         `json:"owner"`
	OwnerPubkey     string         `json:"ownerPubkey"`
	Market          string         `json:"market"`
	MarketID        string         `json:"marketId"`
	MarketIDHash    string         `json:"marketIdHash"`
	EventID         string         `json:"eventId"`
	Position        string         `json:"position"`
	Status          string         `json:"status"`
	IsYes           bool           `json:"isYes"`
	IsBuy           bool           `json:"isBuy"`
	CreatedAt       int64          `json:"createdAt"`
	UpdatedAt       int64          `json:"updatedAt"`
	Contracts       string         `json:"contracts"`
	MaxFillPriceUSD string         `json:"maxFillPriceUsd"`
	MaxBuyPriceUSD  string         `json:"maxBuyPriceUsd"`
	MinSellPriceUSD string         `json:"minSellPriceUsd"`
	FilledAt        int64          `json:"filledAt"`
	FilledContracts string         `json:"filledContracts"`
	AvgFillPriceUSD string         `json:"avgFillPriceUsd"`
	Settled         bool           `json:"settled"`
	OrderID         string         `json:"orderId"`
	SizeUSD         string         `json:"sizeUsd"`
	EventMetadata   EventMetadata  `json:"eventMetadata"`
	MarketMetadata  MarketMetadata `json:"marketMetadata"`
	ExternalOrderID string         `json:"externalOrderId"`
	Bump            int            `json:"bump"`

	Metadata response.Metadata `json:"-"`
}

type OrderStatus struct {
	OrderPubkey     string               `json:"orderPubkey"`
	Status          string               `json:"status"`
	LatestEventType string               `json:"latestEventType"`
	LatestSignature string               `json:"latestSignature"`
	ExternalOrderID string               `json:"externalOrderId"`
	OrderID         string               `json:"orderId"`
	History         []OrderStatusHistory `json:"history"`

	Metadata response.Metadata `json:"-"`
}

type OrderStatusHistory struct {
	EventType       string `json:"eventType"`
	Status          string `json:"status"`
	RawStatus       string `json:"rawStatus"`
	Timestamp       int64  `json:"timestamp"`
	Signature       string `json:"signature"`
	ExternalOrderID string `json:"externalOrderId"`
	OrderID         string `json:"orderId"`
}

// Position values that the API nulls once the market closes are empty strings
// or empty json.Number values.
type Position struct {
	Pubkey                 string         `json:"pubkey"`
	Owner                  string         `json:"owner"`
	OwnerPubkey            string         `json:"ownerPubkey"`
	Market                 string         `json:"market"`
	MarketID               string         `json:"marketId"`
	MarketIDHash           string         `json:"marketIdHash"`
	IsYes                  bool           `json:"isYes"`
	Contracts              string         `json:"contracts"`
	ContractsMicro         string         `json:"contractsMicro"`
	ContractsDecimal       string         `json:"contractsDecimal"`
	Source                 string         `json:"source"`
	MaxSlippageBps         int            `json:"maxSlippageBps"`
	TotalCostUSD           string         `json:"totalCostUsd"`
	SizeUSD                string         `json:"sizeUsd"`
	ValueUSD               string         `json:"valueUsd"`
	AvgPriceUSD            string         `json:"avgPriceUsd"`
	MarkPriceUSD           string         `json:"markPriceUsd"`
	SellPriceUSD           string         `json:"sellPriceUsd"`
	PnLUSD                 string         `json:"pnlUsd"`
	PnLUSDPercent          json.Number    `json:"pnlUsdPercent"`
	PnLUSDAfterFees        string         `json:"pnlUsdAfterFees"`
	PnLUSDAfterFeesPercent json.Number    `json:"pnlUsdAfterFeesPercent"`
	OpenOrders             int            `json:"openOrders"`
	FeesPaidUSD            string         `json:"feesPaidUsd"`
	RealizedPnLUSD         json.Number    `json:"realizedPnlUsd"`
	Claimed                bool           `json:"claimed"`
	ClaimedUSD             string         `json:"claimedUsd"`
	OpenedAt               int64          `json:"openedAt"`
	UpdatedAt              int64          `json:"updatedAt"`
	ClaimableAt            *int64         `json:"claimableAt"`
	PayoutUSD              string         `json:"payoutUsd"`
	Bump                   int            `json:"bump"`
	EventID                string         `json:"eventId"`
	EventMetadata          EventMetadata  `json:"eventMetadata"`
	MarketMetadata         MarketMetadata `json:"marketMetadata"`
	SettlementDate         *int64         `json:"settlementDate"`
	Claimable              bool           `json:"claimable"`

	Metadata response.Metadata `json:"-"`
}

type HistoryEvent struct {
	ID                      int64          `json:"id"`
	EventType               string         `json:"eventType"`
	Signature               string         `json:"signature"`
	Slot                    string         `json:"slot"`
	Timestamp               int64          `json:"timestamp"`
	OrderPubkey             string         `json:"orderPubkey"`
	PositionPubkey          string         `json:"positionPubkey"`
	MarketID                string         `json:"marketId"`
	OwnerPubkey             string         `json:"ownerPubkey"`
	KeeperPubkey            string         `json:"keeperPubkey"`
	ExternalOrderID         string         `json:"externalOrderId"`
	OrderID                 string         `json:"orderId"`
	IsBuy                   bool           `json:"isBuy"`
	IsYes                   bool           `json:"isYes"`
	Contracts               string         `json:"contracts"`
	ContractsMicro          string         `json:"contractsMicro"`
	ContractsDecimal        string         `json:"contractsDecimal"`
	FilledContracts         string         `json:"filledContracts"`
	FilledContractsMicro    string         `json:"filledContractsMicro"`
	FilledContractsDecimal  string         `json:"filledContractsDecimal"`
	ContractsSettled        string         `json:"contractsSettled"`
	ContractsSettledMicro   string         `json:"contractsSettledMicro"`
	ContractsSettledDecimal string         `json:"contractsSettledDecimal"`
	MaxFillPriceUSD         string         `json:"maxFillPriceUsd"`
	AvgFillPriceUSD         string         `json:"avgFillPriceUsd"`
	MaxBuyPriceUSD          string         `json:"maxBuyPriceUsd"`
	MinSellPriceUSD         string         `json:"minSellPriceUsd"`
	DepositAmountUSD        string         `json:"depositAmountUsd"`
	TotalCostUSD            string         `json:"totalCostUsd"`
	FeeUSD                  string         `json:"feeUsd"`
	GrossProceedsUSD        string         `json:"grossProceedsUsd"`
	NetProceedsUSD          string         `json:"netProceedsUsd"`
	TransferAmountToken     string         `json:"transferAmountToken"`
	RealizedPnL             string         `json:"realizedPnl"`
	RealizedPnLBeforeFees   string         `json:"realizedPnlBeforeFees"`
	PayoutAmountUSD         string         `json:"payoutAmountUsd"`
	EventID                 string         `json:"eventId"`
	MarketMetadata          MarketMetadata `json:"marketMetadata"`
	EventMetadata           EventMetadata  `json:"eventMetadata"`
}

type TxMeta struct {
	Blockhash            string `json:"blockhash"`
	LastValidBlockHeight uint64 `json:"lastValidBlockHeight"`
}

// Execution carries the opaque context to pass unchanged to Execute.
type Execution struct {
	Endpoint string          `json:"endpoint"`
	Context  json.RawMessage `json:"context"`
}

// BuildOrder describes an order built by CreateOrder or ClosePosition.
type BuildOrder struct {
	OrderPubkey             string `json:"orderPubkey"`
	OrderATAPubkey          string `json:"orderAtaPubkey"`
	UserPubkey              string `json:"userPubkey"`
	MarketID                string `json:"marketId"`
	MarketIDHash            string `json:"marketIdHash"`
	PositionPubkey          string `json:"positionPubkey"`
	IsBuy                   bool   `json:"isBuy"`
	IsYes                   bool   `json:"isYes"`
	Contracts               string `json:"contracts"`
	ContractsMicro          string `json:"contractsMicro"`
	ContractsDecimal        string `json:"contractsDecimal"`
	NewContracts            string `json:"newContracts"`
	NewContractsMicro       string `json:"newContractsMicro"`
	NewContractsDecimal     string `json:"newContractsDecimal"`
	MaxBuyPriceUSD          string `json:"maxBuyPriceUsd"`
	MinSellPriceUSD         string `json:"minSellPriceUsd"`
	ExternalOrderID         string `json:"externalOrderId"`
	OrderCostUSD            string `json:"orderCostUsd"`
	NewAvgPriceUSD          string `json:"newAvgPriceUsd"`
	NewSizeUSD              string `json:"newSizeUsd"`
	NewPayoutUSD            string `json:"newPayoutUsd"`
	EstimatedProtocolFeeUSD string `json:"estimatedProtocolFeeUsd"`
	EstimatedVenueFeeUSD    string `json:"estimatedVenueFeeUsd"`
	EstimatedTotalFeeUSD    string `json:"estimatedTotalFeeUsd"`
	PayoutUSD               string `json:"payoutUsd"`

	// SlippageBps is an estimate for quoting, not a guaranteed loss.
	SlippageBps    int  `json:"slippageBps"`
	MaxSlippageBps *int `json:"maxSlippageBps"`
}

// ClaimDetails describes the position paid out by ClaimPosition.
type ClaimDetails struct {
	PositionPubkey   string `json:"positionPubkey"`
	MarketPubkey     string `json:"marketPubkey"`
	UserPubkey       string `json:"userPubkey"`
	OwnerPubkey      string `json:"ownerPubkey"`
	IsYes            bool   `json:"isYes"`
	Contracts        string `json:"contracts"`
	ContractsMicro   string `json:"contractsMicro"`
	ContractsDecimal string `json:"contractsDecimal"`
	PayoutAmountUSD  string `json:"payoutAmountUsd"`
}

// TransactionBuild is an unsigned transaction from CreateOrder, ClosePosition,
// CloseAllPositions, or ClaimPosition. Order is set for order builds and Position
// for claims. Sign Transaction for every key in RequiredSigners, then submit it with
// Execute, passing Execution, or through your own RPC.
type TransactionBuild struct {
	Transaction          string        `json:"transaction"`
	TxMeta               *TxMeta       `json:"txMeta"`
	ExternalOrderID      string        `json:"externalOrderId"`
	RequiredSigners      []string      `json:"requiredSigners"`
	Execution            *Execution    `json:"execution"`
	ExecutionModel       string        `json:"executionModel"`
	Settlement           string        `json:"settlement"`
	JupiterSwapRequestID string        `json:"jupiterSwapRequestId"`
	Order                *BuildOrder   `json:"order"`
	Position             *ClaimDetails `json:"position"`

	Metadata response.Metadata `json:"-"`
}
