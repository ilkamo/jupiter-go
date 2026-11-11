package jupiter

import (
	"fmt"
	"time"

	"github.com/ilkamo/jupiter-go/swap"
)

func WithAPIKey(key string) Option {
	return func(c *config) error {
		c.apiKey = key

		return nil
	}
}

func WithBaseURL(baseURL string) Option {
	return func(c *config) error {
		c.baseURL = baseURL

		return nil
	}
}

func WithHTTPClient(client HTTPClient) Option {
	return func(c *config) error {
		c.http = client

		return nil
	}
}

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

// WithoutMaintainerFee selects the current fee-free SDK policy.
// Jupiter and network fees still apply.
func WithoutMaintainerFee() Option {
	return func(c *config) error {
		if c.feeConfigured {
			return fmt.Errorf("conflicting fee options")
		}

		c.feeConfigured = true
		c.fee = swap.Referral{}

		return nil
	}
}

// WithReferral replaces the SDK fee policy for managed orders. Accounts must be provisioned separately.
func WithReferral(account string, bps uint16) Option {
	return func(c *config) error {
		referral := swap.Referral{Account: account, BasisPoints: bps}
		if err := referral.Validate(); err != nil {
			return fmt.Errorf("referral: %w", err)
		}

		if c.feeConfigured {
			return fmt.Errorf("conflicting fee options")
		}

		c.feeConfigured = true
		c.fee = referral

		return nil
	}
}
