# Jupiter-go v1 implementation plan

Status: implementation in progress. See [development status](development-status.md)
for completed components, validation, decisions, and remaining work.

Prepared on 2026-11-11 from the repository review and discussion. API details
below reflect the documentation reviewed during planning and must be verified
against current product documentation before implementation.

## Objective

Release an idiomatic, maintainable Go SDK covering Jupiter's current public REST
APIs, with functional options, typed requests and responses, transparent optional
swap monetization, and a clear migration path from the existing library.

The SDK version and upstream API versions are independent: library v1 will
initially target Swap V2, Tokens V2, Price V3, Trigger V2, and other documented
current APIs. Do not add a `/v1` suffix to the Go module path.

## Design direction

- Expose a root package named `jupiter`, imported from
  `github.com/ilkamo/jupiter-go`, with `NewClient(opts ...Option) (*Client, error)`.
- Provide product services such as `client.Swap`, `client.Tokens`, and
  `client.Trigger`, with product-specific request and response types.
- Use functional options for configuration and request structs for endpoint
  parameters. Validate the complete configuration during construction.
- Keep configuration immutable after construction and clients safe for concurrent
  use. Constructors do not perform network calls.
- Accept `context.Context` as the first argument to network operations.
- Return concrete clients; use small interfaces at dependency boundaries rather
  than exporting an interface for every service.
- Share HTTP transport, authentication plumbing, response handling, and retry
  machinery. Keep product behavior and authentication requirements explicit.
- Keep generated wire bindings internal. Handwritten public types and methods
  define the compatibility contract; avoid exposing generator-specific unions or
  method names.
- Keep signing and RPC dependencies outside the core HTTP implementation.

Usage:

```go
client, err := jupiter.NewClient(
    jupiter.WithAPIKey(apiKey),
    jupiter.WithHTTPClient(httpClient),
    jupiter.WithTimeout(15*time.Second),
)
if err != nil {
    return err
}

order, err := client.Swap.Order(ctx, swap.OrderRequest{
    InputMint:  inputMint,
    OutputMint: outputMint,
    Amount:    "100000000",
    Taker:     walletAddress,
})
```

Proposed layout:

```text
client.go
options.go
errors.go
swap/
tokens/
price/
trigger/
lend/
prediction/
send/
studio/
transaction/
internal/transport/
internal/generated/
openapi/
examples/
docs/
```

Resolve shared transport and error ownership without import cycles: product
packages must not depend on the root package that constructs them. Avoid creating
generic utility packages without a concrete need.

## Coverage contract

The endpoint discovery inventory is recorded in [api-inventory.md](api-inventory.md).
Its source manifest covers all 82 currently indexed API-reference operations.
Full per-operation contract review is required before implementation. For every
endpoint, record the product, upstream version, method/path, schema source and
revision, authentication, pagination, response semantics, idempotency, upstream
stability, implementation status, and test status.

| Area | Intended coverage |
| --- | --- |
| Swap V2 | Order, execute, build, program labels and all other documented operations found in the inventory |
| Tokens V2 | Search, discovery, metadata and remaining documented operations |
| Price V3 | Price retrieval and documented limits |
| Trigger V2 | Wallet authentication, vaults, deposits, price orders, DCA, management and history |
| Lend | Documented Earn and Borrow REST operations, including transaction and instruction responses |
| Prediction | Documented REST operations, with upstream beta status explicit |
| Send | Documented transfer and invite REST operations |
| Studio | Documented token creation and fee REST operations |
| Transaction submission | Documented submission operations and their relationship to build flows |
| Perps and Lock | Separate on-chain integration roadmap; not counted as REST coverage |

Audit the current documentation index for additional public products before
freezing scope. Do not treat this table as an exhaustive endpoint inventory or
advertise all Jupiter products as supported when on-chain integrations are absent.

Trigger V2 unifies limit orders and DCA. The reviewed documentation describes
Recurring V1 as unmaintained and Trigger V1 as receiving critical fixes only.
Trigger V2 also introduces wallet JWT authentication and custodial vaults; the
migration documentation must explain the changed workflow and custody model.

For beta upstream products, decide the compatibility boundary before exposing
public types. An upstream beta label alone does not waive the SDK's v1 semantic
versioning obligations.

## Fees (decided)

The SDK charges no fee and has no maintainer referral. `WithReferral(account,
bps)` lets integrators add their own Jupiter referral (50–255 bps) to managed
`/order` swaps; it is validated, cannot be configured twice, and does not apply to
`/build`, which takes explicit `PlatformFeeBps` and `FeeAccount` per request.
Referral and fee token accounts are provisioned by the integrator; the SDK never
creates them. Upstream fee fields are preserved so applications can show actual
fees before signing.

## Engineering requirements

### Transport and authentication

- Inject an HTTP dependency with `Do(*http.Request) (*http.Response, error)`.
- Support explicit API keys, base URLs, and timeouts without mutating a
  caller-owned HTTP client. Decide whether service-specific URL overrides are
  needed when completing the inventory.
- Respect cancellation and deadlines across requests, backoff, and polling.
- Close response bodies, bound error payloads, and preserve useful response
  metadata such as request IDs and rate-limit information.
- Prevent credential forwarding to unintended hosts or redirects.
- Keep wallet JWT sessions scoped to the relevant wallet and product; do not
  store a mutable global wallet identity on a shared multi-user client.
- Verify keyless access per product instead of enforcing an outdated global
  requirement from an integration example.
- Make logging optional and redact API keys, authorization tokens, signatures
  where sensitive, and transaction payloads. Prefer standard `log/slog` if a
  logging option is needed; allow transport wrappers for instrumentation.

### Types and errors

