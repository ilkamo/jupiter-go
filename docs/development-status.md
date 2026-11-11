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
- No signing adapter is supplied yet. Execute accepts caller-signed base64 bytes
  unchanged. It does not parse transaction messages or claim wallet verification.
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
3. Add fixture tests for build and managed execution. Retry policies and
   request validation are complete for Swap.
4. Add signing adapters and examples for managed execution and custom builds,
   preserving blockhashes and existing signatures. Test transaction versions
   explicitly before declaring adapter support.
5. Implement Tokens and Price, then Trigger V2 authentication/vaults/orders/DCA,
   Lend, Prediction, Send, Studio, and transaction submission. Audit the current
   documentation index for additional APIs. These clients are not implemented.
6. Complete pinned schema tooling, generation checks, fuzz tests, examples, CI,
   migration documentation, and compatibility review from `v1-plan.md`.
7. Resolve the maintainer fee recipient/default and beta API compatibility before
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
