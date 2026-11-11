# v1 release handoff

Start here. This is the to-do list for finishing the v1 release on branch `v1`.
Read `AGENTS.md` for conventions; `docs/development-status.md` is the detailed
history of every step (decisions, tests, validation).

## Decisions already made (owner, 2026-11-11)

- **No SDK fee.** The maintainer-fee concept was removed (`WithoutMaintainerFee`
  is gone). `WithReferral` only adds the integrator's own Jupiter referral to
  managed Swap orders; configuring it twice is an error.
- **Lend, Prediction, Send, and Studio ship as beta**, mirroring upstream. Their
  package docs say so; the README must say so too (see below).
- **The legacy library is dropped.** The old `jupiter/` (oapi-codegen client),
  `solana/` packages, and `_examples/` were deleted; oapi-codegen,
  shopspring/decimal, and testify are no longer dependencies.
- Module path stays `github.com/ilkamo/jupiter-go` (v0 to v1 needs no `/v1`).

## State

- Every operation in `docs/api-inventory.md` (82 indexed plus the spec-only
  Prediction `/forecast`) is implemented with contract tests.
- `go test -race ./...`, `go vet ./...`, and golangci-lint v2.12.2 (CI's version,
  including gofumpt) pass under Go 1.26.0 and 1.27.4. Run the linter with
  `GOTOOLCHAIN=go1.26.0` if your golangci-lint build cannot read newer export
  data. `make check` runs everything CI runs.
- No live requests or funded transactions have been made.

## Done in the release-preparation pass

- `README.md` rewritten (product table with beta labels, no-fee statement,
  options, managed swap with `solanatx`, error/retry/safety model).
- `examples/README.md` header updated; `CHANGELOG.md` added with the v1.0.0 entry
  and the v0 migration table.
- `docs/v1-plan.md` updated (fee decision, milestones); the summary sections of
  `docs/development-status.md` rewritten.
- Godoc examples in `example_test.go` (`ExampleNewClient` and
  `Example_sendInviteCode` run with checked output; quote, managed swap, and
  Trigger session examples compile only, because they need the network).
- Every exported function and method has a doc comment (revive's `exported` rule
  reports none for functions and methods). Exported constants and struct types
  were not documented one by one: `AGENTS.md` forbids comments that restate the
  code, so only those with a non-obvious contract carry comments.
- CI runs golangci-lint v2.12.2 and gofumpt on Go 1.26, and `make vet` plus
  `make test` on Go 1.26 and 1.27, for pull requests and pushes to `main` and `v1`.

## Remaining work, in order

1. **Funded smoke test (tier 3).** Free tiers are done (see "Live verification"
   in `development-status.md`): reads decode, and real transactions parse, sign,
   and simulate. Still needed, with a throwaway wallet holding about 0.05 SOL:
   one managed swap (`examples/managed`, `JUPITER_EXECUTE_SWAP=yes`, lower the
   amount to 0.001 SOL) and one build submitted through tx.jup.ag
   (`examples/build`, `JUPITER_EXECUTE_SWAP=yes`, `JUPITER_SUBMIT_VIA=tx`,
   `SOLANA_RPC_URL`). Also re-check the two Prediction routes that returned 404
   upstream (`pnl-history`, `events/suggested`).
2. **Re-audit the documentation index** (`https://developers.jup.ag/docs/llms.txt`)
   for operations added since the inventory (the Prediction spec already had one,
   `/forecast`), and compare each pinned file in `openapi/` with upstream
   (hashes are in `docs/api-reference-manifest.json`).
3. **Exported API review** before freezing v1: names, request/response shapes,
   and error types across products (the plan's "Review exported symbols" item).
   Beta packages may change later; stable ones may not without v2.
4. **Optional:** `solanatx.Assemble` support for version 1 builds (needs a
   compute-unit-limit policy, since builds return a unit price but no limit);
   fuzz tests for `solanatx.Parse`, `internal/address`, and amount parsing.
5. **Release:** open a PR from `v1` to `main`, let CI pass, merge, and tag
   `v1.0.0`; update the CHANGELOG date.
