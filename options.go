package jupiter

import (
	"fmt"
	"time"

	"github.com/ilkamo/jupiter-go/swap"
)

// WithAPIKey sends key as the x-api-key header on API and transaction requests.
// Presigned upload URLs never receive it.
func WithAPIKey(key string) Option {
	return func(c *config) error {
		c.apiKey = key

		return nil
	}
}

// WithBaseURL overrides the API gateway, https://api.jup.ag, for example for tests
// or a proxy. It must be an absolute HTTP(S) URL without credentials or query.
func WithBaseURL(baseURL string) Option {
	return func(c *config) error {
		c.baseURL = baseURL

		return nil
	}
}

// WithTransactionURL overrides the transaction landing service (transaction.DefaultURL),
// which is separate from the API gateway set by WithBaseURL.
func WithTransactionURL(transactionURL string) Option {
	return func(c *config) error {
		c.transactionURL = transactionURL

		return nil
	}
}

// WithHTTPClient replaces the default HTTP client, which refuses redirects so the
// API key cannot be forwarded. The replacement owns its redirect and credential
// policy and is never modified.
func WithHTTPClient(client HTTPClient) Option {
	return func(c *config) error {
		c.http = client

		return nil
	}
}

// WithTimeout bounds each operation, including retries and waits between them.
// The default is 30 seconds.
func WithTimeout(timeout time.Duration) Option {
	return func(c *config) error {
		c.timeout = timeout

		return nil
	}
}

// WithoutRetries makes every operation a single attempt. Use it when the injected
// HTTP client already retries, so attempts do not multiply.
func WithoutRetries() Option {
	return func(c *config) error {
		c.disableRetries = true

		return nil
	}
}

// WithReferral adds your own Jupiter referral fee, from 50 to 255 basis points, to
// managed Swap orders. The SDK charges no fee of its own. Provision the referral
// account and its token accounts separately: without them Jupiter executes orders
// without collecting the fee. Configuring a referral twice is an error.
func WithReferral(account string, bps uint16) Option {
	return func(c *config) error {
		referral := swap.Referral{Account: account, BasisPoints: bps}
		if err := referral.Validate(); err != nil {
			return fmt.Errorf("referral: %w", err)
		}

		if c.referralConfigured {
			return fmt.Errorf("referral configured more than once")
		}

		c.referralConfigured = true
		c.referral = referral

		return nil
	}
}
