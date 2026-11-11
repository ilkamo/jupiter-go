# v1 decision log

The decisions behind the v1 rewrite and the reason for each. Planned work is
tracked in GitHub issues; conventions are in [AGENTS.md](../AGENTS.md).

## Scope and release policy

- **One root client, typed product clients.** `jupiter.NewClient` with
  functional options returns `Swap`, `Tokens`, `Price`, `Trigger`, `Transaction`,
  `Lend`, `Prediction`, `Send`, and `Studio`. Product packages never import the
  root package, which avoids import cycles.
- **Cover every public REST operation.** The inventory (`api-inventory.md`, with
  hashes in `api-reference-manifest.json`) lists 82 indexed operations plus
  Prediction's spec-only `GET /forecast`; all are implemented.
- **No SDK fee.** The maintainer-fee concept was removed. `WithReferral` only
  adds the integrator's own Jupiter referral to managed swaps and cannot be set
  twice, so the result never depends on option order.
- **Beta products are labelled, not withheld.** Lend, Prediction, Send, and
  Studio are beta upstream; their types may change in minor releases. Swap,
  Tokens, Price, Trigger, and Transaction are frozen at v1.0.0.
- **The legacy library is dropped.** The generated `jupiter` client, the `solana`
  helpers, and `_examples` were removed with their dependencies; v0 receives no
  further maintenance. `CHANGELOG.md` maps v0 calls to v1.
- **Module path unchanged.** `github.com/ilkamo/jupiter-go` moves from v0 to v1
  without a `/v1` suffix.
- **Releases come from commits, not merges.** release-please opens a release
  pull request from Conventional Commit types; merging it tags the version.
  Docs, test, build, and CI commits alone never trigger a release.
- **Docs stay out of the download.** `docs/` (plans, inventory, pinned
  references in `docs/openapi/`, images in `docs/assets/`) has its own `go.mod`,
  so Go excludes it from the library's module zip.

## Transport and configuration

- **No network calls at construction; clients are safe for concurrent use.**
- **The default HTTP client refuses redirects** so the API key cannot be
  forwarded. A caller-supplied client owns its redirect policy and is never
  modified.
- **`WithTimeout` bounds a whole operation**, including retries and waits.
- **Separate service URLs.** tx.jup.ag gets its own transport and
  `WithTransactionURL` instead of overloading the API base URL.
- **Presigned uploads never receive credentials.** External URLs must be HTTPS
  without user info; the API key and request headers are not sent and the
  signed query string is used unchanged.
- **Requests are encoded once and replayed byte-for-byte on retries.**
- **Response metadata everywhere.** Every result carries status, request ID,
  headers, and attempt count, also alongside HTTP errors; validation failures
  return before any request.

## Retries and submission safety

- **Retry by semantics, not HTTP method.** Reads use `ReadOnlyRetry`; operations
  that only build unsigned transactions or return fresh single-use values use
  the short `RepeatableRetry`. Retries cover 429, 5xx, and transport failures,
  with full-jitter backoff and `Retry-After`; a server wait beyond the policy
  limit or the deadline stops retrying.
- **Every operation must name its retry policy**; an unset policy is rejected so
  new operations decide deliberately.
- **Submissions of signed transactions make one attempt** (Swap and Prediction
  execute, Trigger order creation and cancellation confirmation, Studio submit,
  Express Verification execute), because no replay guarantee is documented.
- **Uncertain outcomes are explicit.** Failures after which a transaction may
  still land (transport errors, 5xx, unreadable success bodies) wrap the
  package's `ErrOutcomeUnknown`; 4xx and explicit failure statuses are known
  rejections; an already ended context sends nothing.
- **Explicit success is required.** Execute responses must state success (Swap
  also requires a zero code); anything else is an execution error.
- **tx.jup.ag is the exception**: resubmitting identical signed bytes cannot
  execute twice because Solana deduplicates by signature, so transient failures
  are retried.

## Validation and types

- **Validate locally with established parsers, never with RPC.** Addresses are
  base58 (mr-tron/base58) decoding to 32 bytes; curve membership is not checked
  because program-derived accounts are off-curve. Validated addresses compare as
  strings because base58 encodings are unique.
- **Raw amounts are positive base-10 integers** parsed with `math/big` and sent
  canonically; Lend's signed amounts follow the documented pattern and the
  signed 128-bit range.
