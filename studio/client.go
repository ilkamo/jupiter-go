// Package studio implements Jupiter Studio: launching tokens on Dynamic Bonding
// Curve (DBC) pools and claiming creator trading fees.
//
// The API is beta upstream. Launching is a sequence: CreatePool builds the
// transaction and returns presigned upload URLs, UploadImage and UploadMetadata
// store the token's off-chain metadata, and Submit sends the signed transaction,
// which Jupiter submits on-chain. Uploads go to presigned URLs that never receive
// the API key.
package studio

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"

	"github.com/ilkamo/jupiter-go/internal/address"
	"github.com/ilkamo/jupiter-go/internal/transport"
	"github.com/ilkamo/jupiter-go/response"
)

// Documented values. Fields pass other values through so upstream additions
// remain usable; Jupiter validates them.
const (
	QuoteDecimals6 = 6
	QuoteDecimals9 = 9

	FeeBps100 = 100
	FeeBps200 = 200

	BaseFeeModeLinear      = "linear"
	BaseFeeModeExponential = "exponential"

	ImageJPEG = "image/jpeg"
	ImagePNG  = "image/png"
	ImageGIF  = "image/gif"
	ImageWebP = "image/webp"

	basePath        = "/studio/v1"
	jsonContentType = "application/json"
	decimalBase     = 10
)

var (
	// ErrInvalidAddress is the same value as swap.ErrInvalidAddress.
	ErrInvalidAddress = address.ErrInvalid

	// ErrOutcomeUnknown marks Submit failures after which the launch transaction may
	// still land. Check PoolAddresses for the mint before submitting again.
	ErrOutcomeUnknown = errors.New("studio submission outcome unknown")
)

type Client struct {
	transport *transport.Client
}

// NewClient is called by jupiter.NewClient, which applications use instead; it
// takes the SDK's internal transport.
func NewClient(t *transport.Client) *Client {
	return &Client{transport: t}
}

// LockedVesting configures vested supply. Durations are seconds; numbers keep the
// exact JSON text.
type LockedVesting struct {
	TotalLockedVestingAmount       json.Number `json:"totalLockedVestingAmount"`
	CliffUnlockAmount              json.Number `json:"cliffUnlockAmount"`
	NumberOfVestingPeriod          json.Number `json:"numberOfVestingPeriod"`
	TotalVestingDuration           json.Number `json:"totalVestingDuration"`
	CliffDurationFromMigrationTime json.Number `json:"cliffDurationFromMigrationTime"`
}

// NoLockedVesting returns the zero vesting configuration: no supply is locked.
func NoLockedVesting() LockedVesting {
	const zero = json.Number("0")

	return LockedVesting{
		TotalLockedVestingAmount:       zero,
		CliffUnlockAmount:              zero,
		NumberOfVestingPeriod:          zero,
		TotalVestingDuration:           zero,
		CliffDurationFromMigrationTime: zero,
	}
}

// CurveByMarketCap sets the bonding curve from market caps in quote token units:
// 16000 means a 16,000 USDC (or SOL, or JUP) initial market cap.
type CurveByMarketCap struct {
	QuoteMint          string        `json:"quoteMint"`
	InitialMarketCap   json.Number   `json:"initialMarketCap"`
	MigrationMarketCap json.Number   `json:"migrationMarketCap"`
	TokenQuoteDecimal  int           `json:"tokenQuoteDecimal"`
	LockedVesting      LockedVesting `json:"lockedVestingParam"`
}

// Fee sets the trading fee schedule; FeeBps is the ending fee (FeeBps100 or
// FeeBps200). TotalDuration and BaseFeeMode are optional.
type Fee struct {
	FeeBps        int    `json:"feeBps"`
	TotalDuration int    `json:"totalDuration,omitempty"`
	BaseFeeMode   string `json:"baseFeeMode,omitempty"`
}

// PoolRequest launches a token. IsLPLocked defaults to true upstream when nil.
type PoolRequest struct {
	Curve                 CurveByMarketCap `json:"buildCurveByMarketCapParam"`
	AntiSniping           bool             `json:"antiSniping"`
	Fee                   Fee              `json:"fee"`
	IsLPLocked            *bool            `json:"isLpLocked,omitempty"`
	TokenName             string           `json:"tokenName"`
	TokenSymbol           string           `json:"tokenSymbol"`
	TokenImageContentType string           `json:"tokenImageContentType"`
	Creator               string           `json:"creator"`
}

