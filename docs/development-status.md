# v1 development status

Updated: 2026-11-11. Implementation is in progress; this is not a v1 release.

## Working conventions

Repository-wide instructions are maintained in [`AGENTS.md`](../AGENTS.md).
Long function signatures use one parameter per line.

Continue development on branch `v1`. Use blank lines after error checks, if
blocks, and loops, and between distinct logical steps. Keep assignments next to
their error checks; expand compressed functions and separate standard-library
imports from project imports. Apply this style to examples and tests too.
Handle every returned error explicitly, including decoding, response-body cleanup,
constructor calls in tests, and output writes; never discard an error with `_`.
Use descriptive constants with explicit units for policy limits. Prefer standard
library mechanisms over manual sentinel-byte arithmetic or clever shortcuts.

## API inventory completed

Added [the API inventory](api-inventory.md) and its
[source manifest](api-reference-manifest.json). All 82 API-reference pages linked
from the current documentation index were fetched and catalogued across nine
products. The inventory records operations, servers, declared authentication,
pagination controls, source hashes, and SDK/test coverage. Full schema and replay
semantics remain per-operation implementation prerequisites.

Key design findings: transaction submission uses a separate JSON-RPC endpoint;
Trigger vault registration is a state-changing GET; Tokens includes payment-based
verification operations. These require explicit component boundaries before the
corresponding clients are implemented.

Validation for this documentation step: all references fetched successfully;
unique reference URLs and method/URL pairs; JSON manifest parsed successfully;
`git diff --check` passed. Go code was unchanged, so the previously passing code
validation remains applicable.

## Implemented

- Root `jupiter.NewClient` and functional options for HTTP injection, gateway URL,
  API key, operation timeout, and explicit referral policy.
- Separate HTTP transport and Swap V2 service responsibilities.
- Typed order, execute, build, and program-label operations.
- Exact raw amount strings and `json.Number` for decimal response values.
- Structured HTTP and execution errors; context errors retain their causes.
- Default HTTP redirect refusal, bounded responses, and body cleanup.
- Configuration and request validation, including conflicting fee options.
- Contract tests and a runnable quote-only example in `examples/quote`.
- Pinned official Swap endpoint documentation in `openapi/swap-v2`, fetched from
  the URLs in `v1-plan.md` on the date above. These Markdown files contain their
  OpenAPI definitions; deterministic schema extraction/generation is still pending.

The existing `jupiter` and `solana` subpackages remain available. The new root
package imports only standard-library code and the new Swap service; it does not
load a wallet or mutate transaction messages. Legacy dependencies remain in the
module until the migration strategy is finalized.

## Decisions and limits

- No maintainer fee is enabled: no recipient account has been supplied. The
  default-on policy remains a release decision. `WithoutMaintainerFee` makes the
  current policy explicit; `WithReferral` configures managed-order referrals.
- Build platform fees are explicit request fields and require a token fee account;
  managed-order referrals do not carry into builds.
- Retries are chosen per operation (see "Operation-specific retry policies").
  Execute makes one attempt and reports ambiguous failures explicitly.
- A custom HTTP implementation is responsible for its own redirect policy. The
  constructor documents this; caller-owned HTTP clients are never mutated.
- Signing lives in `solanatx`, separate from the HTTP clients. Execute still
  accepts caller-signed base64 bytes unchanged and does not parse them.
- The public API is provisional until the compatibility review.
- The repository's Go 1.26 baseline is unchanged. Validation uses the installed
  Go 1.27.4 toolchain; support-range CI remains pending.

## Resume here

1. Run `go test -race ./...` and `go vet ./...`; inspect this document's validation
   section before repeating already completed checks.
2. Use `docs/api-inventory.md` and the source manifest as the endpoint checklist.
   Verify full schemas, effective authentication, pagination semantics, and
   idempotency per operation before implementing it. Discovery is complete for
   the 82 indexed references; guide-only operations still need reconciliation.
3. Swap is functionally complete: validation, retries, signing adapters, and
   examples. Remaining Swap work: version 1 assembly for build responses (needs a
   compute-unit-limit policy) and a live smoke test with a funded wallet before
   release.
4. Implement transaction submission and Tokens verification. Audit the current documentation index for additional APIs.
   These are not implemented. Decide the compatibility boundary for upstream beta
   products (Lend, Prediction) during the compatibility review.
5. Complete pinned schema tooling, generation checks, fuzz tests, examples, CI,
   migration documentation, and compatibility review from `v1-plan.md`.
6. Resolve the maintainer fee recipient/default and beta API compatibility before
   release candidates. Publishing and funded on-chain account setup are pending.

## Validation

Contract tests were added for amount precision, referral encoding, null quote
transactions, structured errors, one-attempt execution, cancellation, option
conflicts, and redirect refusal. Validation completed successfully with Go 1.27.4:

