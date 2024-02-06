package gojup

import (
	"context"
	"fmt"

	"github.com/ilkamo/go-jup/openapi"
)

const (
	defaultAPIURL = "https://quote-api.jup.ag/v6"
)

type JupAPI interface {
	GetQuoteWithResponse(
		ctx context.Context,
		params *openapi.GetQuoteParams,
		reqEditors ...openapi.RequestEditorFn,
	) (*openapi.GetQuoteResponse, error)
	PostSwapWithResponse(
		ctx context.Context,
		body openapi.PostSwapJSONRequestBody,
		reqEditors ...openapi.RequestEditorFn,
	) (*openapi.PostSwapResponse, error)
}

// JupClient is a wrapper client for the Jupiter API
type JupClient struct {
	jup JupAPI
}

// NewJupClient creates a new client for the Jupiter API
func NewJupClient(apiURL string, option ...JupClientOption) (JupClient, error) {
	if apiURL == "" {
		apiURL = defaultAPIURL
	}

	c := &JupClient{}

	for _, opt := range option {
		if err := opt(c); err != nil {
			return JupClient{}, fmt.Errorf("could not apply option: %w", err)
		}
	}

	if c.jup == nil {
		jup, err := openapi.NewClientWithResponses(apiURL)
		if err != nil {
			return JupClient{}, fmt.Errorf("could not create Jupiter client: %w", err)
		}
		c.jup = jup
	}

	return *c, nil
}

// JupClientOption is a function that allows to specify options for the client
type JupClientOption func(*JupClient) error

// WithJupAPI sets the Jupiter API for the client
func WithJupAPI(jup JupAPI) JupClientOption {
	return func(c *JupClient) error {
		c.jup = jup
		return nil
	}
}

// GetQuote requests a quote from the Jupiter API
func (c *JupClient) GetQuote(
	ctx context.Context,
	quoteParams openapi.GetQuoteParams,
) (openapi.QuoteResponse, error) {
	resp, err := c.jup.GetQuoteWithResponse(ctx, &quoteParams)
	if err != nil {
		return openapi.QuoteResponse{}, fmt.Errorf("could not get quote: %w", err)
	}

	if resp.JSON200 == nil {
		return openapi.QuoteResponse{}, fmt.Errorf("got nil response")
	}

	return *resp.JSON200, nil
}

// PostSwap requests a swap from the Jupiter API
func (c *JupClient) PostSwap(
	ctx context.Context,
	swapParams openapi.PostSwapJSONRequestBody,
) (openapi.SwapResponse, error) {
	resp, err := c.jup.PostSwapWithResponse(ctx, swapParams)
	if err != nil {
		return openapi.SwapResponse{}, fmt.Errorf("could not post swap: %w", err)
	}

	if resp.JSON200 == nil {
		return openapi.SwapResponse{}, fmt.Errorf("got nil response")
	}

	return *resp.JSON200, nil
}
