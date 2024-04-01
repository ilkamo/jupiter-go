package solana

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

type SubResponse struct {
	Slot           uint64
	InstructionErr error
}

type subscriber struct {
	clientWS *ws.Client
}

func newSubscriber(wsEndpoint string) (subscriber, error) {
	if wsEndpoint == "" {
		return subscriber{}, fmt.Errorf("wsEndpoint is required")
	}

	wsClient, err := ws.Connect(context.Background(), wsEndpoint)
	if err != nil {
		return subscriber{}, fmt.Errorf("could not connect to ws: %w", err)
	}

	return subscriber{wsClient}, nil
}

func (s subscriber) Pull(
	ctx context.Context,
	txID TxID,
	status CommitmentStatus,
) (SubResponse, error) {
	tx, err := solana.SignatureFromBase58(string(txID))
	if err != nil {
		return SubResponse{}, fmt.Errorf("invalid txID: %w", err)
	}

	ct, err := mapToCommitmentType(status)
	if err != nil {
		return SubResponse{}, err
	}

	sub, err := s.clientWS.SignatureSubscribe(tx, ct)
	if err != nil {
		return SubResponse{}, fmt.Errorf("could not subscribe to signature: %w", err)
	}

	defer sub.Unsubscribe()

	select {
	case <-ctx.Done():
		return SubResponse{}, fmt.Errorf("context cancelled")
	case res := <-sub.Response():
		resp := SubResponse{
			Slot: res.Context.Slot,
		}

		if res.Value.Err != nil {
			resp.InstructionErr = fmt.Errorf("transaction confirmed with error: %v", res.Value.Err)
		}

		return resp, nil
	case subErr := <-sub.Err():
		return SubResponse{}, fmt.Errorf("subscription error: %w", subErr)
	}
}

type MonitorResponse struct {
	// Ok is true if the transaction reached the desired commitment status.
	Ok bool
	// InstructionErr is filled if the transaction was confirmed with an error.
	InstructionErr error
}

type monitor struct {
	sub subscriberService
}

func NewMonitor(wsEndpoint string, opts ...MonitorOption) (Monitor, error) {
	m := &monitor{}

	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, fmt.Errorf("could not apply option: %w", err)
		}
	}

	if m.sub == nil {
		sub, err := newSubscriber(wsEndpoint)
		if err != nil {
			return monitor{}, err
		}
		m.sub = sub
	}

	return m, nil
}

// WaitForCommitmentStatus waits for a transaction to reach a specific commitment status.
func (m monitor) WaitForCommitmentStatus(
	ctx context.Context,
	txID TxID,
	status CommitmentStatus,
) (MonitorResponse, error) {
	res, err := m.sub.Pull(ctx, txID, status)
	if err != nil {
		return MonitorResponse{}, err
	}

	return MonitorResponse{
		Ok:             true,
		InstructionErr: res.InstructionErr,
	}, nil
}
