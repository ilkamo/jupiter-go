package swap

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ilkamo/jupiter-go/internal/transport"
)

const (
	quoteMaxAttempts    = 3
	quoteInitialBackoff = 100 * time.Millisecond
	quoteMaxBackoff     = 500 * time.Millisecond
	quoteMaxRetryAfter  = 1 * time.Second
)

// quoteRetryPolicy covers order and build. Repeating these GETs changes no state:
// each attempt returns a fresh quote, blockhash, and request ID, and an unused
// earlier quote simply expires. Waits stay short because the results are
// time-sensitive.
var quoteRetryPolicy = transport.RetryPolicy{
	MaxAttempts:    quoteMaxAttempts,
	InitialBackoff: quoteInitialBackoff,
	MaxBackoff:     quoteMaxBackoff,
	MaxRetryAfter:  quoteMaxRetryAfter,
}

// executeRetryPolicy makes one attempt: the execute documentation gives no replay
// guarantee for repeating a signed transaction and request ID.
var executeRetryPolicy = transport.NoRetry

// ErrExecutionOutcomeUnknown marks execute failures after which the transaction may
// still land: transport failures, timeouts, server errors, and unreadable success
// responses. Check the transaction signature on-chain before requesting a new order;
// a new order creates a new transaction that could also land.
var ErrExecutionOutcomeUnknown = errors.New("swap execution outcome unknown")

// classifyExecuteFailure treats client errors as definitive rejections and every
// other transport failure as an unknown outcome.
func classifyExecuteFailure(err error) error {
	var apiError *transport.APIError
	if errors.As(err, &apiError) && apiError.StatusCode < http.StatusInternalServerError {
		return err
	}

	return fmt.Errorf("%w: %w", ErrExecutionOutcomeUnknown, err)
}
