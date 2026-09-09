# End-to-end CLI cases

The [corpus workflow test](corpus_test.go) builds the developer command with cgo
disabled, plans the pinned Ptah sources, extracts candidates, and repeats extraction
to verify them. Its [golden](corpusdata/ptah-units.golden.json) records all 378 units'
text, context, role, kind, and original source segments. Conflicting split pins and
invented human-corpus status must fail. See the
[fixture provenance](../research/annotation/corpus/testdata/README.md); these units
have no human labels and all belong to development.

The [annotation protocol test](annotation_test.go) builds the research command,
checks packet and decision goldens and an agreement summary, and requires exit 2 for
ambiguous JSON or attempts to count simulated responses as a human corpus.
Its fixtures are documented teaching data, not human annotation evidence.
The decision golden preserves simulation status, unresolved labels, source
bindings, and adjudication selections without copying target prose or rationales.

Run `go test ./e2e -count=1` from the repository root. The suite builds the real CLI
with `CGO_ENABLED=0`, then launches it in a fresh directory for each scenario.
The native CI matrix and `make check` run this package through the module runner
with `-count=1`. This bypasses cached test results while retaining the Go build
cache. The suite builds nested-module commands as subprocesses; changing those
sources can otherwise leave a cached root e2e result valid. A plain
`go test ./e2e` is insufficient when checking such changes. See the
[reproduction](../docs/e2e-freshness.md).
No installed Unswell binary, external service, or separate test framework is needed.

## Read a case

Each directory in `testdata` contains:

- `case.json`: source filenames, expected process exit code, and optional CLI modes.
- Optional `resources` in `case.json`: local config and dictionary files copied without becoming source arguments.
- `sample.go.txt` or another language fixture: source with inline `want` annotations.
- `policy.yaml`, when the case overrides the shared two-rule policy.
- `diagnostics.golden.json`: exact diagnostics, source spans, completion, and gate state.
- `stdout.golden` and `stderr.golden`: complete user-facing process output.
- `features.golden.json`, when `feature_collection` is expected: requested values,
  absence reasons, source segments, and compatibility identities.

For example, the [Go fixture](testdata/go/sample.go.txt) places expectations beside
the comment and string that should trigger them:

```go
// It is important to note that the client retries. // want "filler.announced-importance"
const message = "Certainly! The client opens connections." // want "scaffold.chat-preamble"
```

The annotation lists exact rule IDs expected on that source line. Multiple quoted
IDs represent multiple findings. Every finding must consume one expectation; missing,
extra, duplicate, and wrong-line findings fail the test. Unannotated prose is a
negative expectation. The harness rejects malformed annotations.

Use `want-suppressed "rule.id"` for a raw finding that must have a valid source
permission. A plain `want` requires an active finding. A mismatch in either
direction fails, even when the rule ID and source line are correct.

Before scanning, the harness replaces annotation bytes with spaces to preserve
offsets without feeding expected diagnostics into prose analysis. Fixture suffixes
keep source samples out of Go package discovery; the CLI receives the original
language filename. CRLF and BOM cases materialize those bytes explicitly on every OS.

## Coverage

The cases cover all 18 supported input formats, comments and strings, and clean
rewrites. Source-mapping cases include Markdown code protection and entities,
escaped Unicode, CRLF/BOM coordinates, and interpolation boundaries. Policy cases
check context replacement, reasoned symbol exceptions, and directory exclusions.
Process cases cover stdin, malformed and empty input, invalid configuration, and
`--no-gate`. Exit codes 0, 1, and 2 are explicit expectations.

The [feature collection](testdata/feature_collection/sample.md.txt) case requests
word count and lexical diversity through the CLI. Its golden keeps three counted
words across Markdown emphasis, excludes protected code, and records absent
heading measurements under CRLF. A missing requested collection fails the case.

The [window activations](testdata/window_activations/sample.md.txt) case checks
all six contrast, triad, preface, question/answer, and passive-candidate rules.
Eight `want` annotations cover Markdown, Go comments, and a Go string. Feature
goldens distinguish zero from unavailable context, preserve emphasis with
BOM/CRLF coordinates, and keep independent comments and answers separate.

The [repetition group activations](testdata/repetition_group_activations/sample.md.txt)
case preserves five annotated clusters and their related locations. Exact matches
can link a Go comment to a string, while opener measurements remain unavailable
for both. Markdown goldens retain sentence minima, the first-sentence requirement,
code boundaries, emphasis, and BOM/CRLF coordinates.

The [empty table cells](testdata/markdown_empty_cells/sample.md.txt) case came from
the Ptah evaluation. It preserves detections beside empty cells and protected code,
including Unicode and Markdown emphasis in a file with CRLF and a BOM.

