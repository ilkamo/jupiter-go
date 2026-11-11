// Package send implements Jupiter's Send API: token transfers to anyone through
// an invite code that the recipient claims in Jupiter Mobile.
//
// The API is beta upstream. It only builds transactions: invite codes, key
// derivation, and signing stay on the client and are never sent to Jupiter. A send
// transaction needs signatures from the sender and the invite key; a clawback
// needs only the sender. Unclaimed invites return to the sender at expiry.
package send

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ilkamo/jupiter-go/internal/address"
	"github.com/ilkamo/jupiter-go/internal/amount"
	"github.com/ilkamo/jupiter-go/internal/transport"
	"github.com/ilkamo/jupiter-go/response"
)

const (
	basePath    = "/send/v1"
	decimalBase = 10
)

// ErrInvalidAddress is the same value as swap.ErrInvalidAddress.
var ErrInvalidAddress = address.ErrInvalid

type Client struct {
	transport *transport.Client
}

// NewClient is called by jupiter.NewClient, which applications use instead; it
// takes the SDK's internal transport.
func NewClient(t *transport.Client) *Client {
	return &Client{transport: t}
}

// SendRequest transfers Amount raw units of Mint (wrapped SOL when empty) from
// Sender to the invite whose public key is InviteSigner (InviteCode.Signer).
// Sender also pays network fees and token account rent.
type SendRequest struct {
	InviteSigner string `json:"inviteSigner"`
	Sender       string `json:"sender"`
	Amount       string `json:"amount"`
	Mint         string `json:"mint,omitempty"`
}

// SendTransaction is an unsigned send transaction. Sign it with the sender and
// InviteCode.PrivateKey. Expiry is reported as sent by the API; the reference
// describes a Unix timestamp while its example is an RFC 3339 time.
type SendTransaction struct {
	Transaction      string `json:"tx"`
	Expiry           string `json:"expiry"`
	TotalFeeLamports string `json:"totalFeeLamports"`

	Metadata response.Metadata `json:"-"`
}

// CraftSend builds an unsigned send transaction. Building submits nothing, so
// transient failures are retried.
func (c *Client) CraftSend(ctx context.Context, r SendRequest) (*SendTransaction, error) {
	if err := validate("invite signer", r.InviteSigner); err != nil {
		return nil, err
	}

	if err := validate("sender", r.Sender); err != nil {
		return nil, err
	}

	if r.Mint != "" {
		if err := validate("mint", r.Mint); err != nil {
			return nil, err
		}
	}

	rawAmount, err := amount.ParseRaw(r.Amount)
	if err != nil {
		return nil, err
	}

	r.Amount = rawAmount

	var result SendTransaction
	metadata, err := c.transport.Do(ctx, buildRequest("/craft-send", r), &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

// ClawbackRequest returns an unclaimed invite's funds to Sender. InvitePDA comes
// from InviteCode.PDA or InvitePDA.
type ClawbackRequest struct {
	InvitePDA string `json:"invitePDA"`
	Sender    string `json:"sender"`
}

type ClawbackTransaction struct {
	Transaction string `json:"tx"`

	Metadata response.Metadata `json:"-"`
}

// CraftClawback builds an unsigned clawback transaction for the sender to sign.
func (c *Client) CraftClawback(ctx context.Context, r ClawbackRequest) (*ClawbackTransaction, error) {
	if err := validate("invite PDA", r.InvitePDA); err != nil {
		return nil, err
	}

	if err := validate("sender", r.Sender); err != nil {
		return nil, err
	}

	var result ClawbackTransaction
	metadata, err := c.transport.Do(ctx, buildRequest("/craft-clawback", r), &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

func buildRequest(path string, body any) transport.Request {
	return transport.Request{
		Method: http.MethodPost,
		Path:   basePath + path,
		Body:   body,
		Retry:  transport.RepeatableRetry,
	}
}

func validate(field string, value string) error {
	if err := address.Validate(value); err != nil {
		return fmt.Errorf("%s: %w", field, err)
	}

	return nil
}

// Confirmed accepts the integer the schema declares and the boolean its examples
// return. Known is false when the API reports null.
type Confirmed struct {
	Known bool
	Value bool
}

// UnmarshalJSON implements json.Unmarshaler for boolean, integer, and null.
func (c *Confirmed) UnmarshalJSON(data []byte) error {
	var flag *bool
	if err := json.Unmarshal(data, &flag); err == nil {
		*c = Confirmed{Known: flag != nil, Value: flag != nil && *flag}

		return nil
	}

	var number json.Number
	if err := json.Unmarshal(data, &number); err != nil {
		return fmt.Errorf("confirmed must be a boolean or integer: %w", err)
	}

	value, err := number.Int64()
	if err != nil {
		return fmt.Errorf("confirmed must be a boolean or integer: %w", err)
	}

	*c = Confirmed{Known: true, Value: value != 0}

	return nil
}

// Invite is one Send invite. Nullable fields are empty when unset.
type Invite struct {
	InviteSigner string    `json:"invite_signer"`
	InvitePDA    string    `json:"invite_pda"`
	Sender       string    `json:"sender"`
	Expiry       string    `json:"expiry"`
	Amount       string    `json:"amount"`
	CreationTx   string    `json:"creation_tx"`
	DeletionTx   string    `json:"deletion_tx"`
	Receiver     string    `json:"receiver"`
	Action       string    `json:"action"`
	CreationTime string    `json:"creation_time"`
	DeletionTime string    `json:"deletion_time"`
	Mint         string    `json:"mint"`
	Confirmed    Confirmed `json:"confirmed"`
}

type InvitePage struct {
	Invites     []Invite `json:"invites"`
	HasMoreData bool     `json:"hasMoreData"`

	Metadata response.Metadata `json:"-"`
}

// PendingInvites lists unclaimed invites sent by address. Page starts at 1; zero
// uses the API default.
func (c *Client) PendingInvites(ctx context.Context, sender string, page uint32) (*InvitePage, error) {
	return c.invites(ctx, "/pending-invites", sender, page)
}

// InviteHistory lists invites where address is the sender or recipient.
func (c *Client) InviteHistory(ctx context.Context, address string, page uint32) (*InvitePage, error) {
	return c.invites(ctx, "/invite-history", address, page)
}

func (c *Client) invites(ctx context.Context, path string, owner string, page uint32) (*InvitePage, error) {
	if err := validate("address", owner); err != nil {
		return nil, err
	}

	query := url.Values{"address": {owner}}
	if page > 0 {
		query.Set("page", strconv.FormatUint(uint64(page), decimalBase))
	}

	request := transport.Request{
		Method: http.MethodGet,
		Path:   basePath + path,
		Query:  query,
		Retry:  transport.ReadOnlyRetry,
	}

	var result InvitePage
	metadata, err := c.transport.Do(ctx, request, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}
