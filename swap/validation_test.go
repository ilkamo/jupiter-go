package swap

import (
	"errors"
	"strconv"
	"testing"

	"github.com/mr-tron/base58"

	"github.com/ilkamo/jupiter-go/internal/address"
)

const (
	wrappedSOLMint    = "So11111111111111111111111111111111111111112"
	usdcMint          = "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
	takerAddress      = "GkwFnmMDvn3HGMpJpWBg8tgJxr3NxNvg3AXxvXVPbRGJ"
	payerAddress      = "gasTzr94Pmp4Gf8vknQnqxeYxdgwFjbgdJa4msYRpnB"
	receiverAddress   = "ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL"
	tokenAccount      = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	referralAccount   = "JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4"
	malformedAddress  = "not-a-solana-address"
	defaultTestAmount = "1000000"
	testFeeLamports   = uint64(1000)
)

// shortAddress decodes cleanly as base58 but is one byte shorter than a public key.
var shortAddress = base58.Encode(make([]byte, address.PublicKeyBytes-1))

type validationCase[T any] struct {
	name         string
	configure    func(*T)
	valid        bool
	addressError bool
	parseError   bool
}

func runValidationCases[T any](
	t *testing.T,
	tests []validationCase[T],
	base func() T,
	validate func(T) error,
) {
	t.Helper()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := base()
			test.configure(&request)

			err := validate(request)
			if test.valid {
				if err != nil {
					t.Fatalf("valid request rejected: %v", err)
				}

				return
			}

			if err == nil {
				t.Fatal("invalid request accepted")
			}

			if errors.Is(err, ErrInvalidAddress) != test.addressError {
				t.Fatalf("unexpected address error classification: %v", err)
			}

			_, isParseError := errors.AsType[*strconv.NumError](err)
			if isParseError != test.parseError {
				t.Fatalf("unexpected parse cause: %v", err)
			}
		})
	}
}

func validOrderRequest() OrderRequest {
	return OrderRequest{
		InputMint:  wrappedSOLMint,
		OutputMint: usdcMint,
		Amount:     defaultTestAmount,
	}
}

func encodeOrder(r OrderRequest) error {
	_, err := r.query(Referral{})

	return err
}

