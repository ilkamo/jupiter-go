package jupiter_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"testing"
	"time"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/trigger"
)

const (
	decimalBase      = 10
	orderValidity    = 7 * 24 * time.Hour
	dcaInterval      = time.Hour
	signedDeposit    = "AQ=="
	orderID          = "order/1"
	escapedOrderID   = "order%2F1"
	createdOrderBody = `{"id":"order-1","txSignature":"sig","depositConfirmed":true}`
	submissionBody   = `{"id":"order-1","txSignature":"withdrawal-sig"}`
	cancellationBody = `{"id":"order-1","transaction":"AQ==","requestId":"cancel-1"}`
	dcaCancelBody    = `{
  "id": "dca-1",
  "transaction": "AQ==",
  "requestId": "cancel-2",
  "roundsRemaining": 2,
  "refundAmount": "20000000"
}`
	serverFailureBody = `{"error":"internal"}`
)

func priceSession(t *testing.T) (*fakeTrigger, *trigger.Session) {
	t.Helper()

	fake, client := newFakeTrigger(t)
	session := login(t, client, trigger.SessionOptions{})

	return fake, session
}

func singleOrder() trigger.PriceOrderRequest {
	return trigger.PriceOrderRequest{
		OrderType:                trigger.OrderSubTypeSingle,
		DepositRequestID:         "deposit-1",
		DepositSignedTransaction: signedDeposit,
		InputMint:                wrappedSOLMint,
		InputAmount:              "0001000000000",
		OutputMint:               usdcMint,
		TriggerMint:              wrappedSOLMint,
		TriggerCondition:         trigger.TriggerAbove,
		TriggerPriceUSD:          "200.123456789012345678",
		ExpiresAt:                time.Now().Add(orderValidity),
	}
}

func ocoOrder() trigger.PriceOrderRequest {
	r := singleOrder()
	r.OrderType = trigger.OrderSubTypeOCO
	r.TriggerCondition = ""
	r.TriggerPriceUSD = ""
	r.TakeProfitPriceUSD = "250"
	r.StopLossPriceUSD = "150"

	return r
}

func decodeBody(t *testing.T, raw []byte) map[string]json.RawMessage {
	t.Helper()

	var body map[string]json.RawMessage
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}

	return body
}

func lastRequest(t *testing.T, fake *fakeTrigger) fakeRequest {
	t.Helper()

	requests := fake.recorded()
	if len(requests) == 0 {
		t.Fatal("no requests recorded")
	}

	return requests[len(requests)-1]
}

func TestCreatePriceOrderEncodesExactValues(t *testing.T) {
	fake, session := priceSession(t)
	fake.route(http.MethodPost, "/orders/price", http.StatusOK, createdOrderBody)

	request := singleOrder()
	order, err := session.CreatePriceOrder(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}

	if order.ID != "order-1" || !order.DepositConfirmed {
		t.Fatalf("order = %+v", order)
	}

	body := decodeBody(t, lastRequest(t, fake).body)
	want := map[string]string{
		"orderType":       `"single"`,
		"userPubkey":      `"` + triggerWallet() + `"`,
		"inputAmount":     `"1000000000"`,
		"triggerPriceUsd": `200.123456789012345678`,
		"expiresAt":       strconv.FormatInt(request.ExpiresAt.UnixMilli(), decimalBase),
	}
	for key, value := range want {
		if string(body[key]) != value {
			t.Errorf("%s = %s, want %s", key, body[key], value)
		}
	}

	for _, absent := range []string{"trailingBps", "tpPriceUsd", "slPriceUsd", "slippageBps"} {
		if _, ok := body[absent]; ok {
			t.Errorf("unset field %s was sent", absent)
		}
	}
}

