package transport

import (
	"io"
	"net/http"
)

// Bound response buffering to 8 MiB, including responses without Content-Length.
const maxResponseBodyBytes int64 = 8 * 1024 * 1024

func readResponseBody(body io.ReadCloser, maxBytes int64) ([]byte, error) {
	// Client-side reads have no ResponseWriter. MaxBytesReader still enforces the
	// limit and returns a typed *http.MaxBytesError when the body exceeds it.
	reader := http.MaxBytesReader(nil, body, maxBytes)

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return data, nil
}
