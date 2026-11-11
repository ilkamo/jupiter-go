package amount

import "testing"

func TestParseRaw(t *testing.T) {
	tests := []struct {
		name   string
		amount string
		want   string
	}{
		{name: "positive integer", amount: "123", want: "123"},
		{name: "smallest raw unit", amount: "1", want: "1"},
		{name: "beyond uint64", amount: "18446744073709551616", want: "18446744073709551616"},
		{name: "leading zeros", amount: "00123", want: "123"},
		{name: "explicit positive sign", amount: "+123", want: "123"},
		{name: "empty", amount: ""},
		{name: "zero", amount: "0"},
		{name: "all zeros", amount: "000"},
		{name: "negative", amount: "-1"},
		{name: "fraction", amount: "1.5"},
		{name: "decimal notation", amount: "1.0"},
		{name: "scientific notation", amount: "1e3"},
		{name: "hexadecimal", amount: "0xff"},
		{name: "separator", amount: "1_000"},
		{name: "whitespace", amount: " 123"},
		{name: "invalid", amount: "abc"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			amount, err := ParseRaw(test.amount)
			if test.want == "" {
				if err == nil {
					t.Fatal("invalid amount accepted")
				}

				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if amount != test.want {
				t.Fatalf("amount = %q, want %q", amount, test.want)
			}
		})
	}
}
