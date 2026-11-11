package tokens

import (
	"encoding/json"

	"github.com/ilkamo/jupiter-go/response"
)

// Numeric fields use json.Number to keep upstream precision. An empty json.Number
// means the field was null or absent; explicit zero is preserved as "0". Nullable
// booleans use pointers. Timestamps are kept as the RFC 3339 strings sent upstream.

type ListResponse struct {
	Tokens   []Token
	Metadata response.Metadata
}

// Token is the mint information returned by every token discovery operation.
type Token struct {
	ID                string      `json:"id"`
	Name              string      `json:"name"`
	Symbol            string      `json:"symbol"`
	Icon              string      `json:"icon"`
	Decimals          int         `json:"decimals"`
	TokenProgram      string      `json:"tokenProgram"`
	CreatedAt         string      `json:"createdAt"`
	Twitter           string      `json:"twitter"`
	Telegram          string      `json:"telegram"`
	Website           string      `json:"website"`
	Discord           string      `json:"discord"`
	Instagram         string      `json:"instagram"`
	TikTok            string      `json:"tiktok"`
	OtherURL          string      `json:"otherUrl"`
	Dev               string      `json:"dev"`
	MintAuthority     string      `json:"mintAuthority"`
	FreezeAuthority   string      `json:"freezeAuthority"`
	CircSupply        json.Number `json:"circSupply"`
	TotalSupply       json.Number `json:"totalSupply"`
	Launchpad         string      `json:"launchpad"`
	PartnerConfig     string      `json:"partnerConfig"`
	GraduatedPool     string      `json:"graduatedPool"`
	GraduatedAt       string      `json:"graduatedAt"`
	HolderCount       json.Number `json:"holderCount"`
	FDV               json.Number `json:"fdv"`
	MarketCap         json.Number `json:"mcap"`
	USDPrice          json.Number `json:"usdPrice"`
	PriceBlockID      json.Number `json:"priceBlockId"`
	Liquidity         json.Number `json:"liquidity"`
	APY               *APY        `json:"apy"`
	Stats5m           *SwapStats  `json:"stats5m"`
	Stats1h           *SwapStats  `json:"stats1h"`
	Stats6h           *SwapStats  `json:"stats6h"`
	Stats24h          *SwapStats  `json:"stats24h"`
	FirstPool         *Pool       `json:"firstPool"`
	Audit             *Audit      `json:"audit"`
	OrganicScore      json.Number `json:"organicScore"`
	OrganicScoreLabel string      `json:"organicScoreLabel"`
	IsVerified        *bool       `json:"isVerified"`
	Tags              []string    `json:"tags"`
	UpdatedAt         string      `json:"updatedAt"`
}

// APY is present only for assets listed on Jupiter Lend's Earn.
type APY struct {
	JupEarn json.Number `json:"jupEarn"`
}

type SwapStats struct {
	PriceChange       json.Number `json:"priceChange"`
	HolderChange      json.Number `json:"holderChange"`
	LiquidityChange   json.Number `json:"liquidityChange"`
	VolumeChange      json.Number `json:"volumeChange"`
	BuyVolume         json.Number `json:"buyVolume"`
	SellVolume        json.Number `json:"sellVolume"`
	BuyOrganicVolume  json.Number `json:"buyOrganicVolume"`
	SellOrganicVolume json.Number `json:"sellOrganicVolume"`
	NumBuys           json.Number `json:"numBuys"`
	NumSells          json.Number `json:"numSells"`
	NumTraders        json.Number `json:"numTraders"`
	NumOrganicBuyers  json.Number `json:"numOrganicBuyers"`
	NumNetBuyers      json.Number `json:"numNetBuyers"`
}

type Pool struct {
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`
}

// Audit fields are conditional and vary per token.
type Audit struct {
	// IsSus is set only for tokens flagged as suspicious; check for nil rather than
	// its value. Not every risky token is flagged.
	IsSus                   *bool `json:"isSus"`
	MintAuthorityDisabled   *bool `json:"mintAuthorityDisabled"`
	FreezeAuthorityDisabled *bool `json:"freezeAuthorityDisabled"`

	// Percentages use a 0–100 scale: 25.07 means 25.07%.
	TopHoldersPercentage json.Number `json:"topHoldersPercentage"`
	DevBalancePercentage json.Number `json:"devBalancePercentage"`

	DevMints json.Number `json:"devMints"`
}
