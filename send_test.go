package jupiter_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/mr-tron/base58"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/send"
	"github.com/ilkamo/jupiter-go/solanatx"
)

const (
	generatedInviteCodes = 200
	inviteFixtureCode    = "AbCdEfGh2345"
)

func TestInviteCodeGeneration(t *testing.T) {
	seen := make(map[string]bool, generatedInviteCodes)
	for range generatedInviteCodes {
		code, err := send.NewInviteCode()
		if err != nil {
			t.Fatal(err)
		}

		revealed := code.Reveal()
		if _, err := send.ParseInviteCode(revealed); err != nil {
			t.Fatalf("generated code %q does not parse: %v", revealed, err)
		}

		if seen[revealed] {
			t.Fatalf("duplicate invite code %q", revealed)
		}

		seen[revealed] = true
	}
}

func TestInviteCodeDerivation(t *testing.T) {
	code, err := send.ParseInviteCode(inviteFixtureCode)
	if err != nil {
		t.Fatal(err)
	}

	seed := sha256.Sum256([]byte("invite:" + inviteFixtureCode))
	key := code.PrivateKey()
	if !bytes.Equal(key.Seed(), seed[:]) || len(key) != ed25519.PrivateKeySize {
		t.Fatal("private key is not the SHA-256 seed of the prefixed code")
	}

	public := key.Public().(ed25519.PublicKey)
	if !bytes.Equal(key[ed25519.SeedSize:], public) || code.Signer() != base58.Encode(public) {
		t.Fatal("keypair layout or signer differs from [seed || public key]")
	}
}

func TestInvitePDAMatchesSolanaDerivation(t *testing.T) {
	program := solana.MustPublicKeyFromBase58(send.ProgramID)

	for range generatedInviteCodes {
		code, err := send.NewInviteCode()
		if err != nil {
			t.Fatal(err)
		}

		signer := solana.MustPublicKeyFromBase58(code.Signer())
		want, _, err := solana.FindProgramAddress([][]byte{[]byte("invite"), signer[:]}, program)
		if err != nil {
			t.Fatal(err)
		}

		got, err := code.PDA()
		if err != nil {
			t.Fatal(err)
		}

		if got != want.String() {
			t.Fatalf("PDA = %s, want %s", got, want)
		}
	}
}

func TestInviteCodeIsRedactedWhenFormatted(t *testing.T) {
	code, err := send.ParseInviteCode(inviteFixtureCode)
	if err != nil {
		t.Fatal(err)
	}

	holder := struct{ Code send.InviteCode }{Code: code}
	formatted := fmt.Sprintf("%v %s %+v %#v %v %+v", code, code, code, code, holder, holder)
	if strings.Contains(formatted, inviteFixtureCode) {
		t.Fatalf("invite code leaked: %s", formatted)
	}

	if code.Reveal() != inviteFixtureCode {
		t.Fatal("Reveal did not return the code")
	}
}

func TestParseInviteCodeRejectsMalformedCodes(t *testing.T) {
	malformed := []string{
		"",
		"AbCdEfGh234",
		"AbCdEfGh23456",
		"AbCdEfGh2340",
		"AbCdEfGh234O",
		"AbCdEfGh234 ",
	}
	for _, value := range malformed {
		if _, err := send.ParseInviteCode(value); !errors.Is(err, send.ErrInvalidInviteCode) {
			t.Fatalf("ParseInviteCode(%q) = %v", value, err)
		}
	}
}

// inviteSendTransaction builds an unsigned transaction requiring the sender and
// the invite signer, like a craft-send response.
func inviteSendTransaction(t *testing.T, sender solana.PublicKey, invite solana.PublicKey) string {
	t.Helper()

	instruction := solana.NewInstruction(
		solana.MustPublicKeyFromBase58(send.ProgramID),
		solana.AccountMetaSlice{
			solana.NewAccountMeta(sender, true, true),
			solana.NewAccountMeta(invite, true, true),
		},
		[]byte{1},
	)

	tx, err := solana.NewTransaction(
		[]solana.Instruction{instruction},
		solana.Hash{},
		solana.TransactionPayer(sender),
	)
	if err != nil {
		t.Fatal(err)
	}

	encoded, err := tx.ToBase64()
	if err != nil {
		t.Fatal(err)
	}

	return encoded
}