The [empty table rows](testdata/markdown_empty_rows/sample.md.txt) case requires
a detection on line 5 after a blank data row. Its
[CRLF variant](testdata/markdown_empty_rows_crlf/sample.md.txt) covers repeated rows,
different column counts, a BOM, Unicode, emphasis and protected code boundaries.
The [orphaned delimiter](testdata/markdown_orphaned_delimiter/sample.md.txt) case
requires exit code 2 for a known partial grammar tree, preserving the difference
between an unsupported boundary and a completed clean scan.

The [Markdown boundaries](testdata/markdown_boundaries/heading.md.txt) fixtures
intentionally lack final newlines. They cover EOF headings and the first quote,
code block, or link definition with BOM/CRLF and original SARIF coordinates.
The [included quote](testdata/markdown_quote_included/quote.md.txt) case requires
the same leading quoted phrase to be detected when quote analysis is enabled.

The shared policy limits language-extraction cases to two phrase rules and sets
chat-preamble matching to sentence starts. The `document_position` case checks its
document-start setting separately. This suite does not claim every catalog rule or
cancellation behavior is covered; rule examples and failure-path tests also live
in the root, `extract`, `report`, and `mcp` packages.

JSON goldens retain diagnostic messages and exact byte/Unicode coordinates while
omitting unrelated catalog and build metadata. The suite checks SARIF results against
those same validated diagnostics, confirms incomplete scans are unsuccessful in
SARIF, checks source privacy, and verifies input files remain unchanged.

The [activation collection](testdata/activation_collection/sample.md.txt) case
uses CRLF Markdown to distinguish a disabled rule, an unsupported heading, an
insufficient word count, an observed zero, and a positive activation. Its
`features.golden.json` records actual values and absence reasons beside the usual
annotated detection and source coordinates. No model or quality labels are used.

The [phrase activations](testdata/phrase_activations/sample.md.txt) case adds a
heading match, a clean comparison, an empty dictionary, document-position limits,
protected code, and too few tokens for a comparison. BOM, CRLF, emphasis, and
Unicode exercise the original source coordinates alongside numeric and absent
values. A matched phrase still fails its configured gate.

The [local activations](testdata/local_activations/sample.md.txt) case collects
seven syntax, modifier, connective, insertion, and punctuation signals. Annotated
matches, clean measurements, protected code, unsupported headings, and short
blocks share one real CLI run with BOM/CRLF and mapped Markdown emphasis.

The custom-rule cases load inline packs through the real CLI. They cover matching
Go comments and escaped strings with CRLF, and paragraph conditions with exact
regex locations, Markdown emphasis, a BOM, protected code, and a local exception.
Their annotated findings and JSON/SARIF output use the same checks as builtin rules.

The `config_inheritance` case loads a parent policy and shared term dictionary,
applies two overlapping file overrides, and retains the unexempted occurrence beside
a valid technical term. Its nested Markdown sources include emphasis, protected code,
CRLF, and a BOM. `config_cycle` requires a startup error before any report is written.
The harness verifies that source files, policy, and dependency resources stay unchanged.

The `suppression_*` cases cover sentence and Go region permissions, unused and
misspelled directives, complete multi-location evidence, and explicit file-wide
permission. Goldens retain directive locations, reasons, targets, used IDs, and
the suppressed state of raw findings. SARIF must mark exactly those findings as
accepted source suppressions with reasons. CRLF, a BOM, and Markdown emphasis
exercise the same source mapping as ordinary detections.
The [empty-row suppression case](testdata/suppression_after_empty_rows/guide.md.txt)
checks block and inline permissions after a table whose empty rows require parser
normalization. Their locations still refer to original bytes, and the final
unpermitted finding must fail the gate.

## Review golden changes

`TestBaselineWorkflow` runs annotated files in [baselinedata](baselinedata) through
the built CLI: create, check after unrelated movement and Markdown emphasis, reject
semantic changes and new debt, explicitly update, then report resolved debt.
Its golden preserves matches, stale entries, finding states, raw and effective
scores, and gate acceptance. Each check verifies that source and baseline bytes
remain unchanged. Use `go test ./e2e -run TestBaselineWorkflow -update -count=1` to
review an intentional change to this workflow separately.

`TestCommittedWorkflow` uses [changesdata](changesdata) in a temporary Git repository.
Its inline annotations still require every raw detection, including unchanged debt.
The golden adds change states, compared/selected units, source hashes, scores, and
unchanged gate reasons. The process must pass after unrelated line movement, fail
after joining paragraphs or changing technical meaning, and return an operational
error for an uncommitted rewrite even with `--no-gate`. Both commit IDs are checked
against Git without embedding timestamp-dependent commit hashes in the golden.
Run `go test ./e2e -run TestCommittedWorkflow -update -count=1` for intentional
changes to this workflow.

