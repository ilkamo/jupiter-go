package jupiter

const (
	// DefaultAPIURL is the default Jupiter API provided by the official Jupiter team.
	// For more info visit: https://station.jup.ag/docs/
	DefaultAPIURL = "https://api.jup.ag/swap/v1"

	// LegacyAPIURL is the legacy Jupiter API provided by the official Jupiter team.
	// Old hostnames will be fully deprecated on 1 June 2025.
	LegacyAPIURL = "https://quote-api.jup.ag/v6"

	// JupiterAPIURL provides higher rate limits, but includes a small 0.2% platform fee.
	JupiterAPIURL = "https://public.jupiterapi.com"
)
