package solana

import (
	"testing"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/require"
)

func Test_mapToCommitmentType(t *testing.T) {
	testCases := []struct {
		name    string
		cs      CommitmentStatus
		want    rpc.CommitmentType
		wantErr bool
	}{
		{
			name:    "processed",
			cs:      CommitmentProcessed,
			want:    rpc.CommitmentProcessed,
			wantErr: false,
		},
		{
			name:    "confirmed",
			cs:      CommitmentConfirmed,
			want:    rpc.CommitmentConfirmed,
			wantErr: false,
		},
		{
			name:    "finalized",
			cs:      CommitmentFinalized,
			want:    rpc.CommitmentFinalized,
			wantErr: false,
		},
		{
			name:    "invalid",
			cs:      CommitmentStatus{},
			want:    "",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := mapToCommitmentType(tc.cs)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}
