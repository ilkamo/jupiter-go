package jupiter_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/prediction"
)

const (
	positionAddress = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	orderAddress    = "ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL"
	eventsReply     = `{
  "data": [{
    "eventId": "POLY-45915",
    "isActive": true,
    "metadata": {"title": "Brazil Presidential Election"},
    "liveScore": null,
    "markets": [
      {"marketId": "POLY-1", "resolveAt": 1735689600, "team": null,
       "pricing": {"buyYesPriceUsd": 850000, "sellNoPriceUsd": null}},
      {"marketId": "POLY-2", "resolveAt": "2026-01-01T00:00:00Z"}
    ]
  }],
  "pagination": {"start": 0, "end": 1, "total": 40, "hasNext": true}
}`
	buildReply = `{
  "transaction": "AQ==",
  "txMeta": {"blockhash": "hash", "lastValidBlockHeight": 42},
  "requiredSigners": ["` + takerAddress + `"],
  "execution": {"endpoint": "/execute", "context": {"type":"create_order","nested":{"k":1}}},
  "order": {"orderPubkey": "order", "contractsMicro": "5000000", "slippageBps": 25}
}`
)

const (
	marketReply      = `{"marketId": "POLY-1", "pricing": {"buyYesPriceUsd": 850000}}`
	orderbookReply   = `{"yes": [[85, 120]], "no": [], "yes_dollars": [], "no_dollars": []}`
	forecastReply    = `{"forecast_history": [{"numerical_forecast": 0.62}]}`
	claimReply       = `{"transaction": "AQ==", "txMeta": {}, "position": {"isYes": true}}`
	ordersReply      = `{"data": [{"pubkey": "o", "createdAt": 1}], "pagination": {}}`
	orderStatusReply = `{"status": "filled", "history": [{"eventType": "order_filled"}]}`
	historyReply     = `{"data": [{"id": 7, "eventType": "payout_claimed"}], "pagination": {}}`
	tradesReply      = `{"data": [{"id": "t", "timestamp": 1735689600.5, "side": "yes"}]}`
	leaderboardReply = `{"data": [{"ownerPubkey": "o", "predictionsCount": 12, "periodStart": null}]}`
	vaultInfoReply   = `{"pubkey": "v", "data": {"authority": "a"}, "vaultBalance": "100"}`
)

func predictionPath(suffix string) string {
	return "/prediction/v1" + suffix
}

func uint32Pointer(value uint32) *uint32 {
	return &value
}

