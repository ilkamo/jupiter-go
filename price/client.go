// Package price implements Jupiter's Price V3 API.
package price

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ilkamo/jupiter-go/internal/address"
	"github.com/ilkamo/jupiter-go/internal/listparam"
	"github.com/ilkamo/jupiter-go/internal/transport"
	"github.com/ilkamo/jupiter-go/response"
)

const (
	MaxMints = 50

	pricePath = "/price/v3/"
)

// ErrInvalidAddress identifies mint arguments that are not Solana account addresses.
// It is the same value as swap.ErrInvalidAddress.
var ErrInvalidAddress = address.ErrInvalid

type Client struct {
	transport *transport.Client
}

// NewClient is called by jupiter.NewClient, which applications use instead; it
// takes the SDK's internal transport.
func NewClient(t *transport.Client) *Client {
	return &Client{transport: t}
}

// Price is the USD price of one token. Numbers keep upstream precision; an empty
// PriceChange24h means the API reported null.
type Price struct {
	CreatedAt      string      `json:"createdAt"`
	Liquidity      json.Number `json:"liquidity"`
	USDPrice       json.Number `json:"usdPrice"`
	BlockID        *uint64     `json:"blockId"`
	Decimals       int         `json:"decimals"`
	PriceChange24h json.Number `json:"priceChange24h"`
}

type Response struct {
	// Prices maps mint addresses to prices.
	Prices map[string]Price

	// Missing lists requested mints, in request order, that Jupiter omitted because
	// it has no reliable price for them.
	Missing []string

	Metadata response.Metadata
}

// Get returns USD prices for 1 to MaxMints mint addresses.
func (c *Client) Get(ctx context.Context, mints []string) (*Response, error) {
	if len(mints) == 0 || len(mints) > MaxMints {
		return nil, fmt.Errorf("price requires between 1 and %d mints", MaxMints)
	}

	for i, mint := range mints {
		if err := address.Validate(mint); err != nil {
			return nil, fmt.Errorf("mint %d: %w", i, err)
		}
	}

	request := transport.Request{
		Method: http.MethodGet,
		Path:   pricePath,
		Query:  url.Values{"ids": {listparam.Join(mints)}},
		Retry:  transport.ReadOnlyRetry,
	}

	var result Response
	metadata, err := c.transport.Do(ctx, request, &result.Prices)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	for _, mint := range mints {
		if _, ok := result.Prices[mint]; !ok {
			result.Missing = append(result.Missing, mint)
		}
	}

	return &result, nil
}