`TestTrustedWorkflow` adds a disabled candidate rule and a new source permission in
a real Git repository. The built CLI must still fail under base policy, then pass
after editing the prose. The [trusted fixtures](trusteddata) retain expected raw
detections; their golden records policy hashes, permission trust, source locations,
and full-scan selection. Update only this workflow with
`go test ./e2e -run TestTrustedWorkflow -update -count=1` after reviewing those facts.

After an intentional behavior change, run:

```sh
go test ./e2e -run TestCLI -update -count=1
git diff -- e2e/testdata
go test ./e2e -count=1
```

Review expected rule IDs, source ranges, messages, and gate behavior before accepting
the diff. `-update` still checks annotations, exit codes, source coordinates, and
SARIF parity first. Ordinary runs never create or rewrite missing goldens.

The MCP process tests use the official client and remain in the separate MCP module.
Repository dogfooding also compares CLI and MCP results through the actual protocol.

The [Go analysis module](../goanalysis) adds standard `analysistest` fixtures with
inline `want` annotations, exact byte-range assertions, and a real
`go vet -vettool` test. Its adapter result is compared with the public engine.
These tests participate in the module inventory and native-platform CI.

The `editorial_patterns` case covers the eleven opt-in contextual rules, the
version-2 not-only matcher, and a FAQ file override. It retains the overlapping
praise/triad findings and their related ranges. `editorial_code` covers an escaped
Go string, a reasoned permission, and independent comments/strings that must not be
combined to reach a rhetorical threshold. Both fixtures include CRLF and a BOM.

The `repetition_patterns` case covers five opt-in signals, their related clusters,
heading/summary structure, and clean technical contrasts. `repetition_code` covers
escaped Go strings and a reasoned permission for a repeated comment block. Both
retain exact source coordinates with CRLF and a BOM. Library tests also check
numeric units, modalities, terminology, resource budgets, and concurrent reuse.

`surface_patterns` covers all eight experimental syntax/readability and
formatting rules with file-specific policies, CRLF, a BOM, emphasis, and entities.
`qualifier_activations` covers seven annotated detections from vague praise,
absolute claims, weak intensifiers, and stacked hedging. Feature goldens distinguish
clean zeros from inapplicable headings, qualifications, short fragments, and code.
Go comments/strings and Markdown retain their BOM/CRLF and emphasis coordinates.

`window_phrase_activations` checks repeated section announcements, empty transitions,
and metaphors. Annotated detections and feature goldens cover cluster membership,
an allowed occurrence, clean zeros, protected sentences, headings, short fragments,
and separate Go comments and strings. BOM/CRLF, Unicode, and Markdown emphasis
exercise original coordinates in JSON and SARIF.

`noun_ambiguity` preserves the reported package and method comments as negative
cases in Go and Markdown, alongside genuine stacks with singular and plural heads.

`candidate_activations` covers the seven near/extended repetition and list rules.
Annotated clusters and feature goldens distinguish evaluated zeros from missing
pairs, protected candidates, excluded boundaries, and reference lists. Markdown
headings, summaries, lists, and emphasis retain original coordinates; Go comments
and escaped strings cover the same engine with BOM/CRLF. Collection must preserve
ordinary findings and gate outcomes.

`near_contrasts` checks modal, condition, state, and number-unit differences against
positive near-match clusters. It includes the reported `may retry`/`must retry`
pair, a contrasting sentence beside a valid cluster, Go comments and escaped
strings, Markdown emphasis, and BOM/CRLF. JSON and SARIF retain original locations.

The GitHub matcher tests render real text reports, preserve severity and locations,
and leave compiler diagnostics for setup-go. CI registers the Unswell matcher
after setup-go so it takes precedence for Unswell's own diagnostic lines.
Runner uses ECMAScript regex semantics. These tests use regexp2 in that mode and
retain a counterexample accepted by Go regexp but rejected by ECMAScript. This
test dependency is pure Go and does not enter the CLI or MCP runtime. Also confirm
the annotation levels in live CI when changing the matcher.

`surface_code` checks the same engine through Go comments and escaped
string literals. Unannotated clean prose must produce no extra findings.

`TestDependencyTraceReplay` uses `testdata/dependency-traces.json`, a frozen
GoSpacy/model run on ten source probes. It passes those predictions through the
public engine and a test-only relation rule, checking missing/expected findings,
source coordinates, and model identity. Markdown entities, emphasis, CRLF/BOM,
protected code, Go comments, and escaped literals retain their original ranges.
The JSON includes full trees for inspection. These are model predictions, not
human labels: the `server` finding in `technical-3.txt` deliberately preserves a
questionable parse reproduced by Python. The experiment and its limitations are
in [dependency research](../research/dependencies/README.md).
