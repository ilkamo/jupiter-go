// Package listparam validates and encodes comma-separated query parameter lists.
package listparam

import (
	"fmt"
	"strings"
)

const separator = ","

// ValidateItem rejects a value that would change meaning inside a comma-separated list.
func ValidateItem(field string, item string) error {
	if item == "" {
		return fmt.Errorf("%s: empty value", field)
	}

	if strings.Contains(item, separator) {
		return fmt.Errorf("%s: value %q contains a separator", field, item)
	}

	if strings.TrimSpace(item) != item {
		return fmt.Errorf("%s: value %q has surrounding whitespace", field, item)
	}

	return nil
}

// Validate applies ValidateItem to every item.
func Validate(field string, items []string) error {
	for _, item := range items {
		if err := ValidateItem(field, item); err != nil {
			return err
		}
	}

	return nil
}

// Join encodes validated items as one comma-separated value.
func Join(items []string) string {
	return strings.Join(items, separator)
}
