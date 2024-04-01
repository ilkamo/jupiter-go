package solana

import (
	"testing"

	"github.com/test-go/testify/require"
)

func TestWithMaxRetries(t *testing.T) {
	c := client{}
	require.Zero(t, c.maxRetries)

	opt := WithMaxRetries(10)

	err := opt(&c)
	require.NoError(t, err)
	require.Equal(t, uint(10), c.maxRetries)
}
