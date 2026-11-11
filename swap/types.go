package swap

import (
	"encoding/json"

	"github.com/ilkamo/jupiter-go/response"
)

type RouteStep struct {
	USDValue json.Number `json:"usdValue"`
	SwapInfo SwapInfo    `json:"swapInfo"`
	Percent  json.Number `json:"percent"`
	Bps      json.Number `json:"bps"`
}

type SwapInfo struct {
	AMMKey     string `json:"ammKey"`
	Label      string `json:"label"`
	InputMint  string `json:"inputMint"`
	OutputMint string `json:"outputMint"`
	InAmount   string `json:"inAmount"`
	OutAmount  string `json:"outAmount"`
	FeeAmount  string `json:"feeAmount"`
	FeeMint    string `json:"feeMint"`
}

type PlatformFee struct {
	Amount  string      `json:"amount"`
	FeeBps  json.Number `json:"feeBps"`
	FeeMint string      `json:"feeMint"`
}

type OrderResponse struct {
	Metadata                  response.Metadata `json:"-"`
	SwapUSDValue              json.Number       `json:"swapUsdValue"`
	PriceImpactPct            string            `json:"priceImpactPct"`
	SwapType                  string            `json:"swapType"`
	TotalTime                 json.Number       `json:"totalTime"`
	Taker                     *string           `json:"taker"`
	QuoteID                   string            `json:"quoteId"`
	Maker                     string            `json:"maker"`
	ExpireAt                  string            `json:"expireAt"`
	InputMint                 string            `json:"inputMint"`
	OutputMint                string            `json:"outputMint"`
	InAmount                  string            `json:"inAmount"`
	OutAmount                 string            `json:"outAmount"`
	InUSDValue                json.Number       `json:"inUsdValue"`
	OutUSDValue               json.Number       `json:"outUsdValue"`
	PriceImpact               json.Number       `json:"priceImpact"`
	OtherAmountThreshold      string            `json:"otherAmountThreshold"`
	SwapMode                  string            `json:"swapMode"`
	SlippageBps               json.Number       `json:"slippageBps"`
	RoutePlan                 []RouteStep       `json:"routePlan"`
	Mode                      string            `json:"mode"`
	Router                    string            `json:"router"`
	Transaction               *string           `json:"transaction"`
	TransactionVersion        *uint8            `json:"transactionVersion"`
	RequestID                 string            `json:"requestId"`
	LastValidBlockHeight      string            `json:"lastValidBlockHeight"`
	Gasless                   *bool             `json:"gasless"`
	FeeBps                    json.Number       `json:"feeBps"`
	FeeMint                   string            `json:"feeMint"`
	PlatformFee               *PlatformFee      `json:"platformFee"`
	ReferralAccount           string            `json:"referralAccount"`
	SignatureFeeLamports      json.Number       `json:"signatureFeeLamports"`
	SignatureFeePayer         *string           `json:"signatureFeePayer"`
	PrioritizationFeeLamports json.Number       `json:"prioritizationFeeLamports"`
	PrioritizationFeePayer    *string           `json:"prioritizationFeePayer"`
	RentFeeLamports           json.Number       `json:"rentFeeLamports"`
	RentFeePayer              *string           `json:"rentFeePayer"`
	ErrorCode                 *int              `json:"errorCode"`
	ErrorMessage              string            `json:"errorMessage"`
}

type ExecuteResponse struct {
	Metadata           response.Metadata `json:"-"`
	Status             string            `json:"status"`
	Signature          string            `json:"signature"`
	Slot               string            `json:"slot"`
	Code               *int              `json:"code"`
	Error              string            `json:"error"`
	TotalInputAmount   string            `json:"totalInputAmount"`
	TotalOutputAmount  string            `json:"totalOutputAmount"`
	InputAmountResult  string            `json:"inputAmountResult"`
	OutputAmountResult string            `json:"outputAmountResult"`
	SwapEvents         []SwapEvent       `json:"swapEvents"`
}

type SwapEvent struct {
	InputMint    string `json:"inputMint"`
	InputAmount  string `json:"inputAmount"`
	OutputMint   string `json:"outputMint"`
	OutputAmount string `json:"outputAmount"`
}

type ProgramLabelsResponse struct {
	Metadata response.Metadata
	Labels   map[string]string
}
