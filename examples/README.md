# Examples

Runnable programs for each product. Run them from the repository root with
`go run ./examples/<name>`. Examples that move funds only do so when an explicit
environment flag is set; without it they stop after building or signing.

The SDK charges no fee; Jupiter and network fees still apply. If you operate a
Jupiter referral account, `WithReferral(account, 50)` adds a 0.5% referral to
managed swap orders. Provision the referral and its token accounts separately:
without them orders execute without collecting the fee. Always inspect the
returned fee fields before signing.

## Quote

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

Set `EXPRESS_TOKEN_ID` to also check Express Verification eligibility. Express
submissions cost 1000 JUP (free standard submissions are on verified.jup.ag):
`client.Tokens.Express.CraftPayment` builds the payment in JUP, SOL, USDC, or
JUPUSD; sign it with `solanatx`, then `Execute` with the token details. For
non-JUP payments pass the craft response's `QuotedInputAmount` as
`PaymentAmount`. Execute makes one attempt; `tokens.ErrOutcomeUnknown` marks
failures after which the payment may have landed.

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

## Lend

```sh
JUPITER_API_KEY=your-key WALLET_ADDRESS=... go run ./examples/lend
```

Lists Earn tokens and Borrow vaults and, when `WALLET_ADDRESS` is set, builds an
unsigned Earn deposit. Lend is beta upstream. `client.Lend.Earn` builds deposit,
withdraw, mint, and redeem transactions or single instructions;
`client.Lend.Borrow.Operate` builds one transaction from signed collateral and debt
deltas (`lend.NewPosition`, `lend.WithdrawAllCollateral`, `lend.RepayAllDebt`).
Nothing is submitted: sign the transaction with `solanatx` and send it through
your RPC node. Borrow instruction responses list lookup table addresses only, so
compiling them needs an RPC node; `solanatx.ConvertInstruction` converts the
instructions. Borrow does not wrap native SOL.

## Prediction markets

```sh
JUPITER_API_KEY=your-key go run ./examples/prediction
```

Checks the exchange status and prints trending events with YES prices. The API is
beta upstream and blocks United States and South Korea IP addresses. Prices and
USD values are micro USD (`850000` is $0.85); contract quantities in `*Micro`
fields use 1000000 per contract.

Trading: `client.Prediction.CreateOrder` builds a buy (market, side, deposit in
JupUSD or USDC, 5 USD minimum) or a sell (position and exactly one contract
quantity). `ClosePosition`, `CloseAllPositions`, and `ClaimPosition` build sells
and payout claims. Sign the transaction with `solanatx` for every key in
`RequiredSigners`, then submit it with `Execute`, passing `Execution.Context`
unchanged. Execute makes one attempt, requires an explicit `Success`, and wraps
`prediction.ErrOutcomeUnknown` when the transaction may still land. An executed
order is open, not filled; poll `OrderStatus`.

## Send

```sh
JUPITER_API_KEY=your-key SENDER_ADDRESS=... go run ./examples/send
```

Generates an invite code and builds an unsigned 0.01 SOL send to it. Send is beta
upstream; recipients claim in Jupiter Mobile. Invite codes and their keys stay on
the client: `send.NewInviteCode` generates one, `InviteCode.Signer` and
`InviteCode.PDA` give the addresses the API needs, and `InviteCode.PrivateKey` is
an ed25519 `crypto.Signer` for `solanatx`. A send transaction needs the sender
and invite signatures; a clawback (`CraftClawback`) needs only the sender. An
`InviteCode` prints as redacted; `Reveal` returns the code to share.

## Studio

```sh
JUPITER_API_KEY=your-key SOLANA_PRIVATE_KEY=... TOKEN_NAME=... TOKEN_SYMBOL=... \
  go run ./examples/studio
```

Builds and signs a token launch on a Dynamic Bonding Curve with the documented
"Meme" preset and stops. With `JUPITER_LAUNCH_TOKEN=yes` and a PNG at
`TOKEN_IMAGE`, it uploads the image and metadata to the presigned URLs and
submits, which creates the token on-chain. Uploads go only to HTTPS presigned
URLs and never carry the API key. `Submit` makes one attempt and wraps
`studio.ErrOutcomeUnknown` when the launch may still land; check
`PoolAddresses` for the mint before submitting again. Studio is beta upstream.

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

Set `JUPITER_SIMULATE=yes` with `SOLANA_RPC_URL` to simulate the signed
transaction on your node instead of sending it: signatures are verified and the
swap runs against current mainnet state at no cost. An unfunded wallet fails with
`AccountNotFound` or insufficient funds, which still shows the transaction was
well formed and correctly signed.

Set `JUPITER_SUBMIT_VIA=tx` to submit through Jupiter's landing service
(`client.Transaction.Send`, `https://tx.jup.ag`) instead of the RPC node. Every
transaction sent there must tip at least `transaction.MinTipLamports` (0.001 SOL)
to one of `transaction.TipReceivers`: the example sets `BuildRequest.TipAmount`;
for other transactions add `solanatx.TipInstruction`. The service is send-only,
so confirmation still uses `SOLANA_RPC_URL`. JSON-RPC rejections are
`*transaction.RPCError`; resubmitting the same signed bytes is safe, so transient
failures are retried, and `transaction.ErrOutcomeUnknown` marks failures after
which the transaction may have been forwarded.

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
