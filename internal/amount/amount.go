// Package amount parses raw token amounts sent to Jupiter APIs.
package amount

import (
	"fmt"
	"math/big"
)

const decimalBase = 10

// ParseRaw accepts a positive base-10 integer in raw token units (no decimals or
// exponent) and returns its canonical form. Values beyond 64 bits are supported.
func ParseRaw(value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("amount required")
	}

	raw, ok := new(big.Int).SetString(value, decimalBase)
	if !ok {
		return "", fmt.Errorf("amount must be a base-10 raw integer")
	}

	if raw.Sign() <= 0 {
		return "", fmt.Errorf("amount must be positive")
	}

	return raw.String(), nil
}