- `go test -race ./...` (all packages, including legacy packages and the quote example)
- `go vet ./...`
- `git diff --check`

Additional contract tests verify build instruction preservation, separation of
build fees from managed referrals, and byte-for-byte preservation of the signed
base64 payload supplied to execute. No live swaps or funded transactions were performed.

## Readability and error audit

Reviewed all new Go sources, tests, examples, and planning documents against
`AGENTS.md`. API limits now have shared names for basis points, referral fees,
account counts, blockhash expiry slots, and transaction versions. Default and
example timeouts and quote parameters are named. HTTP status checks use standard
library constants. Ordinary zero/empty checks and explicit fixture data remain
literal where their meaning is direct.

Removed the JSON-character-index shortcut from error tests. Parsing failures now
wrap their causes separately from range validation. Added regression cases for
malformed numeric inputs and values outside the protocol limits. Pinned upstream
Markdown/OpenAPI references retain their original numeric schema constraints;
they are reference snapshots rather than handwritten implementation code.

Protocol mode, priority-level, broadcast-fee, and execution-status strings now
have named constants. Long handwritten Go conditions, calls, and fixtures were
reviewed and wrapped; the conventions are recorded in `AGENTS.md`.

Raw swap amounts are parsed with standard-library `math/big.Int` in base 10,
checked for positivity, and serialized canonically. This replaces custom digit
scanning and supports values beyond machine integer widths. Fractional and
scientific notation are rejected because the API expects raw whole-token units;
decimal arithmetic belongs in fractional price or human-unit conversion code.

## Swap contract and metadata update

Compared the pinned order/build/execute success properties with public response
fields. Added missing order RFQ details, timing, swap USD value, legacy impact/type
fields, route USD value, and build swap mode/slippage. Added the order `swapMode`
parameter with ExactIn validation. Nullable taker and transaction fields retain
pointer representations; exact numbers remain `json.Number`.

Added public `response.Metadata` and changed the shared transport to return it
per call. Every Swap result now carries HTTP status, request ID, and copied headers
on success and response failures. API errors also expose copied headers. Program
labels now return `*ProgramLabelsResponse` with `Labels` and `Metadata`, replacing
the provisional map-only return. There is no shared last-response state.

Tests cover RFQ/route numeric precision, order mode encoding, build mode and zero
slippage, program-label success responses, HTTP versus order IDs, metadata on
failures, and header-copy ownership. Remaining work includes optional-field
presence semantics, address validation, retries, and signing adapters. This step
does not mark the full Swap milestone or v1 release complete.

Validation for this update: `go test -race ./...`, `go vet ./...`, and
`git diff --check` passed. No live trading requests were sent.

## Optional scalar presence and execution confirmation

Committed the preceding inventory and response-metadata work as `d13fab5`.
The next change preserves optional execution code, gasless flag, build transaction
version, blockhash metadata, and timestamp component presence using pointers.
Blockhash metadata now has named types instead of anonymous nested structs.

Execute requires both an explicit Success status and an explicitly supplied zero
code. Missing/null codes, unknown/missing statuses, and contradictory status/code
pairs return an error with the response and metadata intact. ExecutionError.Code
is also a pointer, so missing codes are not reported as upstream zero codes.
Omitted and null optional scalars both mean unavailable; explicit zero/false are
preserved. No distinct null-versus-absent meaning is assumed for these scalars.

Regression coverage includes missing/null codes, unknown status, contradictory
success, explicit false, transaction version zero, and zero-valued blockhash time
components. This does not complete address validation, retry policy, or signing
adapters; these remain the next implementation work.

Validation: `go test -race ./...`, `go vet ./...`, and `git diff --check` passed
for the optional-field update. No live trades were performed.

## Swap request validation

Mint, wallet, payer, receiver, destination, fee, and referral addresses are now
validated before any HTTP request. `internal/address` decodes them with the
established `github.com/mr-tron/base58` parser (already in the module graph via
solana-go, now a direct dependency) and requires exactly 32 decoded bytes. Curve
membership is deliberately not checked: associated token accounts and other
program-derived addresses are off-curve. Base58 encodings are unique, so validated
addresses are compared as strings. Address failures wrap `swap.ErrInvalidAddress`
together with the parser's cause.

Parameter-combination rules now enforced:

- Order and build: input and output mints must differ.
- Order: receiver and payer require a taker; receiver must differ from taker
  (documented). Payer equal to taker is accepted because the docs describe it as
  a no-op rather than an error. `broadcastFeeType` requires a priority fee or Jito
  tip, since the API otherwise ignores it silently.
- Order and build: router names and DEX labels must be non-empty, contain no
  comma separator, and have no surrounding whitespace, so joining them cannot
  change their meaning. Router names are not restricted to the documented set
  (`metis`, `jupiterz`, `dflow`, `okx`, provided as named constants) so routers
  added upstream remain usable without an SDK release.
