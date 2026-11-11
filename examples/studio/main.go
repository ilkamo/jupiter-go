// Command studio prepares a Jupiter Studio token launch on a Dynamic Bonding Curve
// with the documented "Meme" preset: 16k USDC initial and 69k USDC migration
// market caps, no vesting, 1% fee.
//
// It builds and signs the launch transaction and stops. Only with
// JUPITER_LAUNCH_TOKEN=yes does it upload TOKEN_IMAGE (PNG) and metadata and
// submit, which creates the token on-chain.
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
	"github.com/ilkamo/jupiter-go/studio"
)

const (
	operationTimeout   = 60 * time.Second
	usdcMint           = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	initialMarketCap   = "16000"
	migrationMarketCap = "69000"
	launchConfirmation = "yes"
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

	name, symbol := os.Getenv("TOKEN_NAME"), os.Getenv("TOKEN_SYMBOL")
	if name == "" || symbol == "" {
		return fmt.Errorf("TOKEN_NAME and TOKEN_SYMBOL are required")
	}

	client, err := jupiter.NewClient(jupiter.WithAPIKey(os.Getenv("JUPITER_API_KEY")))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()

	creator := wallet.PublicKey().String()
	pool, err := client.Studio.CreatePool(ctx, studio.PoolRequest{
		Curve: studio.CurveByMarketCap{
			QuoteMint:          usdcMint,
			InitialMarketCap:   initialMarketCap,
			MigrationMarketCap: migrationMarketCap,
			TokenQuoteDecimal:  studio.QuoteDecimals6,
			LockedVesting:      studio.NoLockedVesting(),
		},
		Fee:                   studio.Fee{FeeBps: studio.FeeBps100},
		TokenName:             name,
		TokenSymbol:           symbol,
		TokenImageContentType: studio.ImagePNG,
		Creator:               creator,
	})
	if err != nil {
		return fmt.Errorf("create pool transaction: %w", err)
	}

	tx, err := solanatx.Parse(pool.Transaction)
	if err != nil {
		return fmt.Errorf("parse launch transaction: %w", err)
	}

	if err := tx.Sign(ed25519.PrivateKey(wallet)); err != nil {
		return fmt.Errorf("sign launch transaction: %w", err)
	}

	if _, err := fmt.Printf("Prepared launch of %s (mint %s)\n", symbol, pool.Mint); err != nil {
		return err
	}

	if os.Getenv("JUPITER_LAUNCH_TOKEN") != launchConfirmation {
		_, err := fmt.Println("Dry run: set JUPITER_LAUNCH_TOKEN=yes and TOKEN_IMAGE to launch.")

		return err
	}

	return launch(ctx, client, pool, tx, creator, name, symbol)
}

func launch(
	ctx context.Context,
	client *jupiter.Client,
	pool *studio.PoolTransaction,
	tx *solanatx.Transaction,
	creator string,
	name string,
	symbol string,
) error {
	image, err := os.ReadFile(os.Getenv("TOKEN_IMAGE"))
	if err != nil {
		return fmt.Errorf("read TOKEN_IMAGE: %w", err)
	}

	if _, err := client.Studio.UploadImage(ctx, pool.ImagePresignedURL, studio.ImagePNG, image); err != nil {
		return fmt.Errorf("upload image: %w", err)
	}

	metadata := studio.TokenMetadata{Name: name, Symbol: symbol, Image: pool.ImageURL}
	if _, err := client.Studio.UploadMetadata(ctx, pool.MetadataPresignedURL, metadata); err != nil {
		return fmt.Errorf("upload metadata: %w", err)
	}

	result, err := client.Studio.Submit(ctx, studio.SubmitRequest{
		SignedTransaction: tx.Base64(),
		Owner:             creator,
	})
	if errors.Is(err, studio.ErrOutcomeUnknown) {
		return fmt.Errorf("check pool addresses for %s before submitting again: %w", pool.Mint, err)
	}

	if err != nil {
		return fmt.Errorf("submit: %w", err)
	}

	_, err = fmt.Printf("Launched %s with config %s\n", result.Mint, result.ConfigKey)

	return err
}
