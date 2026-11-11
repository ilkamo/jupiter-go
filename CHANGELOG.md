# Changelog

All notable changes to this project are documented here. The project follows
[Semantic Versioning](https://semver.org).

## v1.0.0 (unreleased)

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

### API review before release

- `swap.ErrExecutionOutcomeUnknown` is now `swap.ErrOutcomeUnknown`, matching
  every other product.
- `trigger.Tokens` is now `trigger.AuthTokens` (and `Session.Tokens` is
  `Session.AuthTokens`), so it is not confused with the Tokens product.
- `transaction.TipReceivers` is a function returning a copy, so callers cannot
  alter the shared list.
- `trigger.Session.UpdatePriceOrder` returns `*trigger.OrderUpdate` with response
  metadata instead of a bare ID.

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
