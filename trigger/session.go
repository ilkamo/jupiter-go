package trigger

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ilkamo/jupiter-go/internal/transport"
	"github.com/ilkamo/jupiter-go/response"
)

// refreshMargin refreshes the access token this long before it expires, so a
// request does not start with a token that lapses in flight.
const refreshMargin = time.Minute

var (
	// ErrSessionExpired means the refresh token was rejected (expired, revoked, or
	// reused). Authenticate again to obtain a new session.
	ErrSessionExpired = errors.New("trigger session expired")

	// ErrOutcomeUnknown marks failures of operations that submit a signed
	// transaction (order creation and cancellation confirmation) after which the
	// transaction may still land. Check order history before creating another order.
	ErrOutcomeUnknown = errors.New("trigger submission outcome unknown")

	// ErrSessionClosed means Logout was called on the session.
	ErrSessionClosed = errors.New("trigger session closed")

	// ErrTokenPersistence means the session rotated its tokens but
	// SessionOptions.OnRefresh failed. The session holds the new tokens and stays
	// usable; the stored copy is stale until OnRefresh succeeds.
	ErrTokenPersistence = errors.New("persist rotated trigger tokens")
)

type SessionOptions struct {
	// OnRefresh receives every rotated token pair, for example to persist it for
	// Resume. Rotation invalidates the previous refresh token, so a stored copy that
	// is not updated cannot resume the session. Calls are serialized.
	OnRefresh func(context.Context, Tokens) error
}

// Session holds one wallet's tokens and authorizes Trigger operations for that
// wallet only. It is safe for concurrent use; refreshes are serialized because
// presenting an already rotated refresh token revokes the session.
type Session struct {
	client  *Client
	wallet  string
	options SessionOptions
	now     func() time.Time

	mu     sync.Mutex
	tokens Tokens
	closed bool
}

func newSession(c *Client, wallet string, tokens Tokens, options SessionOptions) *Session {
	return &Session{
		client:  c,
		wallet:  wallet,
		options: options,
		now:     time.Now,
		tokens:  tokens,
	}
}

func (s *Session) Wallet() string {
	return s.wallet
}

// Tokens returns the current token pair, for example to persist it before exit.
func (s *Session) Tokens() Tokens {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.tokens
}

// accessToken returns a token that is not about to expire, refreshing if needed.
func (s *Session) accessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return "", ErrSessionClosed
	}

	if s.tokens.ExpiresAt.Sub(s.now()) > refreshMargin {
		return s.tokens.AccessToken, nil
	}

	if err := s.refreshLocked(ctx); err != nil {
		return "", err
	}

	return s.tokens.AccessToken, nil
}

// refreshAfterRejection refreshes once after the server rejected rejectedToken,
// unless a concurrent request has already replaced it.
func (s *Session) refreshAfterRejection(ctx context.Context, rejectedToken string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return "", ErrSessionClosed
	}

	if s.tokens.AccessToken != rejectedToken {
		return s.tokens.AccessToken, nil
	}

	if err := s.refreshLocked(ctx); err != nil {
		return "", err
	}

	return s.tokens.AccessToken, nil
}

func (s *Session) refreshLocked(ctx context.Context) error {
	tokens, err := s.client.refresh(ctx, s.tokens.RefreshToken)
	if isUnauthorized(err) {
		return fmt.Errorf("%w: %w", ErrSessionExpired, err)
	}

	if err != nil {
		return fmt.Errorf("refresh trigger session: %w", err)
	}

	s.tokens = tokens

	if s.options.OnRefresh != nil {
		if err := s.options.OnRefresh(ctx, tokens); err != nil {
			return fmt.Errorf("%w: %w", ErrTokenPersistence, err)
		}
	}

	return nil
}

// do sends an authorized request. A 401 means the request was not processed, so
// after one refresh it is sent again regardless of the operation's retry policy.
func (s *Session) do(
	ctx context.Context,
	request transport.Request,
	result any,
) (response.Metadata, error) {
	return s.authorized(ctx, request, result, false)
}

// submit sends a request that submits a signed transaction. Failures after which
// the transaction may still land wrap ErrOutcomeUnknown; failures before anything
// was sent, such as an expired session, do not.
func (s *Session) submit(
	ctx context.Context,
	request transport.Request,
	result any,
) (response.Metadata, error) {
	return s.authorized(ctx, request, result, true)
}

func (s *Session) authorized(
	ctx context.Context,
	request transport.Request,
	result any,
	submits bool,
) (response.Metadata, error) {
	token, err := s.accessToken(ctx)
	if err != nil {
		return response.Metadata{}, err
	}

	metadata, err := s.send(ctx, request, result, token, submits)
	if !isUnauthorized(err) {
		return metadata, err
	}

	token, err = s.refreshAfterRejection(ctx, token)
	if err != nil {
		return metadata, err
	}

	return s.send(ctx, request, result, token, submits)
}

func (s *Session) send(
	ctx context.Context,
	request transport.Request,
	result any,
	token string,
	submits bool,
) (response.Metadata, error) {
	// An ended context guarantees nothing is sent, so the outcome is known.
	if err := ctx.Err(); err != nil {
		return response.Metadata{}, err
	}

	metadata, err := s.client.transport.Do(ctx, withBearer(request, token), result)
	if err != nil && submits && transport.MayHaveBeenProcessed(err) {
		return metadata, fmt.Errorf("%w: %w", ErrOutcomeUnknown, err)
	}

	return metadata, err
}

func withBearer(request transport.Request, token string) transport.Request {
	header := request.Header.Clone()
	if header == nil {
		header = make(http.Header)
	}

	header.Set("Authorization", tokenTypeBearer+" "+token)
	request.Header = header

	return request
}

func isUnauthorized(err error) bool {
	apiError, ok := errors.AsType[*transport.APIError](err)

	return ok && apiError.StatusCode == http.StatusUnauthorized
}

// Logout revokes this session's refresh token and closes the session. Access
// tokens already issued remain valid upstream until they expire.
func (s *Session) Logout(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrSessionClosed
	}

	if err := s.client.logout(ctx, s.tokens.RefreshToken); err != nil {
		return fmt.Errorf("logout: %w", err)
	}

	s.closed = true

	return nil
}

// LogoutAll revokes every refresh token issued for the wallet, on all devices, and
// closes this session.
func (s *Session) LogoutAll(ctx context.Context) error {
	request := transport.Request{
		Method: http.MethodPost,
		Path:   basePath + "/auth/logout-all",
		Body:   struct{}{},
		Retry:  repeatableRetry,
	}

	var result successResponse
	if _, err := s.do(ctx, request, &result); err != nil {
		return fmt.Errorf("logout all: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("logout all was not confirmed")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true

	return nil
}