func TestCreatePriceOrderVariants(t *testing.T) {
	minTrailing := uint16(trigger.MinTrailingBps)
	maxSlippage := uint16(trigger.BasisPointsPerWhole)

	tests := []struct {
		name    string
		request trigger.PriceOrderRequest
		present []string
		absent  []string
	}{
		{
			name: "trailing sell-below",
			request: func() trigger.PriceOrderRequest {
				r := singleOrder()
				r.TriggerCondition = trigger.TriggerBelow
				r.TriggerPriceUSD = ""
				r.TrailingBps = &minTrailing
				r.SlippageBps = &maxSlippage
				return r
			}(),
			present: []string{"trailingBps", "slippageBps", "triggerCondition"},
			absent:  []string{"triggerPriceUsd"},
		},
		{
			name:    "oco",
			request: ocoOrder(),
			present: []string{"tpPriceUsd", "slPriceUsd"},
			absent:  []string{"triggerCondition", "triggerPriceUsd"},
		},
		{
			name: "otoco",
			request: func() trigger.PriceOrderRequest {
				r := ocoOrder()
				r.OrderType = trigger.OrderSubTypeOTOCO
				r.TriggerCondition = trigger.TriggerBelow
				r.TriggerPriceUSD = "180"
				return r
			}(),
			present: []string{"tpPriceUsd", "slPriceUsd", "triggerCondition", "triggerPriceUsd"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fake, session := priceSession(t)
			fake.route(http.MethodPost, "/orders/price", http.StatusOK, createdOrderBody)

			if _, err := session.CreatePriceOrder(context.Background(), test.request); err != nil {
				t.Fatal(err)
			}

			body := decodeBody(t, lastRequest(t, fake).body)
			for _, key := range test.present {
				if _, ok := body[key]; !ok {
					t.Errorf("%s missing", key)
				}
			}

			for _, key := range test.absent {
				if _, ok := body[key]; ok {
					t.Errorf("%s sent", key)
				}
			}
		})
	}
}

func TestTriggerSubmissionOutcomes(t *testing.T) {
	confirm := trigger.ConfirmCancelRequest{SignedTransaction: signedDeposit, CancelRequestID: "cancel-1"}

	submissions := []struct {
		name   string
		method string
		path   string
		call   func(context.Context, *trigger.Session) error
	}{
		{
			name:   "create price order",
			method: http.MethodPost,
			path:   "/orders/price",
			call: func(ctx context.Context, s *trigger.Session) error {
				_, err := s.CreatePriceOrder(ctx, singleOrder())
				return err
			},
		},
		{
			name:   "confirm price cancellation",
			method: http.MethodPost,
			path:   "/orders/price/confirm-cancel/order-1",
			call: func(ctx context.Context, s *trigger.Session) error {
				_, err := s.ConfirmCancelPriceOrder(ctx, "order-1", confirm)
				return err
			},
		},
		{
			name:   "create DCA order",
			method: http.MethodPost,
			path:   "/orders/dca",
			call: func(ctx context.Context, s *trigger.Session) error {
				_, err := s.CreateDCAOrder(ctx, dcaOrder())
				return err
			},
		},
		{
			name:   "confirm DCA cancellation",
			method: http.MethodPost,
			path:   "/orders/dca/confirm-cancel/dca-1",
			call: func(ctx context.Context, s *trigger.Session) error {
				_, err := s.ConfirmCancelDCAOrder(ctx, "dca-1", confirm)
				return err
			},
		},
	}

	statuses := []struct {
		status  int
		unknown bool
	}{
		{status: http.StatusInternalServerError, unknown: true},
		{status: http.StatusServiceUnavailable, unknown: true},
		{status: http.StatusBadRequest},
		{status: http.StatusTooManyRequests},
	}

	for _, submission := range submissions {
		for _, status := range statuses {
			t.Run(submission.name+" "+http.StatusText(status.status), func(t *testing.T) {
				fake, session := priceSession(t)
				fake.route(submission.method, submission.path, status.status, serverFailureBody)
				before := len(fake.recorded())

				err := submission.call(context.Background(), session)
				if err == nil {
					t.Fatal("expected failure")
				}

				if errors.Is(err, trigger.ErrOutcomeUnknown) != status.unknown {
					t.Fatalf("unexpected outcome classification: %v", err)
				}

				var apiError *jupiter.APIError
				if !errors.As(err, &apiError) || apiError.StatusCode != status.status {
					t.Fatalf("HTTP error lost: %v", err)
				}

				if attempts := len(fake.recorded()) - before; attempts != singleAttempt {
					t.Fatalf("submission retried: %d attempts", attempts)
				}
			})
		}
	}
}

func TestUpdatePriceOrderSendsOnlyChanges(t *testing.T) {
	fake, session := priceSession(t)
	fake.route(http.MethodPatch, "/orders/price/"+escapedOrderID, http.StatusOK, `{"id":"order/1"}`)

	update := trigger.PriceOrderUpdate{OrderType: trigger.OrderSubTypeSingle, TriggerPriceUSD: "210"}
	id, err := session.UpdatePriceOrder(context.Background(), orderID, update)
	if err != nil {
		t.Fatal(err)
	}

	body := decodeBody(t, lastRequest(t, fake).body)
	if id != orderID || len(body) != 2 || string(body["triggerPriceUsd"]) != "210" {
		t.Fatalf("id=%q body=%v", id, body)
	}
}

