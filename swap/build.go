package swap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ilkamo/jupiter-go/response"
)

type BuildRequest struct {
	InputMint                  string
	OutputMint                 string
	Amount                     string
	Taker                      string
	SlippageBps                string
	Mode                       string
	Dexes                      []string
	ExcludeDexes               []string
	PlatformFeeBps             uint16
	FeeAccount                 string
	MaxAccounts                *uint8
	MaxBytes                   *uint16
	Payer                      string
	DestinationTokenAccount    string
	NativeDestinationAccount   string
	WrapAndUnwrapSOL           *bool
	BlockhashSlotsToExpiry     *uint16
	TipAmount                  string
	ComputeUnitPricePercentile string
	ForJitoBundle              *bool
	TransactionVersion         *uint8
}

type AccountMeta struct {
	Pubkey     string `json:"pubkey"`
	IsSigner   bool   `json:"isSigner"`
	IsWritable bool   `json:"isWritable"`
}

type Instruction struct {
	ProgramID string        `json:"programId"`
	Accounts  []AccountMeta `json:"accounts"`
	Data      string        `json:"data"`
}

type BuildResponse struct {
	Metadata                      response.Metadata   `json:"-"`
	SwapMode                      string              `json:"swapMode"`
	SlippageBps                   json.Number         `json:"slippageBps"`
	InputMint                     string              `json:"inputMint"`
	OutputMint                    string              `json:"outputMint"`
	InAmount                      string              `json:"inAmount"`
	OutAmount                     string              `json:"outAmount"`
	OtherAmountThreshold          string              `json:"otherAmountThreshold"`
	PriceImpactPct                string              `json:"priceImpactPct"`
	RoutePlan                     []RouteStep         `json:"routePlan"`
	ComputeBudgetInstructions     []Instruction       `json:"computeBudgetInstructions"`
	SetupInstructions             []Instruction       `json:"setupInstructions"`
	SwapInstruction               Instruction         `json:"swapInstruction"`
	CleanupInstruction            *Instruction        `json:"cleanupInstruction"`
	OtherInstructions             []Instruction       `json:"otherInstructions"`
	TipInstruction                *Instruction        `json:"tipInstruction"`
	AddressesByLookupTableAddress map[string][]string `json:"addressesByLookupTableAddress"`
	TransactionVersion            *uint8              `json:"transactionVersion"`
	ComputeUnitPrice              string              `json:"computeUnitPrice"`
	BlockhashWithMetadata         *BlockhashMetadata  `json:"blockhashWithMetadata"`
}

type BlockhashMetadata struct {
	Blockhash            []uint8             `json:"blockhash"`
	LastValidBlockHeight *uint64             `json:"lastValidBlockHeight"`
	FetchedAt            *BlockhashTimestamp `json:"fetchedAt"`
}

type BlockhashTimestamp struct {
	Seconds     *uint64 `json:"secs_since_epoch"`
	Nanoseconds *uint32 `json:"nanos_since_epoch"`
}

func (r BuildRequest) validate() error {
	if err := validateMintPair(r.InputMint, r.OutputMint); err != nil {
		return err
	}

	if r.Taker == "" {
		return fmt.Errorf("taker required for build")
	}

	if err := r.validateAccounts(); err != nil {
		return err
	}

	if err := r.validateRouting(); err != nil {
		return err
	}

	if err := r.validateFees(); err != nil {
		return err
	}

	return r.validateTransaction()
}

func (r BuildRequest) validateAccounts() error {
	if err := validateAddress("taker", r.Taker); err != nil {
		return err
	}

	if err := validateOptionalAddress("payer", r.Payer); err != nil {
		return err
	}

	if err := validateOptionalAddress("destination token account", r.DestinationTokenAccount); err != nil {
		return err
	}

	if err := validateOptionalAddress("native destination account", r.NativeDestinationAccount); err != nil {
		return err
	}

	if r.DestinationTokenAccount != "" && r.NativeDestinationAccount != "" {
		return fmt.Errorf("destination accounts are mutually exclusive")
	}

	return nil
}

func (r BuildRequest) validateRouting() error {
	if len(r.Dexes) > 0 && len(r.ExcludeDexes) > 0 {
		return fmt.Errorf("dexes and excludeDexes are mutually exclusive")
	}

	if err := validateLabels("DEXes", r.Dexes); err != nil {
		return err
	}

	if err := validateLabels("excluded DEXes", r.ExcludeDexes); err != nil {
		return err
	}

	if r.Mode != "" && r.Mode != BuildModeFast {
		return fmt.Errorf("invalid build mode %q", r.Mode)
	}

	if r.SlippageBps != "" && r.SlippageBps != SlippageRTSE {
		slippageBps, err := strconv.Atoi(r.SlippageBps)
		if err != nil {
			return fmt.Errorf("parse slippage basis points: %w", err)
		}

		if slippageBps < 0 || slippageBps > BasisPointsPerWhole {
			return fmt.Errorf("slippage must be between zero and %d bps", BasisPointsPerWhole)
		}
	}

	if r.MaxAccounts != nil {
		if *r.MaxAccounts < MinBuildAccounts || *r.MaxAccounts > MaxBuildAccounts {
			return fmt.Errorf(
				"max accounts must be between %d and %d",
				MinBuildAccounts,
				MaxBuildAccounts,
			)
		}
	}

	return nil
}

