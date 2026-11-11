// Command build requests swap instructions, assembles and signs a transaction
// locally, and optionally submits it through a Solana RPC node, or through
// Jupiter's landing service (tx.jup.ag) when JUPITER_SUBMIT_VIA=tx.
//
// It spends real funds only when JUPITER_EXECUTE_SWAP=yes; otherwise it stops after
// signing. Build transactions cannot use Jupiter's execute endpoint. Submitting
// through tx.jup.ag adds the required 0.001 SOL tip; confirmation always uses
// SOLANA_RPC_URL because the landing service is send-only.
package main

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/solanatx"
	"github.com/ilkamo/jupiter-go/swap"
	"github.com/ilkamo/jupiter-go/transaction"
)

const (
	operationTimeout    = 90 * time.Second
	confirmationPoll    = 2 * time.Second
	wrappedSOLMint      = "So11111111111111111111111111111111111111112"
	usdcMint            = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	swapAmountLamports  = "10000000"
	executeConfirmation = "yes"
	submitViaJupiter    = "tx"
)

var errBlockhashExpired = errors.New("blockhash expired before confirmation")

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
		jupiter.WithoutMaintainerFee(),
	)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	taker := wallet.PublicKey().String()
	viaJupiter := os.Getenv("JUPITER_SUBMIT_VIA") == submitViaJupiter

	request := swap.BuildRequest{
		InputMint:  wrappedSOLMint,
		OutputMint: usdcMint,
		Amount:     swapAmountLamports,
		Taker:      taker,
	}
	if viaJupiter {
		// tx.jup.ag rejects transactions without the Jupiter tip.
		request.TipAmount = strconv.Itoa(transaction.MinTipLamports)
	}

	build, err := client.Swap.Build(ctx, request)
	if err != nil {
		return fmt.Errorf("build: %w", err)
	}

	tx, err := solanatx.Assemble(build, taker)
	if err != nil {
		return fmt.Errorf("assemble: %w", err)
	}

	if err := tx.Sign(ed25519.PrivateKey(wallet)); err != nil {
		return fmt.Errorf("sign: %w", err)
	}

	if missing := tx.MissingSigners(); len(missing) > 0 {
		return fmt.Errorf("transaction still needs signatures from %v", missing)
	}

	_, err = fmt.Printf(
		"Signed %s transaction %s for %s raw output units (minimum %s)\n",
		tx.Version(),
		tx.ID(),
		build.OutAmount,
		build.OtherAmountThreshold,
	)
	if err != nil {
		return err
	}

	if os.Getenv("JUPITER_EXECUTE_SWAP") != executeConfirmation {
		_, err = fmt.Println("Dry run: set JUPITER_EXECUTE_SWAP=yes and SOLANA_RPC_URL to submit.")

		return err
	}

	rpcURL := os.Getenv("SOLANA_RPC_URL")
	if rpcURL == "" {
		return fmt.Errorf("SOLANA_RPC_URL is required to submit")
	}

	if build.BlockhashWithMetadata.LastValidBlockHeight == nil {
		return fmt.Errorf("build response has no last valid block height")
	}

	node := rpc.New(rpcURL)
	signature, err := submit(ctx, client, node, tx, viaJupiter)
	if err != nil {
		return fmt.Errorf("check transaction %s before retrying; submission failed: %w", tx.ID(), err)
	}

	lastValidBlockHeight := *build.BlockhashWithMetadata.LastValidBlockHeight
	if err := awaitConfirmation(ctx, node, signature, lastValidBlockHeight); err != nil {
		return fmt.Errorf("transaction %s: %w", signature, err)
	}

	_, err = fmt.Printf("Confirmed %s\n", signature)

	return err
}

// submit sends the signed transaction through tx.jup.ag or the RPC node.
func submit(
	ctx context.Context,
	client *jupiter.Client,
	node *rpc.Client,
	tx *solanatx.Transaction,
	viaJupiter bool,
) (solana.Signature, error) {
	if !viaJupiter {
		return node.SendRawTransactionWithOpts(ctx, tx.Bytes(), rpc.TransactionOpts{
			PreflightCommitment: rpc.CommitmentConfirmed,
		})
	}

	sent, err := client.Transaction.Send(ctx, tx.Base64(), transaction.SendOptions{})
	if err != nil {
		return solana.Signature{}, err
	}

	return solana.SignatureFromBase58(sent.Signature)
}

// awaitConfirmation polls until the transaction is confirmed, fails, or can no longer land.
func awaitConfirmation(
	ctx context.Context,
	node *rpc.Client,
	signature solana.Signature,
	lastValidBlockHeight uint64,
) error {
	ticker := time.NewTicker(confirmationPoll)
	defer ticker.Stop()

	for {
		statuses, err := node.GetSignatureStatuses(ctx, false, signature)
		if err != nil {
			return fmt.Errorf("signature status: %w", err)
		}

		if len(statuses.Value) > 0 && statuses.Value[0] != nil {
			status := statuses.Value[0]
			if status.Err != nil {
				return fmt.Errorf("transaction failed: %v", status.Err)
			}

			switch status.ConfirmationStatus {
			case rpc.ConfirmationStatusConfirmed, rpc.ConfirmationStatusFinalized:
				return nil
			}
		}

		height, err := node.GetBlockHeight(ctx, rpc.CommitmentConfirmed)
		if err != nil {
			return fmt.Errorf("block height: %w", err)
		}

		if height > lastValidBlockHeight {
			return errBlockhashExpired
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("outcome unknown: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}
