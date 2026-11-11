// Command send generates a Send invite code and builds, but does not sign or
// submit, a transfer of 0.01 SOL from SENDER_ADDRESS to that invite.
//
// The invite code is a secret: whoever holds it can claim the funds once the
// transaction lands. Sign the built transaction with the sender and the invite's
// private key (send.InviteCode.PrivateKey) to fund it.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/send"
)

const (
	operationTimeout = 15 * time.Second
	sendLamports     = "10000000"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	sender := os.Getenv("SENDER_ADDRESS")
	if sender == "" {
		return fmt.Errorf("SENDER_ADDRESS is required")
	}

	client, err := jupiter.NewClient(jupiter.WithAPIKey(os.Getenv("JUPITER_API_KEY")))
	if err != nil {
		return err
	}

	code, err := send.NewInviteCode()
	if err != nil {
		return err
	}

	pda, err := code.PDA()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	built, err := client.Send.CraftSend(ctx, send.SendRequest{
		InviteSigner: code.Signer(),
		Sender:       sender,
		Amount:       sendLamports,
	})
	if err != nil {
		return fmt.Errorf("craft send: %w", err)
	}

	_, err = fmt.Printf(
		"Invite %s (account %s): unsigned send built, fees %s lamports, expires %s\n",
		code,
		pda,
		built.TotalFeeLamports,
		built.Expiry,
	)
	if err != nil {
		return err
	}

	// Nothing was submitted, so this code holds no funds yet.
	_, err = fmt.Printf("Invite code to share with the recipient once funded: %s\n", code.Reveal())

	return err
}