func (r BuildRequest) validateFees() error {
	if r.PlatformFeeBps > BasisPointsPerWhole {
		return fmt.Errorf("platform fee exceeds %d bps", BasisPointsPerWhole)
	}

	feeEnabled := r.PlatformFeeBps > 0
	feeAccountProvided := r.FeeAccount != ""
	if feeEnabled != feeAccountProvided {
		return fmt.Errorf("platform fee and fee account must be configured together")
	}

	if err := validateOptionalAddress("fee account", r.FeeAccount); err != nil {
		return err
	}

	if r.TipAmount != "" {
		if _, err := strconv.ParseUint(r.TipAmount, decimalBase, lamportAmountBits); err != nil {
			return fmt.Errorf("parse tip amount in lamports: %w", err)
		}
	}

	switch r.ComputeUnitPricePercentile {
	case "", ComputeUnitPriceMedium, ComputeUnitPriceHigh, ComputeUnitPriceVeryHigh:
		return nil
	}

	percentileBps, err := strconv.Atoi(r.ComputeUnitPricePercentile)
	if err != nil {
		return fmt.Errorf("parse compute unit price percentile: %w", err)
	}

	if percentileBps < 0 || percentileBps > BasisPointsPerWhole {
		return fmt.Errorf(
			"compute unit price percentile must be between zero and %d bps",
			BasisPointsPerWhole,
		)
	}

	return nil
}

func (r BuildRequest) validateTransaction() error {
	if r.TransactionVersion != nil && *r.TransactionVersion > MaxTransactionVersion {
		return fmt.Errorf("unsupported transaction version")
	}

	if r.MaxBytes != nil {
		if *r.MaxBytes < MinBuildBytes || *r.MaxBytes > MaxBuildBytes {
			return fmt.Errorf("max bytes must be between %d and %d", MinBuildBytes, MaxBuildBytes)
		}

		buildsV1 := r.TransactionVersion != nil && *r.TransactionVersion == TransactionVersionV1
		if buildsV1 {
			return fmt.Errorf("max bytes applies only to version 0 transactions")
		}
	}

	if r.BlockhashSlotsToExpiry != nil {
		slots := *r.BlockhashSlotsToExpiry
		if slots < MinBlockhashExpirySlots || slots > MaxBlockhashExpirySlots {
			return fmt.Errorf(
				"blockhash expiry must be between %d and %d slots",
				MinBlockhashExpirySlots,
				MaxBlockhashExpirySlots,
			)
		}
	}

	return nil
}

func (r BuildRequest) query() (url.Values, error) {
	if err := r.validate(); err != nil {
		return nil, err
	}

	amount, err := parseRawAmount(r.Amount)
	if err != nil {
		return nil, err
	}

	q := url.Values{
		"inputMint":  {r.InputMint},
		"outputMint": {r.OutputMint},
		"amount":     {amount},
		"taker":      {r.Taker},
	}

	set(q, "slippageBps", r.SlippageBps)
	set(q, "mode", r.Mode)
	set(q, "dexes", strings.Join(r.Dexes, labelSeparator))
	set(q, "excludeDexes", strings.Join(r.ExcludeDexes, labelSeparator))

	if r.PlatformFeeBps > 0 {
		q.Set("platformFeeBps", strconv.Itoa(int(r.PlatformFeeBps)))
		q.Set("feeAccount", r.FeeAccount)
	}

	if r.MaxAccounts != nil {
		q.Set("maxAccounts", strconv.Itoa(int(*r.MaxAccounts)))
	}

	if r.MaxBytes != nil {
		q.Set("maxBytes", strconv.Itoa(int(*r.MaxBytes)))
	}

	if r.BlockhashSlotsToExpiry != nil {
		q.Set("blockhashSlotsToExpiry", strconv.Itoa(int(*r.BlockhashSlotsToExpiry)))
	}

	if r.TransactionVersion != nil {
		q.Set("transactionVersion", strconv.Itoa(int(*r.TransactionVersion)))
	}

	if r.WrapAndUnwrapSOL != nil {
		q.Set("wrapAndUnwrapSol", strconv.FormatBool(*r.WrapAndUnwrapSOL))
	}

	if r.ForJitoBundle != nil {
		q.Set("forJitoBundle", strconv.FormatBool(*r.ForJitoBundle))
	}

	set(q, "payer", r.Payer)
	set(q, "destinationTokenAccount", r.DestinationTokenAccount)
	set(q, "nativeDestinationAccount", r.NativeDestinationAccount)
	set(q, "tipAmount", r.TipAmount)
	set(q, "computeUnitPricePercentile", r.ComputeUnitPricePercentile)

	return q, nil
}

// Build returns instructions; it does not sign, submit, or apply managed-order referral settings.
func (c *Client) Build(
	ctx context.Context,
	r BuildRequest,
) (*BuildResponse, error) {
	q, err := r.query()
	if err != nil {
		return nil, err
	}

	var result BuildResponse
	metadata, err := c.transport.Do(ctx, http.MethodGet, "/swap/v2/build", q, nil, &result)
	result.Metadata = metadata
	if err != nil {
		return &result, err
	}

	return &result, nil
}
