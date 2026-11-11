package transport

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestReadResponseBodyLimit(t *testing.T) {
	const maxBytes int64 = 4

	for _, test := range []struct {
		name      string
		body      string
		oversized bool
	}{
		{name: "empty", body: ""},
		{name: "below limit", body: "abc"},
		{name: "exact limit", body: "abcd"},
		{name: "above limit", body: "abcde", oversized: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			body := io.NopCloser(strings.NewReader(test.body))
			data, err := readResponseBody(body, maxBytes)

			if test.oversized {
				sizeErr, ok := errors.AsType[*http.MaxBytesError](err)
				if !ok || sizeErr.Limit != maxBytes {
					t.Fatalf("expected size error, got %v", err)
				}

				if data != nil {
					t.Fatal("oversized response returned partial data")
				}

				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if string(data) != test.body {
				t.Fatalf("response changed: %q", data)
			}
		})
	}
}
