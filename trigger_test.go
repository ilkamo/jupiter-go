package jupiter_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mr-tron/base58"

	jupiter "github.com/ilkamo/jupiter-go"
	"github.com/ilkamo/jupiter-go/trigger"
)

const (
	challengeText    = "Sign this message to authenticate with Jupiter. Nonce 1."
	vaultAddress     = "7nE9GJoYHNmtaQvTQpota3KV2oz4pQ2dA6nvYK8EUJHV"
	accessTokenTTL   = 15 * time.Minute
	nearlyExpiredTTL = 30 * time.Second
)

var triggerWalletKey = ed25519.NewKeyFromSeed(bytes.Repeat([]byte{7}, ed25519.SeedSize))

func triggerWallet() string {
	return base58.Encode(triggerWalletKey.Public().(ed25519.PublicKey))
}

// fakeTrigger emulates the documented Trigger auth and vault rules: signed
// challenges, rotating refresh tokens whose reuse is rejected, and bearer checks.
type fakeTrigger struct {
	t *testing.T

	mu               sync.Mutex
	generation       int
	accessToken      string
	refreshToken     string
	revokedAccess    bool
	vaultRegistered  bool
	registerConflict bool
	refreshCalls     int
	verifyBodies     []map[string]string
	logoutTokens     []string
	accessOnly       bool

	routes   map[string]fakeRoute
	requests []fakeRequest
}

// fakeRoute is a canned response for an authorized "METHOD path" route.
type fakeRoute struct {
	status int
	body   string
}

type fakeRequest struct {
	method string
	path   string
	query  map[string][]string
	body   []byte
}

func (f *fakeTrigger) route(method string, path string, status int, body string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.routes == nil {
		f.routes = make(map[string]fakeRoute)
	}

	f.routes[method+" "+path] = fakeRoute{status: status, body: body}
}

func (f *fakeTrigger) recorded() []fakeRequest {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]fakeRequest(nil), f.requests...)
}

func newFakeTrigger(t *testing.T) (*fakeTrigger, *jupiter.Client) {
	t.Helper()

	fake := &fakeTrigger{t: t}
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)

	client, err := jupiter.NewClient(jupiter.WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	return fake, client
}

// issue rotates to a new token pair; callers hold the lock.
func (f *fakeTrigger) issue(ttl time.Duration) trigger.Tokens {
	f.generation++
	f.accessToken = fmt.Sprintf("access-%d", f.generation)
	f.refreshToken = fmt.Sprintf("refresh-%d", f.generation)
	f.revokedAccess = false

	return trigger.Tokens{
		AuthMode:     "access_refresh",
		AccessToken:  f.accessToken,
		RefreshToken: f.refreshToken,
		ExpiresAt:    time.Now().Add(ttl).UTC(),
		TokenType:    "Bearer",
	}
}

func (f *fakeTrigger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		f.t.Errorf("read %s body: %v", r.URL.Path, err)
	}

	body := stringFields(f.t, raw)
	path := strings.TrimPrefix(r.URL.EscapedPath(), "/trigger/v2")
	f.requests = append(f.requests, fakeRequest{
		method: r.Method,
		path:   path,
		query:  r.URL.Query(),
		body:   raw,
	})

	switch path {
	case "/auth/challenge":
		f.reply(w, http.StatusOK, map[string]string{"type": body["type"], "challenge": challengeText})
	case "/auth/verify":
		f.verify(w, body)
	case "/auth/refresh":
		f.refreshCalls++
		if body["refreshToken"] != f.refreshToken {
			f.refreshToken = ""
			f.reply(w, http.StatusUnauthorized, map[string]string{"error": "refresh token reused"})

			return
		}

		f.reply(w, http.StatusOK, f.issue(accessTokenTTL))
	case "/auth/logout":
		f.logoutTokens = append(f.logoutTokens, body["refreshToken"])
		f.reply(w, http.StatusOK, map[string]bool{"success": true})
	default:
		f.authorized(w, r, path, body)
	}
}

func (f *fakeTrigger) verify(w http.ResponseWriter, body map[string]string) {
	f.verifyBodies = append(f.verifyBodies, body)

	if f.accessOnly {
		f.reply(w, http.StatusOK, map[string]string{"authMode": "access_only", "token": "legacy"})

		return
	}

	if body["type"] == trigger.ChallengeMessage {
		signature, err := base58.Decode(body["signature"])
		publicKey, keyErr := base58.Decode(body["walletPubkey"])
		valid := err == nil && keyErr == nil &&
			ed25519.Verify(publicKey, []byte(challengeText), signature)
		if !valid {
			f.reply(w, http.StatusUnauthorized, map[string]string{"error": "Invalid signature"})

			return
		}
	}

	f.reply(w, http.StatusOK, f.issue(accessTokenTTL))
}

