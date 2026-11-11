package transport

import (
	"context"
	"fmt"
	"math"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

const (
	retryAfterHeader      = "Retry-After"
	singleAttempt         = 1
	backoffGrowthFactor   = 2
	retryAfterSecondsBase = 10
	retryAfterSecondsBits = 64
)

// RetryPolicy bounds the attempts made for one operation. The operation timeout
// covers every attempt and every wait, so a policy never extends a deadline.
type RetryPolicy struct {
	// MaxAttempts includes the first attempt; one disables retries.
	MaxAttempts int

	// InitialBackoff and MaxBackoff bound the full-jitter exponential wait used
	// when the server does not send Retry-After.
	InitialBackoff time.Duration
	MaxBackoff     time.Duration

	// MaxRetryAfter is the longest server-requested wait the operation accepts.
	// A longer Retry-After ends retries and returns the last response's error.
	MaxRetryAfter time.Duration
}

// NoRetry makes exactly one attempt. Use it for operations without a documented
// replay guarantee.
var NoRetry = RetryPolicy{MaxAttempts: singleAttempt}

func (p RetryPolicy) validate() error {
	if p.MaxAttempts < singleAttempt {
		return fmt.Errorf("retry policy requires at least %d attempt", singleAttempt)
	}

	if p.MaxAttempts == singleAttempt {
		return nil
	}

	if p.InitialBackoff <= 0 || p.MaxBackoff < p.InitialBackoff {
		return fmt.Errorf("retry policy requires positive backoff with maximum not below initial")
	}

	if p.MaxRetryAfter < 0 {
		return fmt.Errorf("retry policy maximum Retry-After must not be negative")
	}

	return nil
}

// backoff returns a full-jitter wait for the given retry, counting from zero.
func (p RetryPolicy) backoff(retry int) time.Duration {
	ceiling := p.InitialBackoff
	for range retry {
		if ceiling >= p.MaxBackoff/backoffGrowthFactor {
			ceiling = p.MaxBackoff

			break
		}

		ceiling *= backoffGrowthFactor
	}

	return rand.N(ceiling + 1)
}

// retryableStatus reports statuses that indicate the request was not processed
// and may succeed when repeated: rate limiting and transient server failures.
func retryableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

// retryAfter parses the delay-seconds and HTTP-date forms of Retry-After.
func retryAfter(header http.Header, now time.Time) (time.Duration, bool) {
	value := header.Get(retryAfterHeader)
	if value == "" {
		return 0, false
	}

	seconds, err := strconv.ParseInt(value, retryAfterSecondsBase, retryAfterSecondsBits)
	if err == nil {
		maxSeconds := int64(math.MaxInt64 / time.Second)
		if seconds < 0 || seconds > maxSeconds {
			return 0, false
		}

		return time.Duration(seconds) * time.Second, true
	}

	date, err := http.ParseTime(value)
	if err != nil {
		return 0, false
	}

	return max(date.Sub(now), 0), true
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
