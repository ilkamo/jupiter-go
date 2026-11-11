package address

import (
	"bytes"
	"errors"
	"testing"

	"github.com/mr-tron/base58"
)

const maxByteValue = 0xff

func TestValidate(t *testing.T) {
	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{name: "wallet", value: "GkwFnmMDvn3HGMpJpWBg8tgJxr3NxNvg3AXxvXVPbRGJ", valid: true},
		{name: "wrapped SOL mint", value: "So11111111111111111111111111111111111111112", valid: true},
		{name: "all zero key", value: base58.Encode(make([]byte, PublicKeyBytes)), valid: true},
		{
			name:  "all max key",
			value: base58.Encode(bytes.Repeat([]byte{maxByteValue}, PublicKeyBytes)),
			valid: true,
		},
		{name: "one byte short", value: base58.Encode(make([]byte, PublicKeyBytes-1))},
		{name: "one byte long", value: base58.Encode(make([]byte, PublicKeyBytes+1))},
		{
			name:  "nonzero one byte long",
			value: base58.Encode(bytes.Repeat([]byte{maxByteValue}, PublicKeyBytes+1)),
		},
		{name: "empty", value: ""},
		{name: "excluded zero digit", value: "0kwFnmMDvn3HGMpJpWBg8tgJxr3NxNvg3AXxvXVPbRGJ"},
		{name: "excluded capital O", value: "OkwFnmMDvn3HGMpJpWBg8tgJxr3NxNvg3AXxvXVPbRGJ"},
		{name: "excluded capital I", value: "IkwFnmMDvn3HGMpJpWBg8tgJxr3NxNvg3AXxvXVPbRGJ"},
		{name: "excluded lowercase l", value: "lkwFnmMDvn3HGMpJpWBg8tgJxr3NxNvg3AXxvXVPbRGJ"},
		{name: "leading whitespace", value: " GkwFnmMDvn3HGMpJpWBg8tgJxr3NxNvg3AXxvXVPbRGJ"},
		{name: "trailing newline", value: "GkwFnmMDvn3HGMpJpWBg8tgJxr3NxNvg3AXxvXVPbRGJ\n"},
		{name: "non-ASCII", value: "GkwFnmMDvn3HGMpJpWBg8tgJxr3NxNvg3AXxvXVPbRGé"},
		{name: "hexadecimal", value: "0x0000000000000000000000000000000000000000000000000000000000000000"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Validate(test.value)
			if test.valid {
				if err != nil {
					t.Fatalf("valid address rejected: %v", err)
				}

				return
			}

			if !errors.Is(err, ErrInvalid) {
				t.Fatalf("expected ErrInvalid, got %v", err)
			}
		})
	}
}