func TestOrderRequestValidation(t *testing.T) {
	tests := []validationCase[OrderRequest]{
		{name: "quote only", configure: func(*OrderRequest) {}, valid: true},
		{
			name: "taker receiver and payer",
			configure: func(r *OrderRequest) {
				r.Taker = takerAddress
				r.Receiver = receiverAddress
				r.Payer = payerAddress
			},
			valid: true,
		},
		{
			name: "payer equal to taker has no effect",
			configure: func(r *OrderRequest) {
				r.Taker = takerAddress
				r.Payer = takerAddress
			},
			valid: true,
		},
		{
			name:         "malformed input mint",
			configure:    func(r *OrderRequest) { r.InputMint = malformedAddress },
			addressError: true,
		},
		{
			name:         "missing output mint",
			configure:    func(r *OrderRequest) { r.OutputMint = "" },
			addressError: true,
		},
		{
			name:         "short output mint",
			configure:    func(r *OrderRequest) { r.OutputMint = shortAddress },
			addressError: true,
		},
		{
			name:      "identical mints",
			configure: func(r *OrderRequest) { r.OutputMint = r.InputMint },
		},
		{
			name:         "malformed taker",
			configure:    func(r *OrderRequest) { r.Taker = malformedAddress },
			addressError: true,
		},
		{
			name: "malformed receiver",
			configure: func(r *OrderRequest) {
				r.Taker = takerAddress
				r.Receiver = shortAddress
			},
			addressError: true,
		},
		{
			name: "malformed payer",
			configure: func(r *OrderRequest) {
				r.Taker = takerAddress
				r.Payer = malformedAddress
			},
			addressError: true,
		},
		{
			name:      "receiver without taker",
			configure: func(r *OrderRequest) { r.Receiver = receiverAddress },
		},
		{
			name:      "payer without taker",
			configure: func(r *OrderRequest) { r.Payer = payerAddress },
		},
		{
			name: "receiver equal to taker",
			configure: func(r *OrderRequest) {
				r.Taker = takerAddress
				r.Receiver = takerAddress
			},
		},
		{
			name:      "zero amount",
			configure: func(r *OrderRequest) { r.Amount = "0" },
		},
		{
			name:      "unsupported swap mode",
			configure: func(r *OrderRequest) { r.SwapMode = "ExactOut" },
		},
		{
			name:      "maximum slippage",
			configure: func(r *OrderRequest) { r.SlippageBps = new(uint16(BasisPointsPerWhole)) },
			valid:     true,
		},
		{
			name:      "slippage above maximum",
			configure: func(r *OrderRequest) { r.SlippageBps = new(uint16(BasisPointsPerWhole + 1)) },
		},
		{
			name: "maximum transaction version",
			configure: func(r *OrderRequest) {
				r.MaxSupportedTransactionVersion = new(uint8(MaxTransactionVersion))
			},
			valid: true,
		},
		{
			name: "transaction version above maximum",
			configure: func(r *OrderRequest) {
				r.MaxSupportedTransactionVersion = new(uint8(MaxTransactionVersion + 1))
			},
		},
		{
			name: "broadcast fee type with priority fee",
			configure: func(r *OrderRequest) {
				r.BroadcastFeeType = BroadcastFeeExact
				r.PriorityFeeLamports = new(testFeeLamports)
			},
			valid: true,
		},
		{
			name: "broadcast fee type with Jito tip",
			configure: func(r *OrderRequest) {
				r.BroadcastFeeType = BroadcastFeeMaxCap
				r.JitoTipLamports = new(testFeeLamports)
			},
			valid: true,
		},
		{
			name:      "broadcast fee type without fee",
			configure: func(r *OrderRequest) { r.BroadcastFeeType = BroadcastFeeMaxCap },
		},
		{
			name: "unknown broadcast fee type",
			configure: func(r *OrderRequest) {
				r.BroadcastFeeType = "minimum"
				r.PriorityFeeLamports = new(testFeeLamports)
			},
		},
		{
			name: "documented routers excluded",
			configure: func(r *OrderRequest) {
				r.ExcludeRouters = []string{RouterJupiterZ, RouterDflow, RouterOKX}
			},
			valid: true,
		},
		{
			name:      "router added upstream",
			configure: func(r *OrderRequest) { r.ExcludeRouters = []string{"newrouter"} },
			valid:     true,
		},
		{
			name:      "empty router",
			configure: func(r *OrderRequest) { r.ExcludeRouters = []string{""} },
		},
		{
			name:      "router with separator",
			configure: func(r *OrderRequest) { r.ExcludeRouters = []string{"okx,dflow"} },
		},
		{
			name:      "router with whitespace",
			configure: func(r *OrderRequest) { r.ExcludeRouters = []string{"okx "} },
		},
		{
			name:      "excluded DEX labels",
			configure: func(r *OrderRequest) { r.ExcludeDexes = []string{"Raydium", "Orca V2"} },
			valid:     true,
		},
		{
			name:      "empty excluded DEX label",
			configure: func(r *OrderRequest) { r.ExcludeDexes = []string{""} },
		},
		{
			name:      "excluded DEX label with separator",
			configure: func(r *OrderRequest) { r.ExcludeDexes = []string{"Raydium,Orca"} },
		},
		{
			name:      "excluded DEX label with whitespace",
			configure: func(r *OrderRequest) { r.ExcludeDexes = []string{" Raydium"} },
		},
	}

	runValidationCases(t, tests, validOrderRequest, encodeOrder)
}

