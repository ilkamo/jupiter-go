package solanatx

import (
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/gagliardetto/solana-go"

	"github.com/ilkamo/jupiter-go/instruction"
	"github.com/ilkamo/jupiter-go/swap"
)

var ErrMissingBlockhash = errors.New("build response has no usable blockhash")

// Instructions converts a build response to solana-go instructions in this order:
// compute budget, setup, swap, cleanup, other, then tip. The API does not document
// a position for other instructions; callers that need a different order can
// convert individual instructions with ConvertInstruction.
//
// Jupiter's compute budget instructions set the unit price but not the unit limit.
func Instructions(build *swap.BuildResponse) ([]solana.Instruction, error) {
	ordered := make([]swap.Instruction, 0, len(build.ComputeBudgetInstructions)+len(build.SetupInstructions))
	ordered = append(ordered, build.ComputeBudgetInstructions...)
	ordered = append(ordered, build.SetupInstructions...)
	ordered = append(ordered, build.SwapInstruction)

	if build.CleanupInstruction != nil {
		ordered = append(ordered, *build.CleanupInstruction)
	}

	ordered = append(ordered, build.OtherInstructions...)

	if build.TipInstruction != nil {
		ordered = append(ordered, *build.TipInstruction)
	}

	instructions := make([]solana.Instruction, 0, len(ordered))
	for i, ix := range ordered {
		converted, err := ConvertInstruction(ix)
		if err != nil {
			return nil, fmt.Errorf("instruction %d: %w", i, err)
		}

		instructions = append(instructions, converted)
	}

	return instructions, nil
}

// ConvertInstruction converts one Jupiter instruction, such as a swap build or Lend
// instruction, decoding its base64 data.
func ConvertInstruction(ix instruction.Instruction) (solana.Instruction, error) {
	programID, err := solana.PublicKeyFromBase58(ix.ProgramID)
	if err != nil {
		return nil, fmt.Errorf("program ID: %w", err)
	}

	accounts := make(solana.AccountMetaSlice, 0, len(ix.Accounts))
	for i, account := range ix.Accounts {
		address, err := solana.PublicKeyFromBase58(account.Pubkey)
		if err != nil {
			return nil, fmt.Errorf("account %d: %w", i, err)
		}

		accounts = append(accounts, solana.NewAccountMeta(address, account.IsWritable, account.IsSigner))
	}

	data, err := base64.StdEncoding.DecodeString(ix.Data)
	if err != nil {
		return nil, fmt.Errorf("decode instruction data: %w", err)
	}

	return solana.NewInstruction(programID, accounts, data), nil
}

// AddressTables converts the build response's lookup tables for solana-go.
func AddressTables(build *swap.BuildResponse) (map[solana.PublicKey]solana.PublicKeySlice, error) {
	tables := make(map[solana.PublicKey]solana.PublicKeySlice, len(build.AddressesByLookupTableAddress))
	for tableAddress, addresses := range build.AddressesByLookupTableAddress {
		table, err := solana.PublicKeyFromBase58(tableAddress)
		if err != nil {
			return nil, fmt.Errorf("lookup table %q: %w", tableAddress, err)
		}

		keys := make(solana.PublicKeySlice, 0, len(addresses))
		for _, address := range addresses {
			key, err := solana.PublicKeyFromBase58(address)
			if err != nil {
				return nil, fmt.Errorf("lookup table %s address %q: %w", table, address, err)
			}

			keys = append(keys, key)
		}

		tables[table] = keys
	}

	return tables, nil
}

// Assemble builds an unsigned version 0 transaction from a build response using its
// blockhash and lookup tables. Additional instructions run after Jupiter's. The fee
// payer must be the build taker, or the payer when one was requested.
//
// Version 1 builds are rejected: their compute budget belongs in the message config,
// and the build response supplies a unit price but no compute unit limit.
//
// Build transactions cannot use Execute; sign them and submit them through your own RPC.
func Assemble(
	build *swap.BuildResponse,
	feePayer string,
	additional ...solana.Instruction,
) (*Transaction, error) {
	versionSet := build.TransactionVersion != nil
	if versionSet && *build.TransactionVersion != swap.TransactionVersionV0 {
		return nil, fmt.Errorf("%w: %d", ErrUnsupportedVersion, *build.TransactionVersion)
	}

	payer, err := solana.PublicKeyFromBase58(feePayer)
	if err != nil {
		return nil, fmt.Errorf("fee payer: %w", err)
	}

	blockhash, err := buildBlockhash(build)
	if err != nil {
		return nil, err
	}

	instructions, err := Instructions(build)
	if err != nil {
		return nil, err
	}

	instructions = append(instructions, additional...)

	tables, err := AddressTables(build)
	if err != nil {
		return nil, err
	}

	assembled, err := solana.NewTransaction(
		instructions,
		blockhash,
		solana.TransactionPayer(payer),
		solana.TransactionAddressTables(tables),
		solana.TransactionMessageVersion(solana.MessageVersionV0),
	)
	if err != nil {
		return nil, fmt.Errorf("assemble transaction: %w", err)
	}

	// MarshalBinary writes an empty slot for each required signature.
	wire, err := assembled.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("encode transaction: %w", err)
	}

	return parseWire(wire)
}

func buildBlockhash(build *swap.BuildResponse) (solana.Hash, error) {
	metadata := build.BlockhashWithMetadata
	if metadata == nil || len(metadata.Blockhash) != len(solana.Hash{}) {
		return solana.Hash{}, ErrMissingBlockhash
	}

	return solana.HashFromBytes(metadata.Blockhash), nil
}
