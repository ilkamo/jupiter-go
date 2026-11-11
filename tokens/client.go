// Package tokens implements Jupiter's Tokens V2 discovery operations.
//
// Express Verification, which crafts and pays for a submission, is separate:
// Client.Express.
package tokens

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ilkamo/jupiter-go/internal/address"
	"github.com/ilkamo/jupiter-go/internal/listparam"
	"github.com/ilkamo/jupiter-go/internal/transport"
)

// Documented values for Tag and Category. Other values are passed through so
// additions upstream remain usable without an SDK release.
const (
	TagLST      = "lst"
	TagVerified = "verified"
	TagStocks   = "stocks"

	CategoryTopOrganicScore = "toporganicscore"
	CategoryTopTraded       = "toptraded"
	CategoryTopTrending     = "toptrending"

	Interval5m  = "5m"
	Interval1h  = "1h"
	Interval6h  = "6h"
	Interval24h = "24h"
)

const (
	MaxLookupMints   = 100
	MinCategoryLimit = 1
	MaxCategoryLimit = 100

	basePath = "/tokens/v2"
)

// ErrInvalidAddress identifies mint arguments that are not Solana account addresses.
// It is the same value as swap.ErrInvalidAddress.
var ErrInvalidAddress = address.ErrInvalid

type Client struct {
	transport *transport.Client

	// Express submits tokens for paid VRFD Express Verification.
	Express *ExpressClient
}

func NewClient(t *transport.Client) *Client {
	return &Client{
		transport: t,
		Express:   &ExpressClient{transport: t},
	}
}

// Search finds tokens by symbol, name, or a single mint address. Jupiter returns up
// to 20 matches for symbol and name searches. Use Lookup for several mints.
func (c *Client) Search(ctx context.Context, query string) (*ListResponse, error) {
	if err := listparam.ValidateItem("search query", query); err != nil {
		return nil, fmt.Errorf("%w; use Lookup for several mints", err)
	}

	return c.list(ctx, "/search", url.Values{"query": {query}})
}

// Lookup returns token information for up to MaxLookupMints mint addresses. Unknown
// mints are omitted from the result.
func (c *Client) Lookup(ctx context.Context, mints []string) (*ListResponse, error) {
	if len(mints) == 0 || len(mints) > MaxLookupMints {
		return nil, fmt.Errorf("lookup requires between 1 and %d mints", MaxLookupMints)
	}

	for i, mint := range mints {
		if err := address.Validate(mint); err != nil {
			return nil, fmt.Errorf("mint %d: %w", i, err)
		}
	}

	return c.list(ctx, "/search", url.Values{"query": {listparam.Join(mints)}})
}

// Tag returns every token with the tag, such as TagVerified. The result can be large.
func (c *Client) Tag(ctx context.Context, tag string) (*ListResponse, error) {
	if err := listparam.ValidateItem("tag", tag); err != nil {
		return nil, err
	}

	return c.list(ctx, "/tag", url.Values{"query": {tag}})
}

// CategoryRequest selects top tokens in a trading category over an interval.
// Jupiter excludes generic top tokens such as SOL and USDC and returns 50 tokens
// unless Limit is set.
type CategoryRequest struct {
	Category string
	Interval string
	Limit    *uint8
}

func (c *Client) Category(ctx context.Context, r CategoryRequest) (*ListResponse, error) {
	if r.Category == "" || r.Interval == "" {
		return nil, fmt.Errorf("category and interval required")
	}

	query := url.Values{}
	if r.Limit != nil {
		if *r.Limit < MinCategoryLimit || *r.Limit > MaxCategoryLimit {
			return nil, fmt.Errorf("limit must be between %d and %d", MinCategoryLimit, MaxCategoryLimit)
		}

		query.Set("limit", strconv.Itoa(int(*r.Limit)))
	}

	path := "/" + url.PathEscape(r.Category) + "/" + url.PathEscape(r.Interval)

	return c.list(ctx, path, query)
}

// Recent returns tokens whose first pool was created recently; Jupiter returns 30.
func (c *Client) Recent(ctx context.Context) (*ListResponse, error) {
	return c.list(ctx, "/recent", nil)
}

func (c *Client) list(
	ctx context.Context,
	path string,
	query url.Values,
) (*ListResponse, error) {
	request := transport.Request{
		Method: http.MethodGet,
		Path:   basePath + path,
		Query:  query,
		Retry:  transport.ReadOnlyRetry,
	}

	var result ListResponse
	metadata, err := c.transport.Do(ctx, request, &result.Tokens)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}