func (r PoolRequest) validate() error {
	if err := validate("quote mint", r.Curve.QuoteMint); err != nil {
		return err
	}

	if err := validate("creator", r.Creator); err != nil {
		return err
	}

	initial, err := parseNumber("initial market cap", r.Curve.InitialMarketCap, true)
	if err != nil {
		return err
	}

	migration, err := parseNumber("migration market cap", r.Curve.MigrationMarketCap, true)
	if err != nil {
		return err
	}

	if migration.Cmp(initial) <= 0 {
		return fmt.Errorf("migration market cap must exceed the initial market cap")
	}

	vesting := []struct {
		field string
		value json.Number
	}{
		{field: "total locked vesting amount", value: r.Curve.LockedVesting.TotalLockedVestingAmount},
		{field: "cliff unlock amount", value: r.Curve.LockedVesting.CliffUnlockAmount},
		{field: "number of vesting periods", value: r.Curve.LockedVesting.NumberOfVestingPeriod},
		{field: "total vesting duration", value: r.Curve.LockedVesting.TotalVestingDuration},
		{field: "cliff duration", value: r.Curve.LockedVesting.CliffDurationFromMigrationTime},
	}
	for _, item := range vesting {
		if _, err := parseNumber(item.field, item.value, false); err != nil {
			return err
		}
	}

	if r.Fee.FeeBps <= 0 || r.Curve.TokenQuoteDecimal <= 0 {
		return fmt.Errorf("fee basis points and quote decimals are required")
	}

	if r.TokenName == "" || r.TokenSymbol == "" || r.TokenImageContentType == "" {
		return fmt.Errorf("token name, symbol, and image content type are required")
	}

	return nil
}

// parseNumber requires a JSON number that is positive, or non-negative when
// positive is false. Vesting fields are required by the API, so empty is rejected.
func parseNumber(field string, value json.Number, positive bool) (*big.Rat, error) {
	if !json.Valid([]byte(value)) {
		return nil, fmt.Errorf("%s: %q is not a JSON number", field, value)
	}

	number, ok := new(big.Rat).SetString(string(value))
	if !ok {
		return nil, fmt.Errorf("%s: %q is not a number", field, value)
	}

	if number.Sign() < 0 || (positive && number.Sign() == 0) {
		return nil, fmt.Errorf("%s must not be negative or zero", field)
	}

	return number, nil
}

func validate(field string, value string) error {
	if err := address.Validate(value); err != nil {
		return fmt.Errorf("%s: %w", field, err)
	}

	return nil
}

// PoolTransaction is the unsigned launch transaction and the upload targets for
// the token's off-chain metadata.
type PoolTransaction struct {
	Transaction          string `json:"transaction"`
	Mint                 string `json:"mint"`
	ImagePresignedURL    string `json:"imagePresignedUrl"`
	MetadataPresignedURL string `json:"metadataPresignedUrl"`
	ImageURL             string `json:"imageUrl"`

	Metadata response.Metadata `json:"-"`
}

// CreatePool builds the unsigned launch transaction. Building submits nothing, so
// transient failures are retried.
func (c *Client) CreatePool(ctx context.Context, r PoolRequest) (*PoolTransaction, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}

	request := transport.Request{
		Method: http.MethodPost,
		Path:   basePath + "/dbc-pool/create-tx",
		Body:   r,
		Retry:  transport.RepeatableRetry,
	}

	var result PoolTransaction
	metadata, err := c.transport.Do(ctx, request, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}

// UploadImage stores the token image at PoolTransaction.ImagePresignedURL. The
// content type must match PoolRequest.TokenImageContentType. Re-uploading
// overwrites the same object, so transient failures are retried.
func (c *Client) UploadImage(
	ctx context.Context,
	presignedURL string,
	contentType string,
	image []byte,
) (response.Metadata, error) {
	if contentType == "" || len(image) == 0 {
		return response.Metadata{}, fmt.Errorf("image content type and data required")
	}

	return c.upload(ctx, presignedURL, contentType, image)
}

// TokenMetadata is the off-chain metadata referenced by the token's on-chain URI.
// Image should be PoolTransaction.ImageURL.
type TokenMetadata struct {
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Website     string `json:"website"`
	Twitter     string `json:"twitter"`
	Telegram    string `json:"telegram"`
}

