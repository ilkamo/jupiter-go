// Package jupiter is a Go SDK for Jupiter (https://jup.ag), the DEX aggregator and
// trading platform on Solana.
//
// [NewClient] returns a [Client] whose fields cover every public Jupiter REST API:
//
//   - Swap: swap quotes, managed swaps (order and execute), and swap instructions
//     for custom transactions (Swap API V2).
//   - Tokens: token search, metadata, tags, trending categories, and Express
//     Verification (Tokens API V2).
//   - Price: USD token prices (Price API V3).
//   - Trigger: limit orders, OCO and OTOCO, trailing stops, and DCA (Trigger API V2).
//   - Transaction: transaction landing through tx.jup.ag.
//   - Lend, Prediction, Send, and Studio: beta upstream; their types may change in
//     minor releases.
//
// The root package performs no signing and does not depend on a Solana library.
// Package solanatx parses, signs, and assembles transactions with any ed25519
// crypto.Signer. The SDK charges no fee.
//
// A swap quote for 0.1 SOL to USDC:
//
//	client, err := jupiter.NewClient(jupiter.WithAPIKey(apiKey))
//	if err != nil {
//		return err
//	}
//
//	quote, err := client.Swap.Order(ctx, swap.OrderRequest{
//		InputMint:  "So11111111111111111111111111111111111111112",
//		OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
//		Amount:     "100000000",
//	})
//
// Operations validate input before sending, return response metadata with every
// result, retry only when repeating is safe, and report submissions whose outcome
// is uncertain with each package's ErrOutcomeUnknown.
package jupiter