func TestPredictionDiscovery(t *testing.T) {
	escapedEvent := "/prediction/v1/events/POLY%2F1"
	client, recorded := pathServer(t, map[string]string{
		predictionPath("/events"):                           eventsReply,
		predictionPath("/events/search"):                    eventsReply,
		escapedEvent:                                        `{"eventId": "POLY/1"}`,
		predictionPath("/events/scores"):                    `{"data": [{"eventId": "e", "gameId": "g"}]}`,
		predictionPath("/events/e/score"):                   `null`,
		predictionPath("/markets/POLY-1"):                   marketReply,
		predictionPath("/orderbook/POLY-1"):                 orderbookReply,
		predictionPath("/orderbook/POLY-2"):                 `null`,
		predictionPath("/trading-status"):                   `{"trading_active": true}`,
		predictionPath("/forecast"):                         forecastReply,
		predictionPath("/events/suggested/" + takerAddress): eventsReply,
	})

	ctx := context.Background()
	events, err := client.Prediction.Events(ctx, prediction.EventsRequest{
		Provider:       prediction.ProviderPolymarket,
		IncludeMarkets: true,
		Page:           prediction.Page{Start: uint32Pointer(0), End: uint32Pointer(1)},
		Category:       prediction.CategoryPolitics,
		Filter:         prediction.FilterTrending,
		Tags:           "soccer",
	})
	if err != nil {
		t.Fatal(err)
	}

	event := events.Events[0]
	first, second := event.Markets[0], event.Markets[1]
	if event.EventMetadata.Title != "Brazil Presidential Election" || !events.Pagination.HasNext {
		t.Fatalf("event = %+v", event)
	}

	if first.ResolveAt != "1735689600" || second.ResolveAt != "2026-01-01T00:00:00Z" {
		t.Fatalf("resolveAt = %q, %q", first.ResolveAt, second.ResolveAt)
	}

	if first.Pricing.BuyYesPriceUSD.String() != "850000" || first.Pricing.SellNoPriceUSD != "" {
		t.Fatalf("pricing = %+v", first.Pricing)
	}

	query := recorded()[0].query
	wantQuery := map[string]string{
		"provider":       "polymarket",
		"includeMarkets": "true",
		"start":          "0",
		"end":            "1",
		"category":       "politics",
		"filter":         "trending",
		"tags":           "soccer",
	}
	for key, value := range wantQuery {
		if query.Get(key) != value {
			t.Errorf("%s = %q, want %q", key, query.Get(key), value)
		}
	}

	limit := uint8(prediction.MaxSearchLimit)
	search := prediction.SearchRequest{Query: "election", Limit: &limit}
	if _, err := client.Prediction.SearchEvents(ctx, search); err != nil {
		t.Fatal(err)
	}

	single, err := client.Prediction.Event(ctx, prediction.EventRequest{EventID: "POLY/1"})
	if err != nil || single.EventID != "POLY/1" {
		t.Fatalf("event=%+v err=%v", single, err)
	}

	scores, err := client.Prediction.EventScores(ctx, []string{"e", "f"})
	if err != nil || scores.Scores[0].GameID != "g" {
		t.Fatalf("scores=%+v err=%v", scores, err)
	}

	score, err := client.Prediction.EventScore(ctx, "e")
	if err != nil || score.Score != nil {
		t.Fatalf("score=%+v err=%v", score, err)
	}

	market, err := client.Prediction.Market(ctx, "POLY-1")
	if err != nil || market.Pricing.BuyYesPriceUSD.String() != "850000" {
		t.Fatalf("market=%+v err=%v", market, err)
	}

	book, err := client.Prediction.Orderbook(ctx, "POLY-1")
	if err != nil || book.Orderbook.Yes[0][1].String() != "120" {
		t.Fatalf("orderbook=%+v err=%v", book, err)
	}

	empty, err := client.Prediction.Orderbook(ctx, "POLY-2")
	if err != nil || empty.Orderbook != nil {
		t.Fatalf("null orderbook=%+v err=%v", empty, err)
	}

	status, err := client.Prediction.TradingStatus(ctx)
	if err != nil || !status.TradingActive {
		t.Fatalf("status=%+v err=%v", status, err)
	}

	forecast, err := client.Prediction.Forecast(ctx, "POLY-1")
	if err != nil || forecast.History[0].NumericalForecast.String() != "0.62" {
		t.Fatalf("forecast=%+v err=%v", forecast, err)
	}

	if _, err := client.Prediction.SuggestedEvents(ctx, takerAddress, ""); err != nil {
		t.Fatal(err)
	}

	requests := recorded()
	paths := make(map[string]pathRequest, len(requests))
	for _, request := range requests {
		paths[request.path] = request
	}

	if paths["/prediction/v1/events/search"].query.Get("limit") != "20" {
		t.Error("search limit not sent")
	}

	if paths["/prediction/v1/events/scores"].query.Get("eventIds") != "e,f" {
		t.Error("event IDs not joined")
	}

	if paths["/prediction/v1/forecast"].query.Get("marketId") != "POLY-1" {
		t.Error("forecast market not sent")
	}

	if _, ok := paths[escapedEvent]; !ok {
		t.Error("event ID was not path-escaped")
	}
}

