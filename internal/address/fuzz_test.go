package address

import (
	"testing"

	"github.com/mr-tron/base58"
)

// FuzzValidate checks that arbitrary input never panics and that every accepted
// address decodes to exactly PublicKeyBytes and re-encodes to itself, which is
// what lets validated addresses be compared as strings.
func FuzzValidate(f *testing.F) {
	f.Add("So11111111111111111111111111111111111111112")
	f.Add("11111111111111111111111111111111")
	f.Add("")
	f.Add("0OIl")

	f.Fuzz(func(t *testing.T, value string) {
		if err := Validate(value); err != nil {
			return
		}

		decoded, err := base58.Decode(value)
		if err != nil {
			t.Fatalf("accepted %q but it does not decode: %v", value, err)
		}

		if len(decoded) != PublicKeyBytes || base58.Encode(decoded) != value {
			t.Fatalf("accepted %q is not a canonical %d-byte key", value, PublicKeyBytes)
		}
	})
}