// UploadMetadata stores the token metadata at PoolTransaction.MetadataPresignedURL.
func (c *Client) UploadMetadata(
	ctx context.Context,
	presignedURL string,
	metadata TokenMetadata,
) (response.Metadata, error) {
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return response.Metadata{}, fmt.Errorf("encode token metadata: %w", err)
	}

	return c.upload(ctx, presignedURL, jsonContentType, encoded)
}

func (c *Client) upload(
	ctx context.Context,
	presignedURL string,
	contentType string,
	data []byte,
) (response.Metadata, error) {
	request := transport.Request{
		Method:      http.MethodPut,
		ExternalURL: presignedURL,
		Payload:     &transport.Payload{ContentType: contentType, Data: data},
		Retry:       transport.RepeatableRetry,
	}

	return c.transport.Do(ctx, request, nil)
}

// HeaderImage is an optional banner for the Studio token page; it is not on-chain
// metadata.
type HeaderImage struct {
	Filename    string
	ContentType string
	Data        []byte
}

// SubmitRequest submits the signed launch transaction for Owner. Content is
// optional text for the Studio token page, not on-chain metadata.
type SubmitRequest struct {
	SignedTransaction string
	Owner             string
	Content           string
	HeaderImage       *HeaderImage
}

func (r SubmitRequest) form() (*transport.Payload, error) {
	if r.SignedTransaction == "" {
		return nil, fmt.Errorf("signed transaction required")
	}

	if err := validate("owner", r.Owner); err != nil {
		return nil, err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	fields := []struct {
		name  string
		value string
	}{
		{name: "transaction", value: r.SignedTransaction},
		{name: "owner", value: r.Owner},
		{name: "content", value: r.Content},
	}
	for _, field := range fields {
		if err := writer.WriteField(field.name, field.value); err != nil {
			return nil, fmt.Errorf("encode %s field: %w", field.name, err)
		}
	}

	if r.HeaderImage != nil {
		if err := writeFile(writer, "headerImage", *r.HeaderImage); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("finish form: %w", err)
	}

	return &transport.Payload{ContentType: writer.FormDataContentType(), Data: body.Bytes()}, nil
}

// writeFile adds a file part with an explicit content type, which
// multipart.Writer.CreateFormFile does not allow.
func writeFile(writer *multipart.Writer, field string, file HeaderImage) error {
	if file.Filename == "" || file.ContentType == "" {
		return fmt.Errorf("header image filename and content type required")
	}

	disposition := mime.FormatMediaType("form-data", map[string]string{
		"name":     field,
		"filename": file.Filename,
	})
	if disposition == "" {
		return fmt.Errorf("header image filename cannot be encoded")
	}

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", disposition)
	header.Set("Content-Type", file.ContentType)

	part, err := writer.CreatePart(header)
	if err != nil {
		return fmt.Errorf("create header image part: %w", err)
	}

	if _, err := part.Write(file.Data); err != nil {
		return fmt.Errorf("write header image: %w", err)
	}

	return nil
}

type SubmitResult struct {
	Mint      string
	ConfigKey string
	Metadata  response.Metadata
}

type submitResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Mint      string `json:"mint"`
		ConfigKey string `json:"configKey"`
	} `json:"data"`
}

// Submit sends the signed launch transaction, which Jupiter submits on-chain. It
// makes one attempt because no replay guarantee is documented; failures after
// which the launch may have landed wrap ErrOutcomeUnknown.
func (c *Client) Submit(ctx context.Context, r SubmitRequest) (*SubmitResult, error) {
	payload, err := r.form()
	if err != nil {
		return nil, err
	}

	// An ended context guarantees nothing is sent, so the outcome is known.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	request := transport.Request{
		Method:  http.MethodPost,
		Path:    basePath + "/dbc-pool/submit",
		Payload: payload,
		Retry:   transport.NoRetry,
	}

	var decoded submitResponse
	metadata, err := c.transport.Do(ctx, request, &decoded)
	result := &SubmitResult{Mint: decoded.Data.Mint, ConfigKey: decoded.Data.ConfigKey, Metadata: metadata}
	if err != nil {
		if transport.MayHaveBeenProcessed(err) {
			return result, fmt.Errorf("%w: %w", ErrOutcomeUnknown, err)
		}

		return result, err
	}

	// An explicit unsuccessful response is the server's own statement, not an
	// ambiguous outcome.
	if !decoded.Success {
		return result, fmt.Errorf("submission was not successful")
	}

	return result, nil
}

