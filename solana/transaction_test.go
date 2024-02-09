package solana_test

import (
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/test-go/testify/require"

	jupSolana "github.com/ilkamo/jupiter-go/solana"
)

const testTx = "AAEAAQPrM+1WcczVrvBstwqcH1lXpPpbHuKVFpSj9kZOi1GITD6KBh4ENmDzZ4cG9x+7s1w6q77AoogJbaz28WWsI0elAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAANgS9CVZkT3oU8ECpERHXI92vwg8ofvcIVgdQtcOK3NgECAgABDAIAAACghgEAAAAAAA=="

func TestNewTransactionFromBase64(t *testing.T) {
	tx, err := jupSolana.NewTransactionFromBase64(testTx)
	require.NoError(t, err)

	require.Equal(t, "uiYzZ5PCq6C8BRSLSUGBScrXo62bBFbRFP9EkPcaWN9", tx.Message.RecentBlockhash.String())
	require.Len(t, tx.Message.AccountKeys, 3)
	require.Len(t, tx.Message.Instructions, 1)
}

func generateTestNotSignedTx(t *testing.T) solana.Transaction {
	tx, err := jupSolana.NewTransactionFromBase64(testTx)
	require.NoError(t, err)

	return tx
}
