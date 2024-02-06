package jupitergo_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ilkamo/jupiter-go/openapi"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/test-go/testify/require"

	jupitergo "github.com/ilkamo/jupiter-go"
)

type rpcMock struct {
	shouldFailGetLatestBlockhash bool
	shouldFailSendTransaction    bool
}

func (r rpcMock) SendTransactionWithOpts(
	_ context.Context,
	_ *solana.Transaction,
	_ rpc.TransactionOpts,
) (signature solana.Signature, err error) {
	if r.shouldFailSendTransaction {
		return solana.Signature{}, errors.New("mocked error")
	}

	return solana.MustSignatureFromBase58(
		"24jRjMP3medE9iMqVSPRbkwfe9GdPmLfeftKPuwRHZdYTZJ6UyzNMGGKo4BHrTu2zVj4CgFF3CEuzS79QXUo2CMC",
	), nil
}

func (r rpcMock) GetLatestBlockhash(
	_ context.Context,
	_ rpc.CommitmentType,
) (out *rpc.GetLatestBlockhashResult, err error) {
	if r.shouldFailGetLatestBlockhash {
		return nil, errors.New("mocked error")
	}

	return &rpc.GetLatestBlockhashResult{
		Value: &rpc.LatestBlockhashResult{
			LastValidBlockHeight: 123,
			Blockhash:            solana.MustHashFromBase58("uiYzZ5PCq6C8BRSLSUGBScrXo62bBFbRFP9EkPcaWN9"),
		},
	}, nil
}

func TestNewSolanaEngine(t *testing.T) {
	testPk := "5473ZnvEhn35BdcCcPLKnzsyP6TsgqQrNFpn4i2gFegFiiJLyWginpa9GoFn2cy6Aq2EAuxLt2u2bjFDBPvNY6nw"

	wallet, err := jupitergo.NewWalletFromPrivateKeyBase58(testPk)
	require.NoError(t, err)

	t.Run("create new solana engine", func(t *testing.T) {
		_, err := jupitergo.NewSolanaEngine(
			wallet,
			"http://localhost:8899",
			jupitergo.WithMaxRetries(10),
		)
		require.NoError(t, err)
	})

	t.Run("solana engine without rpc endpoint", func(t *testing.T) {
		_, err := jupitergo.NewSolanaEngine(
			wallet,
			"",
			jupitergo.WithMaxRetries(10),
		)
		require.EqualError(t, err, "rpcEndpoint is required when no SolanaClientRPC is provided")
	})

	t.Run("solana engine with rpc endpoint", func(t *testing.T) {
		_, err := jupitergo.NewSolanaEngine(
			wallet,
			"",
			jupitergo.WithSolanaClientRPC(rpcMock{}),
		)
		require.NoError(t, err)
	})

	t.Run("execute valid swap", func(t *testing.T) {
		eng, err := jupitergo.NewSolanaEngine(
			wallet,
			"",
			jupitergo.WithSolanaClientRPC(rpcMock{}),
		)
		require.NoError(t, err)

		txID, err := eng.SendSwapOnChain(context.TODO(), openapi.SwapResponse{
			LastValidBlockHeight: 123,
			SwapTransaction:      testTx,
		})
		require.NoError(t, err)

		expectedTxID := jupitergo.TxID("24jRjMP3medE9iMqVSPRbkwfe9GdPmLfeftKPuwRHZdYTZJ6UyzNMGGKo4BHrTu2zVj4CgFF3CEuzS79QXUo2CMC")
		require.Equal(t, expectedTxID, txID)
	})

	t.Run("execute valid swap", func(t *testing.T) {
		eng, err := jupitergo.NewSolanaEngine(
			wallet,
			"",
			jupitergo.WithSolanaClientRPC(rpcMock{}),
		)
		require.NoError(t, err)

		txID, err := eng.SendSwapOnChain(context.TODO(), openapi.SwapResponse{
			LastValidBlockHeight: 123,
			SwapTransaction:      testTx,
		})
		require.NoError(t, err)

		expectedTxID := jupitergo.TxID("24jRjMP3medE9iMqVSPRbkwfe9GdPmLfeftKPuwRHZdYTZJ6UyzNMGGKo4BHrTu2zVj4CgFF3CEuzS79QXUo2CMC")
		require.Equal(t, expectedTxID, txID)
	})

	t.Run("error when getting the blockhash", func(t *testing.T) {
		eng, err := jupitergo.NewSolanaEngine(
			wallet,
			"",
			jupitergo.WithSolanaClientRPC(rpcMock{shouldFailGetLatestBlockhash: true}),
		)
		require.NoError(t, err)

		_, err = eng.SendSwapOnChain(context.TODO(), openapi.SwapResponse{
			LastValidBlockHeight: 123,
			SwapTransaction:      testTx,
		})
		require.EqualError(t, err, "could not get latest blockhash: mocked error")
	})

	t.Run("error when sending the transaction on chain", func(t *testing.T) {
		eng, err := jupitergo.NewSolanaEngine(
			wallet,
			"",
			jupitergo.WithSolanaClientRPC(rpcMock{shouldFailSendTransaction: true}),
		)
		require.NoError(t, err)

		_, err = eng.SendSwapOnChain(context.TODO(), openapi.SwapResponse{
			LastValidBlockHeight: 123,
			SwapTransaction:      testTx,
		})
		require.EqualError(t, err, "could not send transaction: mocked error")
	})
}