- **Documented combinations are enforced; undocumented ones are not invented.**
  For example, receiver and payer require a taker, but a price-conditional DCA is
  not required to have a price band.
- **Enum-like filters accept upstream additions** (routers, tags, categories,
  markets, history filters) after a format check; documented values are
  constants. Values that determine required parameters (Trigger order types) are
  validated strictly.
- **Exact numbers.** Decimal values use `json.Number` or strings and never pass
  through `float64`. Where schemas and examples disagree on a type, the SDK
  accepts both forms (`lend.ChainID`, `prediction.StringOrNumber`,
  `send.Confirmed`).
- **Presence matters where the API distinguishes it**: nullable scalars that can
  be zero use pointers; opaque payloads such as Prediction's execution context
  stay `json.RawMessage` and are returned unchanged.

## Signing (`solanatx`)

- **The only package that depends on solana-go**, so the root package stays
  free of Solana dependencies.
- **Sign the exact received message bytes** and write only the signer's slot;
  messages, blockhashes, and existing signatures are never rewritten. A
  transaction is rejected unless solana-go re-encodes it byte-for-byte.
- **Any ed25519 `crypto.Signer`** (local keys, HSM, KMS, invite keys).
- **Versions are detected before parsing.** Legacy, v0, and v1 are supported;
  `Assemble` builds v0 only, because v1 builds need a compute-unit limit the
  build response does not provide.
- **Latest solana-go (v1.26.0)**, which added v1 transactions.

## Product decisions

- **Swap:** managed orders need no RPC node; build instructions are ordered
  compute budget, setup, swap, cleanup, other, tip (the docs do not place
  "other").
- **Tokens and Price:** comma-containing searches are rejected in favour of
  `Lookup`; Price reports mints Jupiter omits in `Missing`. Express Verification
  sits on `Tokens.Express` because it pays for submissions.
- **Trigger:** wallet-scoped sessions; the client holds no wallet identity. Only
  `access_refresh` tokens are used; refreshes are serialized because reusing a
  rotated refresh token revokes the session; rotated tokens are reported through
  `OnRefresh`. `EnsureVault` handles both documented register behaviours.
- **Lend:** builders only, no submission; Borrow instructions return lookup
  table addresses only, so assembly is left to the caller's RPC.
- **Prediction:** one `TransactionBuild` type for all builders; the live API
  requires an owner for orders, positions, and history although the spec marks
  it optional, so the SDK requires it.
- **Send:** invite codes and keys never leave the client; `InviteCode` prints
  redacted; the invite PDA uses filippo.io/edwards25519 for the off-curve check.
- **Studio:** multipart submit; uploads to presigned URLs; `Fee` is a POST that
  only reads, so it uses the read-only policy.
- **Transaction:** JSON-RPC errors are typed `*RPCError`; tip receivers are
  exposed through a function returning a copy.

## API settled before freezing

- `swap.ErrOutcomeUnknown` (renamed for consistency with every product).
- `trigger.AuthTokens` (renamed to avoid confusion with the Tokens product).
- `transaction.TipReceivers()` (a copy instead of a mutable exported slice).
- `trigger.Session.UpdatePriceOrder` returns `*OrderUpdate` with metadata.
- Product `NewClient` functions stay exported because the root package calls
  them; they take an internal transport and are documented as such.

## Verification approach

- **Jupiter's APIs are mainnet-only**, so verification used free mainnet reads,
  unsubmitted signed transactions with RPC simulation, and a few tiny funded
  swaps from a throwaway wallet. No wallet addresses, signatures, or balances
  are recorded in this repository.
- **Live behaviour wins over the spec.** Live checks found Prediction's
  `closeTime` sent as a number and owner filters being required; the SDK follows
  the live API. Prediction `pnl-history` and `events/suggested` return 404
  upstream and stay implemented as documented.
- **RPC providers differ.** Some return the slot from `getBlockHeight`, which
  makes fresh blockhashes look expired; examples recommend a dedicated provider.
- **Quality gates:** race tests on Go 1.26 and 1.27, `go vet`, golangci-lint with
  gofumpt, and fuzz targets for transaction parsing, address validation, and
  amount parsing (`make check`, `make fuzz`).
