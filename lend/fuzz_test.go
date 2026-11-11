package lend

import "testing"

// FuzzParseSignedAmount checks that arbitrary input never panics and that accepted
// amounts match the documented pattern and stay within the signed 128-bit range.
func FuzzParseSignedAmount(f *testing.F) {
	seeds := []string{
		"0",
		"-1",
		"1",
		WithdrawAllCollateral,
		"170141183460469231731687303715884105727",
		"+1",
		" 1",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, value string) {
		parsed, err := parseSignedAmount("amount", value)
		if err != nil {
			return
		}

		if !signedAmountPattern.MatchString(value) {
			t.Fatalf("accepted %q outside the documented pattern", value)
		}

		if parsed.Cmp(minSignedAmount) < 0 || parsed.Cmp(maxSignedAmount) > 0 {
			t.Fatalf("accepted %q outside the signed 128-bit range", value)
		}
	})
}
