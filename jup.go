package gojup

import (
	"context"
	"fmt"

	"github.com/ilkamo/go-jup/openapi"
)

const (
	defaultAPIURL = "https://quote-api.jup.ag/v6"
)

// JupiterClient is a wrapper client for the Jupiter API
type JupiterClient struct {
	jup *openapi.ClientWithResponses
}

// NewJupiterClient creates a new client for the Jupiter API
func NewJupiterClient(apiURL string) (JupiterClient, error) {
	if apiURL == "" {
		apiURL = defaultAPIURL
	}

	jup, err := openapi.NewClientWithResponses(apiURL)
	if err != nil {
		return JupiterClient{}, fmt.Errorf("could not create Jupiter client: %w", err)
	}

	c := &JupiterClient{
		jup: jup,
	}

	return *c, nil
}

// GetQuote requests a quote from the Jupiter API
func (c *JupiterClient) GetQuote(
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
func (c *JupiterClient) PostSwap(
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
