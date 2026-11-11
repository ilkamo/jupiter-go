package jupiter_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/lend"
	"github.com/ilkamo/jupiter-go/solanatx"
)

const (
	maxSignedAmount  = "170141183460469231731687303715884105727"
	instructionReply = `{
  "programId": "jup3YeL8QhtSx1e253b2FDvsMNC87fDrgQZivbrndc9",
  "accounts": [{
    "pubkey": "GkwFnmMDvn3HGMpJpWBg8tgJxr3NxNvg3AXxvXVPbRGJ",
    "isSigner": true,
    "isWritable": true
  }],
  "data": "AQID"
}`
)

type lendRequest struct {
	method string
	path   string
	query  url.Values
	body   map[string]json.RawMessage
}

// lendServer serves fixed bodies by path and records every request.
func lendServer(t *testing.T, replies map[string]string) (*jupiter.Client, func() []lendRequest) {
	t.Helper()

	var (
		mu       sync.Mutex
		requests []lendRequest
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}

		recorded := lendRequest{method: r.Method, path: r.URL.Path, query: r.URL.Query()}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &recorded.body); err != nil {
				t.Errorf("decode body: %v", err)
			}
		}

		mu.Lock()
		requests = append(requests, recorded)
		mu.Unlock()

		reply, ok := replies[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
		}

		if _, err := w.Write([]byte(reply)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client, err := jupiter.NewClient(jupiter.WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	return client, func() []lendRequest {
		mu.Lock()
		defer mu.Unlock()

		return append([]lendRequest(nil), requests...)
	}
}

func TestEarnBuildersEncodeRequests(t *testing.T) {
	amount := lend.AmountRequest{Asset: usdcMint, Signer: takerAddress, Amount: "0001000000"}
	shares := lend.SharesRequest{Asset: usdcMint, Signer: takerAddress, Shares: "000500"}

	tests := []struct {
		path     string
		quantity string
		want     string
		call     func(context.Context, *lend.EarnClient) error
	}{
		{
			path:     "/lend/v1/earn/deposit",
			quantity: "amount",
			want:     "1000000",
			call: func(ctx context.Context, c *lend.EarnClient) error {
				_, err := c.Deposit(ctx, amount)
				return err
			},
		},
		{
			path:     "/lend/v1/earn/withdraw",
			quantity: "amount",
			want:     "1000000",
			call: func(ctx context.Context, c *lend.EarnClient) error {
				_, err := c.Withdraw(ctx, amount)
				return err
			},
		},
		{
			path:     "/lend/v1/earn/mint",
			quantity: "shares",
			want:     "500",
			call: func(ctx context.Context, c *lend.EarnClient) error {
				_, err := c.Mint(ctx, shares)
				return err
			},
		},
		{
			path:     "/lend/v1/earn/redeem",
			quantity: "shares",
			want:     "500",
			call: func(ctx context.Context, c *lend.EarnClient) error {
				_, err := c.Redeem(ctx, shares)
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			client, recorded := lendServer(t, map[string]string{test.path: `{"transaction":"AQ=="}`})

			if err := test.call(context.Background(), client.Lend.Earn); err != nil {
				t.Fatal(err)
			}

			requests := recorded()
			if len(requests) != 1 || requests[0].method != http.MethodPost {
				t.Fatalf("requests = %+v", requests)
			}

			body := requests[0].body
			quantityKept := string(body[test.quantity]) == `"`+test.want+`"`
			if !quantityKept || string(body["signer"]) != `"`+takerAddress+`"` {
				t.Fatalf("body = %v", body)
			}
		})
	}
}

func TestEarnInstructionsConvertForSigning(t *testing.T) {
	client, _ := lendServer(t, map[string]string{"/lend/v1/earn/deposit-instructions": instructionReply})

	request := lend.AmountRequest{Asset: usdcMint, Signer: takerAddress, Amount: "1"}
	result, err := client.Lend.Earn.DepositInstructions(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}

	converted, err := solanatx.ConvertInstruction(result.Instruction)
	if err != nil {
		t.Fatal(err)
	}

	data, err := converted.Data()
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "\x01\x02\x03" || !result.Instruction.Accounts[0].IsSigner {
		t.Fatalf("instruction = %+v", result.Instruction)
	}
}

func TestEarnReadsDecodeMixedNumberForms(t *testing.T) {
	const tokens = `[
  {"id": 1, "decimals": 6, "totalAssets": "123456789012345678901234567890",
   "asset": {"chain_id": 0, "price": 171.34}},
  {"id": 2, "asset": {"chain_id": "solana", "price": "0.9999"}}
]`
	client, recorded := lendServer(t, map[string]string{
		"/lend/v1/earn/tokens":    tokens,
		"/lend/v1/earn/positions": `[{"ownerAddress": "owner", "shares": "42", "token": {"id": 1}}]`,
		"/lend/v1/earn/earnings":  `{"address": "token", "earnings": 34483000, "slot": 1}`,
	})

	ctx := context.Background()
	earnTokens, err := client.Lend.Earn.Tokens(ctx)
	if err != nil {
		t.Fatal(err)
	}

	first := earnTokens.Tokens[0]
	second := earnTokens.Tokens[1]
	if first.Asset.ChainID != "0" || first.Asset.Price.String() != "171.34" ||
		first.TotalAssets.String() != "123456789012345678901234567890" {
		t.Fatalf("first token = %+v", first)
	}

	if second.Asset.ChainID != "solana" || second.Asset.Price.String() != "0.9999" {
		t.Fatalf("second token = %+v", second)
	}

	positions, err := client.Lend.Earn.Positions(ctx, []string{takerAddress, payerAddress})
	if err != nil {
		t.Fatal(err)
	}

	if positions.Positions[0].Shares.String() != "42" {
		t.Fatalf("positions = %+v", positions)
	}

	earnings, err := client.Lend.Earn.Earnings(ctx, takerAddress, []string{usdcMint})
	if err != nil {
		t.Fatal(err)
	}

	if len(earnings.Earnings) != 1 || earnings.Earnings[0].Earnings.String() != "34483000" {
		t.Fatalf("earnings = %+v", earnings)
	}

	requests := recorded()
	if got := requests[1].query["users"]; len(got) != 1 || got[0] != takerAddress+","+payerAddress {
		t.Fatalf("users = %v", got)
	}

	if got := requests[2].query["positions"]; len(got) != 1 || got[0] != usdcMint {
		t.Fatalf("positions = %v", got)
	}
}

func TestEarnEarningsAcceptsDocumentedArrayExample(t *testing.T) {
	client, _ := lendServer(t, map[string]string{
		"/lend/v1/earn/earnings": `[{"earnings": 1}, {"earnings": 2}]`,
	})

	earnings, err := client.Lend.Earn.Earnings(context.Background(), takerAddress, []string{usdcMint})
	if err != nil {
		t.Fatal(err)
	}

	if len(earnings.Earnings) != 2 {
		t.Fatalf("earnings = %+v", earnings)
	}
}

func TestBorrowReads(t *testing.T) {
	const vaults = `[{
  "id": 1,
  "supplyToken": {"symbol": "WSOL", "price": "68.403265633557"},
  "collateralFactor": "800",
  "liquiditySupplyData": {"expandDuration": "3600"},
  "liquidityBorrowData": {"lastUpdateTimestamp": 1735689600, "expandDuration": 7200},
  "oracleSources": [{"source": "s", "sourceType": {"chainlink": {}}}],
  "metadata": {"multiply": {"enabled": true, "reduceFactor": 300}}
}]`
	client, recorded := lendServer(t, map[string]string{
		"/lend/v1/borrow/vaults":    vaults,
		"/lend/v1/borrow/positions": `[{"id": 3518, "tick": -2147483648, "config": {"k": "v"}}]`,
	})

	ctx := context.Background()
	result, err := client.Lend.Borrow.Vaults(ctx, lend.VaultsRequest{
		Market: lend.MarketEthena,
		RPCURL: "https://rpc.example.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	vault := result.Vaults[0]
	if vault.CollateralFactor.String() != "800" ||
		vault.LiquiditySupplyData.ExpandDuration.String() != "3600" ||
		vault.LiquidityBorrowData.ExpandDuration.String() != "7200" ||
		!vault.Metadata.Multiply.Enabled ||
		string(vault.OracleSources[0].SourceType) != `{"chainlink": {}}` {
		t.Fatalf("vault = %+v", vault)
	}

	positions, err := client.Lend.Borrow.Positions(ctx, "", []string{takerAddress})
	if err != nil {
		t.Fatal(err)
	}

	if positions.Positions[0].Tick != -2147483648 || positions.Positions[0].Config["k"] != "v" {
		t.Fatalf("positions = %+v", positions)
	}

	requests := recorded()
	vaultQuery := requests[0].query
	if vaultQuery.Get("market") != "ethena" || vaultQuery.Get("rpcUrl") != "https://rpc.example.com" {
		t.Fatalf("vault query = %v", vaultQuery)
	}

	if _, ok := requests[1].query["market"]; ok {
		t.Fatal("empty market was sent instead of using the default")
	}
}

func TestBorrowOperate(t *testing.T) {
	client, recorded := lendServer(t, map[string]string{
		"/lend/v1/borrow/operate": `{"nftId": 9062, "transaction": "AQ=="}`,
		"/lend/v1/borrow/operate-instructions": `{"nftId": 9062, "instructions": [` + instructionReply +
			`], "addressLookupTableAddresses": ["3eB86bwSJanGwLo5WoMQZL7Bk7uubCRggZKshJKJVKwx"]}`,
	})

	ctx := context.Background()
	request := lend.OperateRequest{
		Market:     lend.MarketMain,
		VaultID:    1,
		PositionID: lend.NewPosition,
		Signer:     takerAddress,
		ColAmount:  "00030000000",
		DebtAmount: lend.RepayAllDebt,
	}

	operation, err := client.Lend.Borrow.Operate(ctx, request)
	if err != nil {
		t.Fatal(err)
	}

	instructions, err := client.Lend.Borrow.OperateInstructions(ctx, request)
	if err != nil {
		t.Fatal(err)
	}

	if operation.NFTID != 9062 || len(instructions.Instructions) != 1 ||
		len(instructions.AddressLookupTableAddresses) != 1 {
		t.Fatalf("operation=%+v instructions=%+v", operation, instructions)
	}

	sent := recorded()[0]
	want := map[string]string{
		"vaultId":    "1",
		"positionId": "0",
		"colAmount":  `"30000000"`,
		"debtAmount": `"` + lend.RepayAllDebt + `"`,
	}
	for key, value := range want {
		if string(sent.body[key]) != value {
			t.Errorf("%s = %s, want %s", key, sent.body[key], value)
		}
	}

	if _, ok := sent.body["positionOwner"]; ok || sent.query["market"][0] != "main" {
		t.Fatalf("body=%v query=%v", sent.body, sent.query)
	}
}

func TestLendBuildersRetryTransientFailures(t *testing.T) {
	server, calls := flakyServer(t, http.StatusServiceUnavailable, `{"nftId": 1, "transaction": "AQ=="}`)

	client, err := jupiter.NewClient(jupiter.WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	request := lend.OperateRequest{VaultID: 1, Signer: takerAddress, ColAmount: "1", DebtAmount: "0"}
	result, err := client.Lend.Borrow.Operate(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}

	if calls.Load() != recoveredAttempts || result.Metadata.Attempts != recoveredAttempts {
		t.Fatalf("calls=%d attempts=%d", calls.Load(), result.Metadata.Attempts)
	}
}

func TestOperateAcceptsSignedRangeBoundaries(t *testing.T) {
	for _, value := range []string{lend.WithdrawAllCollateral, maxSignedAmount, "-1"} {
		t.Run(value, func(t *testing.T) {
			client, _ := lendServer(t, map[string]string{"/lend/v1/borrow/operate": `{"nftId": 1}`})

			request := lend.OperateRequest{
				VaultID:    1,
				PositionID: 7,
				Signer:     takerAddress,
				ColAmount:  value,
				DebtAmount: "0",
			}
			if _, err := client.Lend.Borrow.Operate(context.Background(), request); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestInvalidLendRequestsDoNotReachTransport(t *testing.T) {
	client, err := jupiter.NewClient(jupiter.WithHTTPClient(unreachableHTTPClient{t: t}))
	if err != nil {
		t.Fatal(err)
	}

	earn := client.Lend.Earn
	borrow := client.Lend.Borrow
	ctx := context.Background()

	operate := func(configure func(*lend.OperateRequest)) func() error {
		return func() error {
			request := lend.OperateRequest{VaultID: 1, Signer: takerAddress, ColAmount: "1", DebtAmount: "0"}
			configure(&request)
			_, err := borrow.Operate(ctx, request)

			return err
		}
	}

	tests := []struct {
		name string
		call func() error
		want error
	}{
		{
			name: "deposit malformed asset",
			call: func() error {
				request := lend.AmountRequest{Asset: "USDC", Signer: takerAddress, Amount: "1"}
				_, err := earn.Deposit(ctx, request)
				return err
			},
			want: lend.ErrInvalidAddress,
		},
		{
			name: "withdraw zero amount",
			call: func() error {
				request := lend.AmountRequest{Asset: usdcMint, Signer: takerAddress, Amount: "0"}
				_, err := earn.Withdraw(ctx, request)
				return err
			},
		},
		{
			name: "mint malformed signer",
			call: func() error {
				request := lend.SharesRequest{Asset: usdcMint, Signer: "me", Shares: "1"}
				_, err := earn.MintInstructions(ctx, request)
				return err
			},
			want: lend.ErrInvalidAddress,
		},
		{
			name: "redeem fractional shares",
			call: func() error {
				request := lend.SharesRequest{Asset: usdcMint, Signer: takerAddress, Shares: "1.5"}
				_, err := earn.Redeem(ctx, request)
				return err
			},
		},
		{
			name: "positions without users",
			call: func() error {
				_, err := earn.Positions(ctx, nil)
				return err
			},
		},
		{
			name: "earnings without positions",
			call: func() error {
				_, err := earn.Earnings(ctx, takerAddress, nil)
				return err
			},
		},
		{
			name: "borrow positions malformed user",
			call: func() error {
				_, err := borrow.Positions(ctx, "", []string{"user"})
				return err
			},
			want: lend.ErrInvalidAddress,
		},
		{
			name: "market with separator",
			call: func() error {
				_, err := borrow.Vaults(ctx, lend.VaultsRequest{Market: "main,ethena"})
				return err
			},
		},
		{
			name: "relative RPC URL",
			call: func() error {
				_, err := borrow.Vaults(ctx, lend.VaultsRequest{RPCURL: "rpc.example.com"})
				return err
			},
		},
		{name: "no-op operation", call: operate(func(r *lend.OperateRequest) { r.ColAmount = "0" })},
		{name: "explicit plus sign", call: operate(func(r *lend.OperateRequest) { r.ColAmount = "+1" })},
		{name: "fractional amount", call: operate(func(r *lend.OperateRequest) { r.DebtAmount = "1.5" })},
		{name: "whitespace amount", call: operate(func(r *lend.OperateRequest) { r.ColAmount = " 1" })},
		{name: "empty amount", call: operate(func(r *lend.OperateRequest) { r.DebtAmount = "" })},
		{
			name: "below signed range",
			call: operate(func(r *lend.OperateRequest) {
				r.ColAmount = "-170141183460469231731687303715884105729"
			}),
		},
		{
			name: "above signed range",
			call: operate(func(r *lend.OperateRequest) {
				r.DebtAmount = "170141183460469231731687303715884105728"
			}),
		},
		{
			name: "malformed signer",
			call: operate(func(r *lend.OperateRequest) { r.Signer = "signer" }),
			want: lend.ErrInvalidAddress,
		},
		{
			name: "malformed position owner",
			call: operate(func(r *lend.OperateRequest) { r.PositionOwner = "owner" }),
			want: lend.ErrInvalidAddress,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.call()
			if err == nil {
				t.Fatal("invalid request accepted")
			}

			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}

func TestLendFailuresKeepResponseMetadata(t *testing.T) {
	client, _ := lendServer(t, map[string]string{})

	result, err := client.Lend.Earn.Tokens(context.Background())
	if err == nil || result == nil || result.Metadata.StatusCode != http.StatusNotFound {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