func TestOrderQueryEncoding(t *testing.T) {
	request := validOrderRequest()
	request.Amount = "000123"
	request.Taker = takerAddress
	request.ExcludeRouters = []string{RouterDflow, RouterOKX}
	request.ExcludeDexes = []string{"Raydium", "Orca V2"}

	referral := Referral{Account: referralAccount, BasisPoints: MinReferralFeeBps}
	query, err := request.query(referral)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"amount":          "123",
		"taker":           takerAddress,
		"excludeRouters":  "dflow,okx",
		"excludeDexes":    "Raydium,Orca V2",
		"referralAccount": referralAccount,
		"referralFee":     strconv.Itoa(MinReferralFeeBps),
	}
	for key, value := range want {
		if query.Get(key) != value {
			t.Errorf("%s = %q, want %q", key, query.Get(key), value)
		}
	}

	if query.Has("receiver") || query.Has("payer") || query.Has("slippageBps") {
		t.Fatalf("unset optional fields encoded: %s", query.Encode())
	}
}

func validBuildRequest() BuildRequest {
	return BuildRequest{
		InputMint:  wrappedSOLMint,
		OutputMint: usdcMint,
		Amount:     defaultTestAmount,
		Taker:      takerAddress,
	}
}

func encodeBuild(r BuildRequest) error {
	_, err := r.query()

	return err
}

