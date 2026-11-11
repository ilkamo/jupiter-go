// Package prediction implements Jupiter's Prediction Market API: binary YES/NO
// markets on real-world events.
//
// The API is beta upstream and may change before the SDK's compatibility review.
// Jupiter blocks requests from United States and South Korea IP addresses. Order,
// sell, and claim operations build unsigned transactions; Execute submits a signed
// one.
package prediction

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"unicode/utf8"

	"github.com/ilkamo/jupiter-go/internal/address"
	"github.com/ilkamo/jupiter-go/internal/listparam"
	"github.com/ilkamo/jupiter-go/internal/transport"
	"github.com/ilkamo/jupiter-go/response"
)

// Documented filter values. Filters pass other values through, format-checked, so
// upstream additions remain usable.
const (
	ProviderPolymarket = "polymarket"
	ProviderKalshi     = "kalshi"
	ProviderBisonFi    = "bisonfi"

	CategoryAll       = "all"
	CategoryCrypto    = "crypto"
	CategorySports    = "sports"
	CategoryPolitics  = "politics"
	CategoryEsports   = "esports"
	CategoryCulture   = "culture"
	CategoryEconomics = "economics"
	CategoryTech      = "tech"

	FilterNew      = "new"
	FilterLive     = "live"
	FilterTrending = "trending"
	FilterUpcoming = "upcoming"

	SortByVolume  = "volume"
	SortByBeginAt = "beginAt"

	DirectionAscending  = "asc"
	DirectionDescending = "desc"
)

const (
	MinSearchLimit     = 1
	MaxSearchLimit     = 20
	MaxSearchQueryRune = 200

	basePath    = "/prediction/v1"
	decimalBase = 10
)

// ErrInvalidAddress is the same value as swap.ErrInvalidAddress.
var ErrInvalidAddress = address.ErrInvalid

type Client struct {
	transport *transport.Client
}

func NewClient(t *transport.Client) *Client {
	return &Client{transport: t}
}

// Page selects a range of results by index. Nil bounds use the API defaults.
type Page struct {
	Start *uint32
	End   *uint32
}

func (p Page) apply(query url.Values) error {
	if p.Start != nil && p.End != nil && *p.End < *p.Start {
		return fmt.Errorf("page end must not precede start")
	}

	if p.Start != nil {
		query.Set("start", strconv.FormatUint(uint64(*p.Start), decimalBase))
	}

	if p.End != nil {
		query.Set("end", strconv.FormatUint(uint64(*p.End), decimalBase))
	}

	return nil
}

// setFilter adds an optional enum-like filter after a format check.
func setFilter(query url.Values, key string, value string) error {
	if value == "" {
		return nil
	}

	if err := listparam.ValidateItem(key, value); err != nil {
		return err
	}

	query.Set(key, value)

	return nil
}

func pathID(field string, id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("%s required", field)
	}

	return url.PathEscape(id), nil
}

func validatePubkey(field string, value string) error {
	if err := address.Validate(value); err != nil {
		return fmt.Errorf("%s: %w", field, err)
	}

	return nil
}

func (c *Client) get(
	ctx context.Context,
	path string,
	query url.Values,
	result any,
) (response.Metadata, error) {
	request := transport.Request{
		Method: http.MethodGet,
		Path:   basePath + path,
		Query:  query,
		Retry:  transport.ReadOnlyRetry,
	}

	return c.transport.Do(ctx, request, result)
}

type EventList struct {
	Events     []Event    `json:"data"`
	Pagination Pagination `json:"pagination"`

	Metadata response.Metadata `json:"-"`
}

// EventsRequest filters events. IncludeMarkets nests each event's markets.
type EventsRequest struct {
	Provider          string
	IncludeMarkets    bool
	IncludeAllMarkets bool
	Page              Page
	Category          string
	Subcategory       string
	SortBy            string
	SortDirection     string
	Filter            string
	Tags              string
}

