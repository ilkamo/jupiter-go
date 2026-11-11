package solanatx

import (
	"errors"
	"fmt"
	"slices"

	"github.com/gagliardetto/solana-go"
)

// JitoDontFront is Jito's sandwich-mitigation account. Jito's block engine
// rejects any bundle that includes a transaction referencing an account whose
// address starts with "jitodontfront" unless that transaction comes first.
// The account does not need to exist on chain. Add it with WithReadOnlyAccount.
//
// It reduces, but does not prevent, sandwiching: it applies only to Jito's
// block engine. See
// https://docs.jito.wtf/lowlatencytxnsend/#sandwich-mitigation.
const JitoDontFront = "jitodontfront111111111111111111111111111111"

// ErrDuplicateAccount reports an account the instruction already references.
var ErrDuplicateAccount = errors.New("instruction already references the account")

// WithReadOnlyAccount returns a copy of ix that also references account as a
// read-only, non-signing account, for example JitoDontFront. The original
// instruction is unchanged.
//
// The instruction's program must ignore accounts it does not expect. The System
// Program transfer returned by TipInstruction does, so attach the account there
// and pass the result to Assemble instead of setting swap.BuildRequest.TipAmount.
func WithReadOnlyAccount(ix solana.Instruction, account string) (solana.Instruction, error) {
	key, err := solana.PublicKeyFromBase58(account)
	if err != nil {
		return nil, fmt.Errorf("account: %w", err)
	}

	data, err := ix.Data()
	if err != nil {
		return nil, fmt.Errorf("instruction data: %w", err)
	}

	existing := ix.Accounts()
	accounts := make(solana.AccountMetaSlice, 0, len(existing)+1)
	for _, meta := range existing {
		if meta.PublicKey.Equals(key) {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateAccount, account)
		}

		copied := *meta
		accounts = append(accounts, &copied)
	}

	accounts = append(accounts, solana.Meta(key))

	return solana.NewInstruction(ix.ProgramID(), accounts, slices.Clone(data)), nil
}