func TestPriceOrderCancellation(t *testing.T) {
	fake, session := priceSession(t)
	fake.route(http.MethodPost, "/orders/price/cancel/"+escapedOrderID, http.StatusOK, cancellationBody)
	fake.route(http.MethodPost, "/orders/price/confirm-cancel/"+escapedOrderID, http.StatusOK, submissionBody)

	ctx := context.Background()
	cancellation, err := session.CancelPriceOrder(ctx, orderID)
	if err != nil {
		t.Fatal(err)
	}

	confirm := trigger.ConfirmCancelRequest{
		SignedTransaction: cancellation.Transaction,
		CancelRequestID:   cancellation.RequestID,
	}
	submission, err := session.ConfirmCancelPriceOrder(ctx, orderID, confirm)
	if err != nil {
		t.Fatal(err)
	}

	body := stringFields(t, lastRequest(t, fake).body)
	if body["cancelRequestId"] != "cancel-1" || submission.TxSignature != "withdrawal-sig" {
		t.Fatalf("body=%v submission=%+v", body, submission)
	}
}

func dcaOrder() trigger.DCAOrderRequest {
	return trigger.DCAOrderRequest{
		DepositRequestID:         "deposit-2",
		DepositSignedTransaction: signedDeposit,
		InputMint:                usdcMint,
		OutputMint:               wrappedSOLMint,
		InputAmount:              "20000000",
		OrderCount:               trigger.MinDCAOrderCount,
		Interval:                 dcaInterval,
	}
}

func TestDCAOrderLifecycle(t *testing.T) {
	fake, session := priceSession(t)
	fake.route(http.MethodPost, "/orders/dca", http.StatusOK, `{"id":"dca-1","txSignature":"deposit-sig"}`)
	fake.route(http.MethodPost, "/orders/dca/cancel/dca-1", http.StatusOK, dcaCancelBody)

	request := dcaOrder()
	request.OrderType = trigger.DCAPriceConditional
	request.TriggerMint = wrappedSOLMint
	request.MinPriceUSD = "50"
	request.MaxPriceUSD = "1000"
	request.BeginFillAt = time.Now().Add(trigger.MaxDCAStartDelay + time.Hour)

	ctx := context.Background()
	if _, err := session.CreateDCAOrder(ctx, request); err == nil {
		t.Fatal("a start more than 30 days ahead was accepted")
	}

	start := time.Now().Add(time.Hour).Truncate(time.Second).UTC()
	request.BeginFillAt = start

	created, err := session.CreateDCAOrder(ctx, request)
	if err != nil {
		t.Fatal(err)
	}

	body := decodeBody(t, lastRequest(t, fake).body)
	want := map[string]string{
		"intervalSeconds": "3600",
		"orderCount":      "2",
		"orderType":       `"price_conditional"`,
		"minPriceUsd":     "50",
		"beginFillAt":     `"` + start.Format(time.RFC3339Nano) + `"`,
		"userPubkey":      `"` + triggerWallet() + `"`,
	}
	for key, value := range want {
		if string(body[key]) != value {
			t.Errorf("%s = %s, want %s", key, body[key], value)
		}
	}

	if _, ok := body["jlEnabled"]; ok || created.ID != "dca-1" {
		t.Fatalf("created=%+v jlEnabled sent=%v", created, ok)
	}

	cancellation, err := session.CancelDCAOrder(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}

	if cancellation.RoundsRemaining != 2 || cancellation.RefundAmount != "20000000" {
		t.Fatalf("cancellation = %+v", cancellation)
	}
}

