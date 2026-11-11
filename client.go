// Package jupiter provides clients for Jupiter APIs. It performs no signing and has no
// Solana library dependency; package solanatx signs and assembles swap transactions.
package jupiter

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ilkamo/jupiter-go/internal/transport"
	"github.com/ilkamo/jupiter-go/lend"
	"github.com/ilkamo/jupiter-go/prediction"
	"github.com/ilkamo/jupiter-go/price"
	"github.com/ilkamo/jupiter-go/send"
	"github.com/ilkamo/jupiter-go/studio"
	"github.com/ilkamo/jupiter-go/swap"
	"github.com/ilkamo/jupiter-go/tokens"
	"github.com/ilkamo/jupiter-go/transaction"
	"github.com/ilkamo/jupiter-go/trigger"
)

const (
	defaultRequestTimeout = 30 * time.Second
	schemeHTTP            = "http"
	schemeHTTPS           = "https"
)

type HTTPClient = transport.HTTPClient

type APIError = transport.APIError

type Client struct {
	Swap        *swap.Client
	Tokens      *tokens.Client
	Price       *price.Client
	Trigger     *trigger.Client
	Lend        *lend.Client
	Prediction  *prediction.Client
	Send        *send.Client
	Studio      *studio.Client
	Transaction *transaction.Client
}

type config struct {
	baseURL, apiKey    string
	transactionURL     string
	http               HTTPClient
	timeout            time.Duration
	referral           swap.Referral
	referralConfigured bool
	disableRetries     bool
}

type Option func(*config) error

// NewClient constructs a client without network calls. The SDK charges no fee;
// WithReferral adds your own referral fee to managed Swap orders.
// WithTimeout bounds each operation, including its retries. Read-only operations,
// order, and build retry transient failures; Swap execute never retries.
// WithHTTPClient transfers redirect and credential-forwarding responsibility to the caller.
func NewClient(options ...Option) (*Client, error) {
	c := config{
		baseURL:        "https://api.jup.ag",
		transactionURL: transaction.DefaultURL,
		timeout:        defaultRequestTimeout,
		http: &http.Client{
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}

	for _, option := range options {
		if option == nil {
			return nil, fmt.Errorf("nil client option")
		}

		if err := option(&c); err != nil {
			return nil, err
		}
	}

	if err := validateServiceURL("base", c.baseURL); err != nil {
		return nil, err
	}

	if err := validateServiceURL("transaction", c.transactionURL); err != nil {
		return nil, err
	}

	if c.timeout <= 0 || c.http == nil {
		return nil, fmt.Errorf("positive timeout and HTTP client required")
	}

	t := c.transport(c.baseURL)

	return &Client{
		Swap:        swap.NewClient(t, c.referral),
		Tokens:      tokens.NewClient(t),
		Price:       price.NewClient(t),
		Trigger:     trigger.NewClient(t),
		Lend:        lend.NewClient(t),
		Prediction:  prediction.NewClient(t),
		Send:        send.NewClient(t),
		Studio:      studio.NewClient(t),
		Transaction: transaction.NewClient(c.transport(c.transactionURL)),
	}, nil
}

// transport builds a transport for one service host with the shared settings.
func (c config) transport(baseURL string) *transport.Client {
	return &transport.Client{
		BaseURL:        strings.TrimRight(baseURL, "/"),
		APIKey:         c.apiKey,
		HTTP:           c.http,
		Timeout:        c.timeout,
		DisableRetries: c.disableRetries,
	}
}

func validateServiceURL(name string, raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("parse %s URL: %w", name, err)
	}

	unsupportedScheme := u.Scheme != schemeHTTPS && u.Scheme != schemeHTTP
	if u.Host == "" || unsupportedScheme || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("invalid %s URL", name)
	}

	return nil
}
