package swap

import (
	"fmt"

	"github.com/ilkamo/jupiter-go/internal/address"
)

// ErrInvalidAddress identifies request fields that are not Solana account addresses.
var ErrInvalidAddress = address.ErrInvalid

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
