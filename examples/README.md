# New client examples

The root package is the v1 implementation in development. See
[development status](../docs/development-status.md) for current coverage.

The current default collects no maintainer fee. Jupiter and network fees can
still apply. `WithReferral(account, 50)` configures a 0.5% managed-order referral;
the account must be a base58 Solana address. Provision the referral and applicable
token accounts separately. Always inspect returned fee fields before signing.
Missing referral token accounts can prevent referral collection.

Run a quote without signing or sending a transaction:

```sh
JUPITER_API_KEY=your-key go run ./examples/quote
```

A key may be omitted where Jupiter permits keyless access. The example requests
0.1 SOL in raw units and prints the USDC output amount in raw units. It omits a
`taker`, so the response is a quote without a signable transaction.

## Tokens and prices

```sh
JUPITER_API_KEY=your-key go run ./examples/tokens
```

Searches tokens by symbol and prints USD prices for SOL and USDC. `Tokens.Search`
takes one symbol, name, or mint; use `Tokens.Lookup` for up to 100 mints.
`Price.Get` accepts up to 50 mints and lists those without a reliable price in
`Missing`, since Jupiter omits them from the response. Numbers are
`json.Number`; an empty value means the API reported null.

## Trigger sessions

```sh
JUPITER_API_KEY=your-key SOLANA_PRIVATE_KEY=... go run ./examples/trigger
```

Trigger V2 is custodial: deposits move funds into a Jupiter-managed (Privy) vault
until orders fill or are cancelled. The example logs in, resolves the vault,
crafts and signs a deposit, lists active price orders, and logs out. It registers
a vault only when `JUPITER_REGISTER_VAULT=yes`, and creates a single limit order
only when `JUPITER_CREATE_ORDER=yes` and `TRIGGER_PRICE_USD` are set.

Order creation and cancellation confirmation submit signed transactions, so they
make one attempt; failures after which a transaction may have landed wrap
`trigger.ErrOutcomeUnknown`. Check `PriceOrderHistory` or `DCAOrderHistory`
before creating another order. Cancelling is two steps: `CancelPriceOrder` or
`CancelDCAOrder` returns an unsigned withdrawal to sign with `solanatx`, then
`ConfirmCancelPriceOrder` or `ConfirmCancelDCAOrder` submits it. USD prices are
`json.Number` values, sent exactly as given.

`client.Trigger.Login` signs a message challenge with any ed25519
`crypto.Signer`. Hardware wallets that sign only transactions use
`RequestChallenge` with `trigger.ChallengeTransaction`, sign the challenge
transaction with `solanatx`, and call `Verify`. Sessions belong to one wallet and
always use 15-minute access tokens with rotating refresh tokens; the deprecated
24-hour `access_only` mode is not supported. Sessions refresh shortly before
expiry and once after a 401. Pass `SessionOptions.OnRefresh` to persist rotated
tokens and `client.Trigger.Resume` to restore them; a stale refresh token returns
`trigger.ErrSessionExpired`.

## Signing and execution

Signing uses `solanatx`, the only new-SDK package that depends on solana-go.
It accepts any standard `crypto.Signer` with an ed25519 key, including
`ed25519.PrivateKey`, so hardware or remote signers can be plugged in. Signing
covers the exact received message bytes and writes only the signer's own slot;
the message, blockhash, and other signatures are never rewritten. Legacy,
version 0, and version 1 transactions can be signed, so managed orders may set
`MaxSupportedTransactionVersion` to 1. Other versions return
`solanatx.ErrUnsupportedVersion`. `solanatx.Assemble` builds version 0 only.

Both swap examples below read a base58 secret key from `SOLANA_PRIVATE_KEY`
and stop after signing unless `JUPITER_EXECUTE_SWAP=yes` is set. With that flag
they spend real funds (0.01 SOL for USDC), so use a dedicated wallet.

### Managed order

```sh
JUPITER_API_KEY=your-key SOLANA_PRIVATE_KEY=... go run ./examples/managed
```

Requests an order with a taker, signs the returned transaction, and calls
`client.Swap.Execute` with the order's request ID. Gasless and RFQ orders may
list another fee payer or cosigner that Jupiter signs during execution, so the
transaction ID can be unknown before execution. Execute is never retried; if it
returns `swap.ErrExecutionOutcomeUnknown`, check the transaction and wallet
activity before ordering again.

### Custom build

```sh
JUPITER_API_KEY=your-key SOLANA_PRIVATE_KEY=... SOLANA_RPC_URL=https://... \
  go run ./examples/build
```

Requests raw instructions, assembles a version 0 transaction with
`solanatx.Assemble` using the returned blockhash and lookup tables, signs it, and
submits it through your RPC node. Build transactions cannot use `Swap.Execute`.
The example waits for confirmation and stops when the blockhash expires; a
submitted signature alone does not prove success. `solanatx.Assemble` accepts
extra instructions to run after Jupiter's, and `solanatx.Instructions` and
`solanatx.AddressTables` support assembling the transaction yourself.

Every Swap result exposes `Metadata.StatusCode`, `Metadata.RequestID`, and
`Metadata.Headers`. Headers include upstream rate limits and `Retry-After`.
The HTTP request ID is separate from `OrderResponse.RequestID`, which belongs to
order execution. Program labels are available as `result.Labels` alongside the
same metadata. Always check the error before using response data; a non-nil
response may accompany an error so its metadata remains inspectable.

Optional scalar response fields use pointers where zero is meaningful. Check for
nil before reading `Gasless`, execution `Code`, build `TransactionVersion`, or
`BlockhashWithMetadata`. A missing execution code does not confirm success.
`ExecutionError.Code` is nil when the upstream response supplied no code.
