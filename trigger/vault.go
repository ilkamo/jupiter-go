package trigger

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ilkamo/jupiter-go/internal/address"
	"github.com/ilkamo/jupiter-go/internal/amount"
	"github.com/ilkamo/jupiter-go/internal/transport"
	"github.com/ilkamo/jupiter-go/response"
)

const (
	OrderTypePrice = "price"
	OrderTypeDCA   = "dca"

	OrderSubTypeSingle = "single"
	OrderSubTypeOCO    = "oco"
	OrderSubTypeOTOCO  = "otoco"
)

// ErrNoVault means the wallet has no vault yet; see RegisterVault or EnsureVault.
var ErrNoVault = errors.New("trigger vault not registered")

// Vault is the custodial vault that holds deposits for the wallet's orders.
type Vault struct {
	UserPubkey   string `json:"userPubkey"`
	VaultPubkey  string `json:"vaultPubkey"`
	PrivyVaultID string `json:"privyVaultId"`
	PrivyUserID  string `json:"privyUserId"`

	Metadata response.Metadata `json:"-"`
}

// Vault returns the wallet's vault, or an error wrapping ErrNoVault.
func (s *Session) Vault(ctx context.Context) (*Vault, error) {
	request := transport.Request{
		Method: http.MethodGet,
		Path:   basePath + "/vault",
		Retry:  transport.ReadOnlyRetry,
	}

	var vault Vault
	metadata, err := s.do(ctx, request, &vault)
	vault.Metadata = metadata
	if statusIs(err, http.StatusNotFound) {
		return nil, fmt.Errorf("%w: %w", ErrNoVault, err)
	}

	if err != nil {
		return nil, err
	}

	return &vault, nil
}

// RegisterVault creates the wallet's vault. It makes one attempt because the
// operation creates state; an already registered vault returns HTTP 409. Prefer
// EnsureVault, which handles both cases.
func (s *Session) RegisterVault(ctx context.Context) (*Vault, error) {
	request := transport.Request{
		Method: http.MethodGet,
		Path:   basePath + "/vault/register",
		Retry:  transport.NoRetry,
	}

	var vault Vault
	metadata, err := s.do(ctx, request, &vault)
	vault.Metadata = metadata
	if err != nil {
		return nil, err
	}

	return &vault, nil
}

// EnsureVault returns the wallet's vault, registering it first if none exists.
func (s *Session) EnsureVault(ctx context.Context) (*Vault, error) {
	vault, err := s.Vault(ctx)
	if !errors.Is(err, ErrNoVault) {
		return vault, err
	}

	vault, err = s.RegisterVault(ctx)
	if statusIs(err, http.StatusConflict) {
		// Registered concurrently or by an attempt whose response was lost.
		return s.Vault(ctx)
	}

	return vault, err
}

func statusIs(err error, statusCode int) bool {
	var apiError *transport.APIError

	return errors.As(err, &apiError) && apiError.StatusCode == statusCode
}

// DepositRequest crafts a deposit of Amount raw input units from the session's
// wallet into its vault. OrderSubType is required for OrderTypePrice and must be
// empty for OrderTypeDCA. JLMint (Earn While You Wait) applies only to DCA.
type DepositRequest struct {
	InputMint    string
	OutputMint   string
	Amount       string
	OrderType    string
	OrderSubType string
	JLMint       string
}

type depositBody struct {
	InputMint    string `json:"inputMint"`
	OutputMint   string `json:"outputMint"`
	UserAddress  string `json:"userAddress"`
	Amount       string `json:"amount"`
	OrderType    string `json:"orderType"`
	OrderSubType string `json:"orderSubType,omitempty"`
	JLMint       string `json:"jlMint,omitempty"`
}

// Deposit is an unsigned deposit transaction. Sign Transaction with the wallet and
// pass it with RequestID, which is single-use, to the order creation call.
type Deposit struct {
	Transaction        string `json:"transaction"`
	RequestID          string `json:"requestId"`
	ReceiverAddress    string `json:"receiverAddress"`
	Mint               string `json:"mint"`
	Amount             string `json:"amount"`
	TokenDecimals      int    `json:"tokenDecimals"`
	InputTokenAccount  string `json:"inputTokenAccount"`
	OutputTokenAccount string `json:"outputTokenAccount"`
	JLTokenAccount     string `json:"jlTokenAccount"`

	Metadata response.Metadata `json:"-"`
}

func (r DepositRequest) body(wallet string) (depositBody, error) {
	if err := address.Validate(r.InputMint); err != nil {
		return depositBody{}, fmt.Errorf("input mint: %w", err)
	}

	if err := address.Validate(r.OutputMint); err != nil {
		return depositBody{}, fmt.Errorf("output mint: %w", err)
	}

	if r.InputMint == r.OutputMint {
		return depositBody{}, fmt.Errorf("input and output mints must differ")
	}

	rawAmount, err := amount.ParseRaw(r.Amount)
	if err != nil {
		return depositBody{}, err
	}

	switch r.OrderType {
	case OrderTypePrice:
		switch r.OrderSubType {
		case OrderSubTypeSingle, OrderSubTypeOCO, OrderSubTypeOTOCO:
		default:
			return depositBody{}, fmt.Errorf("price deposits require order sub-type single, oco, or otoco")
		}

		if r.JLMint != "" {
			return depositBody{}, fmt.Errorf("JL mint applies only to DCA deposits")
		}
	case OrderTypeDCA:
		if r.OrderSubType != "" {
			return depositBody{}, fmt.Errorf("DCA deposits do not take an order sub-type")
		}

		if r.JLMint != "" {
			if err := address.Validate(r.JLMint); err != nil {
				return depositBody{}, fmt.Errorf("JL mint: %w", err)
			}
		}
	default:
		return depositBody{}, fmt.Errorf("order type must be %q or %q", OrderTypePrice, OrderTypeDCA)
	}

	return depositBody{
		InputMint:    r.InputMint,
		OutputMint:   r.OutputMint,
		UserAddress:  wallet,
		Amount:       rawAmount,
		OrderType:    r.OrderType,
		OrderSubType: r.OrderSubType,
		JLMint:       r.JLMint,
	}, nil
}

// CraftDeposit builds an unsigned deposit transaction from the session's wallet.
// Signing alone moves no funds; the order creation call submits it.
func (s *Session) CraftDeposit(ctx context.Context, r DepositRequest) (*Deposit, error) {
	body, err := r.body(s.wallet)
	if err != nil {
		return nil, err
	}

	request := transport.Request{
		Method: http.MethodPost,
		Path:   basePath + "/deposit/craft",
		Body:   body,
		Retry:  repeatableRetry,
	}

	var deposit Deposit
	metadata, err := s.do(ctx, request, &deposit)
	deposit.Metadata = metadata
	if err != nil {
		return nil, err
	}

	return &deposit, nil
}
