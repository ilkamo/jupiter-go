package jupiter_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/mr-tron/base58"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/internal/address"
	"github.com/ilkamo/jupiter-go/tokens"
)

const tokenFixture = `[{
  "id": "JUPyiwrYJFskUPiHa7hkeR8VUtAeFoSYbKedZNsDvCN",
  "name": "Jupiter",
  "symbol": "JUP",
  "icon": null,
  "decimals": 6,
  "circSupply": 3243891294.880000000000001,
  "holderCount": 0,
  "usdPrice": 0.1415,
  "priceBlockId": 402118601,
  "mcap": null,
  "audit": {"isSus": false, "mintAuthorityDisabled": true, "topHoldersPercentage": 15.45},
  "organicScoreLabel": "high",
  "isVerified": true,
  "tags": ["verified"],
  "stats24h": {"priceChange": -8.27, "numBuys": 33979}
}]`

type recordedRequest struct {
	path  string
	query map[string][]string
}

// recordingServer serves body to every request and records each request's path and query.
func recordingServer(t *testing.T, body string) (*jupiter.Client, func() []recordedRequest) {
	t.Helper()

	var (
		mu       sync.Mutex
		requests []recordedRequest
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests = append(requests, recordedRequest{path: r.URL.EscapedPath(), query: r.URL.Query()})
		mu.Unlock()

		if _, err := w.Write([]byte(body)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client, err := jupiter.NewClient(jupiter.WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	recorded := func() []recordedRequest {
		mu.Lock()
		defer mu.Unlock()

		return append([]recordedRequest(nil), requests...)
	}

	return client, recorded
}

// distinctMints returns count valid, distinct mint addresses.
func distinctMints(count int) []string {
	mints := make([]string, count)
	for i := range mints {
		key := make([]byte, address.PublicKeyBytes)
		key[0] = byte(i + 1)
		mints[i] = base58.Encode(key)
	}

	return mints
}

func TestTokenDecoding(t *testing.T) {
	client, _ := recordingServer(t, tokenFixture)

	result, err := client.Tokens.Search(context.Background(), "JUP")
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Tokens) != 1 {
		t.Fatalf("got %d tokens", len(result.Tokens))
	}

	token := result.Tokens[0]
	if token.Decimals != 6 || token.Icon != "" || token.Symbol != "JUP" {
		t.Fatalf("scalar fields lost: %+v", token)
	}

	if token.CircSupply.String() != "3243891294.880000000000001" || token.USDPrice.String() != "0.1415" {
		t.Fatalf("precision lost: %s %s", token.CircSupply, token.USDPrice)
	}

	if token.HolderCount.String() != "0" || token.MarketCap != "" {
		t.Fatalf("zero or null number changed: %q %q", token.HolderCount, token.MarketCap)
	}

	audit := token.Audit
	if audit == nil || audit.IsSus == nil || *audit.IsSus || audit.FreezeAuthorityDisabled != nil {
		t.Fatalf("audit presence lost: %+v", audit)
	}

	if token.IsVerified == nil || !*token.IsVerified || token.Stats24h == nil || token.Stats5m != nil {
		t.Fatalf("optional objects lost: %+v", token)
	}

	if token.Stats24h.PriceChange.String() != "-8.27" {
		t.Fatalf("stats lost: %+v", token.Stats24h)
	}
}

func TestTokenRequests(t *testing.T) {
	maxLimit := uint8(tokens.MaxCategoryLimit)
	lookupMints := distinctMints(tokens.MaxLookupMints)

	tests := []struct {
		name      string
		call      func(context.Context, *jupiter.Client) error
		wantPath  string
		wantQuery map[string]string
	}{
		{
			name: "search",
			call: func(ctx context.Context, c *jupiter.Client) error {
				_, err := c.Tokens.Search(ctx, "Jupiter Token")
				return err
			},
			wantPath:  "/tokens/v2/search",
			wantQuery: map[string]string{"query": "Jupiter Token"},
		},
		{
			name: "lookup maximum mints",
			call: func(ctx context.Context, c *jupiter.Client) error {
				_, err := c.Tokens.Lookup(ctx, lookupMints)
				return err
			},
			wantPath:  "/tokens/v2/search",
			wantQuery: map[string]string{"query": strings.Join(lookupMints, ",")},
		},
		{
			name: "documented tag",
			call: func(ctx context.Context, c *jupiter.Client) error {
				_, err := c.Tokens.Tag(ctx, tokens.TagVerified)
				return err
			},
			wantPath:  "/tokens/v2/tag",
			wantQuery: map[string]string{"query": "verified"},
		},
		{
			name: "tag added upstream",
			call: func(ctx context.Context, c *jupiter.Client) error {
				_, err := c.Tokens.Tag(ctx, "newtag")
				return err
			},
			wantPath:  "/tokens/v2/tag",
			wantQuery: map[string]string{"query": "newtag"},
		},
		{
			name: "category with maximum limit",
			call: func(ctx context.Context, c *jupiter.Client) error {
				_, err := c.Tokens.Category(ctx, tokens.CategoryRequest{
					Category: tokens.CategoryTopTraded,
					Interval: tokens.Interval24h,
					Limit:    &maxLimit,
				})
				return err
			},
			wantPath:  "/tokens/v2/toptraded/24h",
			wantQuery: map[string]string{"limit": strconv.Itoa(tokens.MaxCategoryLimit)},
		},
		{
			name: "category values are path-escaped",
			call: func(ctx context.Context, c *jupiter.Client) error {
				_, err := c.Tokens.Category(ctx, tokens.CategoryRequest{
					Category: "new/category",
					Interval: tokens.Interval1h,
				})
				return err
			},
			wantPath:  "/tokens/v2/new%2Fcategory/1h",
			wantQuery: map[string]string{},
		},
		{
			name: "recent",
			call: func(ctx context.Context, c *jupiter.Client) error {
				_, err := c.Tokens.Recent(ctx)
				return err
			},
			wantPath:  "/tokens/v2/recent",
			wantQuery: map[string]string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, recorded := recordingServer(t, `[]`)

			if err := test.call(context.Background(), client); err != nil {
				t.Fatal(err)
			}

			requests := recorded()
			if len(requests) != 1 || requests[0].path != test.wantPath {
				t.Fatalf("requests = %+v, want path %s", requests, test.wantPath)
			}

			if len(requests[0].query) != len(test.wantQuery) {
				t.Fatalf("query = %v, want %v", requests[0].query, test.wantQuery)
			}

			for key, value := range test.wantQuery {
				if got := requests[0].query[key]; len(got) != 1 || got[0] != value {
					t.Fatalf("%s = %v, want %q", key, got, value)
				}
			}
		})
	}
}

func TestInvalidTokenRequestsDoNotReachTransport(t *testing.T) {
	client, err := jupiter.NewClient(jupiter.WithHTTPClient(unreachableHTTPClient{t: t}))
	if err != nil {
		t.Fatal(err)
	}

	belowLimit := uint8(tokens.MinCategoryLimit - 1)
	aboveLimit := uint8(tokens.MaxCategoryLimit + 1)

	tests := []struct {
		name         string
		call         func(context.Context) error
		addressError bool
	}{
		{
			name: "empty search",
			call: func(ctx context.Context) error {
				_, err := client.Tokens.Search(ctx, "")
				return err
			},
		},
		{
			name: "search with several mints",
			call: func(ctx context.Context) error {
				_, err := client.Tokens.Search(ctx, wrappedSOLMint+","+usdcMint)
				return err
			},
		},
		{
			name: "lookup without mints",
			call: func(ctx context.Context) error {
				_, err := client.Tokens.Lookup(ctx, nil)
				return err
			},
		},
		{
			name: "lookup above maximum",
			call: func(ctx context.Context) error {
				_, err := client.Tokens.Lookup(ctx, distinctMints(tokens.MaxLookupMints+1))
				return err
			},
		},
		{
			name: "lookup malformed mint",
			call: func(ctx context.Context) error {
				_, err := client.Tokens.Lookup(ctx, []string{wrappedSOLMint, "mint"})
				return err
			},
			addressError: true,
		},
		{
			name: "tag with separator",
			call: func(ctx context.Context) error {
				_, err := client.Tokens.Tag(ctx, "lst,verified")
				return err
			},
		},
		{
			name: "category without interval",
			call: func(ctx context.Context) error {
				_, err := client.Tokens.Category(ctx, tokens.CategoryRequest{
					Category: tokens.CategoryTopTrending,
				})
				return err
			},
		},
		{
			name: "category limit below minimum",
			call: func(ctx context.Context) error {
				_, err := client.Tokens.Category(ctx, tokens.CategoryRequest{
					Category: tokens.CategoryTopTrending,
					Interval: tokens.Interval5m,
					Limit:    &belowLimit,
				})
				return err
			},
		},
		{
			name: "category limit above maximum",
			call: func(ctx context.Context) error {
				_, err := client.Tokens.Category(ctx, tokens.CategoryRequest{
					Category: tokens.CategoryTopTrending,
					Interval: tokens.Interval5m,
					Limit:    &aboveLimit,
				})
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.call(context.Background())
			if err == nil {
				t.Fatal("invalid request accepted")
			}

			if errors.Is(err, tokens.ErrInvalidAddress) != test.addressError {
				t.Fatalf("unexpected address error classification: %v", err)
			}
		})
	}
}

func TestTokensRetryTransientFailures(t *testing.T) {
	server, calls := flakyServer(t, http.StatusServiceUnavailable, `[]`)

	client, err := jupiter.NewClient(jupiter.WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	result, err := client.Tokens.Recent(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if calls.Load() != recoveredAttempts || result.Metadata.Attempts != recoveredAttempts {
		t.Fatalf("calls=%d attempts=%d", calls.Load(), result.Metadata.Attempts)
	}
}
