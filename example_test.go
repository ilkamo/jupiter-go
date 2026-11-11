package jupiter_test

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
	"github.com/ilkamo/jupiter-go/send"
	"github.com/ilkamo/jupiter-go/solanatx"
	"github.com/ilkamo/jupiter-go/swap"
	"github.com/ilkamo/jupiter-go/trigger"
)

const (
	exampleSOL     = "So11111111111111111111111111111111111111112"
	exampleUSDC    = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	exampleTimeout = 15 * time.Second
)

func ExampleNewClient() {
	client, err := jupiter.NewClient(
		jupiter.WithAPIKey(os.Getenv("JUPITER_API_KEY")),
		jupiter.WithTimeout(exampleTimeout),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Construction makes no network calls; product clients are ready to use.
	if _, err := fmt.Println(client.Swap != nil, client.Prediction != nil); err != nil {
		log.Fatal(err)
	}

	// Output: true true
}

// A quote is an order without a taker: no transaction is returned.
func Example_quote() {
	client, err := jupiter.NewClient(jupiter.WithAPIKey(os.Getenv("JUPITER_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}

	quote, err := client.Swap.Order(context.Background(), swap.OrderRequest{
		InputMint:  exampleSOL,
		OutputMint: exampleUSDC,
		Amount:     "100000000",
	})
	if err != nil {
		log.Fatal(err)
	}

	if _, err := fmt.Println("raw USDC out:", quote.OutAmount); err != nil {
		log.Fatal(err)
	}
}

// A managed swap: order with a taker, sign locally, and let Jupiter land it.
func Example_managedSwap() {
	wallet, err := solana.PrivateKeyFromBase58(os.Getenv("SOLANA_PRIVATE_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	client, err := jupiter.NewClient(jupiter.WithAPIKey(os.Getenv("JUPITER_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	order, err := client.Swap.Order(ctx, swap.OrderRequest{
		InputMint:  exampleSOL,
		OutputMint: exampleUSDC,
		Amount:     "10000000",
		Taker:      wallet.PublicKey().String(),
	})
	if err != nil {
		log.Fatal(err)
	}

	tx, err := solanatx.Parse(*order.Transaction)
	if err != nil {
		log.Fatal(err)
	}

	if err := tx.Sign(ed25519.PrivateKey(wallet)); err != nil {
		log.Fatal(err)
	}

	result, err := client.Swap.Execute(ctx, swap.ExecuteRequest{
		RequestID:         order.RequestID,
		SignedTransaction: tx.Base64(),
	})
	if errors.Is(err, swap.ErrExecutionOutcomeUnknown) {
		log.Fatalf("check %s on-chain before ordering again: %v", tx.ID(), err)
	}

	if err != nil {
		log.Fatal(err)
	}

	if _, err := fmt.Println("executed:", result.Signature); err != nil {
		log.Fatal(err)
	}
}

// Trigger V2 operations run on a wallet-scoped session that refreshes itself.
func Example_triggerSession() {
	wallet, err := solana.PrivateKeyFromBase58(os.Getenv("SOLANA_PRIVATE_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	client, err := jupiter.NewClient(jupiter.WithAPIKey(os.Getenv("JUPITER_API_KEY")))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	session, err := client.Trigger.Login(ctx, ed25519.PrivateKey(wallet), trigger.SessionOptions{})
	if err != nil {
		log.Fatal(err)
	}

	history, err := session.PriceOrderHistory(ctx, trigger.HistoryRequest{State: trigger.HistoryActive})
	if err != nil {
		log.Fatal(err)
	}

	if _, err := fmt.Println("active orders:", history.Pagination.Total); err != nil {
		log.Fatal(err)
	}

	if err := session.Logout(ctx); err != nil {
		log.Fatal(err)
	}
}

// Invite codes stay on the client and print redacted; Reveal returns the code.
func Example_sendInviteCode() {
	code, err := send.NewInviteCode()
	if err != nil {
		log.Fatal(err)
	}

	if _, err := fmt.Println(code); err != nil {
		log.Fatal(err)
	}

	// Output: send.InviteCode(redacted)
}