- Build: added the documented `maxBytes` parameter (1–1232); it is rejected with
  transaction version 1, which upstream answers with HTTP 400.
- Execute: `lastValidBlockHeight`, when supplied, must be an unsigned 64-bit
  integer.

`OrderRequest`, `BuildRequest`, and `ExecuteRequest` now have separate
`validate` and query/encoding steps; query construction always validates first.
`swap.Referral.Validate` checks the referral address and fee range and is used by
`WithReferral`. Grouped request struct fields were expanded to one per line.

Boundary tests cover 31/32/33-byte keys, excluded base58 digits, whitespace,
non-ASCII and hexadecimal input, every numeric range edge (slippage, platform fee,
accounts, bytes, blockhash expiry, compute-unit percentile, tip, block height,
referral fee), and each combination rule. A public-API test uses an HTTP client
that fails the test if invoked, proving invalid requests never reach transport.
Test fixtures now use real base58 addresses.

Decision: request validation is limited to format, range, and combination checks
and never requires RPC access. Mint existence, token-account mint or ownership,
and whether `nativeDestinationAccount` is used only with SOL output are left to
the API.

Validation: `go test -race ./...`, `go vet ./...`, `gofmt -l`, and
`git diff --check` passed with Go 1.27.4. No live requests were sent.

## Operation-specific retry policies

Committed request validation as `98f57b3`. The shared transport now runs a
bounded attempt loop driven by a `transport.RetryPolicy` that every
`transport.Request` must specify; an unset policy is rejected before any request
is sent, so new operations must choose their replay semantics explicitly.

Policies follow operation semantics, not the HTTP method:

| Operation | Policy | Reason |
| --- | --- | --- |
| Order, build | 3 attempts, 100–500 ms jittered backoff, Retry-After up to 1 s | Repeating changes no state; each attempt returns a fresh quote, blockhash, and request ID; results are time-sensitive. |
| Program labels | 4 attempts, 250 ms–2 s backoff, Retry-After up to 5 s | Read-only reference data. |
| Execute | One attempt | The pinned documentation gives no replay guarantee for a signed transaction and request ID. |

Retried conditions: HTTP 429, 500, 502, 503, 504, and transport failures while
the operation context is still live. Other statuses, body-read failures, and
decode failures are returned immediately. Waits use full-jitter exponential
backoff (`math/rand/v2`) unless the server sends `Retry-After` (seconds or
HTTP-date), which is honored exactly. A Retry-After above the policy limit, or
any wait that would pass the operation deadline, stops retrying and returns the
last server error. `WithTimeout` bounds the whole operation including waits.
Cancellation during a wait returns both the last server error and the context
error. Request bodies are encoded once and replayed byte-for-byte.

`response.Metadata.Attempts` reports the attempts made. `WithoutRetries()` forces
single attempts for callers whose HTTP client already retries.

Execute failures that may still have landed wrap `swap.ErrExecutionOutcomeUnknown`:
transport failures, timeouts, 5xx responses (the documented 500 body can carry a
signature), and unreadable success bodies. 4xx responses, including 429, are
definitive rejections. If the context has already ended before execute, nothing
is sent and the context error is returned without the unknown marker. The
upstream `APIError` remains reachable with `errors.As`.

Tests cover body replay, attempt limits, each retryable and non-retryable status,
transport failures, Retry-After parsing (both forms, negative, fractional,
overflow) and limits, deadline-aware refusal to wait, cancellation during waits,
policy validation, jitter bounds, `WithoutRetries`, per-operation retry through
the public API, and every execute classification. Transport tests use an
injected wait function and need no real sleeping.

Not covered: no idempotency key or status endpoint is documented for execute, so
the SDK does not resolve unknown outcomes itself. Signature-status checking
belongs with the signing and confirmation adapters (next step).

Validation: `go test -race ./...`, `go vet ./...`, `gofmt -l`, and
`git diff --check` passed with Go 1.27.4. No live requests were sent.

## Signing adapters and Swap examples

Committed retry policies as `1fda6a1`. Added package `solanatx`, the only new-SDK
package that imports solana-go (already a module dependency through the legacy
packages, so no separate module is needed yet). The root `jupiter` package does
not import it.

- `Parse` decodes a base64 wire transaction. It reads the signature count with
  gagliardetto/binary and checks the message version byte itself before parsing,
  because solana-go v1.14 decodes every versioned message as version 0. Legacy and
  v0 are supported; other versions return `ErrUnsupportedVersion`. The message must
  re-encode byte-for-byte through solana-go, which also rejects trailing data, and
  the signature count must equal the required signers.
