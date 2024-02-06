package gojup

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gagliardetto/solana-go/rpc"

	"github.com/ilkamo/go-jup/openapi"
)

const defaultMaxRetries = uint(20)

type TxID string

type SolanaEngine struct {
	logger              *slog.Logger
	maxRetries          uint
	solanaConnectionRPC *rpc.Client
	wallet              Wallet
}

func NewSolanaEngine(
	wallet Wallet,
	rpcEndpoint string,
	opts ...EngineOption,
) (SolanaEngine, error) {
	if rpcEndpoint == "" {
		return SolanaEngine{}, fmt.Errorf("rpcEndpoint is empty")
	}

	e := &SolanaEngine{
		logger:              slog.Default(),
		maxRetries:          defaultMaxRetries,
		solanaConnectionRPC: rpc.New(rpcEndpoint),
		wallet:              wallet,
	}

	for _, opt := range opts {
		if err := opt(e); err != nil {
			return SolanaEngine{}, fmt.Errorf("could not apply option: %w", err)
		}
	}

	return *e, nil
}

// EngineOption is a function that allows to specify options for the client
type EngineOption func(*SolanaEngine) error

// WithLogger sets the logger for the engine
func WithLogger(logger *slog.Logger) EngineOption {
	return func(e *SolanaEngine) error {
		e.logger = logger
		return nil
	}
}

// WithMaxRetries sets the maximum number of retries for the engine when sending a transaction on-chain
func WithMaxRetries(maxRetries uint) EngineOption {
	return func(e *SolanaEngine) error {
		e.maxRetries = maxRetries
		return nil
	}
}

// SendSwapOnChain sends on-chain a swap transaction retrieved from Jupiter
func (e SolanaEngine) SendSwapOnChain(ctx context.Context, swap openapi.SwapResponse) (TxID, error) {
	latestBlockhash, err := e.solanaConnectionRPC.GetLatestBlockhash(ctx, "")
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

	sig, err := e.solanaConnectionRPC.SendTransactionWithOpts(ctx, &tx, rpc.TransactionOpts{
		MaxRetries:          &e.maxRetries,
		MinContextSlot:      &latestBlockhash.Context.Slot,
		PreflightCommitment: rpc.CommitmentProcessed,
	})
	if err != nil {
		return "", fmt.Errorf("could not send transaction: %w", err)
	}

	e.logger.Info("sent transaction", "tx", sig.String())

	return TxID(sig.String()), nil
}
