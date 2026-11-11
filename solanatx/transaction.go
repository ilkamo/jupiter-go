// Package solanatx adapts Jupiter transactions to Solana signing without changing
// their messages. It is the only part of the SDK that depends on solana-go; the root
// jupiter package does not import it.
//
// Signing never rewrites the message, blockhash, or existing signatures: each
// signature covers the exact message bytes received and is written only into its
// signer's slot. Legacy, version 0, and version 1 (SIMD-0385) transactions are
// supported; other versions are rejected with ErrUnsupportedVersion.
package solanatx

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
)

const (
	// versionedMessagePrefix marks a versioned message; the low bits hold the version.
	versionedMessagePrefix byte = 0x80

	version0Number byte = 0
	version1Number byte = 1

	// version1RequiredSignaturesOffset locates the header's required-signature count,
	// directly after the version prefix.
	version1RequiredSignaturesOffset = 1

	// pureEd25519 selects Ed25519 over the full message rather than a prehash.
	pureEd25519 = crypto.Hash(0)
)

var (
	ErrMalformedTransaction = errors.New("malformed transaction")
	ErrUnsupportedVersion   = errors.New("unsupported transaction version")
	ErrUnsupportedSigner    = errors.New("signer does not provide an ed25519 public key")
	ErrSignerNotRequired    = errors.New("signer is not a required transaction signer")
	ErrInvalidSignature     = errors.New("signer returned a signature that does not verify")
)

type Version uint8

const (
	VersionLegacy Version = iota
	Version0
	Version1
)

func (v Version) String() string {
	switch v {
	case VersionLegacy:
		return "legacy"
	case Version0:
		return "v0"
	case Version1:
		return "v1"
	default:
		return fmt.Sprintf("Version(%d)", uint8(v))
	}
}

// Transaction is a serialized Solana transaction whose message bytes are fixed.
// It is not safe for concurrent signing.
type Transaction struct {
	wire    []byte
	layout  wireLayout
	signers []solana.PublicKey
}

// wireLayout locates the message and signature slots. Legacy and version 0
// transactions put a signature count and the signatures before the message;
// version 1 puts exactly the required signatures after the message, uncounted.
type wireLayout struct {
	version          Version
	signatureCount   int
	signaturesOffset int
	messageStart     int
	messageEnd       int
}

// Parse decodes a standard base64 wire transaction, such as OrderResponse.Transaction.
func Parse(encoded string) (*Transaction, error) {
	wire, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("%w: decode base64: %w", ErrMalformedTransaction, err)
	}

	return parseWire(wire)
}

func parseWire(wire []byte) (*Transaction, error) {
	layout, err := locate(wire)
	if err != nil {
		return nil, err
	}

	parsed, err := solana.TransactionFromBytes(wire)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformedTransaction, err)
	}

	message, err := parsed.Message.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("%w: encode message: %w", ErrMalformedTransaction, err)
	}

	// Signatures must cover the received bytes. Requiring an exact re-encoding also
	// rejects trailing data and layouts the parser does not fully understand.
	if !bytes.Equal(message, wire[layout.messageStart:layout.messageEnd]) {
		return nil, fmt.Errorf(
			"%w: message does not re-encode to the received bytes",
			ErrMalformedTransaction,
		)
	}

	required := int(parsed.Message.Header.NumRequiredSignatures)
	if layout.signatureCount != required {
		return nil, fmt.Errorf(
			"%w: %d signatures for %d required signers",
			ErrMalformedTransaction,
			layout.signatureCount,
			required,
		)
	}

	if len(parsed.Message.AccountKeys) < required {
		return nil, fmt.Errorf("%w: fewer account keys than required signers", ErrMalformedTransaction)
	}

	signers := make([]solana.PublicKey, required)
	copy(signers, parsed.Message.AccountKeys[:required])

	return &Transaction{
		wire:    wire,
		layout:  layout,
		signers: signers,
	}, nil
}

// locate reads the version and layout before solana-go parses the transaction, so
// unsupported versions are reported as such rather than as malformed input.
//
// A leading byte with the high bit set is a version prefix. In a legacy or v0
// transaction it would be a signature count of at least 128, which cannot fit in
// any Solana transaction size limit.
func locate(wire []byte) (wireLayout, error) {
	if len(wire) == 0 {
		return wireLayout{}, fmt.Errorf("%w: empty", ErrMalformedTransaction)
	}

	if wire[0]&versionedMessagePrefix == 0 {
		return locateCounted(wire)
	}

	number := wire[0] &^ versionedMessagePrefix
	if number != version1Number {
		return wireLayout{}, fmt.Errorf("%w: %d", ErrUnsupportedVersion, number)
	}

	return locateVersion1(wire)
}

