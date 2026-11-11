package transport

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type doer func(*http.Request) (*http.Response, error)

func (f doer) Do(r *http.Request) (*http.Response, error) {
	return f(r)
}

type failingCloser struct {
	io.Reader
	err error
}

func (b failingCloser) Close() error {
	return b.err
}

func TestMalformedErrorPreservesStatusAndCause(t *testing.T) {
	closeErr := errors.New("close failed")
	client := Client{
		BaseURL: "https://example.com",
		Timeout: time.Second,
		HTTP: doer(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadGateway,
				Header:     http.Header{"X-Request-Id": {"request"}},
				Body:       failingCloser{Reader: strings.NewReader("upstream unavailable"), err: closeErr},
			}, nil
		}),
	}

	var result any
	err := client.Do(context.Background(), "GET", "/", nil, nil, &result)

	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusBadGateway || apiErr.RequestID != "request" {
		t.Fatalf("HTTP context lost: %v", err)
	}

	var syntaxErr *json.SyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Fatalf("decode cause lost: %v", err)
	}

	if !errors.Is(err, closeErr) {
		t.Fatalf("close error lost: %v", err)
	}
}
