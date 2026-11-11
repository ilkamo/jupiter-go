package send

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"slices"

	"filippo.io/edwards25519"
	"github.com/mr-tron/base58"
)

const (
	// InviteCodeLength is the documented invite code length in base58 characters.
	InviteCodeLength = 12

	// ProgramID is the Send program that owns invite accounts.
	ProgramID = "inv1tEtSwRMtM44tbvJGNiTxMvDfPVnX9StyqXfDfks"

	// inviteCodeEntropyBytes random bytes encode to more than InviteCodeLength
	// base58 characters; the code keeps the first InviteCodeLength, as documented.
	inviteCodeEntropyBytes = 13

	inviteKeyPrefix   = "invite:"
	invitePDASeed     = "invite"
	pdaMarker         = "ProgramDerivedAddress"
	lowestPDABumpSeed = 1
)

var ErrInvalidInviteCode = errors.New("invalid invite code")

// InviteCode is the secret that lets its holder claim a Send invite. It derives
// the invite's signing key, so anyone who learns it can take the funds. Printing an
// InviteCode with fmt shows a redacted placeholder; call Reveal to share it with
// the recipient.
type InviteCode struct {
	code string
}

// NewInviteCode generates a random invite code from crypto/rand.
func NewInviteCode() (InviteCode, error) {
	entropy := make([]byte, inviteCodeEntropyBytes)
	if _, err := rand.Read(entropy); err != nil {
		return InviteCode{}, fmt.Errorf("generate invite code: %w", err)
	}

	encoded := base58.Encode(entropy)

	return InviteCode{code: encoded[:InviteCodeLength]}, nil
}

// ParseInviteCode accepts a InviteCodeLength-character base58 invite code.
func ParseInviteCode(code string) (InviteCode, error) {
	if len(code) != InviteCodeLength {
		return InviteCode{}, fmt.Errorf("%w: want %d characters", ErrInvalidInviteCode, InviteCodeLength)
	}

	if _, err := base58.Decode(code); err != nil {
		return InviteCode{}, fmt.Errorf("%w: %w", ErrInvalidInviteCode, err)
	}

	return InviteCode{code: code}, nil
}

// Reveal returns the invite code for delivery to the recipient, for example as a
// link or QR code for Jupiter Mobile. Treat it like a private key.
func (c InviteCode) Reveal() string {
	return c.code
}

func (c InviteCode) String() string {
	return "send.InviteCode(redacted)"
}

// GoString redacts the code for the %#v verb.
func (c InviteCode) GoString() string {
	return c.String()
}

// PrivateKey derives the invite's signing key: the ed25519 seed is
// SHA-256("invite:" + code). It implements crypto.Signer for solanatx.
func (c InviteCode) PrivateKey() ed25519.PrivateKey {
	seed := sha256.Sum256([]byte(inviteKeyPrefix + c.code))

	return ed25519.NewKeyFromSeed(seed[:])
}

// Signer returns the invite's public key, passed to CraftSend as InviteSigner.
func (c InviteCode) Signer() string {
	return base58.Encode(c.PrivateKey().Public().(ed25519.PublicKey))
}

// PDA returns the invite account address passed to CraftClawback.
func (c InviteCode) PDA() (string, error) {
	return InvitePDA(c.Signer())
}

// InvitePDA derives the invite account address from an invite signer public key:
// the program address for seeds "invite" and the signer under ProgramID.
func InvitePDA(inviteSigner string) (string, error) {
	signer, err := base58.Decode(inviteSigner)
	if err != nil || len(signer) != ed25519.PublicKeySize {
		return "", fmt.Errorf("invite signer must be a base58 public key")
	}

	program, err := base58.Decode(ProgramID)
	if err != nil {
		return "", fmt.Errorf("decode program ID: %w", err)
	}

	address, err := findProgramAddress([][]byte{[]byte(invitePDASeed), signer}, program)
	if err != nil {
		return "", err
	}

	return base58.Encode(address), nil
}

// findProgramAddress follows Solana's derivation: try bump seeds from 255 down to
// 1 and return the first SHA-256 hash of seeds, bump, program ID, and the PDA
// marker that is not a valid ed25519 point.
func findProgramAddress(seeds [][]byte, programID []byte) ([]byte, error) {
	var prefix []byte
	for _, seed := range seeds {
		prefix = append(prefix, seed...)
	}

	for bump := math.MaxUint8; bump >= lowestPDABumpSeed; bump-- {
		input := slices.Concat(prefix, []byte{byte(bump)}, programID, []byte(pdaMarker))
		candidate := sha256.Sum256(input)

		if _, err := new(edwards25519.Point).SetBytes(candidate[:]); err != nil {
			return candidate[:], nil
		}
	}

	return nil, fmt.Errorf("no program address found off the ed25519 curve")
}