- Represent raw token amounts exactly with validated integer strings or suitably
  bounded integers. Preserve decimal precision for prices and monetary values;
  do not route them through `float64` calculations.
- Distinguish omitted, null, and explicit zero values where the API does.
- Allow upstream enum additions where possible without unnecessary decoding
  failures, while validating values the SDK sends.
- Distinguish transport errors, HTTP/API errors, and execution failures contained
  in successful HTTP responses. Preserve upstream codes and identifiers.
- Support `errors.Is`/`errors.As`, wrap causes, and preserve context cancellation.
- Make submitted, confirmed, failed, and uncertain execution states explicit.

### Retry and transaction safety

- Use bounded retries with jitter and `Retry-After` support for eligible calls.
- Classify operations by semantics, not only HTTP verb: an order GET can produce
  time-sensitive transaction data.
- Retry writes only with documented idempotency guarantees and the same required
  identifiers/payload. Do not automatically recreate an order after an ambiguous
  execution response.
- Preserve managed transaction messages, blockhashes, and existing signatures.
  The current `solana.SendTransactionOnChain` replaces the blockhash and must not
  be reused unchanged for the managed order flow.
- Keep order, signing, execution, and confirmation separate. Support partial
  signing and reject unsupported transaction versions clearly in adapters.
- Never assume a submitted signature proves successful execution.

### Generation and dependencies

- Pin upstream schema snapshots and generator versions; record provenance.
- Maintain reviewed overlays for schema corrections rather than editing
  generated files. Use handwritten wire code where a reliable schema is absent.
- Make generation deterministic and verify a clean regeneration in CI.
- Keep the public API independent of generator output and upstream renames.
- Choose and document the supported Go versions before changing the existing
  `go.mod` baseline; test the declared support range.
- Decide whether Solana adapters need a separate module for dependency isolation.
  Start with the simplest structure that meets that requirement.

## Delivery milestones

### 1. Contract and foundation

- [x] Verify API inventory and schema sources.
- [x] Settle fee defaults and public fee-mode semantics.
- [ ] Review constructor, service, request, response, and error examples.
- [x] Decide beta API compatibility and legacy package strategy.
- [x] Implement transport, options, validation, authentication boundaries, and
      error handling.

Acceptance: examples establish the intended public API; `httptest` verifies
configuration, encoding, cancellation, response handling, authentication, and
retry eligibility without live credentials.

### 2. Complete Swap V2 workflow

- [x] Implement order, execute, build, and program-label operations.
- [x] Implement fee modes and path-specific fee behavior.
- [x] Provide examples for quote-only, managed execution with external signing,
      and custom builds.
- [x] Establish signing adapter boundaries and transaction version support.

Acceptance: fixture-based tests cover successful and failed executions, partial
signing/message preservation, fee opt-out and replacement, missing referral fee
application, unsupported versions, and ambiguous submission outcomes. No tests
perform real swaps by default.

### 3. Complete REST coverage

- [x] Add Tokens and Price.
- [x] Add Trigger V2, including authenticated wallet sessions and DCA.
- [x] Add Lend.
- [x] Add Prediction, Send, Studio, and transaction submission.
- [x] Add other public REST products identified by the approved inventory.
- [ ] Complete pagination, error fixtures, and examples for each product.

Acceptance: every in-scope endpoint has a typed public method, documented auth
and limitations, and meaningful request/response contract tests. No placeholder
services count as supported APIs.

### 4. Stabilize and release v1

- [ ] Review exported symbols for consistency and compatibility.
- [x] Write a migration guide from the existing generated API and Solana helpers.
- [x] Set the maintenance policy for existing v0 releases and legacy endpoints;
      do not silently redirect legacy calls to incompatible workflows.
- [ ] Complete README, Go documentation, product examples, and fee disclosure.
- [ ] Run formatting, static analysis, race tests, targeted fuzzing, generation
      checks, and example compilation in CI.
- [ ] Make live smoke tests opt-in and avoid funded transactions by default.
- [ ] Publish release candidates, resolve feedback, then tag v1.0.0.

Acceptance: the inventory is complete, migration and fee behavior are documented,
required checks pass, and public compatibility decisions are recorded.

## Current state

Decisions: no SDK fee; Lend, Prediction, Send, and Studio ship as beta; the legacy
`jupiter` and `solana` packages are removed and v0 releases receive no further
maintenance. Remaining work is tracked in [handoff.md](handoff.md).

## References

- [Jupiter API overview](https://developers.jup.ag/docs/get-started#apis)
- [Documentation index](https://developers.jup.ag/docs/llms.txt)
- [Swap overview](https://developers.jup.ag/docs/swap/index.md)
- [Order, execution, and referral fees](https://developers.jup.ag/docs/swap/order-and-execute.md)
- [Build API](https://developers.jup.ag/docs/api-reference/swap/build.md)
- [Trigger V2 and migration differences](https://developers.jup.ag/docs/trigger/index.md)
- [Lend](https://developers.jup.ag/docs/lend/index.md)
- [Prediction](https://developers.jup.ag/docs/prediction/index.md)
- [Jupiter agent skills](https://developers.jup.ag/docs/ai/skills)
- [Integration skill source](https://github.com/jup-ag/agent-skills/blob/main/skills/integrating-jupiter/SKILL.md)
- [Preferred functional-options example](https://github.com/ilkamo/ethparser-go/blob/main/internal/jsonrpc/options.go)

Skills are navigation aids, not the wire contract. During planning, the reviewed
integration skill still routed DCA through Recurring V1 while current Trigger
documentation recommended V2. Resolve discrepancies using current product docs
and schemas, and record the decision in the endpoint inventory.
