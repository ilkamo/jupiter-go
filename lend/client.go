// Package lend implements Jupiter Lend's Earn and Borrow REST APIs.
//
// Both APIs are labelled beta upstream; their types may change before the SDK's
// compatibility review. Operations build unsigned transactions or instructions and
// never submit them: sign and send them through your own RPC node. Native SOL is
// not wrapped or unwrapped for Borrow; fund or close the WSOL account yourself.
package lend

import (
	"encoding/json"
	"fmt"

	"github.com/ilkamo/jupiter-go/instruction"
	"github.com/ilkamo/jupiter-go/internal/address"
	"github.com/ilkamo/jupiter-go/internal/transport"
	"github.com/ilkamo/jupiter-go/response"
)

const basePath = "/lend/v1"

// ErrInvalidAddress is the same value as swap.ErrInvalidAddress.
var ErrInvalidAddress = address.ErrInvalid

type Client struct {
	Earn   *EarnClient
	Borrow *BorrowClient
}

// NewClient is called by jupiter.NewClient, which applications use instead; it
// takes the SDK's internal transport.
func NewClient(t *transport.Client) *Client {
	return &Client{
		Earn:   &EarnClient{transport: t},
		Borrow: &BorrowClient{transport: t},
	}
}

// Transaction is a base64-encoded unsigned transaction to sign and submit.
type Transaction struct {
	Transaction string `json:"transaction"`

	Metadata response.Metadata `json:"-"`
}

// ChainID accepts the string or number forms of Earn's chain_id: the schema declares
// a string, while the documented example returns a number. Numbers keep their text.
type ChainID string

// UnmarshalJSON implements json.Unmarshaler for both chain ID forms.
func (c *ChainID) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*c = ChainID(text)

		return nil
	}

	var number json.Number
	if err := json.Unmarshal(data, &number); err != nil {
		return fmt.Errorf("chain ID must be a string or number: %w", err)
	}

	*c = ChainID(number.String())

	return nil
}

func validateAddresses(field string, addresses []string) error {
	if len(addresses) == 0 {
		return fmt.Errorf("%s required", field)
	}

	for i, value := range addresses {
		if err := address.Validate(value); err != nil {
			return fmt.Errorf("%s %d: %w", field, i, err)
		}
	}

	return nil
}

// Instructions and AccountMeta values are shared with swap build responses and
// can be converted with solanatx.ConvertInstruction.
type (
	Instruction = instruction.Instruction
	AccountMeta = instruction.AccountMeta
)
