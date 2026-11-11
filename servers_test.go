package jupiter_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"

	jupiter "github.com/ilkamo/jupiter-go"
)

type pathRequest struct {
	method string
	path   string
	query  url.Values
	body   map[string]json.RawMessage
}

// pathServer serves fixed bodies by path and records every request.
func pathServer(t *testing.T, replies map[string]string) (*jupiter.Client, func() []pathRequest) {
	t.Helper()

	var (
		mu       sync.Mutex
		requests []pathRequest
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}

		recorded := pathRequest{method: r.Method, path: r.URL.EscapedPath(), query: r.URL.Query()}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &recorded.body); err != nil {
				t.Errorf("decode body: %v", err)
			}
		}

		mu.Lock()
		requests = append(requests, recorded)
		mu.Unlock()

		reply, ok := replies[r.URL.EscapedPath()]
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

	return client, func() []pathRequest {
		mu.Lock()
		defer mu.Unlock()

		return append([]pathRequest(nil), requests...)
	}
}

// statusServer answers every request with status and body, and counts requests.
func statusServer(t *testing.T, status int, body string) (string, func() int32) {
	t.Helper()

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(status)

		if _, err := w.Write([]byte(body)); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	return server.URL, calls.Load
}
