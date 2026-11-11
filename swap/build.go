package swap

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type BuildRequest struct {
	InputMint, OutputMint, Amount, Taker                     string
	SlippageBps                                              string
	Mode                                                     string
	Dexes, ExcludeDexes                                      []string
	PlatformFeeBps                                           uint16
	FeeAccount                                               string
	MaxAccounts                                              *uint8
	Payer, DestinationTokenAccount, NativeDestinationAccount string
	WrapAndUnwrapSOL                                         *bool
	BlockhashSlotsToExpiry                                   *uint16
	TipAmount                                                string
	ComputeUnitPricePercentile                               string
	ForJitoBundle                                            *bool
	TransactionVersion                                       *uint8
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
	TransactionVersion            uint8               `json:"transactionVersion"`
	ComputeUnitPrice              string              `json:"computeUnitPrice"`
	BlockhashWithMetadata         struct {
		Blockhash            []uint8 `json:"blockhash"`
		LastValidBlockHeight uint64  `json:"lastValidBlockHeight"`
		FetchedAt            struct {
			Seconds     uint64 `json:"secs_since_epoch"`
			Nanoseconds uint32 `json:"nanos_since_epoch"`
		} `json:"fetchedAt"`
	} `json:"blockhashWithMetadata"`
}

// Build returns instructions; it does not sign, submit, or apply managed-order referral settings.
func (c *Client) Build(
	ctx context.Context,
	r BuildRequest,
) (*BuildResponse, error) {
	q, err := swapQuery(r.InputMint, r.OutputMint, r.Amount)
	if err != nil {
		return nil, err
	}

	if r.Taker == "" {
		return nil, fmt.Errorf("taker required for build")
	}

	if len(r.Dexes) > 0 && len(r.ExcludeDexes) > 0 {
		return nil, fmt.Errorf("dexes and excludeDexes are mutually exclusive")
	}

	if r.PlatformFeeBps > BasisPointsPerWhole {
		return nil, fmt.Errorf("platform fee exceeds %d bps", BasisPointsPerWhole)
	}

	feeEnabled := r.PlatformFeeBps > 0
	feeAccountProvided := r.FeeAccount != ""
	if feeEnabled != feeAccountProvided {
		return nil, fmt.Errorf("platform fee and fee account must be configured together")
	}

	if r.DestinationTokenAccount != "" && r.NativeDestinationAccount != "" {
		return nil, fmt.Errorf("destination accounts are mutually exclusive")
	}

	if r.Mode != "" && r.Mode != BuildModeFast {
		return nil, fmt.Errorf("invalid build mode")
	}

	if r.SlippageBps != "" && r.SlippageBps != SlippageRTSE {
		slippageBps, err := strconv.Atoi(r.SlippageBps)
		if err != nil {
			return nil, fmt.Errorf("parse slippage basis points: %w", err)
		}

		if slippageBps < 0 || slippageBps > BasisPointsPerWhole {
			return nil, fmt.Errorf("slippage must be between zero and %d bps", BasisPointsPerWhole)
		}
	}

	percentile := r.ComputeUnitPricePercentile
	switch percentile {
	case "", ComputeUnitPriceMedium, ComputeUnitPriceHigh, ComputeUnitPriceVeryHigh:
	default:
		percentileBps, err := strconv.Atoi(percentile)
		if err != nil {
			return nil, fmt.Errorf("parse compute unit price percentile: %w", err)
		}

		if percentileBps < 0 || percentileBps > BasisPointsPerWhole {
			return nil, fmt.Errorf(
				"compute unit price percentile must be between zero and %d bps",
				BasisPointsPerWhole,
			)
		}
	}

	set(q, "taker", r.Taker)
	set(q, "slippageBps", r.SlippageBps)
	set(q, "mode", r.Mode)
	set(q, "dexes", strings.Join(r.Dexes, ","))
	set(q, "excludeDexes", strings.Join(r.ExcludeDexes, ","))

	if r.PlatformFeeBps > 0 {
		q.Set("platformFeeBps", strconv.Itoa(int(r.PlatformFeeBps)))
		q.Set("feeAccount", r.FeeAccount)
	}

	if r.MaxAccounts != nil {
		if *r.MaxAccounts < MinBuildAccounts || *r.MaxAccounts > MaxBuildAccounts {
			return nil, fmt.Errorf("max accounts must be between %d and %d", MinBuildAccounts, MaxBuildAccounts)
		}

		q.Set("maxAccounts", strconv.Itoa(int(*r.MaxAccounts)))
	}

	if r.BlockhashSlotsToExpiry != nil {
		if *r.BlockhashSlotsToExpiry < MinBlockhashExpirySlots ||
			*r.BlockhashSlotsToExpiry > MaxBlockhashExpirySlots {
			return nil, fmt.Errorf(
				"blockhash expiry must be between %d and %d slots",
				MinBlockhashExpirySlots,
				MaxBlockhashExpirySlots,
			)
		}

		q.Set("blockhashSlotsToExpiry", strconv.Itoa(int(*r.BlockhashSlotsToExpiry)))
	}

	if r.TransactionVersion != nil {
		if *r.TransactionVersion > MaxTransactionVersion {
			return nil, fmt.Errorf("unsupported transaction version")
		}

		q.Set("transactionVersion", strconv.Itoa(int(*r.TransactionVersion)))
	}

	if r.WrapAndUnwrapSOL != nil {
		q.Set("wrapAndUnwrapSol", strconv.FormatBool(*r.WrapAndUnwrapSOL))
	}

	if r.ForJitoBundle != nil {
		q.Set("forJitoBundle", strconv.FormatBool(*r.ForJitoBundle))
	}

	if r.TipAmount != "" {
		if _, err := strconv.ParseUint(r.TipAmount, decimalBase, lamportAmountBits); err != nil {
			return nil, fmt.Errorf("parse tip amount in lamports: %w", err)
		}
	}

	set(q, "payer", r.Payer)
	set(q, "destinationTokenAccount", r.DestinationTokenAccount)
	set(q, "nativeDestinationAccount", r.NativeDestinationAccount)
	set(q, "tipAmount", r.TipAmount)
	set(q, "computeUnitPricePercentile", r.ComputeUnitPricePercentile)

	var result BuildResponse
	if err := c.transport.Do(ctx, "GET", "/swap/v2/build", q, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
