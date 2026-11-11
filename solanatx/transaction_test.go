package solanatx

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"testing"

	"github.com/gagliardetto/solana-go"
)

const (
	feePayerSeed byte = iota + 1
	cosignerSeed
	outsiderSeed
	programSeed
	blockhashSeed
)

func testKey(seed byte) ed25519.PrivateKey {
	return ed25519.NewKeyFromSeed(bytes.Repeat([]byte{seed}, ed25519.SeedSize))
}

func testAddress(seed byte) solana.PublicKey {
	return solana.PublicKeyFromBytes(testKey(seed).Public().(ed25519.PublicKey))
}

// unsignedFixture returns a transaction requiring the fee payer and a cosigner.
func unsignedFixture(t *testing.T, options ...solana.TransactionOption) string {
	t.Helper()

	instruction := solana.NewInstruction(
		testAddress(programSeed),
		solana.AccountMetaSlice{
			solana.NewAccountMeta(testAddress(feePayerSeed), true, true),
			solana.NewAccountMeta(testAddress(cosignerSeed), false, true),
		},
		[]byte{1, 2, 3},
	)

	blockhash := solana.HashFromBytes(bytes.Repeat([]byte{blockhashSeed}, len(solana.Hash{})))
	options = append(options, solana.TransactionPayer(testAddress(feePayerSeed)))
	tx, err := solana.NewTransaction([]solana.Instruction{instruction}, blockhash, options...)
	if err != nil {
		t.Fatal(err)
	}

	encoded, err := tx.ToBase64()
	if err != nil {
		t.Fatal(err)
	}

	return encoded
}

func parseFixture(t *testing.T, encoded string) *Transaction {
	t.Helper()

	tx, err := Parse(encoded)
	if err != nil {
		t.Fatal(err)
	}

	return tx
}

func unsignedLegacy(t *testing.T) string {
	t.Helper()

	return unsignedFixture(t)
}

func unsignedVersion1(t *testing.T) string {
	t.Helper()

	return unsignedFixture(t, solana.TransactionV1Config(solana.TransactionConfig{}.
		WithComputeUnitLimit(fixtureComputeUnits).
		WithPriorityFee(fixturePriorityFeeLamports)))
}

const (
	fixtureComputeUnits        = 20_000
	fixturePriorityFeeLamports = 1_000
)

func TestPartialSigningPreservesMessageAndSignatures(t *testing.T) {
	tests := []struct {
		name    string
		fixture func(*testing.T) string
		version Version
	}{
		{name: "legacy", fixture: unsignedLegacy, version: VersionLegacy},
		{name: "version 1", fixture: unsignedVersion1, version: Version1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := parseFixture(t, test.fixture(t))
			message := bytes.Clone(tx.message())

			if tx.Version() != test.version || len(tx.RequiredSigners()) != 2 {
				t.Fatalf("version=%v signers=%v", tx.Version(), tx.RequiredSigners())
			}

			if tx.ID() != "" || len(tx.MissingSigners()) != 2 {
				t.Fatal("unsigned transaction reports signatures")
			}

			if err := tx.Sign(testKey(cosignerSeed)); err != nil {
				t.Fatal(err)
			}

			cosignature := bytes.Clone(tx.signatureSlot(1))
			missing := tx.MissingSigners()
			if len(missing) != 1 || missing[0] != testAddress(feePayerSeed).String() || tx.ID() != "" {
				t.Fatalf("after cosigning: missing=%v id=%q", missing, tx.ID())
			}

			// A second party signs a transaction that already carries a signature.
			resumed := parseFixture(t, tx.Base64())
			if err := resumed.Sign(testKey(feePayerSeed)); err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(resumed.message(), message) {
				t.Fatal("message bytes changed")
			}

			if !bytes.Equal(resumed.signatureSlot(1), cosignature) {
				t.Fatal("existing signature changed")
			}

			if len(resumed.MissingSigners()) != 0 {
				t.Fatalf("still missing: %v", resumed.MissingSigners())
			}

			verified, err := solana.TransactionFromBytes(resumed.Bytes())
			if err != nil {
				t.Fatal(err)
			}

			if err := verified.VerifySignatures(); err != nil {
				t.Fatalf("solana-go rejects signatures: %v", err)
			}

			if resumed.ID() != verified.Signatures[0].String() {
				t.Fatalf("ID %q differs from fee payer signature", resumed.ID())
			}
		})
	}
}

type brokenSigner struct {
	key ed25519.PrivateKey
}

