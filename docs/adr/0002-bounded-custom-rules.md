# ADR 0002: compile declarative rules into the existing engine

Status: accepted for the experimental alpha.

Issue [#10](https://github.com/stokaro/unswell/issues/10) requires team policies over
selected prose without executing code from YAML. Rule packs must compose with Go
extensions and preserve the public engine's offline, source-mapping, capability,
and resource contracts.

## Vale comparison

Reviewed Vale 3.20.0's
[sequence implementation](https://github.com/vale-cli/vale/blob/fdc4cc754f58d953f668586dc1891064a746e0a2/internal/check/sequence.go)
at commit `fdc4cc754f58d953f668586dc1891064a746e0a2`. Its positioned token matching,
negation, bounded skips, occurrence expansion, and target tokens address related
problems. Its constructor and runner also rely on Vale configuration, a regex
wrapper, model registration, and file/report types. Adapting them would require
replacing those boundaries and changing source-location and cost accounting.

We considered a small source adaptation and an independent matcher over Unswell's
existing token model. Choose the independent matcher. It memoizes token-position
states, bounds every gap, and applies the same candidate budget as other rules.
It emits original token segments through `rule.Emitter`. These choices keep the
behavior inspectable without adding a second configuration or reporting engine.

No Vale source or tests are copied. There is therefore no new derived-source
notice. Any later adaptation must retain the upstream copyright after the Go
package declaration, the full MIT notice, the exact revision and filenames, and
the changes made. Prose remains the existing English dependency; no new runtime
dependency is introduced.

## Contract and tradeoffs

Packs compile into immutable `rule.Rule` implementations. Rules and exceptions
share one bounded matcher tree. The CLI loads selected files; `ruleset.Load` and
the engine accept only bytes. Inline packs use the same compiler. Duplicate IDs
fail instead of replacing a builtin implicitly.

The first contract uses exact Penn tags and bounded gaps. It does not promise
compatibility with Vale styles, alternative POS models, UPOS conversion, arbitrary
token regexes, or repeated-token expansion syntax. A sequence can express repeated
requirements explicitly. Counts and Boolean conditions express unit-level policy.
These are documented limits of the DSL, not unimplemented implicit syntax.

Lexical matching is sentence-local; paragraph and document scopes aggregate those
results. Boolean guards and exceptions apply to the entire selected unit. Counting
a predicate without located occurrences is rejected. Regex alternatives must
consume text, avoiding ambiguous zero-width findings. Matches never bridge a
protected token or reintroduce excluded source into density denominators.

Features read existing neutral NLP results. A separate learned-feature registry
remains research work under #56; this implementation adds no statistical model or
probability. POS and chunk evidence is explicitly heuristic. Exact phrase policy
matches do not establish authorship or universal writing quality.

## Acceptance evidence

The `ruleset` tests cover contractions, Unicode, decoded literals, positions,
overlaps, gaps, targets, negation, missing POS, code boundaries, Boolean conditions,
exceptions, density, limits, cancellation, emitter errors, and concurrent results.
CLI tests execute the public pack and reject deliberately incorrect examples.
The external Go consumer registers a compiled pack without internal imports.
Root e2e cases compare exact diagnostics and JSON/SARIF coordinates.

`BenchmarkSequenceScan` measures a complete 100-sentence scan with a bounded gap.
It is a repeatable local benchmark, not evidence that the product's 100,000-word
resource target or editorial precision requirements have been met. Run it with:

```sh
go test ./ruleset -run '^$' -bench BenchmarkSequenceScan -benchmem
go test ./ruleset -run '^$' -fuzz FuzzLoad -fuzztime 10s -parallel 2
```
