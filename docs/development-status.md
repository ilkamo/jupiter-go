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
- Requests currently make one attempt. This avoids unsafe write retries; bounded
  retries for semantically eligible operations remain to be implemented.
- A custom HTTP implementation is responsible for its own redirect policy. The
  constructor documents this; caller-owned HTTP clients are never mutated.
- No signing adapter is supplied yet. Execute accepts caller-signed base64 bytes
  unchanged. It does not parse transaction messages or claim wallet verification.
- The public API is provisional until the compatibility review. Response metadata,
  complete optional fields, and input address validation need further review.
- The repository's Go 1.26 baseline is unchanged. Validation uses the installed
  Go 1.27.4 toolchain; support-range CI remains pending.

## Resume here

1. Run `go test -race ./...` and `go vet ./...`; inspect this document's validation
   section before repeating already completed checks.
2. Finish the full endpoint inventory, including auth, pagination, idempotency,
   schema provenance, and test status. Do not infer total coverage from the four
   implemented Swap operations.
3. Complete Swap response-field parity, success metadata, bounded retry policy,
   amount/address validation, and fixture tests for build and managed execution.
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
