package gojup_test

import (
	"context"
	"errors"
	"testing"

	"github.com/test-go/testify/require"

	gojup "github.com/ilkamo/go-jup"
	"github.com/ilkamo/go-jup/openapi"
)

var testSwapResponse = &openapi.SwapResponse{
	SwapTransaction:      testTx,
	LastValidBlockHeight: 1000,
}

var testQuoteResponse = &openapi.QuoteResponse{
	InAmount:  "100000000",
	OutAmount: "200000000",
}

type jupApiMock struct {
	shouldFail bool
}

func (j jupApiMock) GetQuoteWithResponse(
	_ context.Context,
	_ *openapi.GetQuoteParams,
	_ ...openapi.RequestEditorFn,
) (*openapi.GetQuoteResponse, error) {
	if j.shouldFail {
		return nil, errors.New("mocked error")
	}

	return &openapi.GetQuoteResponse{
		JSON200: testQuoteResponse,
	}, nil
}

func (j jupApiMock) PostSwapWithResponse(
	_ context.Context,
	_ openapi.PostSwapJSONRequestBody,
	_ ...openapi.RequestEditorFn,
) (*openapi.PostSwapResponse, error) {
	if j.shouldFail {
		return nil, errors.New("mocked error")
	}

	return &openapi.PostSwapResponse{
		JSON200: testSwapResponse,
	}, nil
}

func TestJupClient_GetQuote(t *testing.T) {
	slippage := 500

	t.Run("retrieve valid quote", func(t *testing.T) {
		jup := jupApiMock{}
		client, err := gojup.NewJupClient("", gojup.WithJupAPI(jup))
		require.NoError(t, err)

		quote, err := client.GetQuote(context.Background(), openapi.GetQuoteParams{
			InputMint:   "So11111111111111111111111111111111111111112",
			OutputMint:  "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			Amount:      100000000,
			SlippageBps: &slippage,
		})
		require.NoError(t, err)

		require.Equal(t, "100000000", quote.InAmount)
		require.Equal(t, "200000000", quote.OutAmount)
	})

	t.Run("quote error", func(t *testing.T) {
		jup := jupApiMock{
			shouldFail: true,
		}
		client, err := gojup.NewJupClient("", gojup.WithJupAPI(jup))
		require.NoError(t, err)

		_, err = client.GetQuote(context.Background(), openapi.GetQuoteParams{
			InputMint:   "So11111111111111111111111111111111111111112",
			OutputMint:  "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			Amount:      100000000,
			SlippageBps: &slippage,
		})
		require.Error(t, err)
	})
}

func TestJupClient_PostSwap(t *testing.T) {
	t.Run("retrieve valid quote", func(t *testing.T) {
		jup := jupApiMock{}
		client, err := gojup.NewJupClient("", gojup.WithJupAPI(jup))
		require.NoError(t, err)

		quote, err := client.PostSwap(context.Background(), openapi.PostSwapJSONRequestBody{
			QuoteResponse: *testQuoteResponse,
		})
		require.NoError(t, err)

		require.Equal(t, testTx, quote.SwapTransaction)
		require.Equal(t, float32(1000), quote.LastValidBlockHeight)
	})

	t.Run("swap error", func(t *testing.T) {
		jup := jupApiMock{
			shouldFail: true,
		}
		client, err := gojup.NewJupClient("", gojup.WithJupAPI(jup))
		require.NoError(t, err)

		_, err = client.PostSwap(context.Background(), openapi.PostSwapJSONRequestBody{
			QuoteResponse: *testQuoteResponse,
		})
		require.Error(t, err)
	})
}
