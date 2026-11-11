package solanatx

import (
	"bytes"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/gagliardetto/solana-go"

	"github.com/ilkamo/jupiter-go/swap"
)

const (
	computeBudgetSeed byte = iota + 20
	setupSeed
	swapSeed
	cleanupSeed
	otherSeed
	tipSeed
	additionalSeed
	lookupTableSeed
	lookupAccountSeed
)

func instructionFor(programSeed byte, accounts ...swap.AccountMeta) swap.Instruction {
	return swap.Instruction{
		ProgramID: testAddress(programSeed).String(),
		Accounts:  accounts,
		Data:      base64.StdEncoding.EncodeToString([]byte{programSeed}),
	}
}

// testBuild returns a build response whose swap instruction reads one account
// through a lookup table.
func testBuild(t *testing.T) *swap.BuildResponse {
	t.Helper()

	taker := swap.AccountMeta{Pubkey: testAddress(feePayerSeed).String(), IsSigner: true, IsWritable: true}
	lookupAccount := swap.AccountMeta{Pubkey: testAddress(lookupAccountSeed).String()}
	cleanup := instructionFor(cleanupSeed, taker)
	tip := instructionFor(tipSeed, taker)

	return &swap.BuildResponse{
		ComputeBudgetInstructions: []swap.Instruction{instructionFor(computeBudgetSeed)},
		SetupInstructions:         []swap.Instruction{instructionFor(setupSeed, taker)},
		SwapInstruction:           instructionFor(swapSeed, taker, lookupAccount),
		CleanupInstruction:        &cleanup,
		OtherInstructions:         []swap.Instruction{instructionFor(otherSeed)},
		TipInstruction:            &tip,
		AddressesByLookupTableAddress: map[string][]string{
			testAddress(lookupTableSeed).String(): {testAddress(lookupAccountSeed).String()},
		},
		TransactionVersion: new(uint8(swap.TransactionVersionV0)),
		BlockhashWithMetadata: &swap.BlockhashMetadata{
			Blockhash: bytes.Repeat([]byte{blockhashSeed}, len(solana.Hash{})),
		},
	}
}

func assembleFixture(t *testing.T, build *swap.BuildResponse) *Transaction {
	t.Helper()

	tx, err := Assemble(build, testAddress(feePayerSeed).String())
	if err != nil {
		t.Fatal(err)
	}

	return tx
}

func TestAssembleOrdersInstructionsAndUsesLookupTables(t *testing.T) {
	additional := solana.NewInstruction(testAddress(additionalSeed), nil, []byte{additionalSeed})

	tx, err := Assemble(testBuild(t), testAddress(feePayerSeed).String(), additional)
	if err != nil {
		t.Fatal(err)
	}

	if tx.Version() != Version0 || len(tx.MissingSigners()) != 1 {
		t.Fatalf("version=%v missing=%v", tx.Version(), tx.MissingSigners())
	}

	if err := tx.Sign(testKey(feePayerSeed)); err != nil {
		t.Fatal(err)
	}

	decoded, err := solana.TransactionFromBytes(tx.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	if err := decoded.VerifySignatures(); err != nil {
		t.Fatal(err)
	}

	if len(decoded.Message.AddressTableLookups) != 1 {
		t.Fatalf("lookup tables not used: %+v", decoded.Message.AddressTableLookups)
	}

	wantBlockhash := solana.HashFromBytes(bytes.Repeat([]byte{blockhashSeed}, len(solana.Hash{})))
	if decoded.Message.RecentBlockhash != wantBlockhash {
		t.Fatal("blockhash changed")
	}

	wantOrder := []byte{
		computeBudgetSeed,
		setupSeed,
		swapSeed,
		cleanupSeed,
		otherSeed,
		tipSeed,
		additionalSeed,
	}
	if len(decoded.Message.Instructions) != len(wantOrder) {
		t.Fatalf("got %d instructions", len(decoded.Message.Instructions))
	}

	for i, instruction := range decoded.Message.Instructions {
		if !bytes.Equal(instruction.Data, []byte{wantOrder[i]}) {
			t.Fatalf("instruction %d data %v, want %d", i, instruction.Data, wantOrder[i])
		}
	}
}

func TestAssembleWithoutLookupTablesKeepsVersion0(t *testing.T) {
	build := testBuild(t)
	build.AddressesByLookupTableAddress = map[string][]string{}

	tx := assembleFixture(t, build)
	if tx.Version() != Version0 {
		t.Fatalf("version = %v", tx.Version())
	}
}

func TestAssembleRejectsUnusableBuilds(t *testing.T) {
	tests := []struct {
		name      string
		configure func(*swap.BuildResponse)
		feePayer  string
		want      error
	}{
		{
			name: "version 1",
			configure: func(b *swap.BuildResponse) {
				b.TransactionVersion = new(uint8(swap.TransactionVersionV1))
			},
			want: ErrUnsupportedVersion,
		},
		{
			name:      "missing blockhash",
			configure: func(b *swap.BuildResponse) { b.BlockhashWithMetadata = nil },
			want:      ErrMissingBlockhash,
		},
		{
			name: "short blockhash",
			configure: func(b *swap.BuildResponse) {
				b.BlockhashWithMetadata.Blockhash = b.BlockhashWithMetadata.Blockhash[1:]
			},
			want: ErrMissingBlockhash,
		},
		{
			name:      "malformed instruction data",
			configure: func(b *swap.BuildResponse) { b.SwapInstruction.Data = "not base64" },
		},
		{
			name:      "malformed program ID",
			configure: func(b *swap.BuildResponse) { b.SwapInstruction.ProgramID = "program" },
		},
		{
			name: "malformed lookup table",
			configure: func(b *swap.BuildResponse) {
				b.AddressesByLookupTableAddress = map[string][]string{"table": nil}
			},
		},
		{
			name:      "malformed fee payer",
			configure: func(*swap.BuildResponse) {},
			feePayer:  "payer",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			build := testBuild(t)
			test.configure(build)

			feePayer := test.feePayer
			if feePayer == "" {
				feePayer = testAddress(feePayerSeed).String()
			}

			_, err := Assemble(build, feePayer)
			if err == nil {
				t.Fatal("unusable build accepted")
			}

			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}
