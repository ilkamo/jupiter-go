package solana

import (
	"context"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
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
	Close() error
}

type wsService interface {
	SignatureSubscribe(
		signature solana.Signature, // Transaction Signature.
		commitment rpc.CommitmentType, // (optional)
	) (sub *ws.SignatureSubscription, err error)
	Close()
}

type DefaultClient interface {
	SendTransactionOnChain(context.Context, string) (TxID, error)
	CheckSignature(context.Context, TxID) (bool, error)
}

type ClientWithWS interface {
	DefaultClient
	WaitForCommitmentStatus(context.Context, TxID, CommitmentStatus) (bool, error)
}
