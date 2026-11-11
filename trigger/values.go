package trigger

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/url"
)

const (
	TriggerAbove = "above"
	TriggerBelow = "below"

	DCATimeBased        = "time_based"
	DCAPriceConditional = "price_conditional"

	HistoryActive = "active"
	HistoryPast   = "past"

	SortUpdatedAt  = "updated_at"
	SortCreatedAt  = "created_at"
	SortExpiresAt  = "expires_at"
	SortNextFillAt = "next_fill_at"

	DirectionAscending  = "asc"
	DirectionDescending = "desc"
)

const (
	BasisPointsPerWhole = 10000
	MinTrailingBps      = 50
	MaxTrailingBps      = 9000
	MinHistoryLimit     = 1
	MaxHistoryLimit     = 100

	decimalBase = 10
)

// parseUSD validates a positive decimal JSON number, such as a USD price. json.Valid
// rejects fractions, hexadecimal, and special values that big.Rat would accept;
// big.Rat rejects JSON values that are not numbers.
func parseUSD(field string, value json.Number) (*big.Rat, error) {
	if !json.Valid([]byte(value)) {
		return nil, fmt.Errorf("%s: %q is not a JSON number", field, value)
	}

	price, ok := new(big.Rat).SetString(string(value))
	if !ok || price.Sign() <= 0 {
		return nil, fmt.Errorf("%s must be a positive number", field)
	}

	return price, nil
}

// parseOptionalUSD returns nil for an unset price.
func parseOptionalUSD(field string, value json.Number) (*big.Rat, error) {
	if value == "" {
		return nil, nil
	}

	return parseUSD(field, value)
}

func validateSlippage(field string, value *uint16) error {
	if value != nil && *value > BasisPointsPerWhole {
		return fmt.Errorf("%s exceeds %d bps", field, BasisPointsPerWhole)
	}

	return nil
}

func validateTrailing(value *uint16) error {
	if value != nil && (*value < MinTrailingBps || *value > MaxTrailingBps) {
		return fmt.Errorf("trailing distance must be between %d and %d bps", MinTrailingBps, MaxTrailingBps)
	}

	return nil
}

func validateSignedTransaction(field string, value string) error {
	if value == "" {
		return fmt.Errorf("%s required", field)
	}

	if _, err := base64.StdEncoding.DecodeString(value); err != nil {
		return fmt.Errorf("%s: %w", field, err)
	}

	return nil
}

func orderPath(prefix string, id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("order ID required")
	}

	return basePath + prefix + url.PathEscape(id), nil
}
