// Command lend lists Jupiter Lend Earn tokens and Borrow vaults and, when
// WALLET_ADDRESS is set, builds an unsigned Earn deposit for that wallet. It never
// signs or submits a transaction.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/lend"
)

const (
	operationTimeout = 20 * time.Second
	usdcMint         = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	depositAmount    = "1000000"
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

	tokens, err := client.Lend.Earn.Tokens(ctx)
	if err != nil {
		return fmt.Errorf("earn tokens: %w", err)
	}

	for _, token := range tokens.Tokens {
		// Rates use a 1e4 scale: 500 means 5%.
		_, err := fmt.Printf(
			"Earn %s for %s: total rate %s (1e4 scale)\n",
			token.Symbol,
			token.Asset.Symbol,
			token.TotalRate,
		)
		if err != nil {
			return err
		}
	}

	vaults, err := client.Lend.Borrow.Vaults(ctx, lend.VaultsRequest{})
	if err != nil {
		return fmt.Errorf("borrow vaults: %w", err)
	}

	for _, vault := range vaults.Vaults {
		_, err := fmt.Printf(
			"Vault %d: %s collateral, %s debt, collateral factor %s (1e3 scale)\n",
			vault.ID,
			vault.SupplyToken.Symbol,
			vault.BorrowToken.Symbol,
			vault.CollateralFactor,
		)
		if err != nil {
			return err
		}
	}

	wallet := os.Getenv("WALLET_ADDRESS")
	if wallet == "" {
		return nil
	}

	deposit, err := client.Lend.Earn.Deposit(ctx, lend.AmountRequest{
		Asset:  usdcMint,
		Signer: wallet,
		Amount: depositAmount,
	})
	if err != nil {
		return fmt.Errorf("build deposit: %w", err)
	}

	_, err = fmt.Printf("Unsigned deposit transaction: %d base64 characters\n", len(deposit.Transaction))

	return err
}
