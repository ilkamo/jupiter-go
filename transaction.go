package gojup

import (
	"encoding/base64"
	"fmt"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
)

// NewTransactionFromBase64 deserializes a transaction from a base64 string.
func NewTransactionFromBase64(txStr string) (solana.Transaction, error) {
	txBytes, err := base64.StdEncoding.DecodeString(txStr)
	if err != nil {
		return solana.Transaction{}, fmt.Errorf("could not decode transaction: %w", err)
	}

	tx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(txBytes))
	if err != nil {
		return solana.Transaction{}, fmt.Errorf("could not deserialize transaction: %w", err)
	}

	return *tx, nil
}