func (r EventsRequest) query() (url.Values, error) {
	query := url.Values{}
	filters := []struct {
		key   string
		value string
	}{
		{key: "provider", value: r.Provider},
		{key: "category", value: r.Category},
		{key: "subcategory", value: r.Subcategory},
		{key: "sortBy", value: r.SortBy},
		{key: "sortDirection", value: r.SortDirection},
		{key: "filter", value: r.Filter},
		{key: "tags", value: r.Tags},
	}
	for _, filter := range filters {
		if err := setFilter(query, filter.key, filter.value); err != nil {
			return nil, err
		}
	}

	if r.IncludeMarkets {
		query.Set("includeMarkets", strconv.FormatBool(true))
	}

	if r.IncludeAllMarkets {
		query.Set("includeAllMarkets", strconv.FormatBool(true))
	}

	if err := r.Page.apply(query); err != nil {
		return nil, err
	}

	return query, nil
}

func (c *Client) Events(ctx context.Context, r EventsRequest) (*EventList, error) {
	query, err := r.query()
	if err != nil {
		return nil, err
	}

	return c.eventList(ctx, "/events", query)
}

// SearchRequest searches event titles. Query is 1 to MaxSearchQueryRune
// characters; Limit, when set, is MinSearchLimit to MaxSearchLimit.
type SearchRequest struct {
	Query    string
	Provider string
	Limit    *uint8
}

func (c *Client) SearchEvents(ctx context.Context, r SearchRequest) (*EventList, error) {
	length := utf8.RuneCountInString(r.Query)
	if length == 0 || length > MaxSearchQueryRune {
		return nil, fmt.Errorf("search query must be 1 to %d characters", MaxSearchQueryRune)
	}

	query := url.Values{"query": {r.Query}}
	if err := setFilter(query, "provider", r.Provider); err != nil {
		return nil, err
	}

	if r.Limit != nil {
		if *r.Limit < MinSearchLimit || *r.Limit > MaxSearchLimit {
			return nil, fmt.Errorf("limit must be between %d and %d", MinSearchLimit, MaxSearchLimit)
		}

		query.Set("limit", strconv.Itoa(int(*r.Limit)))
	}

	return c.eventList(ctx, "/events/search", query)
}

// SuggestedEvents returns events suggested from a wallet's past orders.
func (c *Client) SuggestedEvents(ctx context.Context, wallet string, provider string) (*EventList, error) {
	if err := validatePubkey("wallet", wallet); err != nil {
		return nil, err
	}

	query := url.Values{}
	if err := setFilter(query, "provider", provider); err != nil {
		return nil, err
	}

	return c.eventList(ctx, "/events/suggested/"+wallet, query)
}