// PoolAddresses lists a launched token's pools; fields are empty when the API
// reports null, for example before graduation to a DAMM v2 pool.
type PoolAddresses struct {
	DAMMv2PoolAddress string `json:"dammv2PoolAddress"`
	DBCPoolAddress    string `json:"dbcPoolAddress"`
	ConfigKey         string `json:"configKey"`

	Metadata response.Metadata `json:"-"`
}

// PoolAddresses returns the pools of a token launched through Studio.
func (c *Client) PoolAddresses(ctx context.Context, mint string) (*PoolAddresses, error) {
	if err := validate("mint", mint); err != nil {
		return nil, err
	}

	request := transport.Request{
		Method: http.MethodGet,
		Path:   basePath + "/dbc-pool/addresses/" + url.PathEscape(mint),
		Retry:  transport.ReadOnlyRetry,
	}

	var decoded struct {
		Success bool          `json:"success"`
		Data    PoolAddresses `json:"data"`
	}
	metadata, err := c.transport.Do(ctx, request, &decoded)
	result := decoded.Data
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	if !decoded.Success {
		return &result, fmt.Errorf("pool address lookup was not successful")
	}

	return &result, nil
}

// Fees are creator quote-token amounts in raw units.
type Fees struct {
	Unclaimed string `json:"unclaimed"`
	Total     string `json:"total"`

	Metadata response.Metadata `json:"-"`
}

type poolBody struct {
	PoolAddress string `json:"poolAddress"`
}

// Fee returns the creator's unclaimed and total trading fees for a DBC pool. It
// is a POST that only reads, so it is retried like other reads.
func (c *Client) Fee(ctx context.Context, poolAddress string) (*Fees, error) {
	if err := validate("pool address", poolAddress); err != nil {
		return nil, err
	}

	request := transport.Request{
		Method: http.MethodPost,
		Path:   basePath + "/dbc/fee",
		Body:   poolBody{PoolAddress: poolAddress},
		Retry:  transport.ReadOnlyRetry,
	}

	var fees Fees
	metadata, err := c.transport.Do(ctx, request, &fees)
	fees.Metadata = metadata
	if err != nil {
		return &fees, err
	}

	return &fees, nil
}

// ClaimFeeRequest claims at most MaxQuoteAmount raw quote units of creator fees.
type ClaimFeeRequest struct {
	OwnerWallet    string
	PoolAddress    string
	MaxQuoteAmount string
}

type claimFeeBody struct {
	OwnerWallet    string      `json:"ownerWallet"`
	PoolAddress    string      `json:"poolAddress"`
	MaxQuoteAmount json.Number `json:"maxQuoteAmount"`
}

type ClaimFeeTransaction struct {
	Transaction string `json:"transaction"`

	Metadata response.Metadata `json:"-"`
}

// ClaimFeeTransaction builds an unsigned fee claim for the owner to sign and send
// through an RPC node.
func (c *Client) ClaimFeeTransaction(ctx context.Context, r ClaimFeeRequest) (*ClaimFeeTransaction, error) {
	if err := validate("owner wallet", r.OwnerWallet); err != nil {
		return nil, err
	}

	if err := validate("pool address", r.PoolAddress); err != nil {
		return nil, err
	}

	maxQuote, ok := new(big.Int).SetString(r.MaxQuoteAmount, decimalBase)
	if !ok || maxQuote.Sign() < 0 {
		return nil, fmt.Errorf("max quote amount must be a non-negative base-10 integer")
	}

	body := claimFeeBody{
		OwnerWallet:    r.OwnerWallet,
		PoolAddress:    r.PoolAddress,
		MaxQuoteAmount: json.Number(maxQuote.String()),
	}

	request := transport.Request{
		Method: http.MethodPost,
		Path:   basePath + "/dbc/fee/create-tx",
		Body:   body,
		Retry:  transport.RepeatableRetry,
	}

	var result ClaimFeeTransaction
	metadata, err := c.transport.Do(ctx, request, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}