func TestPredictionOrderBuilds(t *testing.T) {
	claimPath := "/prediction/v1/positions/" + positionAddress + "/claim"
	client, recorded := pathServer(t, map[string]string{
		predictionPath("/orders"):                       buildReply,
		predictionPath("/positions/" + positionAddress): buildReply,
		predictionPath("/positions"):                    `{"data": [` + buildReply + `, ` + claimReply + `]}`,
		claimPath:                                       claimReply,
	})

	ctx := context.Background()
	buy, err := client.Prediction.CreateOrder(ctx, prediction.OrderRequest{
		Owner:         takerAddress,
		MarketID:      "POLY-1",
		IsYes:         true,
		IsBuy:         true,
		DepositAmount: "0005000000",
		DepositMint:   prediction.MintJupUSD,
	})
	if err != nil {
		t.Fatal(err)
	}

	if buy.Order.ContractsMicro != "5000000" || buy.TxMeta.LastValidBlockHeight != 42 ||
		string(buy.Execution.Context) != `{"type":"create_order","nested":{"k":1}}` {
		t.Fatalf("build = %+v", buy)
	}

	buyBody := recorded()[0].body
	if string(buyBody["depositAmount"]) != `"5000000"` || string(buyBody["isBuy"]) != "true" {
		t.Fatalf("buy body = %v", buyBody)
	}

	for _, absent := range []string{"positionPubkey", "contracts", "contractsMicro", "contractsDecimal"} {
		if _, ok := buyBody[absent]; ok {
			t.Errorf("buy sent %s", absent)
		}
	}

	_, err = client.Prediction.CreateOrder(ctx, prediction.OrderRequest{
		Owner:            takerAddress,
		PositionPubkey:   positionAddress,
		ContractsDecimal: "5.777773",
	})
	if err != nil {
		t.Fatal(err)
	}

	sellBody := recorded()[1].body
	if string(sellBody["contractsDecimal"]) != `"5.777773"` || string(sellBody["isBuy"]) != "false" {
		t.Fatalf("sell body = %v", sellBody)
	}

	if _, ok := sellBody["depositAmount"]; ok {
		t.Error("sell sent a deposit")
	}

	if _, err := client.Prediction.ClosePosition(ctx, takerAddress, positionAddress); err != nil {
		t.Fatal(err)
	}

	all, err := client.Prediction.CloseAllPositions(ctx, takerAddress, 150)
	if err != nil {
		t.Fatal(err)
	}

	if len(all.Builds) != 2 || all.Builds[0].Order == nil || all.Builds[1].Position == nil {
		t.Fatalf("close all = %+v", all)
	}

	claim, err := client.Prediction.ClaimPosition(ctx, takerAddress, positionAddress)
	if err != nil || claim.Position == nil || !claim.Position.IsYes {
		t.Fatalf("claim=%+v err=%v", claim, err)
	}

	requests := recorded()
	closeOne, closeAll := requests[2], requests[3]
	if closeOne.method != http.MethodDelete || string(closeOne.body["ownerPubkey"]) != `"`+takerAddress+`"` {
		t.Fatalf("close position = %+v", closeOne)
	}

	if closeAll.method != http.MethodDelete || string(closeAll.body["minSellPriceSlippageBps"]) != "150" {
		t.Fatalf("close all = %+v", closeAll)
	}
}

