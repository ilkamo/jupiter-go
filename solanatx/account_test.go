package solanatx

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/gagliardetto/solana-go"
)

const jitoDontFrontPrefix = "jitodontfront"

func TestJitoDontFrontIsValidAddress(t *testing.T) {
	key, err := solana.PublicKeyFromBase58(JitoDontFront)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(key.String(), jitoDontFrontPrefix) {
		t.Fatalf("address %s lacks the %q prefix", key, jitoDontFrontPrefix)
	}
}

func TestWithReadOnlyAccount(t *testing.T) {
	payer := testAddress(feePayerSeed)

	tip, err := TipInstruction(payer.String(), tipTestLamports)
	if err != nil {
		t.Fatal(err)
	}

	protected, err := WithReadOnlyAccount(tip, JitoDontFront)
	if err != nil {
		t.Fatal(err)
	}

	if len(tip.Accounts()) != 2 {
		t.Fatalf("original instruction changed: %d accounts", len(tip.Accounts()))
	}

	accounts := protected.Accounts()
	if len(accounts) != 3 || !protected.ProgramID().Equals(tip.ProgramID()) {
		t.Fatalf("program=%s accounts=%d", protected.ProgramID(), len(accounts))
	}

	added := accounts[2]
	if added.PublicKey.String() != JitoDontFront || added.IsSigner || added.IsWritable {
		t.Fatalf("added account = %+v", added)
	}

	originalData, err := tip.Data()
	if err != nil {
		t.Fatal(err)
	}

	protectedData, err := protected.Data()
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(originalData, protectedData) {
		t.Fatalf("data changed: %v != %v", protectedData, originalData)
	}

	build := testBuild(t)

	tx, err := Assemble(build, payer.String(), protected)
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := solana.TransactionFromBytes(tx.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	tables, err := AddressTables(build)
	if err != nil {
		t.Fatal(err)
	}

	if err := decoded.Message.SetAddressTables(tables); err != nil {
		t.Fatal(err)
	}

	key := solana.MustPublicKeyFromBase58(JitoDontFront)
	if !slices.Contains(decoded.Message.AccountKeys, key) {
		t.Fatal("assembled message does not reference the account")
	}

	writable, err := decoded.Message.IsWritable(key)
	if err != nil {
		t.Fatal(err)
	}

	if writable || decoded.Message.IsSigner(key) {
		t.Fatalf("account writable=%v signer=%v", writable, decoded.Message.IsSigner(key))
	}
}

func TestWithReadOnlyAccountRejectsInvalidInputs(t *testing.T) {
	payer := testAddress(feePayerSeed).String()

	tip, err := TipInstruction(payer, tipTestLamports)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := WithReadOnlyAccount(tip, "not-an-address"); err == nil {
		t.Fatal("malformed account accepted")
	}

	_, err = WithReadOnlyAccount(tip, payer)
	if !errors.Is(err, ErrDuplicateAccount) {
		t.Fatalf("duplicate account error = %v", err)
	}
}
