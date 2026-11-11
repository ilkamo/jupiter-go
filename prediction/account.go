package prediction

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/ilkamo/jupiter-go/response"
)

const (
	PnLInterval24h   = "24h"
	PnLInterval1w    = "1w"
	PnLInterval1m    = "1m"
	MinPnLPointCount = 1
	MaxPnLPointCount = 1000

	LeaderboardAllTime = "all_time"
	LeaderboardWeekly  = "weekly"
	LeaderboardMonthly = "monthly"

	MetricPnL     = "pnl"
	MetricVolume  = "volume"
	MetricWinRate = "win_rate"

	MinLeaderboardLimit = 1
	MaxLeaderboardLimit = 100

	MinHistoryEventID = 1
)

type OrderList struct {
	Orders     []Order    `json:"data"`
	Pagination Pagination `json:"pagination"`

	Metadata response.Metadata `json:"-"`
}

// Orders lists orders registered on-chain, optionally for one owner.
func (c *Client) Orders(ctx context.Context, owner string, page Page) (*OrderList, error) {
	query := url.Values{}
	if owner != "" {
		if err := validatePubkey("owner", owner); err != nil {
			return nil, err
		}

		query.Set("ownerPubkey", owner)
	}

	if err := page.apply(query); err != nil {
		return nil, err
	}

	var result OrderList
	metadata, err := c.get(ctx, "/orders", query, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

func (c *Client) Order(ctx context.Context, orderPubkey string) (*Order, error) {
	if err := validatePubkey("order", orderPubkey); err != nil {
		return nil, err
	}

	var order Order
	metadata, err := c.get(ctx, "/orders/"+orderPubkey, nil, &order)
	order.Metadata = metadata
	if err != nil {
		return &order, err
	}

	return &order, nil
}

// OrderStatus reports an order's fill status with its event history. Poll it after
// Execute: an executed order is open, not necessarily filled.
func (c *Client) OrderStatus(ctx context.Context, orderPubkey string) (*OrderStatus, error) {
	if err := validatePubkey("order", orderPubkey); err != nil {
		return nil, err
	}

	var status OrderStatus
	metadata, err := c.get(ctx, "/orders/status/"+orderPubkey, nil, &status)
	status.Metadata = metadata
	if err != nil {
		return &status, err
	}

	return &status, nil
}

// PositionsRequest filters positions. IsYes, when set, selects one side.
type PositionsRequest struct {
	Owner        string
	MarketPubkey string
	MarketID     string
	IsYes        *bool
	Page         Page
}

type PositionList struct {
	Positions  []Position `json:"data"`
	Pagination Pagination `json:"pagination"`

	Metadata response.Metadata `json:"-"`
}

func (c *Client) Positions(ctx context.Context, r PositionsRequest) (*PositionList, error) {
	query := url.Values{}

	for _, pubkey := range []struct {
		field string
		key   string
		value string
	}{
		{field: "owner", key: "ownerPubkey", value: r.Owner},
		{field: "market", key: "marketPubkey", value: r.MarketPubkey},
	} {
		if pubkey.value == "" {
			continue
		}

		if err := validatePubkey(pubkey.field, pubkey.value); err != nil {
			return nil, err
		}

		query.Set(pubkey.key, pubkey.value)
	}

	if r.MarketID != "" {
		query.Set("marketId", r.MarketID)
	}

	if r.IsYes != nil {
		query.Set("isYes", strconv.FormatBool(*r.IsYes))
	}

	if err := r.Page.apply(query); err != nil {
		return nil, err
	}

	var result PositionList
	metadata, err := c.get(ctx, "/positions", query, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

func (c *Client) Position(ctx context.Context, positionPubkey string) (*Position, error) {
	if err := validatePubkey("position", positionPubkey); err != nil {
		return nil, err
	}

	var position Position
	metadata, err := c.get(ctx, "/positions/"+positionPubkey, nil, &position)
	position.Metadata = metadata
	if err != nil {
		return &position, err
	}

	return &position, nil
}

// HistoryRequest filters order, fill, settlement, and claim events. Set EventID
// (at least MinHistoryEventID) to fetch a single event.
type HistoryRequest struct {
	Owner          string
	PositionPubkey string
	EventID        *uint64
	Page           Page
}

type HistoryList struct {
	Events     []HistoryEvent `json:"data"`
	Pagination Pagination     `json:"pagination"`

	Metadata response.Metadata `json:"-"`
}

func (c *Client) History(ctx context.Context, r HistoryRequest) (*HistoryList, error) {
	query := url.Values{}
	if r.Owner != "" {
		if err := validatePubkey("owner", r.Owner); err != nil {
			return nil, err
		}

		query.Set("ownerPubkey", r.Owner)
	}

	if r.PositionPubkey != "" {
		if err := validatePubkey("position", r.PositionPubkey); err != nil {
			return nil, err
		}

		query.Set("positionPubkey", r.PositionPubkey)
	}

	if r.EventID != nil {
		if *r.EventID < MinHistoryEventID {
			return nil, fmt.Errorf("history event ID must be at least %d", MinHistoryEventID)
		}

		query.Set("id", strconv.FormatUint(*r.EventID, decimalBase))
	}

	if err := r.Page.apply(query); err != nil {
		return nil, err
	}

	var result HistoryList
	metadata, err := c.get(ctx, "/history", query, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

// Profile counts are decimal strings and USD values are micro USD.
type Profile struct {
	OwnerPubkey                 string `json:"ownerPubkey"`
	RealizedPnLUSD              string `json:"realizedPnlUsd"`
	TotalVolumeUSD              string `json:"totalVolumeUsd"`
	PredictionsCount            string `json:"predictionsCount"`
	CorrectPredictions          string `json:"correctPredictions"`
	WrongPredictions            string `json:"wrongPredictions"`
	TotalActiveContracts        string `json:"totalActiveContracts"`
	TotalActiveContractsMicro   string `json:"totalActiveContractsMicro"`
	TotalActiveContractsDecimal string `json:"totalActiveContractsDecimal"`
	TotalPositionsValueUSD      string `json:"totalPositionsValueUsd"`

	Metadata response.Metadata `json:"-"`
}

func (c *Client) Profile(ctx context.Context, owner string) (*Profile, error) {
	if err := validatePubkey("owner", owner); err != nil {
		return nil, err
	}

	var profile Profile
	metadata, err := c.get(ctx, "/profiles/"+owner, nil, &profile)
	profile.Metadata = metadata
	if err != nil {
		return &profile, err
	}

	return &profile, nil
}

// PnLHistoryRequest selects Count points spaced by Interval (PnLInterval24h,
// PnLInterval1w, or PnLInterval1m). Empty values use the API defaults.
type PnLHistoryRequest struct {
	Owner    string
	Interval string
	Count    *uint16
}

type PnLPoint struct {
	Timestamp      int64  `json:"timestamp"`
	RealizedPnLUSD string `json:"realizedPnlUsd"`
}

type PnLHistory struct {
	OwnerPubkey string     `json:"ownerPubkey"`
	History     []PnLPoint `json:"history"`

	Metadata response.Metadata `json:"-"`
}

func (c *Client) PnLHistory(ctx context.Context, r PnLHistoryRequest) (*PnLHistory, error) {
	if err := validatePubkey("owner", r.Owner); err != nil {
		return nil, err
	}

	query := url.Values{}
	if err := setFilter(query, "interval", r.Interval); err != nil {
		return nil, err
	}

	if r.Count != nil {
		if *r.Count < MinPnLPointCount || *r.Count > MaxPnLPointCount {
			return nil, fmt.Errorf("count must be between %d and %d", MinPnLPointCount, MaxPnLPointCount)
		}

		query.Set("count", strconv.Itoa(int(*r.Count)))
	}

	var history PnLHistory
	metadata, err := c.get(ctx, "/profiles/"+r.Owner+"/pnl-history", query, &history)
	history.Metadata = metadata
	if err != nil {
		return &history, err
	}

	return &history, nil
}

type Trade struct {
	ID            string         `json:"id"`
	OwnerPubkey   string         `json:"ownerPubkey"`
	MarketID      string         `json:"marketId"`
	Message       string         `json:"message"`
	Timestamp     json.Number    `json:"timestamp"`
	Action        string         `json:"action"`
	Side          string         `json:"side"`
	EventTitle    string         `json:"eventTitle"`
	MarketTitle   string         `json:"marketTitle"`
	AmountUSD     string         `json:"amountUsd"`
	PriceUSD      string         `json:"priceUsd"`
	EventImageURL string         `json:"eventImageUrl"`
	EventID       string         `json:"eventId"`
	IsTeamMarket  bool           `json:"isTeamMarket"`
	MarketOptions []MarketOption `json:"marketOptions"`
}

type TradeList struct {
	Trades []Trade `json:"data"`

	Metadata response.Metadata `json:"-"`
}

// Trades returns recent fills across the exchange.
func (c *Client) Trades(ctx context.Context) (*TradeList, error) {
	var result TradeList
	metadata, err := c.get(ctx, "/trades", nil, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

// LeaderboardRequest selects a Period and Metric; empty values use the API
// defaults (all time, P&L, 100 entries).
type LeaderboardRequest struct {
	Period string
	Metric string
	Limit  *uint8
}

type LeaderboardEntry struct {
	OwnerPubkey        string      `json:"ownerPubkey"`
	RealizedPnLUSD     string      `json:"realizedPnlUsd"`
	TotalVolumeUSD     string      `json:"totalVolumeUsd"`
	PredictionsCount   json.Number `json:"predictionsCount"`
	CorrectPredictions json.Number `json:"correctPredictions"`
	WrongPredictions   json.Number `json:"wrongPredictions"`
	WinRatePct         string      `json:"winRatePct"`
	Period             string      `json:"period"`
	PeriodStart        string      `json:"periodStart"`
	PeriodEnd          string      `json:"periodEnd"`
}

type Leaderboard struct {
	Entries []LeaderboardEntry `json:"data"`

	Metadata response.Metadata `json:"-"`
}

func (c *Client) Leaderboards(ctx context.Context, r LeaderboardRequest) (*Leaderboard, error) {
	query := url.Values{}
	if err := setFilter(query, "period", r.Period); err != nil {
		return nil, err
	}

	if err := setFilter(query, "metric", r.Metric); err != nil {
		return nil, err
	}

	if r.Limit != nil {
		if *r.Limit < MinLeaderboardLimit || *r.Limit > MaxLeaderboardLimit {
			return nil, fmt.Errorf(
				"limit must be between %d and %d",
				MinLeaderboardLimit,
				MaxLeaderboardLimit,
			)
		}

		query.Set("limit", strconv.Itoa(int(*r.Limit)))
	}

	var result Leaderboard
	metadata, err := c.get(ctx, "/leaderboards", query, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

// VaultInfo describes the settlement vault that pays out claims.
type VaultInfo struct {
	Pubkey       string            `json:"pubkey"`
	Data         map[string]string `json:"data"`
	VaultBalance string            `json:"vaultBalance"`

	Metadata response.Metadata `json:"-"`
}

func (c *Client) VaultInfo(ctx context.Context) (*VaultInfo, error) {
	var info VaultInfo
	metadata, err := c.get(ctx, "/vault-info", nil, &info)
	info.Metadata = metadata
	if err != nil {
		return &info, err
	}

	return &info, nil
}
