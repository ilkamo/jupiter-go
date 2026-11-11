package listparam

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name  string
		items []string
		valid bool
	}{
		{name: "none", items: nil, valid: true},
		{name: "labels with inner spaces", items: []string{"Raydium", "Orca V2"}, valid: true},
		{name: "empty", items: []string{""}},
		{name: "separator", items: []string{"a,b"}},
		{name: "leading whitespace", items: []string{" a"}},
		{name: "trailing newline", items: []string{"a\n"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := Validate("field", test.items); (err == nil) != test.valid {
				t.Fatalf("Validate(%q) = %v", test.items, err)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	if got := Join([]string{"a", "b c"}); got != "a,b c" {
		t.Fatalf("Join = %q", got)
	}
}