func TestBuildRequestValidation(t *testing.T) {
	tests := []validationCase[BuildRequest]{
		{name: "minimal", configure: func(*BuildRequest) {}, valid: true},
		{
			name:      "missing taker",
			configure: func(r *BuildRequest) { r.Taker = "" },
		},
		{
			name:         "malformed taker",
			configure:    func(r *BuildRequest) { r.Taker = shortAddress },
			addressError: true,
		},
		{
			name:         "malformed input mint",
			configure:    func(r *BuildRequest) { r.InputMint = malformedAddress },
			addressError: true,
		},
		{
			name:      "identical mints",
			configure: func(r *BuildRequest) { r.InputMint = r.OutputMint },
		},
		{
			name:         "malformed payer",
			configure:    func(r *BuildRequest) { r.Payer = malformedAddress },
			addressError: true,
		},
		{
			name:         "malformed destination token account",
			configure:    func(r *BuildRequest) { r.DestinationTokenAccount = shortAddress },
			addressError: true,
		},
		{
			name:         "malformed native destination account",
			configure:    func(r *BuildRequest) { r.NativeDestinationAccount = malformedAddress },
			addressError: true,
		},
		{
			name: "both destination accounts",
			configure: func(r *BuildRequest) {
				r.DestinationTokenAccount = tokenAccount
				r.NativeDestinationAccount = receiverAddress
			},
		},
		{
			name: "maximum platform fee with account",
			configure: func(r *BuildRequest) {
				r.PlatformFeeBps = BasisPointsPerWhole
				r.FeeAccount = tokenAccount
			},
			valid: true,
		},
		{
			name: "platform fee above maximum",
			configure: func(r *BuildRequest) {
				r.PlatformFeeBps = BasisPointsPerWhole + 1
				r.FeeAccount = tokenAccount
			},
		},
		{
			name:      "platform fee without account",
			configure: func(r *BuildRequest) { r.PlatformFeeBps = 1 },
		},
		{
			name:      "fee account without platform fee",
			configure: func(r *BuildRequest) { r.FeeAccount = tokenAccount },
		},
		{
			name: "malformed fee account",
			configure: func(r *BuildRequest) {
				r.PlatformFeeBps = 1
				r.FeeAccount = malformedAddress
			},
			addressError: true,
		},
		{
			name: "dexes and excluded dexes",
			configure: func(r *BuildRequest) {
				r.Dexes = []string{"Raydium"}
				r.ExcludeDexes = []string{"Orca V2"}
			},
		},
		{
			name:      "dex label with separator",
			configure: func(r *BuildRequest) { r.Dexes = []string{"Raydium,Orca"} },
		},
		{
			name:      "invalid mode",
			configure: func(r *BuildRequest) { r.Mode = "slow" },
		},
		{
			name:      "real-time slippage",
			configure: func(r *BuildRequest) { r.SlippageBps = SlippageRTSE },
			valid:     true,
		},
		{
			name:      "zero slippage",
			configure: func(r *BuildRequest) { r.SlippageBps = "0" },
			valid:     true,
		},
		{
			name:      "maximum slippage",
			configure: func(r *BuildRequest) { r.SlippageBps = strconv.Itoa(BasisPointsPerWhole) },
			valid:     true,
		},
		{
			name:      "slippage above maximum",
			configure: func(r *BuildRequest) { r.SlippageBps = strconv.Itoa(BasisPointsPerWhole + 1) },
		},
		{
			name:       "malformed slippage",
			configure:  func(r *BuildRequest) { r.SlippageBps = "invalid" },
			parseError: true,
		},
		{
			name:      "minimum accounts",
			configure: func(r *BuildRequest) { r.MaxAccounts = new(uint8(MinBuildAccounts)) },
			valid:     true,
		},
		{
			name:      "maximum accounts",
			configure: func(r *BuildRequest) { r.MaxAccounts = new(uint8(MaxBuildAccounts)) },
			valid:     true,
		},
		{
			name:      "accounts below minimum",
			configure: func(r *BuildRequest) { r.MaxAccounts = new(uint8(MinBuildAccounts - 1)) },
		},
		{
			name:      "accounts above maximum",
			configure: func(r *BuildRequest) { r.MaxAccounts = new(uint8(MaxBuildAccounts + 1)) },
		},
		{
			name:      "minimum bytes",
			configure: func(r *BuildRequest) { r.MaxBytes = new(uint16(MinBuildBytes)) },
			valid:     true,
		},
		{
			name: "maximum bytes for version 0",
			configure: func(r *BuildRequest) {
				r.MaxBytes = new(uint16(MaxBuildBytes))
				r.TransactionVersion = new(uint8(TransactionVersionV0))
			},
			valid: true,
		},
		{
			name:      "bytes below minimum",
			configure: func(r *BuildRequest) { r.MaxBytes = new(uint16(MinBuildBytes - 1)) },
		},
		{
			name:      "bytes above maximum",
			configure: func(r *BuildRequest) { r.MaxBytes = new(uint16(MaxBuildBytes + 1)) },
		},
		{
			name: "bytes with version 1",
			configure: func(r *BuildRequest) {
				r.MaxBytes = new(uint16(MaxBuildBytes))
				r.TransactionVersion = new(uint8(TransactionVersionV1))
			},
		},
		{
			name: "transaction version above maximum",
			configure: func(r *BuildRequest) {
				r.TransactionVersion = new(uint8(MaxTransactionVersion + 1))
			},
		},
		{
			name: "minimum blockhash expiry",
			configure: func(r *BuildRequest) {
				r.BlockhashSlotsToExpiry = new(uint16(MinBlockhashExpirySlots))
			},
			valid: true,
		},
		{
			name: "maximum blockhash expiry",
			configure: func(r *BuildRequest) {
				r.BlockhashSlotsToExpiry = new(uint16(MaxBlockhashExpirySlots))
			},
			valid: true,
		},
		{
			name: "blockhash expiry below minimum",
			configure: func(r *BuildRequest) {
				r.BlockhashSlotsToExpiry = new(uint16(MinBlockhashExpirySlots - 1))
			},
		},
		{
			name: "blockhash expiry above maximum",
			configure: func(r *BuildRequest) {
				r.BlockhashSlotsToExpiry = new(uint16(MaxBlockhashExpirySlots + 1))
			},
		},
		{
			name:      "maximum tip",
			configure: func(r *BuildRequest) { r.TipAmount = "18446744073709551615" },
			valid:     true,
		},
		{
			name:       "tip beyond uint64",
			configure:  func(r *BuildRequest) { r.TipAmount = "18446744073709551616" },
			parseError: true,
		},
		{
			name:       "malformed tip",
			configure:  func(r *BuildRequest) { r.TipAmount = "invalid" },
			parseError: true,
		},
		{
			name:      "named compute unit price percentile",
			configure: func(r *BuildRequest) { r.ComputeUnitPricePercentile = ComputeUnitPriceVeryHigh },
			valid:     true,
		},
		{
			name: "maximum compute unit price percentile",
			configure: func(r *BuildRequest) {
				r.ComputeUnitPricePercentile = strconv.Itoa(BasisPointsPerWhole)
			},
			valid: true,
		},
		{
			name:      "negative compute unit price percentile",
			configure: func(r *BuildRequest) { r.ComputeUnitPricePercentile = "-1" },
		},
		{
			name: "compute unit price percentile above maximum",
			configure: func(r *BuildRequest) {
				r.ComputeUnitPricePercentile = strconv.Itoa(BasisPointsPerWhole + 1)
			},
		},
	}

	runValidationCases(t, tests, validBuildRequest, encodeBuild)
}

