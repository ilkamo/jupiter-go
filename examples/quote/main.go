package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/swap"
)

const (
	quoteTimeout        = 10 * time.Second
	wrappedSOLMint      = "So11111111111111111111111111111111111111112"
	usdcMint            = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	quoteAmountLamports = "100000000"
)

func main() {
	client, err := jupiter.NewClient(
		jupiter.WithAPIKey(os.Getenv("JUPITER_API_KEY")),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), quoteTimeout)
	defer cancel()

	quote, err := client.Swap.Order(ctx, swap.OrderRequest{
		InputMint:  wrappedSOLMint,
		OutputMint: usdcMint,
		Amount:     quoteAmountLamports,
	})
	if err != nil {
		log.Fatal(err)
	}

	_, err = fmt.Printf(
		"Output: %s raw units; fee: %s bps in %s\n",
		quote.OutAmount,
		quote.FeeBps,
		quote.FeeMint,
	)
	if err != nil {
		log.Fatal(err)
	}
}