func TestPredictionExecute(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		succeeds  bool
		unknown   bool
		execution bool
	}{
		{
			name:     "success",
			status:   http.StatusOK,
			body:     `{"status": "Success", "signature": "sig"}`,
			succeeds: true,
		},
		{
			name:      "failed status",
			status:    http.StatusOK,
			body:      `{"status": "Failed", "error": "slippage"}`,
			execution: true,
		},
		{name: "missing status", status: http.StatusOK, body: `{}`, execution: true},
		{
			name:    "server error",
			status:  http.StatusInternalServerError,
			body:    `{"message": "boom"}`,
			unknown: true,
		},
		{name: "rejected", status: http.StatusBadRequest, body: `{"type": "invalid_request_error"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, calls := statusServer(t, test.status, test.body)

			client, err := jupiter.NewClient(jupiter.WithBaseURL(server))
			if err != nil {
				t.Fatal(err)
			}

			request := prediction.ExecuteRequest{
				SignedTransaction: "AQ==",
				Context:           []byte(`{"type":"create_order","nested":{"k":1}}`),
				RequestID:         "client-1",
			}
			result, err := client.Prediction.Execute(context.Background(), request)

			if calls() != singleAttempt {
				t.Fatalf("execute attempts = %d", calls())
			}

			if test.succeeds {
				if err != nil || result.Signature != "sig" {
					t.Fatalf("result=%+v err=%v", result, err)
				}

				return
			}

			if err == nil {
				t.Fatal("expected failure")
			}

			if errors.Is(err, prediction.ErrOutcomeUnknown) != test.unknown {
				t.Fatalf("unexpected outcome classification: %v", err)
			}

			_, isExecution := errors.AsType[*prediction.ExecutionError](err)
			if isExecution != test.execution {
				t.Fatalf("unexpected execution error: %v", err)
			}
		})
	}
}

func TestPredictionAccountReads(t *testing.T) {
	position := `{"pubkey": "p", "valueUsd": null, "pnlUsdPercent": null, "claimableAt": null,
"settlementDate": 1735689600, "realizedPnlUsd": -1500000, "contractsMicro": "5777773"}`
	pnlPath := predictionPath("/profiles/" + takerAddress + "/pnl-history")
	client, recorded := pathServer(t, map[string]string{
		predictionPath("/orders"):                        ordersReply,
		predictionPath("/orders/" + orderAddress):        `{"pubkey": "o", "status": "filled"}`,
		predictionPath("/orders/status/" + orderAddress): orderStatusReply,
		predictionPath("/positions"):                     `{"data": [` + position + `], "pagination": {}}`,
		predictionPath("/positions/" + positionAddress):  position,
		predictionPath("/history"):                       historyReply,
		predictionPath("/profiles/" + takerAddress):      `{"ownerPubkey": "o", "predictionsCount": "12"}`,
		pnlPath:                         `{"history": [{"timestamp": 1, "realizedPnlUsd": "-5"}]}`,
		predictionPath("/trades"):       tradesReply,
		predictionPath("/leaderboards"): leaderboardReply,
		predictionPath("/vault-info"):   vaultInfoReply,
	})

	ctx := context.Background()
	isYes := false
	eventID := uint64(prediction.MinHistoryEventID)
	count := uint16(prediction.MaxPnLPointCount)
	limit := uint8(prediction.MaxLeaderboardLimit)

	steps := []func() error{
		func() error {
			_, err := client.Prediction.Orders(ctx, takerAddress, prediction.Page{End: uint32Pointer(10)})
			return err
		},
		func() error {
			_, err := client.Prediction.Order(ctx, orderAddress)
			return err
		},
		func() error {
			status, err := client.Prediction.OrderStatus(ctx, orderAddress)
			if err == nil && status.History[0].EventType != "order_filled" {
				t.Errorf("status = %+v", status)
			}
			return err
		},
		func() error {
			positions, err := client.Prediction.Positions(ctx, prediction.PositionsRequest{
				Owner:    takerAddress,
				MarketID: "POLY-1",
				IsYes:    &isYes,
			})
			if err == nil {
				p := positions.Positions[0]
				if p.ValueUSD != "" || p.ClaimableAt != nil || *p.SettlementDate != 1735689600 ||
					p.RealizedPnLUSD.String() != "-1500000" || p.ContractsMicro != "5777773" {
					t.Errorf("position = %+v", p)
				}
			}
			return err
		},
		func() error {
			_, err := client.Prediction.Position(ctx, positionAddress)
			return err
		},
		func() error {
			request := prediction.HistoryRequest{Owner: takerAddress, EventID: &eventID}
			_, err := client.Prediction.History(ctx, request)
			return err
		},
		func() error {
			_, err := client.Prediction.Profile(ctx, takerAddress)
			return err
		},
		func() error {
			_, err := client.Prediction.PnLHistory(ctx, prediction.PnLHistoryRequest{
				Owner:    takerAddress,
				Interval: prediction.PnLInterval1m,
				Count:    &count,
			})
			return err
		},
		func() error {
			trades, err := client.Prediction.Trades(ctx)
			if err == nil && trades.Trades[0].Timestamp.String() != "1735689600.5" {
				t.Errorf("trades = %+v", trades)
			}
			return err
		},
		func() error {
			_, err := client.Prediction.Leaderboards(ctx, prediction.LeaderboardRequest{
				Period: prediction.LeaderboardWeekly,
				Metric: prediction.MetricWinRate,
				Limit:  &limit,
			})
			return err
		},
		func() error {
			info, err := client.Prediction.VaultInfo(ctx)
			if err == nil && info.Data["authority"] != "a" {
				t.Errorf("vault = %+v", info)
			}
			return err
		},
	}

	for i, step := range steps {
		if err := step(); err != nil {
			t.Fatalf("step %d: %v", i, err)
		}
	}

	queries := make(map[string]string)
	for _, request := range recorded() {
		queries[request.path] = request.query.Encode()
	}

	positionsQuery := "isYes=false&marketId=POLY-1&ownerPubkey=" + takerAddress
	want := map[string]string{
		predictionPath("/orders"):       "end=10&ownerPubkey=" + takerAddress,
		predictionPath("/positions"):    positionsQuery,
		predictionPath("/history"):      "id=1&ownerPubkey=" + takerAddress,
		predictionPath("/leaderboards"): "limit=100&metric=win_rate&period=weekly",
		pnlPath:                         "count=1000&interval=1m",
	}
	for path, query := range want {
		if queries[path] != query {
			t.Errorf("%s query = %q, want %q", path, queries[path], query)
		}
	}
}

func TestInvalidPredictionRequestsDoNotReachTransport(t *testing.T) {
	client, err := jupiter.NewClient(jupiter.WithHTTPClient(unreachableHTTPClient{t: t}))
	if err != nil {
		t.Fatal(err)
	}

	p := client.Prediction
	ctx := context.Background()
	overLimit := uint8(prediction.MaxSearchLimit + 1)
	zeroCount := uint16(prediction.MinPnLPointCount - 1)
	zeroEvent := uint64(prediction.MinHistoryEventID - 1)

	order := func(configure func(*prediction.OrderRequest)) func() error {
		return func() error {
			request := prediction.OrderRequest{
				Owner:         takerAddress,
				MarketID:      "POLY-1",
				IsYes:         true,
				IsBuy:         true,
				DepositAmount: "5000000",
			}
			configure(&request)
			_, err := p.CreateOrder(ctx, request)

			return err
		}
	}

	sell := func(configure func(*prediction.OrderRequest)) func() error {
		return order(func(r *prediction.OrderRequest) {
			r.IsBuy = false
			r.MarketID = ""
			r.DepositAmount = ""
			r.PositionPubkey = positionAddress
			r.ContractsMicro = "1000000"
			configure(r)
		})
	}

	tests := []struct {
		name string
		call func() error
		want error
	}{
		{name: "buy without market", call: order(func(r *prediction.OrderRequest) { r.MarketID = "" })},
		{name: "buy zero deposit", call: order(func(r *prediction.OrderRequest) { r.DepositAmount = "0" })},
		{name: "buy with contracts", call: order(func(r *prediction.OrderRequest) { r.Contracts = "1" })},
		{
			name: "buy malformed mint",
			call: order(func(r *prediction.OrderRequest) { r.DepositMint = "USDC" }),
			want: prediction.ErrInvalidAddress,
		},
		{
			name: "malformed owner",
			call: order(func(r *prediction.OrderRequest) { r.Owner = "me" }),
			want: prediction.ErrInvalidAddress,
		},
		{
			name: "sell without quantity",
			call: sell(func(r *prediction.OrderRequest) { r.ContractsMicro = "" }),
		},
		{
			name: "sell with two quantities",
			call: sell(func(r *prediction.OrderRequest) { r.Contracts = "1" }),
		},
		{name: "sell with deposit", call: sell(func(r *prediction.OrderRequest) { r.DepositAmount = "1" })},
		{
			name: "sell without position",
			call: sell(func(r *prediction.OrderRequest) { r.PositionPubkey = "" }),
			want: prediction.ErrInvalidAddress,
		},
		{
			name: "sell fractional decimal",
			call: sell(func(r *prediction.OrderRequest) {
				r.ContractsMicro = ""
				r.ContractsDecimal = "1/2"
			}),
		},
		{
			name: "sell negative decimal",
			call: sell(func(r *prediction.OrderRequest) {
				r.ContractsMicro = ""
				r.ContractsDecimal = "-1"
			}),
		},
		{
			name: "close all slippage above maximum",
			call: func() error {
				_, err := p.CloseAllPositions(ctx, takerAddress, prediction.BasisPointsPerWhole+1)
				return err
			},
		},
		{
			name: "claim malformed position",
			call: func() error {
				_, err := p.ClaimPosition(ctx, takerAddress, "position")
				return err
			},
			want: prediction.ErrInvalidAddress,
		},
		{
			name: "execute unsigned",
			call: func() error {
				_, err := p.Execute(ctx, prediction.ExecuteRequest{SignedTransaction: "%%"})
				return err
			},
		},
		{
			name: "execute invalid context",
			call: func() error {
				request := prediction.ExecuteRequest{SignedTransaction: "AQ==", Context: []byte("{")}
				_, err := p.Execute(ctx, request)
				return err
			},
		},
		{
			name: "empty search",
			call: func() error {
				_, err := p.SearchEvents(ctx, prediction.SearchRequest{})
				return err
			},
		},
		{
			name: "search query too long",
			call: func() error {
				query := strings.Repeat("é", prediction.MaxSearchQueryRune+1)
				_, err := p.SearchEvents(ctx, prediction.SearchRequest{Query: query})
				return err
			},
		},
		{
			name: "search limit above maximum",
			call: func() error {
				_, err := p.SearchEvents(ctx, prediction.SearchRequest{Query: "x", Limit: &overLimit})
				return err
			},
		},
		{
			name: "inverted page",
			call: func() error {
				page := prediction.Page{Start: uint32Pointer(5), End: uint32Pointer(4)}
				_, err := p.Events(ctx, prediction.EventsRequest{Page: page})
				return err
			},
		},
		{
			name: "filter with separator",
			call: func() error {
				_, err := p.Events(ctx, prediction.EventsRequest{Category: "crypto,sports"})
				return err
			},
		},
		{
			name: "missing market ID",
			call: func() error {
				_, err := p.Market(ctx, "")
				return err
			},
		},
		{
			name: "scores without events",
			call: func() error {
				_, err := p.EventScores(ctx, nil)
				return err
			},
		},
		{
			name: "orders malformed owner",
			call: func() error {
				_, err := p.Orders(ctx, "owner", prediction.Page{})
				return err
			},
			want: prediction.ErrInvalidAddress,
		},
		{
			name: "history event zero",
			call: func() error {
				_, err := p.History(ctx, prediction.HistoryRequest{EventID: &zeroEvent})
				return err
			},
		},
		{
			name: "pnl count zero",
			call: func() error {
				request := prediction.PnLHistoryRequest{Owner: takerAddress, Count: &zeroCount}
				_, err := p.PnLHistory(ctx, request)
				return err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.call()
			if err == nil {
				t.Fatal("invalid request accepted")
			}

			if errors.Is(err, prediction.ErrOutcomeUnknown) {
				t.Fatalf("validation failure reported as unknown outcome: %v", err)
			}

			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}
