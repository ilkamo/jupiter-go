package solana_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/test-go/testify/require"

	jupSolana "github.com/ilkamo/jupiter-go/solana"
)

type rpcMock struct {
	shouldFailGetLatestBlockhash bool
	shouldFailSendTransaction    bool
	shouldFailGetSignatureStatus bool
	shoultFailGetTokenBalance    bool
}

var (
	testSignature       = "24jRjMP3medE9iMqVSPRbkwfe9GdPmLfeftKPuwRHZdYTZJ6UyzNMGGKo4BHrTu2zVj4CgFF3CEuzS79QXUo2CMC"
	processingSignature = "24jRjMP3medE9iMqVSPRbkwfe9GdPmLfeftKPuwRHZdYTZJ6UyzNMGGKo4BHrTu2zVj4CgFF3CEuzS79QXUo2CPC"
)

func (r rpcMock) SendTransactionWithOpts(
	_ context.Context,
	_ *solana.Transaction,
	_ rpc.TransactionOpts,
) (signature solana.Signature, err error) {
	if r.shouldFailSendTransaction {
		return solana.Signature{}, errors.New("mocked error")
	}

	return solana.MustSignatureFromBase58(testSignature), nil
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

func (r rpcMock) GetSignatureStatuses(
	_ context.Context,
	_ bool,
	sign ...solana.Signature,
) (out *rpc.GetSignatureStatusesResult, err error) {
	if r.shouldFailGetSignatureStatus {
		return nil, errors.New("mocked error")
	}

	if sign[0].Equals(solana.MustSignatureFromBase58(processingSignature)) {
		return &rpc.GetSignatureStatusesResult{
			Value: []*rpc.SignatureStatusesResult{
				{
					ConfirmationStatus: rpc.ConfirmationStatusProcessed,
				},
			},
		}, nil
	}

	return &rpc.GetSignatureStatusesResult{
		Value: []*rpc.SignatureStatusesResult{
			{
				ConfirmationStatus: rpc.ConfirmationStatusFinalized,
			},
		},
	}, nil
}

func (r rpcMock) GetTokenAccountBalance(
	_ context.Context,
	_ solana.PublicKey,
	_ rpc.CommitmentType,
) (out *rpc.GetTokenAccountBalanceResult, err error) {
	if r.shoultFailGetTokenBalance {
		return nil, errors.New("mocked error")
	}

	return &rpc.GetTokenAccountBalanceResult{
		Value: &rpc.UiTokenAmount{
			Amount:   "1000000000",
			Decimals: 9,
		},
	}, nil
}

func (r rpcMock) Close() error {
	return nil
}

func TestNewClient(t *testing.T) {
	testPk := "5473ZnvEhn35BdcCcPLKnzsyP6TsgqQrNFpn4i2gFegFiiJLyWginpa9GoFn2cy6Aq2EAuxLt2u2bjFDBPvNY6nw"

	wallet, err := jupSolana.NewWalletFromPrivateKeyBase58(testPk)
	require.NoError(t, err)

	t.Run("create new solana client", func(t *testing.T) {
		_, err := jupSolana.NewClient(
			wallet,
			"http://localhost:8899",
			jupSolana.WithMaxRetries(10),
		)
		require.NoError(t, err)
	})

	t.Run("solana client without rpc endpoint", func(t *testing.T) {
		_, err := jupSolana.NewClient(
			wallet,
			"",
			jupSolana.WithMaxRetries(10),
		)
		require.EqualError(t, err, "rpcEndpoint is required when no RPC service is provided")
	})

	t.Run("solana client with rpc endpoint", func(t *testing.T) {
		_, err := jupSolana.NewClient(
			wallet,
			"",
			jupSolana.WithClientRPC(rpcMock{}),
		)
		require.NoError(t, err)
	})

	t.Run("execute valid swap", func(t *testing.T) {
		c, err := jupSolana.NewClient(
			wallet,
			"",
			jupSolana.WithClientRPC(rpcMock{}),
		)
		require.NoError(t, err)

		txID, err := c.SendTransactionOnChain(context.TODO(), testTx)
		require.NoError(t, err)

		expectedTxID := jupSolana.TxID(testSignature)
		require.Equal(t, expectedTxID, txID)
	})

	t.Run("execute valid swap", func(t *testing.T) {
		c, err := jupSolana.NewClient(
			wallet,
			"",
			jupSolana.WithClientRPC(rpcMock{}),
		)
		require.NoError(t, err)

		txID, err := c.SendTransactionOnChain(context.TODO(), testTx)
		require.NoError(t, err)

		expectedTxID := jupSolana.TxID(testSignature)
		require.Equal(t, expectedTxID, txID)
	})

	t.Run("error when getting the blockhash", func(t *testing.T) {
		c, err := jupSolana.NewClient(
			wallet,
			"",
			jupSolana.WithClientRPC(rpcMock{shouldFailGetLatestBlockhash: true}),
		)
		require.NoError(t, err)

		_, err = c.SendTransactionOnChain(context.TODO(), testTx)
		require.EqualError(t, err, "could not get latest blockhash: mocked error")
	})

	t.Run("error when sending the transaction on chain", func(t *testing.T) {
		c, err := jupSolana.NewClient(
			wallet,
			"",
			jupSolana.WithClientRPC(rpcMock{shouldFailSendTransaction: true}),
		)
		require.NoError(t, err)

		_, err = c.SendTransactionOnChain(context.TODO(), testTx)
		require.EqualError(t, err, "could not send transaction: mocked error")
	})

	t.Run("error when getting the signature status", func(t *testing.T) {
		c, err := jupSolana.NewClient(
			wallet,
			"",
			jupSolana.WithClientRPC(rpcMock{shouldFailGetSignatureStatus: true}),
		)
		require.NoError(t, err)

		_, err = c.CheckSignature(
			context.TODO(),
			jupSolana.TxID(testSignature),
		)
		require.EqualError(t, err, "could not get signature status: mocked error")
	})

	t.Run("transaction still in process when getting signature status", func(t *testing.T) {
		c, err := jupSolana.NewClient(
			wallet,
			"",
			jupSolana.WithClientRPC(rpcMock{}),
		)
		require.NoError(t, err)

		_, err = c.CheckSignature(
			context.TODO(),
			jupSolana.TxID(processingSignature),
		)
		require.EqualError(t, err, "transaction not finalized yet")
	})

	t.Run("transaction confirmed when getting signature status", func(t *testing.T) {
		c, err := jupSolana.NewClient(
			wallet,
			"",
			jupSolana.WithClientRPC(rpcMock{}),
		)
		require.NoError(t, err)

		confirmed, err := c.CheckSignature(
			context.TODO(),
			jupSolana.TxID(testSignature),
		)
		require.NoError(t, err)
		require.True(t, confirmed)
	})

	t.Run("error when getting token balance", func(t *testing.T) {
		c, err := jupSolana.NewClient(
			wallet,
			"",
			jupSolana.WithClientRPC(rpcMock{shoultFailGetTokenBalance: true}),
		)
		require.NoError(t, err)

		_, err = c.GetTokenAccountBalance(
			context.TODO(),
			"invalid token account address",
		)
		require.EqualError(t, err,
			"could not parse token account public key: decode: invalid base58 digit ('l')",
		)

		_, err = c.GetTokenAccountBalance(
			context.TODO(),
			"9K4NT8o4VyXv8RiHWfr7tchGEbsrV7KHYwMQDSgt1pnZ",
		)
		require.EqualError(t, err, "could not get token account balance: mocked error")
	})

	t.Run("get token account balance", func(t *testing.T) {
		c, err := jupSolana.NewClient(
			wallet,
			"",
			jupSolana.WithClientRPC(rpcMock{}),
		)
		require.NoError(t, err)

		balance, err := c.GetTokenAccountBalance(
			context.TODO(),
			"9K4NT8o4VyXv8RiHWfr7tchGEbsrV7KHYwMQDSgt1pnZ",
		)
		require.NoError(t, err)
		require.Equal(t, "1000000000", balance.Amount.String())
		require.Equal(t, uint8(9), balance.Decimals)
	})
}