func TestTriggerHistory(t *testing.T) {
	const priceHistory = `{
  "orders": [{
    "id": "order-1",
    "orderType": "single",
    "orderState": "open",
    "userPubkey": "wallet",
    "inputMint": "in",
    "outputMint": "out",
    "triggerPriceUsd": 200.000000000000000001,
    "expiresAt": null,
    "createdAt": 1735689600000,
    "updatedAt": 1735689600001,
    "outputAmount": null,
    "fillPercent": 0,
    "events": [{"type": "deposit", "timestamp": 1735689600000, "state": "success", "amount": "1"}]
  }],
  "pagination": {"total": 41, "limit": 20, "offset": 20}
}`
	const dcaRecord = `{
  "id": "dca-1",
  "orderType": "time_based",
  "triggerMint": null,
  "jlYieldUsd": null,
  "fillPercent": 0.5,
  "lastFillAt": null,
  "numberOfRounds": 2,
  "roundsFilled": 1,
  "events": [{"type": "fill", "roundNumber": 1, "state": "success", "attemptCount": 1}]
}`

	fake, session := priceSession(t)
	fake.route(http.MethodGet, "/orders/history", http.StatusOK, priceHistory)
	dcaHistoryBody := `{"orders":[` + dcaRecord + `],"pagination":{}}`
	fake.route(http.MethodGet, "/orders/history/dca", http.StatusOK, dcaHistoryBody)
	fake.route(http.MethodGet, "/orders/history/dca/dca-1", http.StatusOK, dcaRecord)

	limit := uint8(trigger.MaxHistoryLimit)
	ctx := context.Background()
	history, err := session.PriceOrderHistory(ctx, trigger.HistoryRequest{
		State:     trigger.HistoryActive,
		Mint:      wrappedSOLMint,
		Limit:     &limit,
		Offset:    20,
		Sort:      trigger.SortExpiresAt,
		Direction: trigger.DirectionAscending,
	})
	if err != nil {
		t.Fatal(err)
	}

	query := lastRequest(t, fake).query
	wantQuery := map[string]string{
		"state":  "active",
		"mint":   wrappedSOLMint,
		"limit":  "100",
		"offset": "20",
		"sort":   "expires_at",
		"dir":    "asc",
	}
	for key, value := range wantQuery {
		if got := query[key]; len(got) != 1 || got[0] != value {
			t.Errorf("%s = %v, want %s", key, got, value)
		}
	}

	order := history.Orders[0]
	if order.TriggerPriceUSD.String() != "200.000000000000000001" || order.ExpiresAt != nil {
		t.Fatalf("price record lost precision or null: %+v", order)
	}

	if order.CreatedAt != 1735689600000 || order.OutputAmount != nil || history.Pagination.Total != 41 {
		t.Fatalf("record = %+v pagination = %+v", order, history.Pagination)
	}

	if len(order.Events) != 1 || order.Events[0].Amount != "1" {
		t.Fatalf("events = %+v", order.Events)
	}

	dcaHistory, err := session.DCAOrderHistory(ctx, trigger.HistoryRequest{Sort: trigger.SortNextFillAt})
	if err != nil {
		t.Fatal(err)
	}

	if len(dcaHistory.Orders) != 1 || dcaHistory.Orders[0].FillPercent.String() != "0.5" {
		t.Fatalf("dca history = %+v", dcaHistory)
	}

	dca, err := session.DCAOrder(ctx, "dca-1")
	if err != nil {
		t.Fatal(err)
	}

	nullsKept := dca.TriggerMint == nil && dca.JLYieldUSD == "" && dca.LastFillAt == nil
	if !nullsKept || dca.Events[0].AttemptCount != 1 {
		t.Fatalf("dca record = %+v", dca)
	}
}

