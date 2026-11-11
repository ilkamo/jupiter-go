package amount

import (
	"strings"
	"testing"
)

// FuzzParseRaw checks that arbitrary input never panics and that accepted amounts
// come back canonical: positive digits without leading zeros, stable on re-parse.
func FuzzParseRaw(f *testing.F) {
	for _, seed := range []string{"1", "00123", "+5", "-1", "1.5", "1e3", "18446744073709551616", ""} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		parsed, err := ParseRaw(value)
		if err != nil {
			return
		}

		if parsed == "" || parsed[0] == '0' || strings.Trim(parsed, "0123456789") != "" {
			t.Fatalf("ParseRaw(%q) = %q is not a canonical positive integer", value, parsed)
		}

		again, err := ParseRaw(parsed)
		if err != nil || again != parsed {
			t.Fatalf("canonical form %q is not stable: %q, %v", parsed, again, err)
		}
	})
}
