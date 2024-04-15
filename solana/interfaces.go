package solana

import (
	"context"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type rpcService interface {
	SendTransactionWithOpts(
		ctx context.Context,
		transaction *solana.Transaction,
		opts rpc.TransactionOpts,
	) (signature solana.Signature, err error)
	GetLatestBlockhash(
		ctx context.Context,
		commitment rpc.CommitmentType,
	) (out *rpc.GetLatestBlockhashResult, err error)
	GetSignatureStatuses(
		ctx context.Context,
		searchTransactionHistory bool,
		transactionSignatures ...solana.Signature,
	) (out *rpc.GetSignatureStatusesResult, err error)
	GetTokenAccountBalance(
		ctx context.Context,
		account solana.PublicKey,
		commitment rpc.CommitmentType, // optional
	) (out *rpc.GetTokenAccountBalanceResult, err error)
	Close() error
}

type Client interface {
	SendTransactionOnChain(context.Context, string) (TxID, error)
	CheckSignature(context.Context, TxID) (bool, error)
	GetTokenAccountBalance(context.Context, string) (TokenAccount, error)
}

type subscriberService interface {
	Pull(
		ctx context.Context,
		txID TxID,
		status CommitmentStatus,
	) (SubResponse, error)
}

type Monitor interface {
	WaitForCommitmentStatus(context.Context, TxID, CommitmentStatus) (MonitorResponse, error)
}
