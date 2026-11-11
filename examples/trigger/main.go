// Command trigger authenticates a wallet with the Trigger V2 API, resolves its
// vault, crafts and signs a deposit, and lists active orders.
//
// Trigger V2 is custodial: deposited funds sit in a Jupiter-managed vault until
// orders fill or are cancelled. The example registers a vault only when
// JUPITER_REGISTER_VAULT=yes, and creates a limit order (selling 0.1 SOL for USDC
// when SOL rises above TRIGGER_PRICE_USD) only when JUPITER_CREATE_ORDER=yes.
package main

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gagliardetto/solana-go"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/solanatx"
	"github.com/ilkamo/jupiter-go/trigger"
)

const (
	operationTimeout = 60 * time.Second
	orderValidity    = 7 * 24 * time.Hour
	wrappedSOLMint   = "So11111111111111111111111111111111111111112"
	usdcMint         = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	depositAmount    = "100000000"
	confirmation     = "yes"
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
		if os.Getenv("JUPITER_REGISTER_VAULT") != confirmation {
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

	tx, err := solanatx.Parse(deposit.Transaction)
	if err != nil {
		return fmt.Errorf("parse deposit: %w", err)
	}

	if err := tx.Sign(ed25519.PrivateKey(wallet)); err != nil {
		return fmt.Errorf("sign deposit: %w", err)
	}

	if os.Getenv("JUPITER_CREATE_ORDER") == confirmation {
		if err := createOrder(ctx, session, deposit, tx); err != nil {
			return err
		}
	} else {
		_, err := fmt.Println("Signed the deposit but did not submit it: set JUPITER_CREATE_ORDER=yes.")
		if err != nil {
			return err
		}
	}

	return listActiveOrders(ctx, session)
}

func createOrder(
	ctx context.Context,
	session *trigger.Session,
	deposit *trigger.Deposit,
	tx *solanatx.Transaction,
) error {
	triggerPrice := os.Getenv("TRIGGER_PRICE_USD")
	if triggerPrice == "" {
		return fmt.Errorf("TRIGGER_PRICE_USD is required to create an order")
	}

	order, err := session.CreatePriceOrder(ctx, trigger.PriceOrderRequest{
		OrderType:                trigger.OrderSubTypeSingle,
		DepositRequestID:         deposit.RequestID,
		DepositSignedTransaction: tx.Base64(),
		InputMint:                wrappedSOLMint,
		InputAmount:              deposit.Amount,
		OutputMint:               usdcMint,
		TriggerMint:              wrappedSOLMint,
		TriggerCondition:         trigger.TriggerAbove,
		TriggerPriceUSD:          json.Number(triggerPrice),
		ExpiresAt:                time.Now().Add(orderValidity),
	})
	if errors.Is(err, trigger.ErrOutcomeUnknown) {
		return fmt.Errorf("check active orders before creating another: %w", err)
	}

	if err != nil {
		return fmt.Errorf("create order: %w", err)
	}

	_, err = fmt.Printf(
		"Created order %s (deposit %s, confirmed=%t)\n",
		order.ID,
		order.TxSignature,
		order.DepositConfirmed,
	)

	return err
}

func listActiveOrders(ctx context.Context, session *trigger.Session) error {
	history, err := session.PriceOrderHistory(ctx, trigger.HistoryRequest{State: trigger.HistoryActive})
	if err != nil {
		return fmt.Errorf("order history: %w", err)
	}

	for _, order := range history.Orders {
		_, err := fmt.Printf(
			"%s %s %s at $%s\n",
			order.ID,
			order.OrderState,
			order.TriggerCondition,
			order.TriggerPriceUSD,
		)
		if err != nil {
			return err
		}
	}

	_, err = fmt.Printf("%d active orders\n", history.Pagination.Total)

	return err
}
