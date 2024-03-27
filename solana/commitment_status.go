package solana

import (
	"errors"

	"github.com/gagliardetto/solana-go/rpc"
)

type CommitmentStatus struct {
	s string
}

func (cs CommitmentStatus) String() string {
	return cs.s
}

// For more information, see https://docs.solanalabs.com/consensus/commitments
var (
	CommitmentFinalized = CommitmentStatus{"finalized"}
	CommitmentConfirmed = CommitmentStatus{"confirmed"}
	CommitmentProcessed = CommitmentStatus{"processed"}
)

func mapToCommitmentType(cs CommitmentStatus) (rpc.CommitmentType, error) {
	switch cs {
	case CommitmentFinalized:
		return rpc.CommitmentFinalized, nil
	case CommitmentConfirmed:
		return rpc.CommitmentConfirmed, nil
	case CommitmentProcessed:
		return rpc.CommitmentProcessed, nil
	}

	return "", errors.New("invalid CommitmentStatus")
}
