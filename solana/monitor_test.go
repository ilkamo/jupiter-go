package solana_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ilkamo/jupiter-go/solana"
)

type subscriberMock struct {
	withInstructionError bool
	withError            bool
}

func (s subscriberMock) Pull(
	_ context.Context,
	_ solana.TxID,
	_ solana.CommitmentStatus,
) (solana.SubResponse, error) {
	if s.withError {
		return solana.SubResponse{}, errors.New("mock error")
	}

	if s.withInstructionError {
		return solana.SubResponse{
			Slot:           123,
			InstructionErr: errors.New("mock instruction error"),
		}, nil
	}

	return solana.SubResponse{
		Slot: 123,
	}, nil
}

func Test_monitor_WaitForCommitmentStatus(t *testing.T) {
	t.Run("invalid monitor ws endpoint", func(t *testing.T) {
		_, err := solana.NewMonitor("")
		require.Error(t, err)
	})

	t.Run("subscriber error", func(t *testing.T) {
		sub := subscriberMock{
			withError: true,
		}

		monitor, err := solana.NewMonitor("", solana.WithMonitorSubscriber(sub))
		require.NoError(t, err)

		_, err = monitor.WaitForCommitmentStatus(context.Background(), "txID", solana.CommitmentProcessed)
		require.EqualError(t, err, "mock error")
	})

	t.Run("confirmed with instruction error", func(t *testing.T) {
		sub := subscriberMock{
			withInstructionError: true,
		}

		monitor, err := solana.NewMonitor("", solana.WithMonitorSubscriber(sub))
		require.NoError(t, err)

		res, err := monitor.WaitForCommitmentStatus(context.Background(), "txID", solana.CommitmentProcessed)
		require.NoError(t, err)

		require.True(t, res.Ok)
		require.EqualError(t, res.InstructionErr, "mock instruction error")
	})

	t.Run("successfully confirmed", func(t *testing.T) {
		sub := subscriberMock{}

		monitor, err := solana.NewMonitor("", solana.WithMonitorSubscriber(sub))
		require.NoError(t, err)

		resp, err := monitor.WaitForCommitmentStatus(context.Background(), "txID", solana.CommitmentProcessed)
		require.NoError(t, err)
		require.True(t, resp.Ok)
		require.Nil(t, resp.InstructionErr)
	})
}
