// Command build requests swap instructions, assembles and signs a transaction
// locally, and optionally submits it through a Solana RPC node, or through
// Jupiter's landing service (tx.jup.ag) when JUPITER_SUBMIT_VIA=tx.
//
// It spends real funds only when JUPITER_EXECUTE_SWAP=yes; otherwise it stops after
// signing. JUPITER_SIMULATE=yes instead simulates the signed transaction on
// SOLANA_RPC_URL, which verifies the signatures and runs it against current
// mainnet state without sending it. Build transactions cannot use Jupiter's execute endpoint. Submitting
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
	operationTimeout     = 90 * time.Second
	confirmationPoll     = 2 * time.Second
	wrappedSOLMint       = "So11111111111111111111111111111111111111112"
	usdcMint             = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	defaultSwapLamports  = "10000000"
	executeConfirmation  = "yes"
	simulateConfirmation = "yes"
	simulationLogLines   = 5
	decimalBase          = 10
	submitViaJupiter     = "tx"
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
		Amount:     swapAmount(),
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

	if os.Getenv("JUPITER_SIMULATE") == simulateConfirmation {
		return simulate(ctx, tx)
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

// simulate runs the signed transaction on the RPC node without sending it. An
// unfunded wallet fails with an insufficient-funds error, which still shows that
// the transaction was well formed and correctly signed.
func simulate(ctx context.Context, tx *solanatx.Transaction) error {
	rpcURL := os.Getenv("SOLANA_RPC_URL")
	if rpcURL == "" {
		return fmt.Errorf("SOLANA_RPC_URL is required to simulate")
	}

	options := &rpc.SimulateTransactionOpts{
		SigVerify:  true,
		Commitment: rpc.CommitmentConfirmed,
	}

	result, err := rpc.New(rpcURL).SimulateRawTransactionWithOpts(ctx, tx.Bytes(), options)
	if err != nil {
		return fmt.Errorf("simulate: %w", err)
	}

	simulation := result.Value
	units := "unknown"
	if simulation.UnitsConsumed != nil {
		units = strconv.FormatUint(*simulation.UnitsConsumed, decimalBase)
	}

	outcome := "succeeded"
	if simulation.Err != nil {
		outcome = fmt.Sprintf("failed: %v", simulation.Err)
	}

	if _, err := fmt.Printf("Simulation %s, %s compute units\n", outcome, units); err != nil {
		return err
	}

	logs := simulation.Logs[max(0, len(simulation.Logs)-simulationLogLines):]
	for _, line := range logs {
		if _, err := fmt.Println("  " + line); err != nil {
			return err
		}
	}

	return nil
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

// awaitConfirmation polls until the transaction is confirmed, fails, or can no longer
// land. SOLANA_RPC_URL must report the true block height: some providers return the
// slot from getBlockHeight, which makes a fresh blockhash look expired.
func awaitConfirmation(
	ctx context.Context,
	node *rpc.Client,
	signature solana.Signature,
	lastValidBlockHeight uint64,
) error {
	ticker := time.NewTicker(confirmationPoll)
	defer ticker.Stop()

	for {
		confirmed, err := checkConfirmed(ctx, node, signature)
		if err != nil || confirmed {
			return err
		}

		height, err := node.GetBlockHeight(ctx, rpc.CommitmentConfirmed)
		if err != nil {
			return fmt.Errorf("block height: %w", err)
		}

		if height > lastValidBlockHeight {
			// The transaction may have landed since the last check.
			confirmed, err := checkConfirmed(ctx, node, signature)
			if err != nil || confirmed {
				return err
			}

			return errBlockhashExpired
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("outcome unknown: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

// checkConfirmed reports whether the transaction reached confirmed commitment, or
// an error if it failed on-chain.
func checkConfirmed(ctx context.Context, node *rpc.Client, signature solana.Signature) (bool, error) {
	statuses, err := node.GetSignatureStatuses(ctx, false, signature)
	if err != nil {
		return false, fmt.Errorf("signature status: %w", err)
	}

	if len(statuses.Value) == 0 || statuses.Value[0] == nil {
		return false, nil
	}

	status := statuses.Value[0]
	if status.Err != nil {
		return false, fmt.Errorf("transaction failed: %v", status.Err)
	}

	switch status.ConfirmationStatus {
	case rpc.ConfirmationStatusConfirmed, rpc.ConfirmationStatusFinalized:
		return true, nil
	default:
		return false, nil
	}
}

// swapAmount returns SWAP_AMOUNT_LAMPORTS, or 0.01 SOL when it is unset.
func swapAmount() string {
	if amount := os.Getenv("SWAP_AMOUNT_LAMPORTS"); amount != "" {
		return amount
	}

	return defaultSwapLamports
}
