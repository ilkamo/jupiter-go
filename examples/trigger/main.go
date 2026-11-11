// Command trigger authenticates a wallet with the Trigger V2 API, resolves its
// vault, and crafts (but does not submit) a deposit.
//
// Trigger V2 is custodial: deposited funds sit in a Jupiter-managed vault until
// orders fill or are cancelled. This example moves no funds. It registers a vault
// only when JUPITER_REGISTER_VAULT=yes.
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
	"github.com/ilkamo/jupiter-go/trigger"
)

const (
	operationTimeout     = 30 * time.Second
	wrappedSOLMint       = "So11111111111111111111111111111111111111112"
	usdcMint             = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	depositAmount        = "100000000"
	registerConfirmation = "yes"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() (resultErr error) {
	wallet, err := solana.PrivateKeyFromBase58(os.Getenv("SOLANA_PRIVATE_KEY"))
	if err != nil {
		return fmt.Errorf("load SOLANA_PRIVATE_KEY: %w", err)
	}

	client, err := jupiter.NewClient(jupiter.WithAPIKey(os.Getenv("JUPITER_API_KEY")))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	session, err := client.Trigger.Login(ctx, ed25519.PrivateKey(wallet), trigger.SessionOptions{})
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}

	defer func() {
		if err := session.Logout(ctx); err != nil {
			resultErr = errors.Join(resultErr, err)
		}
	}()

	vault, err := session.Vault(ctx)
	if errors.Is(err, trigger.ErrNoVault) {
		if os.Getenv("JUPITER_REGISTER_VAULT") != registerConfirmation {
			_, err = fmt.Println("No vault: set JUPITER_REGISTER_VAULT=yes to register one.")

			return err
		}

		vault, err = session.EnsureVault(ctx)
	}

	if err != nil {
		return fmt.Errorf("vault: %w", err)
	}

	if _, err := fmt.Printf("Wallet %s uses vault %s\n", session.Wallet(), vault.VaultPubkey); err != nil {
		return err
	}

	deposit, err := session.CraftDeposit(ctx, trigger.DepositRequest{
		InputMint:    wrappedSOLMint,
		OutputMint:   usdcMint,
		Amount:       depositAmount,
		OrderType:    trigger.OrderTypePrice,
		OrderSubType: trigger.OrderSubTypeSingle,
	})
	if err != nil {
		return fmt.Errorf("craft deposit: %w", err)
	}

	_, err = fmt.Printf(
		"Crafted deposit %s of %s raw units to %s; not signed or submitted\n",
		deposit.RequestID,
		deposit.Amount,
		deposit.ReceiverAddress,
	)

	return err
}
