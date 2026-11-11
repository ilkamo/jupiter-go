package swap

import (
	"errors"
	"fmt"

	"github.com/ilkamo/jupiter-go/internal/transport"
)

// quoteRetryPolicy covers order and build. Repeating these GETs changes no state:
// each attempt returns a fresh quote, blockhash, and request ID, and an unused
// earlier quote simply expires.
var quoteRetryPolicy = transport.RepeatableRetry

// executeRetryPolicy makes one attempt: the execute documentation gives no replay
// guarantee for repeating a signed transaction and request ID.
var executeRetryPolicy = transport.NoRetry

// ErrOutcomeUnknown marks execute failures after which the transaction may
// still land: transport failures, timeouts, server errors, and unreadable success
// responses. Check the transaction signature on-chain before requesting a new order;
// a new order creates a new transaction that could also land.
var ErrOutcomeUnknown = errors.New("swap execution outcome unknown")

func classifyExecuteFailure(err error) error {
	if !transport.MayHaveBeenProcessed(err) {
		return err
	}

	return fmt.Errorf("%w: %w", ErrOutcomeUnknown, err)
}
