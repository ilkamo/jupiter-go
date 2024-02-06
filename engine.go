package jupitergo

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"github.com/ilkamo/jupiter-go/openapi"
)

const defaultMaxRetries = uint(20)

type TxID string

type SolanaClientRPC interface {
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
}

type SolanaEngine struct {
	maxRetries      uint
	solanaClientRPC SolanaClientRPC
	wallet          Wallet
}

func NewSolanaEngine(
	wallet Wallet,
	rpcEndpoint string,
	opts ...EngineOption,
) (SolanaEngine, error) {
	e := &SolanaEngine{
		maxRetries: defaultMaxRetries,
		wallet:     wallet,
	}

	for _, opt := range opts {
		if err := opt(e); err != nil {
			return SolanaEngine{}, fmt.Errorf("could not apply option: %w", err)
		}
	}

	if e.solanaClientRPC == nil {
		if rpcEndpoint == "" {
			return SolanaEngine{}, fmt.Errorf("rpcEndpoint is required when no SolanaClientRPC is provided")
		}

		rpcClient := rpc.New(rpcEndpoint)
		e.solanaClientRPC = rpcClient
	}

	return *e, nil
}

// EngineOption is a function that allows to specify options for the client
type EngineOption func(*SolanaEngine) error

// WithMaxRetries sets the maximum number of retries for the engine when sending a transaction on-chain
func WithMaxRetries(maxRetries uint) EngineOption {
	return func(e *SolanaEngine) error {
		e.maxRetries = maxRetries
		return nil
	}
}

// WithSolanaClientRPC sets the Solana client RPC for the engine
func WithSolanaClientRPC(clientRPC SolanaClientRPC) EngineOption {
	return func(e *SolanaEngine) error {
		e.solanaClientRPC = clientRPC
		return nil
	}
}

// SendSwapOnChain sends on-chain a swap transaction retrieved from Jupiter
func (e SolanaEngine) SendSwapOnChain(ctx context.Context, swap openapi.SwapResponse) (TxID, error) {
	latestBlockhash, err := e.solanaClientRPC.GetLatestBlockhash(ctx, "")
	if err != nil {
		return "", fmt.Errorf("could not get latest blockhash: %w", err)
	}

	tx, err := NewTransactionFromBase64(swap.SwapTransaction)
	if err != nil {
		return "", fmt.Errorf("could not deserialize swap transaction: %w", err)
	}

	tx.Message.RecentBlockhash = latestBlockhash.Value.Blockhash

	tx, err = e.wallet.SignTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("could not sign swap transaction: %w", err)
	}

	sig, err := e.solanaClientRPC.SendTransactionWithOpts(ctx, &tx, rpc.TransactionOpts{
		MaxRetries:          &e.maxRetries,
		MinContextSlot:      &latestBlockhash.Context.Slot,
		PreflightCommitment: rpc.CommitmentProcessed,
	})
	if err != nil {
		return "", fmt.Errorf("could not send transaction: %w", err)
	}

	return TxID(sig.String()), nil
}

// CheckSignature checks if a transaction with the given signature has been confirmed on-chain
func (e SolanaEngine) CheckSignature(ctx context.Context, tx TxID) (bool, error) {
	sig, err := solana.SignatureFromBase58(string(tx))
	if err != nil {
		return false, fmt.Errorf("could not convert signature from base58: %w", err)
	}

	status, err := e.solanaClientRPC.GetSignatureStatuses(ctx, false, sig)
	if err != nil {
		return false, fmt.Errorf("could not get signature status: %w", err)
	}

	if len(status.Value) == 0 {
		return false, fmt.Errorf("could not confirm transaction: no valid status")
	}

	if status.Value[0].ConfirmationStatus != rpc.ConfirmationStatusFinalized {
		return false, fmt.Errorf("transaction not finalized yet")
	}

	if status.Value[0].Err != nil {
		return true, fmt.Errorf("transaction confirmed with error: %s", status.Value[0].Err)
	}

	return true, nil
}
