package solana

import (
	"fmt"

	"github.com/gagliardetto/solana-go"
)

// Wallet is a wrapper around a solana.Wallet
type Wallet struct {
	*solana.Wallet
}

func NewWalletFromPrivateKeyBase58(privateKey string) (Wallet, error) {
	w, err := solana.WalletFromPrivateKeyBase58(privateKey)
	if err != nil {
		return Wallet{}, err
	}

	return Wallet{w}, nil
}

// SignTransaction signs a transaction with the wallet's private key.
func (w Wallet) SignTransaction(tx solana.Transaction) (solana.Transaction, error) {
	txMessageBytes, err := tx.Message.MarshalBinary()
	if err != nil {
		return solana.Transaction{}, fmt.Errorf("could not serialize transaction: %w", err)
	}

	signature, err := w.PrivateKey.Sign(txMessageBytes)
	if err != nil {
		return solana.Transaction{}, fmt.Errorf("could not sign transaction: %w", err)
	}

	tx.Signatures = []solana.Signature{signature}

	return tx, nil
}
