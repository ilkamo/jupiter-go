package solana_test

import (
	"testing"

	"github.com/test-go/testify/require"

	jupSolana "github.com/ilkamo/jupiter-go/solana"
)

func TestNewWalletFromPrivateKeyBase58(t *testing.T) {
	testPk := "5473ZnvEhn35BdcCcPLKnzsyP6TsgqQrNFpn4i2gFegFiiJLyWginpa9GoFn2cy6Aq2EAuxLt2u2bjFDBPvNY6nw"

	t.Run("valid private key", func(t *testing.T) {
		wallet, err := jupSolana.NewWalletFromPrivateKeyBase58(testPk)
		require.NoError(t, err)
		require.Equal(t, testPk, wallet.PrivateKey.String())
	})

	t.Run("invalid private key", func(t *testing.T) {
		_, err := jupSolana.NewWalletFromPrivateKeyBase58("invalid")
		require.Error(t, err)
	})
}

func TestWallet_SignTransaction(t *testing.T) {
	tx := generateTestNotSignedTx(t)
	require.Len(t, tx.Signatures, 0)

	testPk := "5473ZnvEhn35BdcCcPLKnzsyP6TsgqQrNFpn4i2gFegFiiJLyWginpa9GoFn2cy6Aq2EAuxLt2u2bjFDBPvNY6nw"

	wallet, err := jupSolana.NewWalletFromPrivateKeyBase58(testPk)
	require.NoError(t, err)

	signedTx, err := wallet.SignTransaction(tx)
	require.NoError(t, err)

	require.Len(t, signedTx.Signatures, 1)
}