- `Sign` accepts standard-library `crypto.Signer` values with ed25519 keys, so
  `ed25519.PrivateKey`, HSMs, and remote signers fit. It signs the exact received
  message bytes (solana-go's own signing re-encodes the message) and writes only
  each signer's slot. Every signature is verified before any is written, so a
  failure leaves the transaction unchanged. Errors: `ErrSignerNotRequired`,
  `ErrInvalidSignature`, `ErrUnsupportedSigner`, plus wrapped signer failures.
- `RequiredSigners`, `MissingSigners`, `ID` (fee-payer signature; empty until the
  fee payer signs, which for gasless orders happens during execution), `Base64`,
  and `Bytes`.
- `Assemble` builds an unsigned v0 transaction from a build response using its
  blockhash and lookup tables, plus optional extra instructions. It rejects v1
  builds and missing or wrong-length blockhashes. `Instructions`,
  `ConvertInstruction`, and `AddressTables` support custom assembly.

Decision: build instructions are ordered compute budget, setup, swap, cleanup,
other, tip. The pinned docs do not place `otherInstructions`; this is recorded in
the `Instructions` documentation. Jupiter's compute budget instructions set only
the unit price, not the limit.

Examples: `examples/managed` (order, sign, execute) and `examples/build` (build,
assemble, sign, submit through `SOLANA_RPC_URL`, poll confirmation until the
blockhash expires). Both stop after signing unless `JUPITER_EXECUTE_SWAP=yes`.
`examples/quote` covers quote-only use and `WithoutMaintainerFee`. The examples
README documents signing, gasless transaction IDs, and unknown outcomes.

Tests: partial signing by two parties with message and existing-signature
preservation; cross-check with solana-go `VerifySignatures` and transaction ID;
atomic rejection for outsiders, non-verifying signatures, non-ed25519 keys, and
signer failures; unsupported versions 1 and 127; malformed base64, empty,
missing message, truncation, trailing data, and signature-count mismatch;
assembly order, lookup-table use, blockhash preservation, forced v0 without
lookups, and rejection of v1, missing/short blockhash, and malformed fields.

Not done: examples were compiled and vetted but not run against the live API or
chain. v1 transactions are rejected rather than supported.

Validation: `go test -race ./...`, `go vet ./...`, `gofmt -l`, and
`git diff --check` passed with Go 1.27.4.

## solana-go v1.26.0 upgrade and version 1 signing

Upgraded `github.com/gagliardetto/solana-go` from v1.14.0 to v1.26.0, the latest
release (`go list -m -versions`). `go mod tidy` also raised indirect dependencies
and testify (v1.12.1) as required. gagliardetto/binary v0.8.0 remains current.
The `_examples` legacy modules depend on the published jupiter-go v0.2.2 and were
not changed.

v1.26 parses and encodes SIMD-0385 version 1 transactions, whose layout is the
message followed by exactly the required signatures with no count. `solanatx`
now locates both layouts itself before parsing: a high-bit leading byte is a
version prefix (a counted layout would need at least 128 signatures, which no
transaction size limit allows), 0x81 selects version 1, and other prefixes return
`ErrUnsupportedVersion`. A version 1 message inside a counted layout is also
unsupported. Byte-exact re-encoding and signature-count checks apply to all
layouts. `Version1` was added.

`Assemble` now selects version 0 with `solana.TransactionMessageVersion` instead
of calling `Message.SetVersion`, whose new error return was being discarded. It
relies on `MarshalBinary` to write empty signature slots. Version 1 builds remain
rejected: their compute budget belongs in the message config, and the build
response provides a unit price but no compute unit limit.

Legacy package fixes required by the upgrade: `solana.subscriber.Pull` uses the
new `SignatureSubscription.Recv` (the `Response` channel was removed) and wraps
the context error; a legacy test now matches `base58.ErrInvalidChar` with
`errors.Is` instead of the old library's error text.

Tests: version 1 partial signing with message and existing-signature
preservation, solana-go signature verification and ID cross-check; unsupported
prefixes 0x82 and 0xff and version 1/127 messages in a counted layout; version 1
missing signature bytes, trailing data, and header-only input.

Validation: `go test -race ./...`, `go vet ./...`, `gofmt -l`, and
`git diff --check` passed with Go 1.27.4.

## Tokens and Price clients

Committed the solana-go upgrade as `894971d` and signing adapters as `2319b9e`.
Pinned the Tokens V2 discovery and Price V3 references in `openapi/tokens-v2`
and `openapi/price-v3` on 2026-11-11; their SHA-256 hashes match the inventory
manifest, which now marks the five operations implemented.

`client.Tokens` (package `tokens`):

- `Search(query)` for a symbol, name, or single mint. Comma-containing queries
  are rejected and pointed at `Lookup`, because the API treats commas as a
  mint-only list.
- `Lookup(mints)` for 1–100 validated mint addresses, sent as that list.
- `Tag(tag)`, `Category(CategoryRequest)` with optional limit 1–100, and
  `Recent()`.
- Documented tags, categories, and intervals are constants, but other values are
  accepted (format-checked, path-escaped for path segments), matching the router
  decision, so upstream additions keep working.
- Verification operations are deliberately absent: they craft and pay for
  transactions and need their own workflow.

`client.Price` (package `price`): `Get(mints)` for 1–50 validated mints. The
response maps mint to `Price` and reports requested mints Jupiter omitted (no
reliable price) in `Missing`, in request order.

Types: numbers are `json.Number` to keep precision; an empty value means null or
absent, explicit zero is preserved. Nullable booleans and `blockId` are pointers;
`Audit.IsSus` must be checked for presence. Timestamps stay as upstream strings
to avoid decode failures on format drift. All operations use the shared
`transport.ReadOnlyRetry` policy (moved from the Swap package, unchanged values).
Comma-list validation moved from `swap` to `internal/listparam` and is shared.
Both `ErrInvalidAddress` re-exports are the same value as `swap.ErrInvalidAddress`.

Example: `examples/tokens` (search and prices; nothing signed).

Tests: decoding of precision, null versus zero numbers, audit presence, and
optional objects; exact path and query encoding for every operation, including
100 lookup mints, the maximum category limit, upstream-added tags, and escaped
path segments; rejection before transport of empty or comma queries, 0 or 101
mints, malformed mints (`ErrInvalidAddress`), separator tags, missing interval,
and limits 0 and 101; read-only retry; price missing-mint reporting, null
`blockId` and change, 50-mint maximum, and 0, 51, and malformed inputs.

Validation: `go test -race ./...`, `go vet ./...`, `gofmt -l`, and
`git diff --check` passed with Go 1.27.4. No live requests were sent.

## Trigger V2 sessions, vault, and deposits

Committed shared helpers as `b5221bb` and Tokens/Price as `50830fd`. Fetched all 18
Trigger references (hashes match the manifest) and pinned the 8 implemented ones
plus the authentication, error, and deposit guides in `openapi/trigger-v2`.

Design (package `trigger`, `client.Trigger`):

- `Client` holds no wallet identity. Authenticated operations run on a
  `Session`, which belongs to one wallet and is safe for concurrent use.
- `Login(ctx, crypto.Signer, SessionOptions)` runs the message challenge with
  any ed25519 signer, verifying the signature locally before sending it.
  Hardware wallets use `RequestChallenge` (`ChallengeTransaction`), sign with
  `solanatx`, then `Verify`. The trigger package does not import solana-go.
- Only `access_refresh` is requested and accepted; the deprecated 24-hour
  `access_only` mode returns `ErrUnexpectedTokenType`.
- Sessions refresh one minute before expiry and once after a 401 (a rejected
  request was not processed, so it is resent once regardless of retry policy).
  Refreshes run under the session mutex: presenting a rotated refresh token
  revokes the session upstream, so concurrent refreshes must not happen. A
  request whose token was already replaced by a concurrent refresh reuses the
  new token instead of refreshing again.
- `SessionOptions.OnRefresh` receives each rotated pair for persistence; `Resume`
  restores a stored pair. A persistence failure returns `ErrTokenPersistence`
  but the session keeps the new tokens. A rejected refresh returns
  `ErrSessionExpired`. `Logout`/`LogoutAll` close the session
  (`ErrSessionClosed` afterwards).
- `Vault` (`ErrNoVault` on 404), `RegisterVault`, and `EnsureVault`. The
  register reference says repeat calls return the vault, while the error guide
  documents 409; `EnsureVault` handles both by re-reading the vault on 409.
- `CraftDeposit` always uses the session wallet as `userAddress`, validates mints
  and the raw amount, requires a sub-type for price deposits, forbids it for DCA,
  and allows `JLMint` only for DCA. Order types and sub-types are validated
  strictly because they determine required parameters.

Retry policies: verify (single-use challenge) and vault registration (creates
state) make one attempt; challenge, refresh (documented grace window), logout,
logout-all, and deposit crafting use a short repeatable policy; vault lookup
uses `transport.ReadOnlyRetry`. `transport.Request.Header` was added for the
bearer token. Raw amount parsing moved to `internal/amount` for Swap and Trigger.

Example: `examples/trigger` (login, vault, craft deposit without submitting,
logout; registers a vault only with `JUPITER_REGISTER_VAULT=yes`).

Tests use an in-memory server that verifies challenge signatures, rotates refresh
tokens, revokes reused ones, and enforces bearer tokens: login with
access_refresh, rejection of access_only, transaction-challenge verification,
`EnsureVault` (missing, existing, concurrent 409), proactive refresh with
`OnRefresh`, refresh once after 401, eight concurrent requests sharing one
refresh, expired session, persistence failure keeping rotated tokens, logout and
logout-all closing the session, deposit wallet and amount canonicalization,
retry choices, and 15 invalid inputs rejected before transport.

Validation: `go test -race ./...`, `go vet ./...`, `gofmt -l`, and
`git diff --check` passed with Go 1.27.4. No live requests were sent.

## Trigger V2 orders

Committed the helpers as `aafb852` and sessions as `b0ca797`. All 18 Trigger
operations are now implemented; the remaining 10 references and the order guides
are pinned in `openapi/trigger-v2`.

Session methods:

- `CreatePriceOrder` (single, oco, otoco, trailing stop). Per-variant rules from
  the create-order guide: single needs a condition and exactly one of trigger
  price or trailing distance (50–9000 bps); a trailing order triggers on the input
  mint for `below` and the output mint for `above`; oco takes no trigger fields
  and needs take-profit above stop-loss; otoco needs condition, trigger price,
  and the pair; single takes no pair fields; only single can trail. Slippage
  ≤ 10000 bps, a future `ExpiresAt` (sent as Unix milliseconds), distinct mints,
  a base64 signed deposit, and a raw amount are required. `userPubkey` is always
  the session wallet.
- `UpdatePriceOrder` validates the same field rules and rejects empty updates.
- `CancelPriceOrder`/`ConfirmCancelPriceOrder`, `CreateDCAOrder`,
  `CancelDCAOrder`/`ConfirmCancelDCAOrder`, `PriceOrderHistory`,
  `DCAOrderHistory`, and `DCAOrder`. Order IDs are path-escaped.
- DCA: at least 2 rounds, a whole-second interval from 1 minute to 365 days, a
  start at most 30 days ahead (RFC 3339), price bounds and trigger mint only for
  price-conditional (trigger mint required there, minimum below maximum), and
  Earn While You Wait only time-based with both the flag and JL mint. A price
  band without bounds is not rejected because the docs do not require one.

USD prices are `json.Number`, validated as positive JSON numbers (`json.Valid`
plus `big.Rat`, which rejects fractions, hexadecimal, and infinities) and sent
unchanged. History filters (state, sort, direction) accept upstream additions,
format-checked like tags; limit is 1–100. Price-order timestamps are Unix
milliseconds (`int64`, pointers when nullable); DCA timestamps are strings.

Submission safety: order creation and cancellation confirmation submit signed
transactions, make one attempt, and wrap `ErrOutcomeUnknown` when the
transaction may have landed (`transport.MayHaveBeenProcessed`: anything but an
HTTP status below 500). Swap execute now uses the same helper. The create-order
response mentions "idempotent retries" while the error guide lists duplicate
deposits as 400, so automatic retries are not assumed. Session errors before
sending, and an already ended context, are reported without the marker. Update
(absolute values) and cancel step one (crafts a withdrawal; DCA cancel is
documented idempotent) retry briefly; history uses `ReadOnlyRetry`.

Example: `examples/trigger` now signs the deposit with `solanatx`, creates a
single limit order only with `JUPITER_CREATE_ORDER=yes` and `TRIGGER_PRICE_USD`,
and lists active orders.

Tests: exact create bodies (precise USD numbers, canonical amounts, millisecond
expiry, session wallet, omitted unset fields) for single, trailing, oco, and
otoco; outcome classification for all four submission operations across 500,
503, 400, and 429 with single attempts; update sending only changed fields with
an escaped ID; price cancellation round trip; DCA creation fields and
cancellation refund; history query encoding and decoding of nulls, precision,
pagination, and events; 46 invalid inputs rejected before transport and never
marked unknown; DCA interval and Earn While You Wait boundaries accepted.

Validation: `go test -race ./...`, `go vet ./...`, `gofmt -l`, and
`git diff --check` passed with Go 1.27.4. No live requests were sent.

## errors.AsType adoption

Replaced all 17 `errors.As` calls in handwritten code (3 production helpers,
14 test assertions) with Go 1.26's type-safe `errors.AsType`, available under the
module's `go 1.26.0` directive. Behavior is unchanged. The convention is recorded
in `AGENTS.md`. Validation: `go test -race ./...`, `go vet ./...`, `gofmt -l`,
and `git diff --check` passed.

## Lend (Earn and Borrow)

Committed the `errors.AsType` refactor as `b700a6e`. All 15 Lend references
(hashes match the manifest) and both API guides are pinned in `openapi/lend-v1`.
Both guides label the APIs beta; the package documents this, and the
compatibility boundary is left to the review.

`client.Lend.Earn`: `Deposit`, `Withdraw`, `Mint`, `Redeem` (unsigned
transactions), their `*Instructions` variants (one instruction each), `Tokens`,
`Positions(users)`, and `Earnings(user, positions)`. Amounts and shares are
positive raw integers; assets, signers, users, and positions are validated.
`client.Lend.Borrow`: `Vaults(VaultsRequest)`, `Positions(market, users)`,
`Operate`, and `OperateInstructions`. Operate amounts follow the documented
`^-?[0-9]+$` pattern and the signed 128-bit range, are canonicalized, and at least
one must be non-zero. `WithdrawAllCollateral`/`RepayAllDebt` name the MIN_I128
sentinel; `NewPosition` is position 0. Markets (`main`, `ethena` constants) pass
through like other enum filters; an empty market uses the default. The optional
`rpcUrl` must be absolute and is documented as being sent to Jupiter.

Schema conflicts handled explicitly: numeric fields declared as strings appear as
numbers in examples (and some are declared string-or-integer), so all numeric
quantities are `json.Number`, which accepts both and keeps exact text. Earn
`chain_id` is declared a string but the example is `0`; `lend.ChainID` accepts
both. Earnings is declared an object but the example is an array; both decode
into a slice.

Shared refactors: `instruction.Instruction`/`AccountMeta` (swap keeps aliases;
`solanatx.ConvertInstruction` takes the shared type, so Lend instructions convert
too) and `transport.RepeatableRetry`, now used by Swap order/build, the
repeatable Trigger operations, and Lend builders. Lend submits nothing, so there
are no unknown-outcome cases; reads use `ReadOnlyRetry`.

Consistency fix: Trigger operations now return their result (with metadata)
alongside HTTP errors, like Swap, Tokens, Price, and Lend; validation failures
still return nil.

Example: `examples/lend` (read-only; builds an unsigned deposit when
`WALLET_ADDRESS` is set).

Tests: request bodies and canonical quantities for all Earn builders; single
instruction decoding and conversion; mixed number forms, chain IDs, and both
earnings shapes; borrow vault decoding (string and integer durations, raw oracle
source type), market and RPC URL encoding, default market omission; operate body
with sentinel and canonical amounts, instruction responses with lookup table
addresses; builder retry; signed-range boundaries accepted; 18 invalid inputs
rejected before transport; metadata kept on Lend and Trigger failures.

Validation: `go test -race ./...`, `go vet ./...`, `gofmt -l`, and
`git diff --check` passed with Go 1.27.4. No live requests were sent.

## Prediction markets

Committed the instruction/retry refactor as `d5aa288`, the Trigger metadata fix
as `14d38cf`, and Lend as `f1b4929`. Fetched the Prediction OpenAPI spec linked
from the documentation index and all 27 reference pages (hashes match the
manifest), and pinned them with the trading guides in `openapi/prediction-v1`.
The spec contains `GET /forecast`, which has no reference page and was missing
from the inventory; it is implemented and listed as spec-only. The API is beta
and geo-restricted (US, South Korea); the package documents both.

`client.Prediction` covers discovery (`Events`, `SearchEvents`, `Event`,
`SuggestedEvents`, `EventScores`, `EventScore`, `EventMarkets`, `EventMarket`,
`Market`, `Orderbook`, `TradingStatus`, `Forecast`), trading (`CreateOrder`,
`ClosePosition`, `CloseAllPositions`, `ClaimPosition`, `Execute`), and account
reads (`Orders`, `Order`, `OrderStatus`, `Positions`, `Position`, `History`,
`Profile`, `PnLHistory`, `Trades`, `Leaderboards`, `VaultInfo`).

Decisions:

- Builders share one `TransactionBuild` type; `Order` is set for order and sell
  builds and `Position` for claims, which also covers the mixed close-all array.
  `Execution.Context` is `json.RawMessage` and is sent back byte-for-byte.
- `Execute` makes one attempt, requires an explicit `Success` status
  (`*ExecutionError` otherwise), and wraps `ErrOutcomeUnknown` via
  `transport.MayHaveBeenProcessed`; an already ended context sends nothing.
  Builders use `RepeatableRetry`; reads use `ReadOnlyRetry`.
- Orders: buys need a market, a positive raw deposit, and an optional valid mint
  (`MintJupUSD`, `MintUSDC`); sells need a position and exactly one quantity
  (whole, micro, or a positive decimal validated like USD prices). The 5 USD
  minimum is left to the API.
- Field types follow the schema: declared strings (micro USD, i128 strings) stay
  strings, declared numbers are `json.Number`, nullable integers are pointers,
  `resolveAt` uses `StringOrNumber`, and null scores and orderbooks are nil.
- Filters (provider, category, filter, sort, tags, intervals, periods, metrics)
  pass upstream additions through with format checks; documented values are
  constants. Search queries are 1–200 characters (runes); pages reject an end
  before the start; IDs are path-escaped; wallets, positions, and orders are
  validated addresses.

Test helpers: the recording server moved to `servers_test.go` as `pathServer`
(escaped paths) with `statusServer`.

Example: `examples/prediction` (read-only trending events and prices).

Tests: discovery encoding and decoding (mixed `resolveAt`, null pricing, null
score and orderbook, escaped IDs, joined event IDs); buy and sell bodies with
canonical quantities and no cross-mode fields; DELETE close bodies; mixed
close-all results; claims; Execute success, failed and missing status, 500 as
unknown, 400 as known, single attempts, and opaque context; account reads with
query encoding and nullable fields; 25 invalid inputs rejected before transport
and never marked unknown.

Validation: `go test -race ./...`, `go vet ./...`, `gofmt -l`, and
`git diff --check` passed with Go 1.27.4. No live requests were sent.

## Send

Committed the shared test servers as `33bd08f` and Prediction as `b376116`.
The 4 Send references (hashes match the manifest) and its guides are pinned in
`openapi/send-v1`. Send is beta upstream.

Invite codes (`send.InviteCode`), implemented from the invite-code and clawback
guides with the standard library plus established parsers:

- `NewInviteCode` takes 13 bytes from `crypto/rand`, base58-encodes them, and keeps
  the first 12 characters, as documented. `ParseInviteCode` requires 12 base58
  characters.
- `PrivateKey` uses `SHA-256("invite:" + code)` as the ed25519 seed;
  `ed25519.NewKeyFromSeed` produces the documented `[seed || public key]` layout.
  It implements `crypto.Signer`, so `solanatx` signs with it.
- `PDA`/`InvitePDA` derive the invite account for seeds `"invite"` and the signer
  under program `inv1tEtSwRMtM44tbvJGNiTxMvDfPVnX9StyqXfDfks`, trying bumps 255 to 1
  and using `filippo.io/edwards25519` (new direct dependency, v1.2.0) to reject
  on-curve candidates. The package stays free of solana-go.
- `InviteCode` formats as a redacted placeholder with `%v`, `%s`, `%+v`, and
  `%#v`, including inside structs; `Reveal` returns the code deliberately.

`client.Send`: `CraftSend` (canonical amount, optional mint; wrapped SOL by
default), `CraftClawback`, `PendingInvites`, and `InviteHistory` (page from 1;
zero uses the default). Builders use `RepeatableRetry`, reads `ReadOnlyRetry`.
Schema conflicts: `confirmed` is declared an integer but returned as a boolean in
examples (`send.Confirmed` accepts both and null); `expiry` is described as a
Unix timestamp but shown as RFC 3339, so it stays a string.

Example: `examples/send` (generates a code and builds an unsigned send).

Tests: 200 generated codes parse and are unique; derivation matches the
documented seed and key layout; PDAs match solana-go `FindProgramAddress` for
200 random invites; redaction in every format verb; malformed codes rejected; a
built send transaction signed by both sender and invite key with `solanatx`;
clawback body, invite paging, and all `confirmed` forms; invalid inputs rejected
before transport.

Validation: `go test -race ./...`, `go vet ./...`, `gofmt -l`, and
`git diff --check` passed with Go 1.27.4. No live requests were sent.

## Studio

Committed Send as `2ca80f7`. Fetched the Studio OpenAPI spec and its 5 reference
pages (hashes match the manifest) and pinned them with the guides in
`openapi/studio-v1`. Studio is beta upstream.

Transport additions, kept generic:

- `Request.Payload` sends pre-encoded bytes with their content type (used for
  multipart forms); it is exclusive with the JSON `Body`.
- `Request.ExternalURL` targets an absolute HTTPS URL outside the API, such as a
  presigned upload. The API key, Accept header, and request headers are never
  sent there, user info is rejected, and the URL's query (the presigned
  signature) is used unchanged. Path, query, and headers cannot be combined with
  it. The default client still refuses redirects.