func (f *fakeTrigger) authorized(
	w http.ResponseWriter,
	r *http.Request,
	path string,
	body map[string]string,
) {
	bearer := "Bearer " + f.accessToken
	if f.accessToken == "" || f.revokedAccess || r.Header.Get("Authorization") != bearer {
		f.reply(w, http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})

		return
	}

	vault := map[string]string{
		"userPubkey":   triggerWallet(),
		"vaultPubkey":  vaultAddress,
		"privyVaultId": "vault-uuid",
	}

	switch path {
	case "/vault":
		if !f.vaultRegistered {
			f.reply(w, http.StatusNotFound, map[string]string{"error": "Vault not found"})

			return
		}

		f.reply(w, http.StatusOK, vault)
	case "/vault/register":
		if f.vaultRegistered || f.registerConflict {
			f.vaultRegistered = true
			f.reply(w, http.StatusConflict, map[string]string{"error": "Vault already registered"})

			return
		}

		f.vaultRegistered = true
		f.reply(w, http.StatusCreated, vault)
	case "/deposit/craft":
		if body["userAddress"] != triggerWallet() {
			f.reply(w, http.StatusForbidden, map[string]string{"error": "User address mismatch"})

			return
		}

		f.reply(w, http.StatusOK, map[string]any{
			"transaction":     "AQ==",
			"requestId":       "deposit-request",
			"receiverAddress": vaultAddress,
			"mint":            body["inputMint"],
			"amount":          body["amount"],
			"tokenDecimals":   9,
		})
	case "/auth/logout-all":
		f.refreshToken = ""
		f.reply(w, http.StatusOK, map[string]bool{"success": true})
	default:
		route, ok := f.routes[r.Method+" "+path]
		if !ok {
			f.reply(w, http.StatusNotFound, map[string]string{"error": "unknown path " + path})

			return
		}

		w.WriteHeader(route.status)

		if _, err := w.Write([]byte(route.body)); err != nil {
			f.t.Errorf("write response: %v", err)
		}
	}
}

// stringFields returns the string-valued top-level fields of a JSON object body.
func stringFields(t *testing.T, raw []byte) map[string]string {
	fields := make(map[string]string)
	if len(raw) == 0 {
		return fields
	}

	var values map[string]json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		t.Errorf("decode body: %v", err)

		return fields
	}

	for key, value := range values {
		var text string
		if err := json.Unmarshal(value, &text); err == nil {
			fields[key] = text
		}
	}

	return fields
}

func (f *fakeTrigger) reply(w http.ResponseWriter, status int, payload any) {
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		f.t.Errorf("write response: %v", err)
	}
}

func login(t *testing.T, client *jupiter.Client, options trigger.SessionOptions) *trigger.Session {
	t.Helper()

	session, err := client.Trigger.Login(context.Background(), triggerWalletKey, options)
	if err != nil {
		t.Fatal(err)
	}

	return session
}

func TestTriggerLoginRequestsRefreshTokens(t *testing.T) {
	fake, client := newFakeTrigger(t)
	session := login(t, client, trigger.SessionOptions{})

	if session.Wallet() != triggerWallet() || session.Tokens().AccessToken != "access-1" {
		t.Fatalf("wallet=%s tokens=%+v", session.Wallet(), session.Tokens())
	}

	body := fake.verifyBodies[0]
	if body["authMode"] != "access_refresh" || body["type"] != trigger.ChallengeMessage {
		t.Fatalf("verify body = %v", body)
	}
}

func TestTriggerRejectsDeprecatedAccessOnlyTokens(t *testing.T) {
	fake, client := newFakeTrigger(t)
	fake.accessOnly = true

	_, err := client.Trigger.Login(context.Background(), triggerWalletKey, trigger.SessionOptions{})
	if !errors.Is(err, trigger.ErrUnexpectedTokenType) {
		t.Fatalf("expected unexpected token type, got %v", err)
	}
}