func (s brokenSigner) Public() crypto.PublicKey {
	return s.key.Public()
}

func (s brokenSigner) Sign(io.Reader, []byte, crypto.SignerOpts) ([]byte, error) {
	return make([]byte, ed25519.SignatureSize), nil
}

type failingSigner struct {
	key ed25519.PrivateKey
	err error
}

func (s failingSigner) Public() crypto.PublicKey {
	return s.key.Public()
}

func (s failingSigner) Sign(io.Reader, []byte, crypto.SignerOpts) ([]byte, error) {
	return nil, s.err
}

func TestSignRejectsWithoutChangingTransaction(t *testing.T) {
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	signerFailure := errors.New("hardware wallet unavailable")

	tests := []struct {
		name    string
		signers []crypto.Signer
		want    error
	}{
		{
			name:    "not a required signer",
			signers: []crypto.Signer{testKey(feePayerSeed), testKey(outsiderSeed)},
			want:    ErrSignerNotRequired,
		},
		{
			name:    "signature does not verify",
			signers: []crypto.Signer{testKey(feePayerSeed), brokenSigner{key: testKey(cosignerSeed)}},
			want:    ErrInvalidSignature,
		},
		{
			name:    "non-ed25519 key",
			signers: []crypto.Signer{testKey(feePayerSeed), ecdsaKey},
			want:    ErrUnsupportedSigner,
		},
		{
			name:    "signer failure",
			signers: []crypto.Signer{failingSigner{key: testKey(cosignerSeed), err: signerFailure}},
			want:    signerFailure,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx := parseFixture(t, unsignedLegacy(t))
			before := tx.Bytes()

			err := tx.Sign(test.signers...)
			if !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}

			if !bytes.Equal(tx.wire, before) {
				t.Fatal("failed signing changed the transaction")
			}
		})
	}
}

func TestParseRejectsUnsupportedVersions(t *testing.T) {
	version0 := assembleFixture(t, testBuild(t))

	tests := []struct {
		name     string
		position int
		prefix   byte
	}{
		{
			name:     "version 1 message in counted layout",
			position: version0.layout.messageStart,
			prefix:   versionedMessagePrefix | version1Number,
		},
		{
			name:     "highest version in counted layout",
			position: version0.layout.messageStart,
			prefix:   versionedMessagePrefix | 0x7f,
		},
		{name: "version 2 prefix", position: 0, prefix: versionedMessagePrefix | 2},
		{name: "highest version prefix", position: 0, prefix: versionedMessagePrefix | 0x7f},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wire := version0.Bytes()
			wire[test.position] = test.prefix

			_, err := Parse(base64.StdEncoding.EncodeToString(wire))
			if !errors.Is(err, ErrUnsupportedVersion) {
				t.Fatalf("expected unsupported version, got %v", err)
			}
		})
	}
}

func TestParseRejectsMalformedTransactions(t *testing.T) {
	valid, err := base64.StdEncoding.DecodeString(unsignedLegacy(t))
	if err != nil {
		t.Fatal(err)
	}

	version1, err := base64.StdEncoding.DecodeString(unsignedVersion1(t))
	if err != nil {
		t.Fatal(err)
	}

	// The legacy fixture starts with a one-byte signature count and two signatures.
	const signatureCountBytes = 1
	firstSignatureEnd := signatureCountBytes + ed25519.SignatureSize
	oneSignature := append([]byte{1}, valid[firstSignatureEnd:]...)
	encode := base64.StdEncoding.EncodeToString

	tests := []struct {
		name    string
		encoded string
	}{
		{name: "not base64", encoded: "not base64"},
		{name: "empty", encoded: ""},
		{name: "signatures without message", encoded: encode(valid[:firstSignatureEnd])},
		{name: "truncated message", encoded: encode(valid[:len(valid)-1])},
		{name: "trailing data", encoded: encode(append(bytes.Clone(valid), 0))},
		{name: "signature count mismatch", encoded: encode(oneSignature)},
		{name: "version 1 missing signature byte", encoded: encode(version1[:len(version1)-1])},
		{name: "version 1 trailing data", encoded: encode(append(bytes.Clone(version1), 0))},
		{name: "version 1 header only", encoded: encode(version1[:version1RequiredSignaturesOffset+1])},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.encoded)
			if !errors.Is(err, ErrMalformedTransaction) {
				t.Fatalf("expected malformed transaction, got %v", err)
			}
		})
	}
}
