// Command tokens searches for a token and prints USD prices without signing
// anything. With EXPRESS_TOKEN_ID set, it also checks whether that token can be
// submitted for paid Express Verification; it never pays.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jupiter "github.com/ilkamo/jupiter-go"
)

const (
	operationTimeout = 15 * time.Second
	searchQuery      = "JUP"
	wrappedSOLMint   = "So11111111111111111111111111111111111111112"
	usdcMint         = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	client, err := jupiter.NewClient(jupiter.WithAPIKey(os.Getenv("JUPITER_API_KEY")))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	found, err := client.Tokens.Search(ctx, searchQuery)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	for _, token := range found.Tokens {
		verified := token.IsVerified != nil && *token.IsVerified
		_, err := fmt.Printf("%s %s (%s) verified=%t\n", token.Symbol, token.Name, token.ID, verified)
		if err != nil {
			return err
		}
	}

	prices, err := client.Price.Get(ctx, []string{wrappedSOLMint, usdcMint})
	if err != nil {
		return fmt.Errorf("price: %w", err)
	}

	for mint, price := range prices.Prices {
		if _, err := fmt.Printf("%s: $%s\n", mint, price.USDPrice); err != nil {
			return err
		}
	}

	if len(prices.Missing) > 0 {
		if _, err := fmt.Printf("No reliable price for %v\n", prices.Missing); err != nil {
			return err
		}
	}

	tokenID := os.Getenv("EXPRESS_TOKEN_ID")
	if tokenID == "" {
		return nil
	}

	eligibility, err := client.Tokens.Express.CheckEligibility(ctx, tokenID)
	if err != nil {
		return fmt.Errorf("express eligibility: %w", err)
	}

	_, err = fmt.Printf(
		"Express verification: can verify %t (%s), can update metadata %t (%s)\n",
		eligibility.CanVerify,
		eligibility.VerificationError,
		eligibility.CanMetadata,
		eligibility.MetadataError,
	)

	return err
}
