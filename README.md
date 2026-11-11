# jupiter-go: Jupiter Go SDK for Solana

### Swap tokens, get prices, place limit orders and DCA on Solana from Go, with the [Jupiter](https://jup.ag) APIs

[![Go](https://github.com/ilkamo/jupiter-go/actions/workflows/workflow.yml/badge.svg)](https://github.com/ilkamo/jupiter-go/actions/workflows/workflow.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/ilkamo/jupiter-go.svg)](https://pkg.go.dev/github.com/ilkamo/jupiter-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/ilkamo/jupiter-go)](https://goreportcard.com/report/ilkamo/jupiter-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

<img align="right" width="200" src="docs/assets/jup-gopher.png" alt="jupiter-go mascot: the Go gopher holding a JUP coin">

**jupiter-go** is a Go (Golang) SDK for [Jupiter](https://jup.ag), the leading
DEX aggregator and trading platform on Solana. Use it to get swap quotes and
execute token swaps through Jupiter's Swap API (Ultra / Swap V2), look up token
metadata and USD prices, place limit orders and dollar-cost averaging (DCA)
orders, earn yield and borrow with Jupiter Lend, trade prediction markets, and
land transactions fast through Jupiter's landing service, all from idiomatic,
typed Go code.

## Features

- **Complete Jupiter API coverage:** Swap V2, Tokens V2, Price V3, Trigger V2
  (limit orders, OCO, OTOCO, DCA), Lend, Prediction, Send, Studio, and
  transaction landing (`tx.jup.ag`).
- **Managed swaps without an RPC node:** order, sign locally, and let Jupiter
  land the transaction.
- **Custom swap transactions:** build instructions, add your own, and submit
  through your RPC node or Jupiter's landing infrastructure.
- **Safe signing** with any ed25519 `crypto.Signer` (local keys, HSM, KMS):
  messages and existing signatures are never modified. Legacy, v0, and v1
  Solana transactions are supported.
- **Production behavior:** typed errors, request IDs and rate-limit headers on
  every result, per-operation retries with backoff and `Retry-After`, and no
  blind retries of anything that moves funds.
- **Exact numbers:** raw token amounts and prices never pass through `float64`.
- **No SDK fee**, MIT licensed, verified against Solana mainnet.

## Supported Jupiter APIs

| Product | Package | Status |
| - | - | - |
| Swap V2 (order, execute, build, program labels) | [`swap`](swap) | Stable |
| Tokens V2 (search, tags, categories, Express Verification) | [`tokens`](tokens) | Stable |
| Price V3 | [`price`](price) | Stable |
| Trigger V2 (limit orders, OCO/OTOCO, DCA) | [`trigger`](trigger) | Stable |
| Transaction landing (`tx.jup.ag`) | [`transaction`](transaction) | Stable |
| Lend (Earn and Borrow) | [`lend`](lend) | **Beta** |
| Prediction markets | [`prediction`](prediction) | **Beta** |
| Send (invite transfers) | [`send`](send) | **Beta** |
| Studio (token launches) | [`studio`](studio) | **Beta** |

> **Beta products.** Jupiter labels Lend, Prediction, Send, and Studio as beta and
> may change them without notice. Their Go types follow upstream and can change in
> minor releases until Jupiter declares them stable.

**The SDK charges no fee.** You only pay Jupiter's and the network's fees. If you
run your own Jupiter referral account, `WithReferral` adds your fee to managed swap
orders.

## Installation

```bash
go get github.com/ilkamo/jupiter-go
```

Go 1.26 or newer is required.

## Quick start

```go
client, err := jupiter.NewClient(jupiter.WithAPIKey(os.Getenv("JUPITER_API_KEY")))
if err != nil {
	return err
}

quote, err := client.Swap.Order(ctx, swap.OrderRequest{
	InputMint:  "So11111111111111111111111111111111111111112", // SOL
	OutputMint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
	Amount:     "100000000",                                    // raw units: 0.1 SOL
})
if err != nil {
	return err
}

fmt.Println("output:", quote.OutAmount)
```

Get an API key at [developers.jup.ag/portal](https://developers.jup.ag/portal).

## Client options

| Option | Purpose |
| - | - |
| `WithAPIKey(key)` | Jupiter API key sent as `x-api-key`. |
| `WithTimeout(d)` | Bounds each operation, including its retries (default 30s). |
| `WithHTTPClient(c)` | Your own `Do(*http.Request)` implementation. The default refuses redirects so the API key never leaves Jupiter's host; a custom client owns that policy. |
| `WithoutRetries()` | One attempt per operation, for clients that already retry. |
| `WithReferral(account, bps)` | Your Jupiter referral fee (50–255 bps) on managed swap orders. |
| `WithBaseURL(url)` / `WithTransactionURL(url)` | Override `https://api.jup.ag` / `https://tx.jup.ag`, for example for tests. |

Clients are safe for concurrent use and make no network calls when constructed.

## Managed swap

`Order` with a `Taker` returns a transaction; sign it and hand it back to Jupiter.
Signing lives in [`solanatx`](solanatx), which accepts any ed25519 `crypto.Signer`
(local keys, HSMs, remote signers) and never alters the message or other signatures.

```go
order, err := client.Swap.Order(ctx, swap.OrderRequest{
	InputMint:  sol,
	OutputMint: usdc,
	Amount:     "10000000",
	Taker:      wallet.PublicKey().String(),
})
if err != nil {
	return err
}

tx, err := solanatx.Parse(*order.Transaction)
if err != nil {
	return err
}

if err := tx.Sign(ed25519.PrivateKey(wallet)); err != nil {
	return err
}

result, err := client.Swap.Execute(ctx, swap.ExecuteRequest{
	RequestID:         order.RequestID,
	SignedTransaction: tx.Base64(),
})
```

For full control use `client.Swap.Build`, assemble with `solanatx.Assemble`, and
submit through your RPC node or `client.Transaction.Send` (tx.jup.ag, which needs
a 0.001 SOL tip: set `BuildRequest.TipAmount` or add `solanatx.TipInstruction`).

## Other products at a glance

```go
// Tokens and prices
tokens, err := client.Tokens.Search(ctx, "JUP")
prices, err := client.Price.Get(ctx, []string{sol, usdc}) // prices.Missing lists unpriced mints

// Trigger V2: a wallet-scoped session with rotating refresh tokens
session, err := client.Trigger.Login(ctx, ed25519.PrivateKey(wallet), trigger.SessionOptions{})
vault, err := session.EnsureVault(ctx)

// Lend (beta): unsigned Earn deposit
deposit, err := client.Lend.Earn.Deposit(ctx, lend.AmountRequest{Asset: usdc, Signer: owner, Amount: "1000000"})

// Prediction (beta): markets priced in micro USD
events, err := client.Prediction.Events(ctx, prediction.EventsRequest{Filter: prediction.FilterTrending})

// Send (beta): invite codes never leave your process and print redacted
code, err := send.NewInviteCode()

// Studio (beta): token launch transaction plus presigned upload URLs
pool, err := client.Studio.CreatePool(ctx, poolRequest)
```

Runnable programs for every product are in [`examples/`](examples). Examples that
move funds only do so behind an explicit environment flag.

## Errors, retries, and safety

- **HTTP errors** are `*jupiter.APIError` (status, Jupiter error code and message,
  request ID, headers). Validation errors are returned before any request is sent;
  malformed addresses match `errors.Is(err, swap.ErrInvalidAddress)`.
- **Response metadata.** Every result carries `Metadata` (status, request ID,
  headers including rate limits, attempts), also alongside errors.
- **Retries follow each operation's semantics**, not its HTTP method: reads and
  operations that only build unsigned transactions retry transient failures with
  jittered backoff and honour `Retry-After`.
- **Submissions never retry blindly.** Operations that submit signed transactions
  (Swap and Prediction execute, Trigger order creation and cancellation, Studio
  submit, Express Verification) make one attempt. When a transaction might still
  land, the error wraps that package's `ErrOutcomeUnknown`: check on-chain state
  before trying again. `transaction.Send` retries because resubmitting identical
  signed bytes cannot execute twice.
- **Exact numbers.** Raw token amounts are validated integer strings; prices and
  USD values are `json.Number` or strings, never `float64`.

## FAQ

### How do I swap tokens on Solana in Go?

Call `client.Swap.Order` with the input mint, output mint, raw amount, and your
wallet as `Taker`; sign the returned transaction with `solanatx.Parse(...).Sign`;
then call `client.Swap.Execute`. Jupiter routes the swap across Solana DEXes and
lands the transaction for you. See [Managed swap](#managed-swap) and
[`examples/managed`](examples/managed).

### How do I get a Solana token price in Go?

`client.Price.Get(ctx, mints)` returns USD prices from Jupiter's Price API V3
for up to 50 mints; mints without a reliable price are listed in `Missing`.
`client.Swap.Order` without a `Taker` returns a swap quote.

### Can I place limit orders or DCA orders on Jupiter from Go?

Yes. `client.Trigger.Login` opens a wallet session for the Trigger V2 API, which
supports single limit orders, take-profit/stop-loss pairs (OCO), OTOCO,
trailing stops, and time-based or price-conditional DCA. See
[`examples/trigger`](examples/trigger).

### Do I need a Jupiter API key?

Requests work without a key at a lower rate limit. For production, get a key
at [developers.jup.ag/portal](https://developers.jup.ag/portal) and pass it with
`jupiter.WithAPIKey`.

### Do I need a Solana RPC node?

Not for managed swaps (`Swap.Order` and `Swap.Execute`) or any read API. You
need one to confirm transactions you submit yourself, for example custom swaps
built with `Swap.Build`; use a dedicated RPC provider for that.

### Which Solana transaction versions are supported?

`solanatx` parses and signs legacy, version 0 (with address lookup tables), and
version 1 transactions without modifying the message or other signatures.

### Is jupiter-go an official Jupiter SDK?

No. It is an independent, community-maintained open-source library. It charges
no fee and has been verified with real swaps on Solana mainnet.

### Which Go versions are supported?

Go 1.26 and newer. CI tests Go 1.26 and 1.27.

## Migrating from v0

v1 is a rewrite. The `jupiter` (generated client) and `solana` subpackages are
removed. See [CHANGELOG.md](CHANGELOG.md#migrating-from-v0) for the mapping.

## Contribute

Contributions are welcome. Read [AGENTS.md](AGENTS.md) for the code conventions and
run `make check` (vet, golangci-lint, race tests) before opening a pull request.

## License

This library is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

This is an unofficial, community-maintained library and is not affiliated with or
endorsed by Jupiter. The Go gopher was designed by
[Renée French](https://reneefrench.blogspot.com/) and is licensed under
[CC BY 4.0](https://creativecommons.org/licenses/by/4.0/); the mascot artwork in
`docs/assets/` is a derivative of it and carries the same license.

## Donate

If you find this library useful and want to support its development, consider donating some JUP/Solana to the following address:

`BXzmfHxfEMcMj8hDccUNdrwXVNeybyfb2iV2nktE1VnJ`

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=ilkamo/jupiter-go&type=Date)](https://www.star-history.com/#ilkamo/jupiter-go&Date)
