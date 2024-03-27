package solana

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

const defaultMaxRetries = uint(20)

type TxID string

type client struct {
	maxRetries uint
	clientRPC  rpcService
	clientWS   wsService
	wallet     Wallet
}

func newClient(
	wallet Wallet,
	rpcEndpoint string,
	opts ...ClientOption,
) (*client, error) {
	c := &client{
		maxRetries: defaultMaxRetries,
		wallet:     wallet,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("could not apply option: %w", err)
		}
	}

	if c.clientRPC == nil {
		if rpcEndpoint == "" {
			return nil, fmt.Errorf("rpcEndpoint is required when no RPC service is provided")
		}

		rpcClient := rpc.New(rpcEndpoint)
		c.clientRPC = rpcClient
	}

	return c, nil
}

func newClientWithWS(
	wallet Wallet,
	rpcEndpoint string,
	wsEndpoint string,
	opts ...ClientOption,
) (*client, error) {
	c, err := newClient(wallet, rpcEndpoint, opts...)
	if err != nil {
		return nil, err
	}

	if c.clientWS == nil {
		if wsEndpoint == "" {
			return nil, fmt.Errorf("wsEndpoint is required when no WS service is provided")
		}

		wsClient, err := ws.Connect(context.Background(), wsEndpoint)
		if err != nil {
			return nil, fmt.Errorf("could not connect to ws: %w", err)
		}

		c.clientWS = wsClient
	}

	return c, nil
}

// NewClient creates a new Solana client with the given wallet and RPC endpoint.
// If you want to monitor your transactions using a websocket endpoint, use NewClientWithWS.
func NewClient(
	wallet Wallet,
	rpcEndpoint string,
	opts ...ClientOption,
) (DefaultClient, error) {
	return newClient(wallet, rpcEndpoint, opts...)
}

// SendTransactionOnChain sends a transaction on-chain.
func (e client) SendTransactionOnChain(ctx context.Context, txBase64 string) (TxID, error) {
	latestBlockhash, err := e.clientRPC.GetLatestBlockhash(ctx, "")
	if err != nil {
		return "", fmt.Errorf("could not get latest blockhash: %w", err)
	}

	tx, err := NewTransactionFromBase64(txBase64)
	if err != nil {
		return "", fmt.Errorf("could not deserialize swap transaction: %w", err)
	}

	tx.Message.RecentBlockhash = latestBlockhash.Value.Blockhash

	tx, err = e.wallet.SignTransaction(tx)
	if err != nil {
		return "", fmt.Errorf("could not sign swap transaction: %w", err)
	}

	sig, err := e.clientRPC.SendTransactionWithOpts(ctx, &tx, rpc.TransactionOpts{
		MaxRetries:          &e.maxRetries,
		MinContextSlot:      &latestBlockhash.Context.Slot,
		PreflightCommitment: rpc.CommitmentProcessed,
	})
	if err != nil {
		return "", fmt.Errorf("could not send transaction: %w", err)
	}

	return TxID(sig.String()), nil
}

// CheckSignature checks if a transaction with the given signature has been confirmed on-chain.
func (e client) CheckSignature(ctx context.Context, tx TxID) (bool, error) {
	sig, err := solana.SignatureFromBase58(string(tx))
	if err != nil {
		return false, fmt.Errorf("could not convert signature from base58: %w", err)
	}

	status, err := e.clientRPC.GetSignatureStatuses(ctx, false, sig)
	if err != nil {
		return false, fmt.Errorf("could not get signature status: %w", err)
	}

	if len(status.Value) == 0 {
		return false, fmt.Errorf("could not confirm transaction: no valid status")
	}

	if status.Value[0] == nil || status.Value[0].ConfirmationStatus != rpc.ConfirmationStatusFinalized {
		return false, fmt.Errorf("transaction not finalized yet")
	}

	if status.Value[0].Err != nil {
		return true, fmt.Errorf("transaction confirmed with error: %s", status.Value[0].Err)
	}

	return true, nil
}

// Close closes the client.
func (e client) Close() error {
	if e.clientRPC != nil {
		return e.clientRPC.Close()
	}

	return nil
}

type clientWithWS struct {
	*client
}

// NewClientWithWS creates a new Solana client with the given wallet, RPC and WebSocket endpoints.
func NewClientWithWS(
	wallet Wallet,
	rpcEndpoint string,
	wsEndpoint string,
	opts ...ClientOption,
) (ClientWithWS, error) {
	defaultClient, err := newClientWithWS(wallet, rpcEndpoint, wsEndpoint, opts...)
	if err != nil {
		return nil, err
	}

	return clientWithWS{defaultClient}, nil
}

// WaitForCommitmentStatus waits for a transaction to reach a specific commitment status.
func (c clientWithWS) WaitForCommitmentStatus(
	ctx context.Context,
	txID TxID,
	status CommitmentStatus,
) (bool, error) {
	tx, err := solana.SignatureFromBase58(string(txID))
	if err != nil {
		return false, fmt.Errorf("invalid txID: %w", err)
	}

	ct, err := mapToCommitmentType(status)
	if err != nil {
		return false, err
	}

	sub, err := c.clientWS.SignatureSubscribe(tx, ct)
	if err != nil {
		return false, fmt.Errorf("could not subscribe to signature: %w", err)
	}

	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return false, fmt.Errorf("context cancelled")
		case res := <-sub.Response():
			if res.Value.Err != nil {
				return false, fmt.Errorf("transaction confirmed with error: %s", res.Value.Err)
			}
			return true, nil
		case subErr := <-sub.Err():
			return false, fmt.Errorf("subscription error: %w", subErr)
		}
	}
}

// Close closes the client.
func (c clientWithWS) Close() error {
	if err := c.client.Close(); err != nil {
		return err
	}

	if c.client.clientWS != nil {
		c.client.clientWS.Close()
	}

	return nil
}