- A nil result ignores the response body.
- `Do` now prepares the request once (URL, body, content type, credentials) and
  replays it per attempt.

`client.Studio`: `CreatePool` (unsigned launch transaction plus presigned URLs),
`UploadImage` and `UploadMetadata` (presigned PUTs; overwriting is harmless, so
`RepeatableRetry`), `Submit` (multipart form built with `mime/multipart`, with an
explicit content type for the header image; one attempt; `ErrOutcomeUnknown` for
ambiguous failures; an explicit `success: false` is a known failure),
`PoolAddresses`, `Fee` (a POST that only reads, so `ReadOnlyRetry`), and
`ClaimFeeTransaction`. Requests validate addresses, positive market caps with
migration above initial, required non-negative vesting numbers
(`NoLockedVesting` helper), and required fields; documented enum values are
constants and other values pass through to the API, as for other products.

Example: `examples/studio` (builds and signs; launches only with
`JUPITER_LAUNCH_TOKEN=yes`).

Tests: transport guarantees for external URLs (no key or headers, exact
presigned query, HTTPS and user-info checks, path/header conflicts), payload
content types, and body/payload exclusivity; the full launch flow over TLS
checking exact curve numbers, omitted LP lock, upload methods, bodies, content
types, the absence of the API key on uploads and its presence on API calls, and
the parsed multipart form; submit outcomes (500 unknown; 400 and explicit
failure known; single attempts); pool and fee reads with nulls; fee read retry;
14 invalid inputs rejected before transport.

Validation: `go test -race ./...`, `go vet ./...`, `gofmt -l`, and
`git diff --check` passed with Go 1.27.4. No live requests were sent.
