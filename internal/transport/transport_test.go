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
	request := Request{Method: http.MethodGet, Path: "/", Retry: NoRetry}
	metadata, err := client.Do(context.Background(), request, &result)

	if metadata.StatusCode != http.StatusBadGateway {
		t.Fatalf("missing failure metadata: %+v", metadata)
	}

	var apiErr *APIError
	httpContextKept := errors.As(err, &apiErr) &&
		apiErr.StatusCode == http.StatusBadGateway &&
		apiErr.RequestID == "request"
	if !httpContextKept {
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

func TestMetadataDoesNotAliasHTTPHeaders(t *testing.T) {
	headers := http.Header{"X-Request-Id": {"original"}}
	client := Client{
		BaseURL: "https://example.com",
		Timeout: time.Second,
		HTTP: doer(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     headers,
				Body:       io.NopCloser(strings.NewReader(`{}`)),
			}, nil
		}),
	}

	var result map[string]string
	request := Request{Method: http.MethodGet, Path: "/", Retry: NoRetry}
	metadata, err := client.Do(context.Background(), request, &result)
	if err != nil {
		t.Fatal(err)
	}

	headers.Set("x-request-id", "changed")
	if metadata.Headers.Get("x-request-id") != "original" {
		t.Fatal("metadata aliases the HTTP response headers")
	}
}
