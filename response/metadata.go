// Package response exposes HTTP metadata shared by Jupiter API services.
package response

import "net/http"

// Metadata belongs to one response. Headers are copied from the HTTP response
// and include rate-limit and Retry-After values without lossy interpretation.
type Metadata struct {
	StatusCode int
	RequestID  string
	Headers    http.Header
}