func (c *Client) eventList(ctx context.Context, path string, query url.Values) (*EventList, error) {
	var result EventList
	metadata, err := c.get(ctx, path, query, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

// EventRequest selects one event and whether to nest its markets.
type EventRequest struct {
	EventID           string
	IncludeMarkets    bool
	IncludeAllMarkets bool
}

func (c *Client) Event(ctx context.Context, r EventRequest) (*Event, error) {
	id, err := pathID("event ID", r.EventID)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	if r.IncludeMarkets {
		query.Set("includeMarkets", strconv.FormatBool(true))
	}

	if r.IncludeAllMarkets {
		query.Set("includeAllMarkets", strconv.FormatBool(true))
	}

	var event Event
	metadata, err := c.get(ctx, "/events/"+id, query, &event)
	event.Metadata = metadata
	if err != nil {
		return &event, err
	}

	return &event, nil
}

type MarketList struct {
	Markets    []Market   `json:"data"`
	Pagination Pagination `json:"pagination"`

	Metadata response.Metadata `json:"-"`
}

func (c *Client) EventMarkets(ctx context.Context, eventID string, page Page) (*MarketList, error) {
	id, err := pathID("event ID", eventID)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	if err := page.apply(query); err != nil {
		return nil, err
	}

	var result MarketList
	metadata, err := c.get(ctx, "/events/"+id+"/markets", query, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

// EventMarket returns a market scoped to its event.
func (c *Client) EventMarket(ctx context.Context, eventID string, marketID string) (*Market, error) {
	event, err := pathID("event ID", eventID)
	if err != nil {
		return nil, err
	}

	market, err := pathID("market ID", marketID)
	if err != nil {
		return nil, err
	}

	return c.market(ctx, "/events/"+event+"/markets/"+market)
}

// Market returns current market details. Re-fetch it right before ordering:
// prices move.
func (c *Client) Market(ctx context.Context, marketID string) (*Market, error) {
	id, err := pathID("market ID", marketID)
	if err != nil {
		return nil, err
	}

	return c.market(ctx, "/markets/"+id)
}

func (c *Client) market(ctx context.Context, path string) (*Market, error) {
	var market Market
	metadata, err := c.get(ctx, path, nil, &market)
	market.Metadata = metadata
	if err != nil {
		return &market, err
	}

	return &market, nil
}

type GameScores struct {
	Scores []GameScore `json:"data"`

	Metadata response.Metadata `json:"-"`
}

// EventScores returns live scores for sports events; events without a score are
// omitted.
func (c *Client) EventScores(ctx context.Context, eventIDs []string) (*GameScores, error) {
	if len(eventIDs) == 0 {
		return nil, fmt.Errorf("event IDs required")
	}

	if err := listparam.Validate("event IDs", eventIDs); err != nil {
		return nil, err
	}

	query := url.Values{"eventIds": {listparam.Join(eventIDs)}}

	var result GameScores
	metadata, err := c.get(ctx, "/events/scores", query, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

type EventScore struct {
	// Score is nil when the event has no score.
	Score    *GameScore
	Metadata response.Metadata
}

func (c *Client) EventScore(ctx context.Context, eventID string) (*EventScore, error) {
	id, err := pathID("event ID", eventID)
	if err != nil {
		return nil, err
	}

	var result EventScore
	metadata, err := c.get(ctx, "/events/"+id+"/score", nil, &result.Score)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

type OrderbookResult struct {
	// Orderbook is nil when the API reports no orderbook.
	Orderbook *Orderbook
	Metadata  response.Metadata
}

func (c *Client) Orderbook(ctx context.Context, marketID string) (*OrderbookResult, error) {
	id, err := pathID("market ID", marketID)
	if err != nil {
		return nil, err
	}

	var result OrderbookResult
	metadata, err := c.get(ctx, "/orderbook/"+id, nil, &result.Orderbook)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

type TradingStatus struct {
	TradingActive bool `json:"trading_active"`

	Metadata response.Metadata `json:"-"`
}

// TradingStatus reports whether the exchange accepts orders; check it before
// building one.
func (c *Client) TradingStatus(ctx context.Context) (*TradingStatus, error) {
	var status TradingStatus
	metadata, err := c.get(ctx, "/trading-status", nil, &status)
	status.Metadata = metadata
	if err != nil {
		return &status, err
	}

	return &status, nil
}

type ForecastPoint struct {
	MarketTicker         string      `json:"market_ticker"`
	EventTicker          string      `json:"event_ticker"`
	EndPeriodTS          json.Number `json:"end_period_ts"`
	PeriodInterval       json.Number `json:"period_interval"`
	NumericalForecast    json.Number `json:"numerical_forecast"`
	RawNumericalForecast json.Number `json:"raw_numerical_forecast"`
	FormattedForecast    string      `json:"formatted_forecast"`
}

type Forecast struct {
	History []ForecastPoint `json:"forecast_history"`

	Metadata response.Metadata `json:"-"`
}

// Forecast returns the forecast time series recorded for a market.
func (c *Client) Forecast(ctx context.Context, marketID string) (*Forecast, error) {
	if marketID == "" {
		return nil, fmt.Errorf("market ID required")
	}

	var forecast Forecast
	metadata, err := c.get(ctx, "/forecast", url.Values{"marketId": {marketID}}, &forecast)
	forecast.Metadata = metadata
	if err != nil {
		return &forecast, err
	}

	return &forecast, nil
}