func TestBuildQueryEncoding(t *testing.T) {
	request := validBuildRequest()
	request.MaxBytes = new(uint16(MaxBuildBytes))
	request.Dexes = []string{"Raydium", "Orca V2"}

	query, err := request.query()
	if err != nil {
		t.Fatal(err)
	}

	if query.Get("maxBytes") != strconv.Itoa(MaxBuildBytes) ||
		query.Get("dexes") != "Raydium,Orca V2" ||
		query.Get("taker") != takerAddress {
		t.Fatalf("unexpected build query: %s", query.Encode())
	}

	if query.Has("feeAccount") || query.Has("transactionVersion") {
		t.Fatalf("unset optional fields encoded: %s", query.Encode())
	}
}

func validExecuteRequest() ExecuteRequest {
	return ExecuteRequest{
		RequestID:         "order",
		SignedTransaction: "AQ==",
	}
}

func validateExecute(r ExecuteRequest) error {
	return r.validate()
}

func TestExecuteRequestValidation(t *testing.T) {
	tests := []validationCase[ExecuteRequest]{
		{name: "minimal", configure: func(*ExecuteRequest) {}, valid: true},
		{
			name:      "maximum block height",
			configure: func(r *ExecuteRequest) { r.LastValidBlockHeight = "18446744073709551615" },
			valid:     true,
		},
		{
			name:      "missing request ID",
			configure: func(r *ExecuteRequest) { r.RequestID = "" },
		},
		{
			name:      "missing transaction",
			configure: func(r *ExecuteRequest) { r.SignedTransaction = "" },
		},
		{
			name:      "malformed transaction",
			configure: func(r *ExecuteRequest) { r.SignedTransaction = "not base64" },
		},
		{
			name:       "block height beyond uint64",
			configure:  func(r *ExecuteRequest) { r.LastValidBlockHeight = "18446744073709551616" },
			parseError: true,
		},
		{
			name:       "negative block height",
			configure:  func(r *ExecuteRequest) { r.LastValidBlockHeight = "-1" },
			parseError: true,
		},
	}

	runValidationCases(t, tests, validExecuteRequest, validateExecute)
}

func TestReferralValidation(t *testing.T) {
	tests := []validationCase[Referral]{
		{name: "minimum fee", configure: func(*Referral) {}, valid: true},
		{
			name:      "maximum fee",
			configure: func(r *Referral) { r.BasisPoints = MaxReferralFeeBps },
			valid:     true,
		},
		{
			name:      "fee below minimum",
			configure: func(r *Referral) { r.BasisPoints = MinReferralFeeBps - 1 },
		},
		{
			name:      "fee above maximum",
			configure: func(r *Referral) { r.BasisPoints = MaxReferralFeeBps + 1 },
		},
		{
			name:         "missing account",
			configure:    func(r *Referral) { r.Account = "" },
			addressError: true,
		},
		{
			name:         "malformed account",
			configure:    func(r *Referral) { r.Account = malformedAddress },
			addressError: true,
		},
	}

	base := func() Referral {
		return Referral{Account: referralAccount, BasisPoints: MinReferralFeeBps}
	}

	runValidationCases(t, tests, base, Referral.Validate)
}
