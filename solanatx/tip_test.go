package solanatx

import (
	"encoding/binary"
	"slices"
	"testing"

	"github.com/gagliardetto/solana-go"

	"github.com/ilkamo/jupiter-go/transaction"
)

const (
	systemTransferIndex = 2
	instructionTagBytes = 4
	tipTestLamports     = transaction.MinTipLamports + 1
)

func TestTipInstruction(t *testing.T) {
	payer := testAddress(feePayerSeed)

	tip, err := TipInstruction(payer.String(), tipTestLamports)
	if err != nil {
		t.Fatal(err)
	}

	accounts := tip.Accounts()
	if !tip.ProgramID().Equals(solana.SystemProgramID) || len(accounts) != 2 {
		t.Fatalf("program=%s accounts=%d", tip.ProgramID(), len(accounts))
	}

	if !accounts[0].PublicKey.Equals(payer) || !accounts[0].IsSigner || !accounts[0].IsWritable {
		t.Fatalf("funding account = %+v", accounts[0])
	}

	if !slices.Contains(transaction.TipReceivers(), accounts[1].PublicKey.String()) {
		t.Fatalf("receiver %s is not a documented tip account", accounts[1].PublicKey)
	}

	data, err := tip.Data()
	if err != nil {
		t.Fatal(err)
	}

	if binary.LittleEndian.Uint32(data) != systemTransferIndex ||
		binary.LittleEndian.Uint64(data[instructionTagBytes:]) != tipTestLamports {
		t.Fatalf("transfer data = %v", data)
	}

	tx, err := Assemble(testBuild(t), payer.String(), tip)
	if err != nil {
		t.Fatal(err)
	}

	if tx.Version() != Version0 {
		t.Fatalf("tipped build version = %v", tx.Version())
	}
}

func TestTipInstructionRejectsInvalidInputs(t *testing.T) {
	payer := testAddress(feePayerSeed).String()
	if _, err := TipInstruction(payer, transaction.MinTipLamports-1); err == nil {
		t.Fatal("tip below the minimum accepted")
	}

	if _, err := TipInstruction("payer", transaction.MinTipLamports); err == nil {
		t.Fatal("malformed payer accepted")
	}
}
