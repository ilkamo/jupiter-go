package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ilkamo/jupiter-go/solana"
)

func main() {
	txID := "2JehHv9a31i2PS1hpkTfXJs7u4ydPmUgkAgecSgYJj8YNJjMjsmLcTHco6grmecyj4kzuz4NoBe11CiF5LhYRS5i"

	monitor, err := solana.NewMonitor(
		"wss://api.mainnet-beta.solana.com",
	)
	if err != nil {
		panic(err)
	}

	// Set a timeout for the context so that the program doesn't hang indefinitely.
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	resp, err := monitor.WaitForCommitmentStatus(
		ctx,
		solana.TxID(txID),
		solana.CommitmentFinalized,
	)
	if err != nil {
		panic(err)
	}

	if resp.InstructionErr != nil {
		fmt.Printf("finalized with instruction error: %s\n", resp.InstructionErr)
	}
}
