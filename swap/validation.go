package swap

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ilkamo/jupiter-go/internal/address"
)

// ErrInvalidAddress identifies request fields that are not Solana account addresses.
var ErrInvalidAddress = address.ErrInvalid

const labelSeparator = ","

func validateAddress(field string, value string) error {
	if err := address.Validate(value); err != nil {
		return fmt.Errorf("%s: %w", field, err)
	}

	return nil
}

func validateOptionalAddress(field string, value string) error {
	if value == "" {
		return nil
	}

	return validateAddress(field, value)
}

func validateMintPair(inputMint string, outputMint string) error {
	if err := validateAddress("input mint", inputMint); err != nil {
		return err
	}

	if err := validateAddress("output mint", outputMint); err != nil {
		return err
	}

	if inputMint == outputMint {
		return fmt.Errorf("input and output mints must differ")
	}

	return nil
}

// parseRawAmount accepts positive base-10 integers in raw token units and returns them canonically.
func parseRawAmount(amount string) (string, error) {
	if amount == "" {
		return "", fmt.Errorf("amount required")
	}

	rawAmount, ok := new(big.Int).SetString(amount, decimalBase)
	if !ok {
		return "", fmt.Errorf("amount must be a base-10 raw integer")
	}

	if rawAmount.Sign() <= 0 {
		return "", fmt.Errorf("amount must be positive")
	}

	return rawAmount.String(), nil
}

// validateLabels rejects entries that would change meaning when joined into a comma-separated list.
func validateLabels(field string, labels []string) error {
	for _, label := range labels {
		if label == "" {
			return fmt.Errorf("%s: empty label", field)
		}

		if strings.Contains(label, labelSeparator) {
			return fmt.Errorf("%s: label %q contains a separator", field, label)
		}

		if strings.TrimSpace(label) != label {
			return fmt.Errorf("%s: label %q has surrounding whitespace", field, label)
		}
	}

	return nil
}
