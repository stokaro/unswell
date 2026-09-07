# ADR 0004: source suppressions preserve raw evidence

Status: accepted; implementation and acceptance are tracked by #12.

Source exceptions need a reason, a specific rule, and a structural target. A line
number alone cannot identify a wrapped sentence or comment block. The engine must
retain the original findings and scores while applying the exception to its gate.

## Ownership and syntax

Extractors collect directives from actual comments recognized by the source grammar.
Markdown HTML comments also carry directives. String literals, fenced code, excluded
quotes, and ordinary prose do not create policy directives. Extracted directives
retain their source byte ranges and are excluded from prose and word counts.

A shared resolver binds directives after NLP has identified sentences. Support
`unswell-disable-next-sentence`, `unswell-disable-next-block`, and paired
`unswell-disable` / `unswell-enable`. Rule IDs are explicit comma-separated names.
Opening directives use `--` followed by a reason. Require at least two words by
default; this checks that a reason was supplied, not whether it justifies the policy.
Reviewers remain responsible for that decision.

Next-unit directives bind to the first complete eligible unit after the directive.
In Go, that unit belongs to a comment block, never a statement or string literal.
Other source languages can target extracted comments or strings. Regions target
complete blocks or sentences between their paired boundaries. Regions must nest
correctly, and their closing IDs must match the opening IDs. Overlapping permissions
for the same rule are errors instead of creating redundant, apparently used records.

Use nonprotected token ranges to bind complete units, retaining their full original
contours for evidence and audit. Compute these bounds once per document. This lets
an inline comment precede a sentence whose raw contour includes that same comment.

The current inline Markdown grammar rejects internal double hyphens even in code
examples. Normalize those bytes only inside grammar-recognized directive comments
or delimited code spans, then require an error-free tree with identical node kinds
and ranges. PowerShell comment-only scripts need a synthetic empty statement after
the original bytes. Apply that only after proving the original tree contains solely
valid comments and whitespace. Neither normalization changes source bytes or accepts
unrelated syntax errors.

File-wide permission requires both `suppressions.allow_file_wide: true` and an
explicit `unswell-disable-file` directive. Wildcards and an unpaired region remain
errors. The default policy requires reasons and rejects unused directives. Unknown
commands, unknown rule IDs, missing targets, invalid pairing, and resource exhaustion
produce an incomplete analysis and exit code 2. No clock or expiration is introduced.

## Findings and scoring

Run the existing rules once on the original extracted prose. A finding is suppressed
only when every evidence segment is covered by a permission for its exact rule ID.
Separate directives may jointly cover a multi-location finding. Partial coverage
does not remove an aggregate whose other evidence remains active; use a region that
covers its complete evidence or change the relevant rule policy explicitly.

Keep raw findings, source locations, fingerprints, and contribution traces. Record
the directive, its reason, targets, and the finding IDs it permits. Recompute the
effective heuristic index from the remaining findings, including correlation and
caps. Subtracting the original contribution would be incorrect when a suppressed
finding previously masked another correlated signal. Derived threshold diagnostics
use only the effective index and never feed back into scoring.

Hard-rule gates ignore only permitted findings. The existing raw probability retains
its meaning; a suppression does not authorize a transformed probability. Models and
probability gates remain governed by their own later applicability contract.

## Interfaces and verification

The public engine, CLI, and MCP share the resolver and the same completed result.
All reports retain an audit path to the raw findings and reasons. JSON and SARIF
retain every finding; display limits and source inclusion remain presentation choices.
Saved reports render without parsing directives again or reading source files.

Tests cover exact spans, CRLF/BOM, structural targets, pairing, typos, unused records,
bounded work, multi-location evidence, raw/effective scores, and unrelated gate
failures. Include root CLI goldens, parser fuzzing, race tests, and CLI/MCP parity.
Suppression changes remain policy-sensitive inputs for the trusted-policy work in #15.
