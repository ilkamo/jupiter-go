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

## Remaining work, in order

1. **Rewrite `README.md`.** It still documents the deleted legacy API. Cover:
   install, `jupiter.NewClient` and options (`WithAPIKey`, `WithTimeout`,
   `WithHTTPClient`, `WithBaseURL`, `WithTransactionURL`, `WithoutRetries`,
   `WithReferral`), one short example per product, a **beta notice for Lend,
   Prediction, Send, and Studio**, the "no SDK fee" statement, signing with
   `solanatx`, error handling (`errors.Is`/`errors.AsType`, `ErrOutcomeUnknown`
   per product, `response.Metadata`), retry behaviour, and links to `examples/`.
   Keep `assets/jup-gopher.png` if desired.
2. **Update `examples/README.md`.** Remove the "maintainer fee" sentence at the
   top; state that the SDK charges no fee.
3. **Add `CHANGELOG.md`** with a v1.0.0 entry: full rewrite, legacy packages
   removed (breaking), product list, beta products, Go 1.26 minimum. Add a short
   migration section mapping old `jupiter.NewClientWithResponses` +
   `solana.NewClient` usage to `client.Swap.Order`/`Execute` + `solanatx`.
4. **Clean up `docs/v1-plan.md`.** Replace the "Swap monetization policy" section
   with the no-fee decision; tick completed milestones; mark the legacy package
   decision done. In `docs/development-status.md`, update the "Decisions and
   limits" and "Resume here" sections at the top to point to this file (the step
   log below them is history and can stay).
5. **Godoc polish.** Add `Example` test functions (`example_test.go`) for the
   root package and key flows (quote, managed swap with `solanatx`, Trigger
   login) so pkg.go.dev shows them. Check every exported identifier has a doc
   comment (`golangci-lint` with `revive`'s exported rule can help).
6. **CI.** `.github/workflows/workflow.yml` already runs golangci-lint v2.12.2 and
   `make test` on Go 1.26. Consider adding `go vet`, a Go 1.27 matrix entry, and
   running on pushes to `v1`.
7. **Live smoke test** with a funded wallet before tagging (quote, managed swap,
   build + `tx.jup.ag` submit). Record results in `development-status.md`.
8. **Re-audit the documentation index** (`https://developers.jup.ag/docs/llms.txt`)
   for operations added since the inventory; the Prediction spec already had one
   (`/forecast`). Also compare each pinned file in `openapi/` against upstream.
9. **Optional follow-ups:** `solanatx.Assemble` support for version 1 builds
   (needs a compute-unit-limit policy); fuzz tests for `solanatx.Parse`,
   `internal/address`, and amount parsing.
10. **Release:** merge `v1` into `main` via PR, tag `v1.0.0`.
