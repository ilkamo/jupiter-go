// Package trigger implements Jupiter's Trigger V2 API: wallet authentication,
// vaults, and deposits.
//
// Trigger V2 is custodial: deposits move funds into a Privy-managed vault that
// Jupiter controls on the wallet's behalf until orders fill or are cancelled.
//
// Authenticated operations run on a Session, which belongs to exactly one wallet.
// The Client holds no wallet identity and can serve many sessions concurrently.
package trigger

import (
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/mr-tron/base58"

	"github.com/ilkamo/jupiter-go/internal/address"
	"github.com/ilkamo/jupiter-go/internal/transport"
)

const (
	ChallengeMessage     = "message"
	ChallengeTransaction = "transaction"

	authModeAccessRefresh = "access_refresh"
	tokenTypeBearer       = "Bearer"

	basePath = "/trigger/v2"

	// pureEd25519 selects Ed25519 over the full message rather than a prehash.
	pureEd25519 = crypto.Hash(0)
)

// repeatableRetry covers operations whose repetition is harmless: requesting a
// challenge or crafting a deposit returns a fresh single-use value, logout is
// documented as idempotent, and refresh tolerates a repeated request within a
// short grace window, so waits stay short.
var repeatableRetry = transport.RepeatableRetry

var (
	// ErrInvalidAddress is the same value as swap.ErrInvalidAddress.
	ErrInvalidAddress = address.ErrInvalid

	ErrUnsupportedSigner   = errors.New("signer does not provide an ed25519 public key")
	ErrUnexpectedTokenType = errors.New("unexpected token response")
)

type Client struct {
	transport *transport.Client
}

// NewClient is called by jupiter.NewClient, which applications use instead; it
// takes the SDK's internal transport.
func NewClient(t *transport.Client) *Client {
	return &Client{transport: t}
}

// Challenge is a wallet-ownership challenge that expires after five minutes.
// Message is set for ChallengeMessage; Transaction (base64, never submitted
// on-chain) is set for ChallengeTransaction.
type Challenge struct {
	Type        string `json:"type"`
	Message     string `json:"challenge"`
	Transaction string `json:"transaction"`
}

type challengeRequest struct {
	WalletPubkey string `json:"walletPubkey"`
	Type         string `json:"type"`
}

// RequestChallenge starts authentication. Use ChallengeTransaction for hardware
// wallets that can only sign transactions.
func (c *Client) RequestChallenge(
	ctx context.Context,
	wallet string,
	challengeType string,
) (*Challenge, error) {
	if err := address.Validate(wallet); err != nil {
		return nil, fmt.Errorf("wallet: %w", err)
	}

	if challengeType != ChallengeMessage && challengeType != ChallengeTransaction {
		return nil, fmt.Errorf("challenge type must be %q or %q", ChallengeMessage, ChallengeTransaction)
	}

	request := transport.Request{
		Method: http.MethodPost,
		Path:   basePath + "/auth/challenge",
		Body:   challengeRequest{WalletPubkey: wallet, Type: challengeType},
		Retry:  repeatableRetry,
	}

	var challenge Challenge
	if _, err := c.transport.Do(ctx, request, &challenge); err != nil {
		return nil, err
	}

	if challenge.Type != challengeType {
		return nil, fmt.Errorf("requested a %s challenge, received %q", challengeType, challenge.Type)
	}

	return &challenge, nil
}

// VerifyRequest submits a signed challenge. Set Signature (base58) for a message
// challenge or SignedTransaction (base64) for a transaction challenge.
type VerifyRequest struct {
	Wallet            string
	Type              string
	Signature         string
	SignedTransaction string
}

type verifyBody struct {
	Type              string `json:"type"`
	WalletPubkey      string `json:"walletPubkey"`
	Signature         string `json:"signature,omitempty"`
	SignedTransaction string `json:"signedTransaction,omitempty"`
	AuthMode          string `json:"authMode"`
}

// AuthTokens is an access and refresh token pair. Treat it as a secret: keep it out of
// logs. The refresh token rotates on every refresh.
type AuthTokens struct {
	AuthMode     string    `json:"authMode"`
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
	TokenType    string    `json:"tokenType"`
}

func (t AuthTokens) validate() error {
	if t.AuthMode != authModeAccessRefresh || t.TokenType != tokenTypeBearer {
		return fmt.Errorf(
			"%w: auth mode %q, token type %q",
			ErrUnexpectedTokenType,
			t.AuthMode,
			t.TokenType,
		)
	}

	if t.AccessToken == "" || t.RefreshToken == "" || t.ExpiresAt.IsZero() {
		return fmt.Errorf("%w: missing token or expiry", ErrUnexpectedTokenType)
	}

	return nil
}