func TestTriggerTransactionChallengeVerification(t *testing.T) {
	fake, client := newFakeTrigger(t)

	ctx := context.Background()
	challenge, err := client.Trigger.RequestChallenge(ctx, triggerWallet(), trigger.ChallengeTransaction)
	if err != nil {
		t.Fatal(err)
	}

	if challenge.Type != trigger.ChallengeTransaction {
		t.Fatalf("challenge = %+v", challenge)
	}

	verify := trigger.VerifyRequest{
		Wallet:            triggerWallet(),
		Type:              trigger.ChallengeTransaction,
		SignedTransaction: "AQ==",
	}
	if _, err := client.Trigger.Verify(ctx, verify, trigger.SessionOptions{}); err != nil {
		t.Fatal(err)
	}

	body := fake.verifyBodies[0]
	if body["signedTransaction"] != "AQ==" || body["signature"] != "" {
		t.Fatalf("verify body = %v", body)
	}
}

func TestTriggerEnsureVault(t *testing.T) {
	tests := []struct {
		name             string
		registered       bool
		registerConflict bool
	}{
		{name: "registers a missing vault"},
		{name: "returns an existing vault", registered: true},
		{name: "recovers from a concurrent registration", registerConflict: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fake, client := newFakeTrigger(t)
			fake.vaultRegistered = test.registered
			fake.registerConflict = test.registerConflict
			session := login(t, client, trigger.SessionOptions{})

			vault, err := session.EnsureVault(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			if vault.VaultPubkey != vaultAddress || vault.UserPubkey != triggerWallet() {
				t.Fatalf("vault = %+v", vault)
			}
		})
	}
}

func TestTriggerVaultReportsMissingVault(t *testing.T) {
	_, client := newFakeTrigger(t)
	session := login(t, client, trigger.SessionOptions{})

	_, err := session.Vault(context.Background())
	if !errors.Is(err, trigger.ErrNoVault) {
		t.Fatalf("expected ErrNoVault, got %v", err)
	}
}

func TestTriggerRefreshesBeforeExpiryAndReportsRotation(t *testing.T) {
	fake, client := newFakeTrigger(t)
	fake.vaultRegistered = true

	fake.mu.Lock()
	tokens := fake.issue(nearlyExpiredTTL)
	fake.mu.Unlock()

	var rotated []trigger.Tokens
	options := trigger.SessionOptions{
		OnRefresh: func(_ context.Context, tokens trigger.Tokens) error {
			rotated = append(rotated, tokens)

			return nil
		},
	}

	session, err := client.Trigger.Resume(triggerWallet(), tokens, options)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := session.Vault(context.Background()); err != nil {
		t.Fatal(err)
	}

	if len(rotated) != 1 || rotated[0].RefreshToken != "refresh-2" {
		t.Fatalf("rotations = %+v", rotated)
	}

	if session.Tokens().AccessToken != "access-2" {
		t.Fatalf("session kept stale tokens: %+v", session.Tokens())
	}
}

func TestTriggerRefreshesOnceAfterRejection(t *testing.T) {
	fake, client := newFakeTrigger(t)
	fake.vaultRegistered = true
	session := login(t, client, trigger.SessionOptions{})

	fake.mu.Lock()
	fake.revokedAccess = true
	fake.mu.Unlock()

	if _, err := session.Vault(context.Background()); err != nil {
		t.Fatal(err)
	}

	if fake.refreshCalls != 1 {
		t.Fatalf("refresh calls = %d", fake.refreshCalls)
	}
}

