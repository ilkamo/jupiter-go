package jupiter_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/tokens"
)

const (
	expressBase   = "/tokens/v2/verify/express"
	solPaymentFix = `{
  "transaction": "AQ==", "requestId": "req_1", "mint": "JUP", "amount": "1000000000",
  "tokenUsdRate": 0.42, "feeLamports": 5000, "code": 0, "gasless": true,
  "inputMint": "So11111111111111111111111111111111111111112", "inputDecimals": 9,
  "quotedInputAmount": "2840000000", "maxInputAmount": "2840000000"
}`
)

func expressRequest() tokens.ExpressRequest {
	return tokens.ExpressRequest{
		Transaction:     "AQ==",
		RequestID:       "req_1",
		SenderAddress:   takerAddress,
		TokenID:         usdcMint,
		TwitterHandle:   "https://x.com/project",
		Description:     "Established project",
		PaymentCurrency: tokens.PaymentSOL,
		PaymentAmount:   "002840000000",
		JUPOutputAmount: "1000000000",
	}
}

func TestExpressEligibilityAndPayment(t *testing.T) {
	client, recorded := pathServer(t, map[string]string{
		expressBase + "/check-eligibility": `{"canVerify": true, "metadataError": "pending"}`,
		expressBase + "/craft-txn":         solPaymentFix,
	})

	ctx := context.Background()
	eligibility, err := client.Tokens.Express.CheckEligibility(ctx, usdcMint)
	if err != nil {
		t.Fatal(err)
	}

	if !eligibility.CanVerify || eligibility.CanMetadata || eligibility.MetadataError != "pending" {
		t.Fatalf("eligibility=%+v err=%v", eligibility, err)
	}

	payment, err := client.Tokens.Express.CraftPayment(ctx, takerAddress, tokens.PaymentSOL)
	if err != nil {
		t.Fatal(err)
	}

	sizedSwap := payment.QuotedInputAmount == "2840000000" && payment.TokenUSDRate.String() == "0.42"
	if !sizedSwap || payment.Code == nil {
		t.Fatalf("payment = %+v", payment)
	}

	if _, err := client.Tokens.Express.CraftPayment(ctx, takerAddress, ""); err != nil {
		t.Fatal(err)
	}

	requests := recorded()
	if requests[0].query.Get("tokenId") != usdcMint {
		t.Fatalf("eligibility query = %v", requests[0].query)
	}

	craftQuery := requests[1].query
	if craftQuery.Get("paymentCurrency") != "SOL" || craftQuery.Get("senderAddress") != takerAddress {
		t.Fatalf("craft query = %v", requests[1].query)
	}

	if _, ok := requests[2].query["paymentCurrency"]; ok {
		t.Fatal("empty currency was sent instead of the JUP default")
	}
}

func TestExpressCraftFailureCode(t *testing.T) {
	client, _ := pathServer(t, map[string]string{
		expressBase + "/craft-txn": `{"code": -1, "error": "insufficient balance"}`,
	})

	_, err := client.Tokens.Express.CraftPayment(context.Background(), takerAddress, "")
	expressError, ok := errors.AsType[*tokens.ExpressError](err)
	if !ok || expressError.Code == nil || *expressError.Code != -1 {
		t.Fatalf("expected express error, got %v", err)
	}
}

func TestExpressExecuteEncodesRequest(t *testing.T) {
	client, recorded := pathServer(t, map[string]string{
		expressBase + "/execute": `{"status": "Success", "signature": "sig", "verificationCreated": true}`,
	})

	request := expressRequest()
	useSupply := false
	request.TokenMetadata = &tokens.TokenMetadataUpdate{
		TokenID:              usdcMint,
		Website:              "https://project.example",
		UseCirculatingSupply: &useSupply,
	}

	result, err := client.Tokens.Express.Execute(context.Background(), request)
	if err != nil || result.Signature != "sig" || !result.VerificationCreated {
		t.Fatalf("result=%+v err=%v", result, err)
	}

	body := recorded()[0].body
	if string(body["paymentAmount"]) != `"2840000000"` || string(body["paymentCurrency"]) != `"SOL"` {
		t.Fatalf("body = %v", body)
	}

	if _, ok := body["senderTwitterHandle"]; ok {
		t.Fatal("unset sender handle was sent")
	}

	var metadata map[string]json.RawMessage
	if err := json.Unmarshal(body["tokenMetadata"], &metadata); err != nil {
		t.Fatal(err)
	}

	if string(metadata["useCirculatingSupply"]) != "false" || len(metadata) != 3 {
		t.Fatalf("metadata = %v", metadata)
	}
}

