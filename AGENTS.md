# Repository working conventions

## Development workflow

- Develop the new v1 SDK on the `v1` branch.
- Follow `docs/v1-plan.md` and keep `docs/development-status.md` current with
  completed work, validation results, decisions, limitations, and next steps so
  another session can resume the work.
- Provide clear, runnable examples for implemented workflows.

## Go readability and design

- Use idiomatic modern Go, separate component responsibilities, and favor code
  that is straightforward to read and maintain.
- Add blank lines after error checks, conditional blocks, and loops, and between
  distinct pieces of logic. Keep a call and its error check together.
- Expand compressed functions and compound one-line logic into readable blocks.
- When a function signature becomes long, put each parameter on its own line,
  with the opening parenthesis on the declaration line and the closing parenthesis
  on a separate line. Expand grouped parameters so each row names one parameter.
  Keep short signatures compact when they remain easy to read.
- Keep handwritten Go lines within roughly 110 columns. Split long conditions,
  calls, literals, signatures, and comments at logical boundaries; use named
  intermediate values or switches when that improves clarity. Keep error checks
  adjacent to their calls. Do not rewrite pinned upstream snapshots to enforce
  handwritten-code style.
- Use named constants for protocol string values that are compared, selected, or
  validated, such as modes, execution statuses, and priority levels. Reuse those
  constants in examples and tests. Ordinary empty-string checks and independent
  wire fixtures can remain literal.
- Separate standard-library imports from project and third-party imports, and
  run `gofmt` on changed Go files.
- Add comments only when they explain a useful contract, constraint, or reason
  that the code itself does not make clear. Avoid comments that restate the code.

## Errors and implementation choices

- Never add technical debt, workarounds, or hacks, not even temporarily. When a
  change does not fit the existing structure, rework the existing code first so
  that it does, then make the change.
- Never leave returned errors unchecked or discard them with `_`. This applies
  to production code, examples, tests, decoding, constructors, output writes,
  and resource cleanup.
- Handle cleanup errors explicitly, preserving the original failure when both
  the operation and cleanup fail.
- Preserve useful error context and causes so callers can use `errors.Is` and
  `errors.As`.
- Avoid magic numbers and clever or hacky shortcuts. Give policy values
  descriptive names and explicit units, and prefer standard-library mechanisms
  over manual tricks such as sentinel-byte arithmetic for response limits.
- Verify changes with appropriate tests and checks; record the results honestly.
