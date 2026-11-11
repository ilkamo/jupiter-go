// Command managed requests a managed order, signs it locally, and optionally executes it.
//
// It spends real funds only when JUPITER_EXECUTE_SWAP=yes; otherwise it stops after
// signing and prints the transaction ID.
package main

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gagliardetto/solana-go"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/solanatx"
	"github.com/ilkamo/jupiter-go/swap"
)

const (
	operationTimeout    = 30 * time.Second
	wrappedSOLMint      = "So11111111111111111111111111111111111111112"
	usdcMint            = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	swapAmountLamports  = "10000000"
	executeConfirmation = "yes"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	wallet, err := solana.PrivateKeyFromBase58(os.Getenv("SOLANA_PRIVATE_KEY"))
	if err != nil {
		return fmt.Errorf("load SOLANA_PRIVATE_KEY: %w", err)
	}

	client, err := jupiter.NewClient(
		jupiter.WithAPIKey(os.Getenv("JUPITER_API_KEY")),
	)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	order, err := client.Swap.Order(ctx, swap.OrderRequest{
		InputMint:  wrappedSOLMint,
		OutputMint: usdcMint,
		Amount:     swapAmountLamports,
		Taker:      wallet.PublicKey().String(),
	})
	if err != nil {
		return fmt.Errorf("order: %w", err)
	}

	if order.Transaction == nil || *order.Transaction == "" {
		return fmt.Errorf("order %s returned no transaction", order.RequestID)
	}

	tx, err := solanatx.Parse(*order.Transaction)
	if err != nil {
		return fmt.Errorf("parse order transaction: %w", err)
	}

	if err := tx.Sign(ed25519.PrivateKey(wallet)); err != nil {
		return fmt.Errorf("sign: %w", err)
	}

	// Gasless and RFQ orders can name another fee payer or cosigner, whose signature
	// Jupiter adds during execution. The transaction ID is then unknown until execution.
	transactionID := tx.ID()
	if transactionID == "" {
		transactionID = "(assigned when the fee payer signs)"
	}

	_, err = fmt.Printf(
		"Signed %s transaction %s for %s raw output units (fee %s bps); awaiting: %v\n",
		tx.Version(),
		transactionID,
		order.OutAmount,
		order.FeeBps,
		tx.MissingSigners(),
	)
	if err != nil {
		return err
	}

	if os.Getenv("JUPITER_EXECUTE_SWAP") != executeConfirmation {
		_, err = fmt.Println("Dry run: set JUPITER_EXECUTE_SWAP=yes to execute.")

		return err
	}

	result, err := client.Swap.Execute(ctx, swap.ExecuteRequest{
		RequestID:         order.RequestID,
		SignedTransaction: tx.Base64(),
	})
	if errors.Is(err, swap.ErrExecutionOutcomeUnknown) {
		return fmt.Errorf(
			"check transaction %s and wallet activity before ordering again: %w",
			transactionID,
			err,
		)
	}

	if err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	_, err = fmt.Printf(
		"Executed %s: %s in, %s out\n",
		result.Signature,
		result.TotalInputAmount,
		result.TotalOutputAmount,
	)

	return err
}