func locateCounted(wire []byte) (wireLayout, error) {
	decoder := bin.NewBinDecoder(wire)
	signatureCount, err := decoder.ReadCompactU16()
	if err != nil {
		return wireLayout{}, fmt.Errorf("%w: read signature count: %w", ErrMalformedTransaction, err)
	}

	signaturesOffset := int(decoder.Position())
	messageStart := signaturesOffset + signatureCount*ed25519.SignatureSize
	if messageStart >= len(wire) {
		return wireLayout{}, fmt.Errorf("%w: missing message", ErrMalformedTransaction)
	}

	version := VersionLegacy
	prefix := wire[messageStart]
	if prefix&versionedMessagePrefix != 0 {
		// Version 1 messages never use the counted layout.
		number := prefix &^ versionedMessagePrefix
		if number != version0Number {
			return wireLayout{}, fmt.Errorf("%w: %d in a counted layout", ErrUnsupportedVersion, number)
		}

		version = Version0
	}

	return wireLayout{
		version:          version,
		signatureCount:   signatureCount,
		signaturesOffset: signaturesOffset,
		messageStart:     messageStart,
		messageEnd:       len(wire),
	}, nil
}

func locateVersion1(wire []byte) (wireLayout, error) {
	if len(wire) <= version1RequiredSignaturesOffset {
		return wireLayout{}, fmt.Errorf("%w: missing message header", ErrMalformedTransaction)
	}

	signatureCount := int(wire[version1RequiredSignaturesOffset])
	messageEnd := len(wire) - signatureCount*ed25519.SignatureSize
	if messageEnd <= version1RequiredSignaturesOffset {
		return wireLayout{}, fmt.Errorf("%w: missing signatures", ErrMalformedTransaction)
	}

	return wireLayout{
		version:          Version1,
		signatureCount:   signatureCount,
		signaturesOffset: messageEnd,
		messageStart:     0,
		messageEnd:       messageEnd,
	}, nil
}

func (t *Transaction) Version() Version {
	return t.layout.version
}

// RequiredSigners returns the base58 signer addresses in signature-slot order.
// The first is the fee payer.
func (t *Transaction) RequiredSigners() []string {
	signers := make([]string, len(t.signers))
	for i, signer := range t.signers {
		signers[i] = signer.String()
	}

	return signers
}

// MissingSigners returns required signers whose signature slot is still empty.
func (t *Transaction) MissingSigners() []string {
	var missing []string
	for i, signer := range t.signers {
		if isEmptySignature(t.signatureSlot(i)) {
			missing = append(missing, signer.String())
		}
	}

	return missing
}

// ID returns the transaction signature that identifies it on-chain: the fee payer's
// signature in base58. It is empty until the fee payer has signed.
func (t *Transaction) ID() string {
	if len(t.signers) == 0 {
		return ""
	}

	slot := t.signatureSlot(0)
	if isEmptySignature(slot) {
		return ""
	}

	return solana.SignatureFromBytes(slot).String()
}

// Sign adds a signature from each signer to its own slot. Every signature is
// verified before any is written, so a failure leaves the transaction unchanged.
func (t *Transaction) Sign(signers ...crypto.Signer) error {
	message := t.message()
	signatures := make(map[int][]byte, len(signers))

	for _, signer := range signers {
		publicKey, ok := signer.Public().(ed25519.PublicKey)
		if !ok || len(publicKey) != ed25519.PublicKeySize {
			return ErrUnsupportedSigner
		}

		address := solana.PublicKeyFromBytes(publicKey)
		index := t.signerIndex(address)
		if index < 0 {
			return fmt.Errorf("%w: %s", ErrSignerNotRequired, address)
		}

		signature, err := signer.Sign(rand.Reader, message, pureEd25519)
		if err != nil {
			return fmt.Errorf("sign with %s: %w", address, err)
		}

		if !ed25519.Verify(publicKey, message, signature) {
			return fmt.Errorf("%w: %s", ErrInvalidSignature, address)
		}

		signatures[index] = signature
	}

	for index, signature := range signatures {
		copy(t.signatureSlot(index), signature)
	}

	return nil
}

// Base64 returns the wire transaction for ExecuteRequest.SignedTransaction.
func (t *Transaction) Base64() string {
	return base64.StdEncoding.EncodeToString(t.wire)
}

// Bytes returns a copy of the wire transaction, for example for RPC submission.
func (t *Transaction) Bytes() []byte {
	return bytes.Clone(t.wire)
}

func (t *Transaction) signerIndex(address solana.PublicKey) int {
	for i, signer := range t.signers {
		if signer.Equals(address) {
			return i
		}
	}

	return -1
}

func (t *Transaction) message() []byte {
	return t.wire[t.layout.messageStart:t.layout.messageEnd]
}

func (t *Transaction) signatureSlot(index int) []byte {
	start := t.layout.signaturesOffset + index*ed25519.SignatureSize

	return t.wire[start : start+ed25519.SignatureSize]
}

func isEmptySignature(slot []byte) bool {
	var empty [ed25519.SignatureSize]byte

	return bytes.Equal(slot, empty[:])
}
