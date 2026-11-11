package jupiter_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/transaction"
)

const (
	txAPIKey       = "tx-secret"
	sentSignature  = "5eykt4UsFv8P8NJdTREpY1vzqKqZKvdpKuc147dw2N9d"
	successRPC     = `{"jsonrpc": "2.0", "id": 1, "result": "` + sentSignature + `"}`
	missingTipRPC  = `{"jsonrpc": "2.0", "id": 1, "error": {"code": -1014, "message": "missing tip"}}`
	wrongIDRPC     = `{"jsonrpc": "2.0", "id": 7, "result": "` + sentSignature + `"}`
	signedTxBase64 = "AQID"
)

type txReply struct {
	status int
	body   string
}

// txServers starts a transaction service replying in order (repeating the last
// reply) and an API server that must not be called.
func txServers(t *testing.T, replies ...txReply) (*jupiter.Client, func() []*http.Request, func() [][]byte) {
	t.Helper()

	var (
		mu       sync.Mutex
		requests []*http.Request
		bodies   [][]byte
	)

	txServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}

		mu.Lock()
		requests = append(requests, r)
		bodies = append(bodies, body)
		reply := replies[min(len(requests), len(replies))-1]
		mu.Unlock()

		w.WriteHeader(reply.status)

		if _, err := w.Write([]byte(reply.body)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(txServer.Close)

	apiServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("transaction submission reached the API gateway")
	}))
	t.Cleanup(apiServer.Close)

	client, err := jupiter.NewClient(
		jupiter.WithBaseURL(apiServer.URL),
		jupiter.WithTransactionURL(txServer.URL),
		jupiter.WithAPIKey(txAPIKey),
	)
	if err != nil {
		t.Fatal(err)
	}

	recordedRequests := func() []*http.Request {
		mu.Lock()
		defer mu.Unlock()

		return append([]*http.Request(nil), requests...)
	}

	recordedBodies := func() [][]byte {
		mu.Lock()
		defer mu.Unlock()

		return append([][]byte(nil), bodies...)
	}

	return client, recordedRequests, recordedBodies
}

func TestTransactionSendEncodesJSONRPC(t *testing.T) {
	tests := []struct {
		name      string
		options   transaction.SendOptions
		wantParam string
	}{
		{name: "default routing", wantParam: `{"encoding":"base64"}`},
		{
			name:      "SWQoS only",
			options:   transaction.SendOptions{SWQoSOnly: true},
			wantParam: `{"encoding":"base64","swqosOnly":true}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, requests, bodies := txServers(t, txReply{status: http.StatusOK, body: successRPC})

			result, err := client.Transaction.Send(context.Background(), signedTxBase64, test.options)
			if err != nil {
				t.Fatal(err)
			}

			if result.Signature != sentSignature {
				t.Fatalf("signature = %q", result.Signature)
			}

			sent := requests()[0]
			if sent.URL.Path != "/" || sent.Header.Get("x-api-key") != txAPIKey {
				t.Fatalf("path=%q key=%q", sent.URL.Path, sent.Header.Get("x-api-key"))
			}

			var body struct {
				JSONRPC string            `json:"jsonrpc"`
				ID      int               `json:"id"`
				Method  string            `json:"method"`
				Params  []json.RawMessage `json:"params"`
			}
			if err := json.Unmarshal(bodies()[0], &body); err != nil {
				t.Fatal(err)
			}

			if body.JSONRPC != "2.0" || body.Method != "sendTransaction" || len(body.Params) != 2 {
				t.Fatalf("envelope = %+v", body)
			}

			if string(body.Params[0]) != `"`+signedTxBase64+`"` || string(body.Params[1]) != test.wantParam {
				t.Fatalf("params = %s, %s", body.Params[0], body.Params[1])
			}
		})
	}
}

func TestTransactionSendOutcomes(t *testing.T) {
	tests := []struct {
		name     string
		replies  []txReply
		attempts int
		rpcCode  int
		unknown  bool
		succeeds bool
	}{
		{
			name: "transient failure then success",
			replies: []txReply{
				{status: http.StatusServiceUnavailable},
				{status: http.StatusOK, body: successRPC},
			},
			attempts: 2,
			succeeds: true,
		},
		{
			name:     "JSON-RPC rejection",
			replies:  []txReply{{status: http.StatusOK, body: missingTipRPC}},
			attempts: 1,
			rpcCode:  transaction.CodeMissingTip,
		},
		{
			name:     "persistent server error",
			replies:  []txReply{{status: http.StatusInternalServerError, body: `{}`}},
			attempts: 3,
			unknown:  true,
		},
		{
			name:     "unexpected response ID",
			replies:  []txReply{{status: http.StatusOK, body: wrongIDRPC}},
			attempts: 1,
			unknown:  true,
		},
		{
			name:     "unauthorized",
			replies:  []txReply{{status: http.StatusUnauthorized, body: `{"code": 401}`}},
			attempts: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, _, bodies := txServers(t, test.replies...)

			_, err := client.Transaction.Send(context.Background(), signedTxBase64, transaction.SendOptions{})

			sent := bodies()
			if len(sent) != test.attempts {
				t.Fatalf("attempts = %d, want %d", len(sent), test.attempts)
			}

			for _, body := range sent[1:] {
				if !bytes.Equal(body, sent[0]) {
					t.Fatal("a retry changed the submitted bytes")
				}
			}

			if test.succeeds {
				if err != nil {
					t.Fatal(err)
				}

				return
			}

			if err == nil {
				t.Fatal("expected failure")
			}

			if errors.Is(err, transaction.ErrOutcomeUnknown) != test.unknown {
				t.Fatalf("unexpected outcome classification: %v", err)
			}

			rpcError, isRPC := errors.AsType[*transaction.RPCError](err)
			if isRPC != (test.rpcCode != 0) || (isRPC && rpcError.Code != test.rpcCode) {
				t.Fatalf("unexpected RPC error: %v", err)
			}
		})
	}
}

func TestInvalidTransactionInputs(t *testing.T) {
	client, err := jupiter.NewClient(jupiter.WithHTTPClient(unreachableHTTPClient{t: t}))
	if err != nil {
		t.Fatal(err)
	}

	for _, signed := range []string{"", "not base64"} {
		_, err := client.Transaction.Send(context.Background(), signed, transaction.SendOptions{})
		if err == nil {
			t.Fatalf("Send(%q) accepted", signed)
		}
	}

	invalidURLs := []string{"ftp://tx.example.com", "https://tx.example.com?key=1", "tx.example.com"}
	for _, invalid := range invalidURLs {
		if _, err := jupiter.NewClient(jupiter.WithTransactionURL(invalid)); err == nil {
			t.Fatalf("transaction URL %q accepted", invalid)
		}
	}
}
