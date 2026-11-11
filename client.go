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
	"github.com/ilkamo/jupiter-go/price"
	"github.com/ilkamo/jupiter-go/swap"
	"github.com/ilkamo/jupiter-go/tokens"
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
	Swap    *swap.Client
	Tokens  *tokens.Client
	Price   *price.Client
	Trigger *trigger.Client
	Lend    *lend.Client
}

type config struct {
	baseURL, apiKey string
	http            HTTPClient
	timeout         time.Duration
	fee             swap.Referral
	feeConfigured   bool
	disableRetries  bool
}

type Option func(*config) error

// NewClient constructs a client without network calls or a maintainer fee.
// WithTimeout bounds each operation, including its retries. Read-only operations,
// order, and build retry transient failures; Swap execute never retries.
// WithHTTPClient transfers redirect and credential-forwarding responsibility to the caller.
func NewClient(options ...Option) (*Client, error) {
	c := config{
		baseURL: "https://api.jup.ag",
		timeout: defaultRequestTimeout,
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

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}

	unsupportedScheme := u.Scheme != schemeHTTPS && u.Scheme != schemeHTTP
	if u.Host == "" || unsupportedScheme || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("invalid base URL")
	}

	if c.timeout <= 0 || c.http == nil {
		return nil, fmt.Errorf("positive timeout and HTTP client required")
	}

	t := &transport.Client{
		BaseURL:        strings.TrimRight(c.baseURL, "/"),
		APIKey:         c.apiKey,
		HTTP:           c.http,
		Timeout:        c.timeout,
		DisableRetries: c.disableRetries,
	}

	return &Client{
		Swap:    swap.NewClient(t, c.fee),
		Tokens:  tokens.NewClient(t),
		Price:   price.NewClient(t),
		Trigger: trigger.NewClient(t),
		Lend:    lend.NewClient(t),
	}, nil
}