// Verify exchanges a signed challenge for tokens and opens a session. It always
// requests access_refresh tokens; the deprecated access_only mode is not supported.
// Verify makes one attempt because a challenge can be used only once.
func (c *Client) Verify(
	ctx context.Context,
	r VerifyRequest,
	options SessionOptions,
) (*Session, error) {
	if err := address.Validate(r.Wallet); err != nil {
		return nil, fmt.Errorf("wallet: %w", err)
	}

	body := verifyBody{
		Type:         r.Type,
		WalletPubkey: r.Wallet,
		AuthMode:     authModeAccessRefresh,
	}

	switch r.Type {
	case ChallengeMessage:
		if r.Signature == "" || r.SignedTransaction != "" {
			return nil, fmt.Errorf("message verification requires only a signature")
		}

		body.Signature = r.Signature
	case ChallengeTransaction:
		if r.SignedTransaction == "" || r.Signature != "" {
			return nil, fmt.Errorf("transaction verification requires only a signed transaction")
		}

		body.SignedTransaction = r.SignedTransaction
	default:
		return nil, fmt.Errorf("challenge type must be %q or %q", ChallengeMessage, ChallengeTransaction)
	}

	request := transport.Request{
		Method: http.MethodPost,
		Path:   basePath + "/auth/verify",
		Body:   body,
		Retry:  transport.NoRetry,
	}

	var tokens AuthTokens
	if _, err := c.transport.Do(ctx, request, &tokens); err != nil {
		return nil, err
	}

	return c.Resume(r.Wallet, tokens, options)
}

// Login authenticates with a message challenge signed by signer, which must hold
// the wallet's ed25519 key. Hardware wallets that sign only transactions should use
// RequestChallenge, sign the challenge transaction, and call Verify.
func (c *Client) Login(
	ctx context.Context,
	signer crypto.Signer,
	options SessionOptions,
) (*Session, error) {
	publicKey, ok := signer.Public().(ed25519.PublicKey)
	if !ok || len(publicKey) != ed25519.PublicKeySize {
		return nil, ErrUnsupportedSigner
	}

	wallet := base58.Encode(publicKey)
	challenge, err := c.RequestChallenge(ctx, wallet, ChallengeMessage)
	if err != nil {
		return nil, fmt.Errorf("request challenge: %w", err)
	}

	message := []byte(challenge.Message)
	signature, err := signer.Sign(rand.Reader, message, pureEd25519)
	if err != nil {
		return nil, fmt.Errorf("sign challenge: %w", err)
	}

	if !ed25519.Verify(publicKey, message, signature) {
		return nil, fmt.Errorf("signer returned a challenge signature that does not verify")
	}

	verify := VerifyRequest{
		Wallet:    wallet,
		Type:      ChallengeMessage,
		Signature: base58.Encode(signature),
	}

	return c.Verify(ctx, verify, options)
}

// Resume opens a session from previously stored tokens without a network call.
// If the refresh token has been rotated or revoked since, the first refresh fails
// with ErrSessionExpired.
func (c *Client) Resume(
	wallet string,
	tokens AuthTokens,
	options SessionOptions,
) (*Session, error) {
	if err := address.Validate(wallet); err != nil {
		return nil, fmt.Errorf("wallet: %w", err)
	}

	if err := tokens.validate(); err != nil {
		return nil, err
	}

	return newSession(c, wallet, tokens, options), nil
}

type refreshTokenBody struct {
	RefreshToken string `json:"refreshToken"`
}

type successResponse struct {
	Success bool `json:"success"`
}

func (c *Client) logout(ctx context.Context, refreshToken string) error {
	request := transport.Request{
		Method: http.MethodPost,
		Path:   basePath + "/auth/logout",
		Body:   refreshTokenBody{RefreshToken: refreshToken},
		Retry:  repeatableRetry,
	}

	var result successResponse
	if _, err := c.transport.Do(ctx, request, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("logout was not confirmed")
	}

	return nil
}

func (c *Client) refresh(ctx context.Context, refreshToken string) (AuthTokens, error) {
	request := transport.Request{
		Method: http.MethodPost,
		Path:   basePath + "/auth/refresh",
		Body:   refreshTokenBody{RefreshToken: refreshToken},
		Retry:  repeatableRetry,
	}

	var tokens AuthTokens
	if _, err := c.transport.Do(ctx, request, &tokens); err != nil {
		return AuthTokens{}, err
	}

	if err := tokens.validate(); err != nil {
		return AuthTokens{}, err
	}

	return tokens, nil
}