func TestInvalidTriggerOrdersDoNotReachTransport(t *testing.T) {
	client, err := jupiter.NewClient(jupiter.WithHTTPClient(unreachableHTTPClient{t: t}))
	if err != nil {
		t.Fatal(err)
	}

	tokens := trigger.Tokens{
		AuthMode:     "access_refresh",
		AccessToken:  "access",
		RefreshToken: "refresh",
		ExpiresAt:    time.Now().Add(accessTokenTTL),
		TokenType:    "Bearer",
	}

	session, err := client.Trigger.Resume(triggerWallet(), tokens, trigger.SessionOptions{})
	if err != nil {
		t.Fatal(err)
	}

	below := uint16(trigger.MinTrailingBps - 1)
	above := uint16(trigger.MaxTrailingBps + 1)
	trailing := uint16(trigger.MinTrailingBps)
	slippage := uint16(trigger.BasisPointsPerWhole + 1)
	historyLimit := uint8(trigger.MaxHistoryLimit + 1)

	price := func(
		base func() trigger.PriceOrderRequest,
		configure func(*trigger.PriceOrderRequest),
	) func() error {
		return func() error {
			request := base()
			configure(&request)
			_, err := session.CreatePriceOrder(context.Background(), request)

			return err
		}
	}

	dca := func(configure func(*trigger.DCAOrderRequest)) func() error {
		return func() error {
			request := dcaOrder()
			configure(&request)
			_, err := session.CreateDCAOrder(context.Background(), request)

			return err
		}
	}

	update := func(u trigger.PriceOrderUpdate) func() error {
		return func() error {
			_, err := session.UpdatePriceOrder(context.Background(), orderID, u)

			return err
		}
	}

	tests := []struct {
		name string
		call func() error
		want error
	}{
		{
			name: "unknown order type",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.OrderType = "limit"
			}),
		},
		{
			name: "missing deposit request",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.DepositRequestID = ""
			}),
		},
		{
			name: "unsigned deposit",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.DepositSignedTransaction = "not base64"
			}),
		},
		{
			name: "identical mints",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.OutputMint = r.InputMint
			}),
		},
		{
			name: "malformed trigger mint",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) { r.TriggerMint = "SOL" }),
			want: trigger.ErrInvalidAddress,
		},
		{
			name: "past expiry",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.ExpiresAt = time.Now().Add(-time.Minute)
			}),
		},
		{
			name: "missing expiry",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.ExpiresAt = time.Time{}
			}),
		},
		{
			name: "single without condition",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.TriggerCondition = ""
			}),
		},
		{
			name: "unknown condition",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.TriggerCondition = "cross"
			}),
		},
		{
			name: "single without price or trail",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.TriggerPriceUSD = ""
			}),
		},
		{
			name: "single with price and trail",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.TrailingBps = &trailing
			}),
		},
		{
			name: "trailing below on output mint",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.TriggerCondition = trigger.TriggerBelow
				r.TriggerPriceUSD = ""
				r.TrailingBps = &trailing
				r.TriggerMint = r.OutputMint
			}),
		},
		{
			name: "trailing above on input mint",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.TriggerPriceUSD = ""
				r.TrailingBps = &trailing
			}),
		},
		{
			name: "trail below minimum",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.TriggerPriceUSD = ""
				r.TrailingBps = &below
			}),
		},
		{
			name: "trail above maximum",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.TriggerPriceUSD = ""
				r.TrailingBps = &above
			}),
		},
		{
			name: "single with take profit",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.TakeProfitPriceUSD = "250"
			}),
		},
		{
			name: "slippage above maximum",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.SlippageBps = &slippage
			}),
		},
		{
			name: "negative price",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.TriggerPriceUSD = "-1"
			}),
		},
		{
			name: "zero price",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.TriggerPriceUSD = "0"
			}),
		},
		{
			name: "fraction price",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.TriggerPriceUSD = "1/2"
			}),
		},
		{
			name: "non-number price",
			call: price(singleOrder, func(r *trigger.PriceOrderRequest) {
				r.TriggerPriceUSD = "Infinity"
			}),
		},
		{
			name: "oco with trigger condition",
			call: price(ocoOrder, func(r *trigger.PriceOrderRequest) {
				r.TriggerCondition = trigger.TriggerAbove
			}),
		},
		{
			name: "oco without stop loss",
			call: price(ocoOrder, func(r *trigger.PriceOrderRequest) {
				r.StopLossPriceUSD = ""
			}),
		},
		{
			name: "oco take profit not above stop loss",
			call: price(ocoOrder, func(r *trigger.PriceOrderRequest) {
				r.TakeProfitPriceUSD = "150"
			}),
		},
		{
			name: "oco trailing",
			call: price(ocoOrder, func(r *trigger.PriceOrderRequest) {
				r.TrailingBps = &trailing
			}),
		},
		{
			name: "otoco without trigger price",
			call: price(ocoOrder, func(r *trigger.PriceOrderRequest) {
				r.OrderType = trigger.OrderSubTypeOTOCO
				r.TriggerCondition = trigger.TriggerBelow
			}),
		},
		{
			name: "update without changes",
			call: update(trigger.PriceOrderUpdate{OrderType: trigger.OrderSubTypeSingle}),
		},
		{
			name: "update oco trigger price",
			call: update(trigger.PriceOrderUpdate{OrderType: trigger.OrderSubTypeOCO, TriggerPriceUSD: "1"}),
		},
		{
			name: "update inverted pair",
			call: update(trigger.PriceOrderUpdate{
				OrderType:          trigger.OrderSubTypeOCO,
				TakeProfitPriceUSD: "100",
				StopLossPriceUSD:   "200",
			}),
		},
		{
			name: "update without order ID",
			call: func() error {
				u := trigger.PriceOrderUpdate{OrderType: trigger.OrderSubTypeSingle, TriggerPriceUSD: "1"}
				_, err := session.UpdatePriceOrder(context.Background(), "", u)
				return err
			},
		},
		{
			name: "confirm without cancel request",
			call: func() error {
				r := trigger.ConfirmCancelRequest{SignedTransaction: signedDeposit}
				_, err := session.ConfirmCancelPriceOrder(context.Background(), orderID, r)
				return err
			},
		},
		{
			name: "confirm unsigned withdrawal",
			call: func() error {
				r := trigger.ConfirmCancelRequest{CancelRequestID: "cancel-1", SignedTransaction: "%%"}
				_, err := session.ConfirmCancelDCAOrder(context.Background(), orderID, r)
				return err
			},
		},
		{name: "one DCA round", call: dca(func(r *trigger.DCAOrderRequest) { r.OrderCount = 1 })},
		{
			name: "interval below minimum",
			call: dca(func(r *trigger.DCAOrderRequest) {
				r.Interval = trigger.MinDCAInterval - time.Second
			}),
		},
		{
			name: "interval above maximum",
			call: dca(func(r *trigger.DCAOrderRequest) {
				r.Interval = trigger.MaxDCAInterval + time.Second
			}),
		},
		{
			name: "fractional interval",
			call: dca(func(r *trigger.DCAOrderRequest) {
				r.Interval = 90*time.Second + time.Millisecond
			}),
		},
		{name: "time-based price band", call: dca(func(r *trigger.DCAOrderRequest) { r.MinPriceUSD = "1" })},
		{name: "unknown DCA type", call: dca(func(r *trigger.DCAOrderRequest) { r.OrderType = "weekly" })},
		{
			name: "price-conditional without trigger mint",
			call: dca(func(r *trigger.DCAOrderRequest) {
				r.OrderType = trigger.DCAPriceConditional
				r.MinPriceUSD = "1"
			}),
			want: trigger.ErrInvalidAddress,
		},
		{
			name: "inverted price band",
			call: dca(func(r *trigger.DCAOrderRequest) {
				r.OrderType = trigger.DCAPriceConditional
				r.TriggerMint = wrappedSOLMint
				r.MinPriceUSD = "100"
				r.MaxPriceUSD = "100"
			}),
		},
		{name: "earn without mint", call: dca(func(r *trigger.DCAOrderRequest) { r.JLEnabled = true })},
		{name: "earn mint without flag", call: dca(func(r *trigger.DCAOrderRequest) { r.JLMint = usdcMint })},
		{
			name: "earn on price-conditional",
			call: dca(func(r *trigger.DCAOrderRequest) {
				r.OrderType = trigger.DCAPriceConditional
				r.TriggerMint = wrappedSOLMint
				r.JLEnabled = true
				r.JLMint = usdcMint
			}),
		},
		{
			name: "history limit above maximum",
			call: func() error {
				request := trigger.HistoryRequest{Limit: &historyLimit}
				_, err := session.PriceOrderHistory(context.Background(), request)
				return err
			},
		},
		{
			name: "history malformed mint",
			call: func() error {
				_, err := session.DCAOrderHistory(context.Background(), trigger.HistoryRequest{Mint: "SOL"})
				return err
			},
			want: trigger.ErrInvalidAddress,
		},
		{
			name: "history state with separator",
			call: func() error {
				request := trigger.HistoryRequest{State: "active,past"}
				_, err := session.PriceOrderHistory(context.Background(), request)
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

			if errors.Is(err, trigger.ErrOutcomeUnknown) {
				t.Fatalf("validation failure reported as unknown outcome: %v", err)
			}

			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}

func TestDCABoundariesAreAccepted(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*trigger.DCAOrderRequest)
	}{
		{
			name: "minimum interval",
			configure: func(r *trigger.DCAOrderRequest) {
				r.Interval = trigger.MinDCAInterval
			},
		},
		{
			name: "maximum interval",
			configure: func(r *trigger.DCAOrderRequest) {
				r.Interval = trigger.MaxDCAInterval
			},
		},
		{
			name: "earn while you wait",
			configure: func(r *trigger.DCAOrderRequest) {
				r.OrderType = trigger.DCATimeBased
				r.JLEnabled = true
				r.JLMint = wrappedSOLMint
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fake, session := priceSession(t)
			fake.route(http.MethodPost, "/orders/dca", http.StatusOK, submissionBody)

			request := dcaOrder()
			test.configure(&request)

			if _, err := session.CreateDCAOrder(context.Background(), request); err != nil {
				t.Fatal(err)
			}
		})
	}
}