func TestCraftSendSignedBySenderAndInvite(t *testing.T) {
	code, err := send.ParseInviteCode(inviteFixtureCode)
	if err != nil {
		t.Fatal(err)
	}

	senderKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{3}, ed25519.SeedSize))
	sender := solana.PublicKeyFromBytes(senderKey.Public().(ed25519.PublicKey))
	invite := solana.MustPublicKeyFromBase58(code.Signer())
	reply := `{"tx":"` + inviteSendTransaction(t, sender, invite) +
		`","expiry":"2025-03-22T10:30:00.000Z","totalFeeLamports":"5000"}`

	client, recorded := pathServer(t, map[string]string{"/send/v1/craft-send": reply})

	built, err := client.Send.CraftSend(context.Background(), send.SendRequest{
		InviteSigner: code.Signer(),
		Sender:       sender.String(),
		Amount:       "0010000000",
	})
	if err != nil {
		t.Fatal(err)
	}

	body := recorded()[0].body
	signerSent := string(body["inviteSigner"]) == `"`+code.Signer()+`"`
	if string(body["amount"]) != `"10000000"` || !signerSent {
		t.Fatalf("body = %v", body)
	}

	if _, ok := body["mint"]; ok {
		t.Fatal("empty mint was sent instead of the wrapped SOL default")
	}

	tx, err := solanatx.Parse(built.Transaction)
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.Sign(senderKey, code.PrivateKey()); err != nil {
		t.Fatal(err)
	}

	if missing := tx.MissingSigners(); len(missing) != 0 || built.TotalFeeLamports != "5000" {
		t.Fatalf("missing=%v build=%+v", missing, built)
	}
}

func TestSendClawbackAndInvites(t *testing.T) {
	code, err := send.ParseInviteCode(inviteFixtureCode)
	if err != nil {
		t.Fatal(err)
	}

	pda, err := code.PDA()
	if err != nil {
		t.Fatal(err)
	}

	invites := `{"invites": [
  {"invite_pda": "a", "confirmed": true, "deletion_tx": null},
  {"invite_pda": "b", "confirmed": 0},
  {"invite_pda": "c", "confirmed": null}
], "hasMoreData": true}`
	client, recorded := pathServer(t, map[string]string{
		"/send/v1/craft-clawback":  `{"tx": "AQ=="}`,
		"/send/v1/pending-invites": invites,
		"/send/v1/invite-history":  `{"invites": [], "hasMoreData": false}`,
	})

	ctx := context.Background()
	request := send.ClawbackRequest{InvitePDA: pda, Sender: takerAddress}
	if _, err := client.Send.CraftClawback(ctx, request); err != nil {
		t.Fatal(err)
	}

	pending, err := client.Send.PendingInvites(ctx, takerAddress, 2)
	if err != nil {
		t.Fatal(err)
	}

	confirmedTrue := pending.Invites[0].Confirmed
	confirmedZero := pending.Invites[1].Confirmed
	confirmedNull := pending.Invites[2].Confirmed
	decoded := confirmedTrue.Known && confirmedTrue.Value &&
		confirmedZero.Known && !confirmedZero.Value &&
		!confirmedNull.Known
	if !decoded || !pending.HasMoreData {
		t.Fatalf("invites = %+v", pending.Invites)
	}

	if _, err := client.Send.InviteHistory(ctx, takerAddress, 0); err != nil {
		t.Fatal(err)
	}

	requests := recorded()
	if string(requests[0].body["invitePDA"]) != `"`+pda+`"` {
		t.Fatalf("clawback body = %v", requests[0].body)
	}

	if requests[1].query.Get("page") != "2" || requests[1].query.Get("address") != takerAddress {
		t.Fatalf("pending query = %v", requests[1].query)
	}

	if _, ok := requests[2].query["page"]; ok {
		t.Fatal("page zero was sent instead of the default")
	}
}

func TestInvalidSendRequestsDoNotReachTransport(t *testing.T) {
	client, err := jupiter.NewClient(jupiter.WithHTTPClient(unreachableHTTPClient{t: t}))
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	valid := send.SendRequest{InviteSigner: payerAddress, Sender: takerAddress, Amount: "1"}

	tests := []struct {
		name         string
		configure    func(*send.SendRequest)
		addressError bool
	}{
		{
			name:         "malformed invite signer",
			configure:    func(r *send.SendRequest) { r.InviteSigner = "invite" },
			addressError: true,
		},
		{
			name:      "zero amount",
			configure: func(r *send.SendRequest) { r.Amount = "0" },
		},
		{
			name:         "malformed mint",
			configure:    func(r *send.SendRequest) { r.Mint = "USDC" },
			addressError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := valid
			test.configure(&request)

			_, err := client.Send.CraftSend(ctx, request)
			if err == nil {
				t.Fatal("invalid request accepted")
			}

			if errors.Is(err, send.ErrInvalidAddress) != test.addressError {
				t.Fatalf("unexpected address error classification: %v", err)
			}
		})
	}

	clawback := send.ClawbackRequest{InvitePDA: "pda", Sender: takerAddress}
	if _, err := client.Send.CraftClawback(ctx, clawback); !errors.Is(err, send.ErrInvalidAddress) {
		t.Fatalf("malformed clawback PDA: %v", err)
	}

	if _, err := client.Send.PendingInvites(ctx, "sender", 1); !errors.Is(err, send.ErrInvalidAddress) {
		t.Fatalf("malformed invites address: %v", err)
	}
}