func TestExpressExecuteOutcomes(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		unknown bool
		express bool
	}{
		{
			name:    "failed status",
			status:  http.StatusOK,
			body:    `{"status": "Failed", "code": -1005}`,
			express: true,
		},
		{name: "missing status", status: http.StatusOK, body: `{}`, express: true},
		{name: "server error", status: http.StatusInternalServerError, body: `{}`, unknown: true},
		{name: "conflict", status: http.StatusConflict, body: `{"error": "pending submission"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, calls := statusServer(t, test.status, test.body)

			client, err := jupiter.NewClient(jupiter.WithBaseURL(server))
			if err != nil {
				t.Fatal(err)
			}

			_, err = client.Tokens.Express.Execute(context.Background(), expressRequest())
			if err == nil {
				t.Fatal("expected failure")
			}

			if calls() != singleAttempt {
				t.Fatalf("execute attempts = %d", calls())
			}

			if errors.Is(err, tokens.ErrOutcomeUnknown) != test.unknown {
				t.Fatalf("unexpected outcome classification: %v", err)
			}

			if _, isExpress := errors.AsType[*tokens.ExpressError](err); isExpress != test.express {
				t.Fatalf("unexpected express error: %v", err)
			}
		})
	}
}

func TestInvalidExpressRequestsDoNotReachTransport(t *testing.T) {
	client, err := jupiter.NewClient(jupiter.WithHTTPClient(unreachableHTTPClient{t: t}))
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		configure func(*tokens.ExpressRequest)
		want      error
	}{
		{name: "non-JUP without amount", configure: func(r *tokens.ExpressRequest) { r.PaymentAmount = "" }},
		{name: "zero payment", configure: func(r *tokens.ExpressRequest) { r.PaymentAmount = "0" }},
		{name: "missing description", configure: func(r *tokens.ExpressRequest) { r.Description = "" }},
		{name: "unsigned transaction", configure: func(r *tokens.ExpressRequest) { r.Transaction = "%%" }},
		{name: "missing request ID", configure: func(r *tokens.ExpressRequest) { r.RequestID = "" }},
		{
			name:      "malformed token",
			configure: func(r *tokens.ExpressRequest) { r.TokenID = "token" },
			want:      tokens.ErrInvalidAddress,
		},
		{
			name: "malformed metadata token",
			configure: func(r *tokens.ExpressRequest) {
				r.TokenMetadata = &tokens.TokenMetadataUpdate{TokenID: "token"}
			},
			want: tokens.ErrInvalidAddress,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := expressRequest()
			test.configure(&request)

			_, err := client.Tokens.Express.Execute(context.Background(), request)
			if err == nil {
				t.Fatal("invalid request accepted")
			}

			if errors.Is(err, tokens.ErrOutcomeUnknown) {
				t.Fatalf("validation failure reported as unknown outcome: %v", err)
			}

			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}

func TestExpressJUPPaymentNeedsNoAmount(t *testing.T) {
	client, recorded := pathServer(t, map[string]string{
		expressBase + "/execute": `{"status": "Success", "signature": "sig"}`,
	})

	request := expressRequest()
	request.PaymentCurrency = tokens.PaymentJUP
	request.PaymentAmount = ""
	request.JUPOutputAmount = ""

	if _, err := client.Tokens.Express.Execute(context.Background(), request); err != nil {
		t.Fatalf("JUP payment without amount rejected: %v", err)
	}

	if _, ok := recorded()[0].body["paymentAmount"]; ok {
		t.Fatal("empty payment amount was sent")
	}
}
