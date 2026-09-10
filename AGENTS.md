# Unswell development

Use American English. Keep runtime messages and user documentation in English.

The CLI calls the public engine. Rules, scoring, and policy belong to the library;
reporters consume a completed result. Library packages must not read the working
directory or environment, print, exit, start processes, or use the network.

Use original UTF-8 byte spans with an explicit source map. Preserve negation,
numbers, versions, identifiers, and protected code boundaries. A score is an index,
never a probability. Missing capabilities and operational errors cannot pass a gate.

Use standard testing and quicktest imported as `qt`; do not use testify. Maintain
package and exported API comments. Keep the public package ledger current.
Do not add blanket lint exclusions or weaken limits for an individual algorithm.

Write blackbox tests by default, in `package <name>_test`, through the public API.
Keep their filenames as `*_test.go`. Do not export production internals just to
make them accessible to tests.

Whitebox tests are exceptions. Each such file must use `*_internal_test.go` and
include a `// White-box tests: ...` comment after the `package` clause and before
imports or declarations. Explain the specific internal behavior under test and
why the public API cannot provide the required evidence. A generic testing
statement or shared test helper is not a justification. For example, in
`objective_internal_test.go`:

```go
package model

// White-box tests: Compare private gradients and Hessians with finite differences;
// the public fitting API exposes only the fitted model, not these derivatives.
```

Repository checks enforce the package convention, exception filename, and the
presence and placement of the explanation across all Go modules. Review must
assess whether the explanation warrants whitebox access.

Run `make check` before publishing. It measures coverage in place of the plain test
run, so the module suite executes once and the area targets in
[`docs/validation.md`](docs/validation.md) gate the result. Run `make race` for
concurrency changes and `make fuzz` when changing parsers, decoders, or other
input handling. `scripts/fuzz.sh` discovers targets with `go test -list`, so a new
`Fuzz` function joins the campaign without editing a list.

Backward compatibility is not required during alpha development. Change APIs and
artifact formats when needed, update their documentation and consumers, and reject
unsupported formats explicitly. Do not add compatibility shims or maintain API
snapshots solely to preserve earlier alpha releases. Current-contract validation,
source integrity, and library boundaries remain required.

Every repository policy check needs a negative test proving it rejects a violation.
Document alpha limitations and preserve later requirements in the roadmap.

Docker work must use an explicit remote context according to the user's global
instructions. Remove only resources created for this task.
