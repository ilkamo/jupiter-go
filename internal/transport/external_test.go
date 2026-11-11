package transport

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

const presignedURL = "https://uploads.example.com/image?X-Signature=a%2Fb%3D&X-Expires=60"

type capturedRequest struct {
	request *http.Request
	body    string
}

func capturingClient(t *testing.T, captured *[]capturedRequest) *Client {
	t.Helper()

	return &Client{
		BaseURL: "https://api.example.com",
		APIKey:  "secret-key",
		Timeout: time.Second,
		HTTP: doer(func(r *http.Request) (*http.Response, error) {
			data, err := io.ReadAll(r.Body)
			if err != nil {
				return nil, err
			}

			*captured = append(*captured, capturedRequest{request: r, body: string(data)})

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("not json")),
			}, nil
		}),
	}
}

func TestExternalRequestsNeverCarryCredentials(t *testing.T) {
	var captured []capturedRequest
	client := capturingClient(t, &captured)

	request := Request{
		Method:      http.MethodPut,
		ExternalURL: presignedURL,
		Payload:     &Payload{ContentType: "image/png", Data: []byte("png")},
		Retry:       NoRetry,
	}

	// A nil result ignores the non-JSON upload response.
	if _, err := client.Do(context.Background(), request, nil); err != nil {
		t.Fatal(err)
	}

	sent := captured[0]
	if sent.request.Header.Get("x-api-key") != "" || sent.request.Header.Get("Accept") != "" {
		t.Fatalf("credentials or API headers sent: %v", sent.request.Header)
	}

	if sent.request.URL.String() != presignedURL {
		t.Fatalf("presigned URL changed: %s", sent.request.URL)
	}

	if sent.request.Header.Get("Content-Type") != "image/png" || sent.body != "png" {
		t.Fatalf("payload changed: %v %q", sent.request.Header, sent.body)
	}
}

func TestPayloadRequestsKeepCredentials(t *testing.T) {
	var captured []capturedRequest
	client := capturingClient(t, &captured)

	request := Request{
		Method:  http.MethodPost,
		Path:    "/submit",
		Payload: &Payload{ContentType: "multipart/form-data; boundary=x", Data: []byte("form")},
		Retry:   NoRetry,
	}

	if _, err := client.Do(context.Background(), request, nil); err != nil {
		t.Fatal(err)
	}

	sent := captured[0].request
	contentType := sent.Header.Get("Content-Type")
	if sent.Header.Get("x-api-key") != "secret-key" || contentType != "multipart/form-data; boundary=x" {
		t.Fatalf("headers = %v", sent.Header)
	}
}

func TestInvalidRequestTargets(t *testing.T) {
	tests := []struct {
		name    string
		request Request
	}{
		{
			name:    "body and payload",
			request: Request{Body: struct{}{}, Payload: &Payload{ContentType: "text/plain"}},
		},
		{
			name:    "payload without content type",
			request: Request{Payload: &Payload{Data: []byte("x")}},
		},
		{
			name:    "plain HTTP external URL",
			request: Request{ExternalURL: "http://uploads.example.com/image"},
		},
		{
			name:    "relative external URL",
			request: Request{ExternalURL: "/image"},
		},
		{
			name:    "external URL with user info",
			request: Request{ExternalURL: "https://user:pass@uploads.example.com/image"},
		},
		{
			name:    "external URL with API path",
			request: Request{ExternalURL: presignedURL, Path: "/image"},
		},
		{
			name: "external URL with headers",
			request: Request{
				ExternalURL: presignedURL,
				Header:      http.Header{"Authorization": {"Bearer token"}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var captured []capturedRequest
			client := capturingClient(t, &captured)

			request := test.request
			request.Method = http.MethodPut
			request.Retry = NoRetry

			if _, err := client.Do(context.Background(), request, nil); err == nil {
				t.Fatal("invalid request accepted")
			}

			if len(captured) != 0 {
				t.Fatal("invalid request was sent")
			}
		})
	}
}