func TestTriggerConcurrentRequestsShareOneRefresh(t *testing.T) {
	const concurrentRequests = 8

	fake, client := newFakeTrigger(t)
	fake.vaultRegistered = true

	fake.mu.Lock()
	tokens := fake.issue(nearlyExpiredTTL)
	fake.mu.Unlock()

	session, err := client.Trigger.Resume(triggerWallet(), tokens, trigger.SessionOptions{})
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errs := make(chan error, concurrentRequests)
	for range concurrentRequests {
		wg.Go(func() {
			_, err := session.Vault(context.Background())
			errs <- err
		})
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	if fake.refreshCalls != 1 {
		t.Fatalf("refresh calls = %d; concurrent refreshes would reuse a rotated token", fake.refreshCalls)
	}
}

func TestTriggerRejectedRefreshExpiresSession(t *testing.T) {
	fake, client := newFakeTrigger(t)
	session := login(t, client, trigger.SessionOptions{})

	fake.mu.Lock()
	fake.revokedAccess = true
	fake.refreshToken = "rotated-elsewhere"
	fake.mu.Unlock()

	_, err := session.Vault(context.Background())
	if !errors.Is(err, trigger.ErrSessionExpired) {
		t.Fatalf("expected expired session, got %v", err)
	}
}

func TestTriggerPersistenceFailureKeepsRotatedTokens(t *testing.T) {
	fake, client := newFakeTrigger(t)
	fake.vaultRegistered = true

	fake.mu.Lock()
	tokens := fake.issue(nearlyExpiredTTL)
	fake.mu.Unlock()

	storeFailure := errors.New("disk full")
	options := trigger.SessionOptions{
		OnRefresh: func(context.Context, trigger.Tokens) error {
			return storeFailure
		},
	}

	session, err := client.Trigger.Resume(triggerWallet(), tokens, options)
	if err != nil {
		t.Fatal(err)
	}

	_, err = session.Vault(context.Background())
	if !errors.Is(err, trigger.ErrTokenPersistence) || !errors.Is(err, storeFailure) {
		t.Fatalf("expected persistence failure, got %v", err)
	}

	if session.Tokens().RefreshToken != "refresh-2" {
		t.Fatal("session discarded the rotated tokens")
	}
}

func TestTriggerLogout(t *testing.T) {
	fake, client := newFakeTrigger(t)
	session := login(t, client, trigger.SessionOptions{})

	if err := session.Logout(context.Background()); err != nil {
		t.Fatal(err)
	}

	if len(fake.logoutTokens) != 1 || fake.logoutTokens[0] != "refresh-1" {
		t.Fatalf("logout tokens = %v", fake.logoutTokens)
	}

	if _, err := session.Vault(context.Background()); !errors.Is(err, trigger.ErrSessionClosed) {
		t.Fatalf("closed session still usable: %v", err)
	}

	if err := session.Logout(context.Background()); !errors.Is(err, trigger.ErrSessionClosed) {
		t.Fatalf("second logout: %v", err)
	}
}

func TestTriggerLogoutAll(t *testing.T) {
	_, client := newFakeTrigger(t)
	session := login(t, client, trigger.SessionOptions{})

	if err := session.LogoutAll(context.Background()); err != nil {
		t.Fatal(err)
	}

	if _, err := session.Vault(context.Background()); !errors.Is(err, trigger.ErrSessionClosed) {
		t.Fatalf("closed session still usable: %v", err)
	}
}

func TestTriggerCraftDepositUsesSessionWallet(t *testing.T) {
	_, client := newFakeTrigger(t)
	session := login(t, client, trigger.SessionOptions{})

	deposit, err := session.CraftDeposit(context.Background(), trigger.DepositRequest{
		InputMint:    wrappedSOLMint,
		OutputMint:   usdcMint,
		Amount:       "0001000000000",
		OrderType:    trigger.OrderTypePrice,
		OrderSubType: trigger.OrderSubTypeSingle,
	})
	if err != nil {
		t.Fatal(err)
	}

	if deposit.RequestID != "deposit-request" || deposit.Amount != "1000000000" {
		t.Fatalf("deposit = %+v", deposit)
	}
}

func TestInvalidTriggerRequestsDoNotReachTransport(t *testing.T) {
	client, err := jupiter.NewClient(jupiter.WithHTTPClient(unreachableHTTPClient{t: t}))
	if err != nil {
		t.Fatal(err)
	}

	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	validTokens := trigger.Tokens{
		AuthMode:     "access_refresh",
		AccessToken:  "access",
		RefreshToken: "refresh",
		ExpiresAt:    time.Now().Add(accessTokenTTL),
		TokenType:    "Bearer",
	}

	session, err := client.Trigger.Resume(triggerWallet(), validTokens, trigger.SessionOptions{})
	if err != nil {
		t.Fatal(err)
	}

	validDeposit := trigger.DepositRequest{
		InputMint:    wrappedSOLMint,
		OutputMint:   usdcMint,
		Amount:       "1",
		OrderType:    trigger.OrderTypePrice,
		OrderSubType: trigger.OrderSubTypeOCO,
	}

	deposit := func(configure func(*trigger.DepositRequest)) func(context.Context) error {
		return func(ctx context.Context) error {
			request := validDeposit
			configure(&request)
			_, err := session.CraftDeposit(ctx, request)

			return err
		}
	}

	tests := []struct {
		name string
		call func(context.Context) error
		want error
	}{
		{
			name: "challenge with malformed wallet",
			call: func(ctx context.Context) error {
				_, err := client.Trigger.RequestChallenge(ctx, "wallet", trigger.ChallengeMessage)
				return err
			},
			want: trigger.ErrInvalidAddress,
		},
		{
			name: "unknown challenge type",
			call: func(ctx context.Context) error {
				_, err := client.Trigger.RequestChallenge(ctx, triggerWallet(), "email")
				return err
			},
		},
		{
			name: "message verification without signature",
			call: func(ctx context.Context) error {
				request := trigger.VerifyRequest{Wallet: triggerWallet(), Type: trigger.ChallengeMessage}
				_, err := client.Trigger.Verify(ctx, request, trigger.SessionOptions{})
				return err
			},
		},
		{
			name: "transaction verification with signature",
			call: func(ctx context.Context) error {
				request := trigger.VerifyRequest{
					Wallet:            triggerWallet(),
					Type:              trigger.ChallengeTransaction,
					Signature:         "sig",
					SignedTransaction: "AQ==",
				}
				_, err := client.Trigger.Verify(ctx, request, trigger.SessionOptions{})
				return err
			},
		},
		{
			name: "login with non-ed25519 signer",
			call: func(ctx context.Context) error {
				_, err := client.Trigger.Login(ctx, ecdsaKey, trigger.SessionOptions{})
				return err
			},
			want: trigger.ErrUnsupportedSigner,
		},
		{
			name: "resume deprecated tokens",
			call: func(context.Context) error {
				tokens := validTokens
				tokens.AuthMode = "access_only"
				_, err := client.Trigger.Resume(triggerWallet(), tokens, trigger.SessionOptions{})
				return err
			},
			want: trigger.ErrUnexpectedTokenType,
		},
		{
			name: "resume without expiry",
			call: func(context.Context) error {
				tokens := validTokens
				tokens.ExpiresAt = time.Time{}
				_, err := client.Trigger.Resume(triggerWallet(), tokens, trigger.SessionOptions{})
				return err
			},
			want: trigger.ErrUnexpectedTokenType,
		},
		{
			name: "deposit malformed mint",
			call: deposit(func(r *trigger.DepositRequest) { r.InputMint = "SOL" }),
			want: trigger.ErrInvalidAddress,
		},
		{
			name: "deposit identical mints",
			call: deposit(func(r *trigger.DepositRequest) { r.OutputMint = r.InputMint }),
		},
		{
			name: "deposit zero amount",
			call: deposit(func(r *trigger.DepositRequest) { r.Amount = "0" }),
		},
		{
			name: "price deposit without sub-type",
			call: deposit(func(r *trigger.DepositRequest) { r.OrderSubType = "" }),
		},
		{
			name: "DCA deposit with sub-type",
			call: deposit(func(r *trigger.DepositRequest) { r.OrderType = trigger.OrderTypeDCA }),
		},
		{
			name: "price deposit with JL mint",
			call: deposit(func(r *trigger.DepositRequest) { r.JLMint = usdcMint }),
		},
		{
			name: "DCA deposit with malformed JL mint",
			call: deposit(func(r *trigger.DepositRequest) {
				r.OrderType = trigger.OrderTypeDCA
				r.OrderSubType = ""
				r.JLMint = "jlUSDC"
			}),
			want: trigger.ErrInvalidAddress,
		},
		{
			name: "unknown order type",
			call: deposit(func(r *trigger.DepositRequest) { r.OrderType = "limit" }),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.call(context.Background())
			if err == nil {
				t.Fatal("invalid request accepted")
			}

			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}

func TestTriggerRetryPolicies(t *testing.T) {
	t.Run("challenge retries", func(t *testing.T) {
		server, calls := flakyServer(t, http.StatusServiceUnavailable, `{"type":"message","challenge":"c"}`)

		client, err := jupiter.NewClient(jupiter.WithBaseURL(server.URL))
		if err != nil {
			t.Fatal(err)
		}

		ctx := context.Background()
		_, err = client.Trigger.RequestChallenge(ctx, triggerWallet(), trigger.ChallengeMessage)
		if err != nil {
			t.Fatal(err)
		}

		if calls.Load() != recoveredAttempts {
			t.Fatalf("calls = %d", calls.Load())
		}
	})

	t.Run("verify makes one attempt", func(t *testing.T) {
		server, calls := flakyServer(t, http.StatusServiceUnavailable, `{}`)

		client, err := jupiter.NewClient(jupiter.WithBaseURL(server.URL))
		if err != nil {
			t.Fatal(err)
		}

		request := trigger.VerifyRequest{
			Wallet:    triggerWallet(),
			Type:      trigger.ChallengeMessage,
			Signature: "signature",
		}
		_, err = client.Trigger.Verify(context.Background(), request, trigger.SessionOptions{})
		if err == nil {
			t.Fatal("expected the transient failure to be returned")
		}

		if calls.Load() != singleAttempt {
			t.Fatalf("verify retried a single-use challenge: %d calls", calls.Load())
		}
	})
}
