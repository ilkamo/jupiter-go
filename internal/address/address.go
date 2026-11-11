// Package address validates Solana account addresses sent to Jupiter APIs.
package address

import (
	"errors"
	"fmt"

	"github.com/mr-tron/base58"
)

// PublicKeyBytes is the decoded size of a Solana account address.
const PublicKeyBytes = 32

var ErrInvalid = errors.New("invalid Solana address")

// Validate reports whether value is the base58 encoding of a 32-byte account address.
// Base58 encodings are unique, so validated addresses can be compared as strings.
// Curve membership is not checked: program-derived accounts, including associated
// token accounts, are valid addresses that lie off the ed25519 curve.
func Validate(value string) error {
	decoded, err := base58.Decode(value)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalid, err)
	}

	if len(decoded) != PublicKeyBytes {
		return fmt.Errorf(
			"%w: decoded to %d bytes, want %d",
			ErrInvalid,
			len(decoded),
			PublicKeyBytes,
		)
	}

	return nil
}
