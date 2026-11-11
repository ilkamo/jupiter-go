# Changelog

All notable changes to this project are documented here. The project follows
[Semantic Versioning](https://semver.org).

## [1.1.0](https://github.com/ilkamo/jupiter-go/compare/v1.0.1...v1.1.0) (2026-11-11)


### Features

* **solanatx:** add Jito sandwich mitigation for assembled transactions ([#64](https://github.com/ilkamo/jupiter-go/issues/64)) ([5a860a5](https://github.com/ilkamo/jupiter-go/commit/5a860a5b0fb3a5b94adb9ba1da35a7f88ec52f64)), closes [#47](https://github.com/ilkamo/jupiter-go/issues/47)

## [1.0.1](https://github.com/ilkamo/jupiter-go/compare/v1.0.0...v1.0.1) (2026-11-11)


### Bug fixes

* **deps:** bump klauspost/compress to v1.20.2 for GO-2026-6865 ([#62](https://github.com/ilkamo/jupiter-go/issues/62)) ([64d0bf8](https://github.com/ilkamo/jupiter-go/commit/64d0bf8705b822bc168119ea084e6616c63d093f))

## v1.0.0 (2026-11-11)

v1 is a rewrite around a single root client, `jupiter.NewClient`, with typed
product clients. It targets Jupiter's current APIs (Swap V2, Tokens V2, Price V3,
Trigger V2, Lend, Prediction, Send, Studio, and tx.jup.ag).

### Breaking changes

- The `jupiter` (generated Swap V1 client) and `solana` subpackages are removed,
  together with their dependencies (oapi-codegen runtime, shopspring/decimal).
- Go 1.26 or newer is required.

### Added

- `jupiter.NewClient` with functional options: API key, timeout, custom HTTP
  client, base and transaction URLs, retry opt-out, and an optional referral.
- Product clients: `Swap`, `Tokens` (including paid Express Verification),
  `Price`, `Trigger`, `Transaction`, and the beta `Lend`, `Prediction`, `Send`,
  and `Studio`.
- `solanatx`: parse, partially sign, and inspect legacy, v0, and v1 transactions
  with any ed25519 `crypto.Signer`; assemble swap builds; build Jupiter tips.
- Per-operation retry policies with jittered backoff and `Retry-After`; single
  attempts and `ErrOutcomeUnknown` for operations that submit signed
  transactions; response metadata (request ID, headers, attempts) on every result.
- Input validation before any request: Solana addresses, raw integer amounts,
  documented ranges and parameter combinations.

### Notes

- The SDK charges no fee. `WithReferral` only adds your own Jupiter referral to
  managed swap orders.
- Lend, Prediction, Send, and Studio are beta upstream; their types may change in
  minor releases until Jupiter declares them stable.

## Migrating from v0

| v0 | v1 |
| - | - |
| `jupiter.NewClientWithResponses(jupiter.DefaultAPIURL, jupiter.WithRequestEditorFn(...))` | `jupiter.NewClient(jupiter.WithAPIKey(key))` |
| `jupClient.QuoteGetWithResponse(ctx, &jupiter.QuoteGetParams{...})` | `client.Swap.Order(ctx, swap.OrderRequest{...})` (quote only without `Taker`) |
| `jupClient.SwapPostWithResponse(...)` then `solanaClient.SendTransactionOnChain(...)` | `client.Swap.Order` with `Taker`, sign with `solanatx.Parse(...).Sign(key)`, then `client.Swap.Execute` (Jupiter lands it) |
| Swap instructions and your own transaction | `client.Swap.Build`, `solanatx.Assemble`, then your RPC node or `client.Transaction.Send` |
| `solana.NewWalletFromPrivateKeyBase58(key)` | `solana.PrivateKeyFromBase58(key)` from solana-go, used as `ed25519.PrivateKey(key)` |
| `solanaClient.CheckSignature` / `Monitor.WaitForCommitmentStatus` | Poll `GetSignatureStatuses` on your RPC node (see `examples/build`), or check `swap.ExecuteResponse` for managed swaps |
| Prioritization fee and Jito tip request fields | `OrderRequest.PriorityFeeLamports`, `JitoTipLamports`, `BroadcastFeeType`; `BuildRequest.ComputeUnitPricePercentile`, `TipAmount` |

Key behaviour changes:

- Managed swaps (`Order` and `Execute`) no longer need an RPC node: Jupiter
  submits the signed transaction.
- Signing never rewrites the blockhash or message; v0's `SendTransactionOnChain`
  replaced the blockhash, which managed orders do not allow.
- Amounts are raw integer strings and decimal values are `json.Number`, so no
  precision is lost through `float64`.
