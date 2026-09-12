# Reasoned source permissions

A source suppression permits a specific rule finding without deleting its evidence.
Use one when wording must remain, such as a quoted contract or a repeated definition.
It does not change other rules or declare the wording free of editorial problems.

## Syntax and targets

Put the directive in an actual source comment or a Markdown HTML comment:

```markdown
<!-- unswell-disable-next-block filler.announced-importance -- Quoted customer wording; preserve it exactly. -->

It is important to note that the client retries.
```

```go
// unswell-disable-next-block filler.announced-importance -- Required wording in the external contract.
// It is important to note that the client retries.
func Retry() {}
```

| Command | Target |
| --- | --- |
| `unswell-disable-next-sentence` | First complete eligible sentence after the comment |
| `unswell-disable-next-block` | First complete eligible prose block after the comment |
| `unswell-disable` followed by `unswell-enable` | Complete blocks or sentences between the paired comments |
| `unswell-disable-file` | Every eligible block in the file, only with explicit policy permission |

Targets follow extracted structure, not physical lines. A wrapped sentence remains
one sentence. A block may be a paragraph, heading, list item, table cell, comment,
or string selected by the extraction policy. In Go, sentence, block, and region
directives target comment prose; they do not select statements or string literals.
An explicitly permitted file-wide directive includes all extracted contexts.

An inline Markdown directive can precede the next sentence in the same paragraph.
Visible prose determines whether a complete unit falls after the directive. Its
reported source contour can still include protected comments or formatting. A
directive inside an unfinished sentence does not permit that sentence's earlier prose.

List multiple exact IDs with commas and no intervening spaces:

```markdown
<!-- unswell-disable filler.announced-importance,repetition.exact-sentence -- Preserve the quoted contract section. -->

It is important to note that the client retries.

It is important to note that the client retries.

<!-- unswell-enable filler.announced-importance,repetition.exact-sentence -->
```

The closing ID set must match the innermost opening set; its order is irrelevant.
The opener supplies the reason, so `enable` takes no reason. Different-rule regions
may nest. Overlapping targets for the same rule are errors, including two next-block
directives that would select the same block. Every listed rule permission must be used.

Commands are case-sensitive. IDs must exist in the engine's rule catalog, including
registered custom rules. Wildcards, `all`, and derived `gate.*` IDs are invalid.
Each directive occupies one logical comment line. Comment delimiters and conventional
leading `*` decoration in block comments are removed before validation.

Strings, fenced or inline Markdown code, indented comment examples, excluded quotes,
raw HTML contents, and plain-text examples cannot create permissions. Directives themselves do not
contribute prose words or editorial findings. Source comments remain policy-sensitive
even when ordinary comment prose is disabled in a source-language context set.

A saved report lists a suppression comment with the reason `suppression-directive`.
A comment for another tool, such as `nolint`, has the reason `directive`. The
[directive comments](inputs.md#directive-comments) section lists those prefixes.

## Policy and errors

These defaults apply without an explicit configuration section:

```yaml
version: 1
suppressions:
  require_reason: true
  allow_file_wide: false
  reject_unused: true
```

The reason follows ` -- ` and must contain at least two words containing letters
when required. This rejects empty reasons and placeholders such as `TODO`; it cannot
judge whether an explanation justifies the exception. Review that decision as a
policy change. Setting `require_reason: false` permits an absent reason.

Settings inherit through the existing configuration plan and support file overrides:

```yaml
version: 1
overrides:
  - files: [contracts/**]
    suppressions:
      allow_file_wide: true
```

File-wide permission still requires an explicit `unswell-disable-file` directive,
exact rule IDs, and a reason. An unpaired region is always an error. Unswell does
not read a clock or implement expiring permissions.

Unknown commands or IDs, malformed reasons, missing targets, invalid pairs,
overlapping permissions, and unused rule permissions produce incomplete analysis
and CLI exit code 2. `--no-gate` does not conceal them. `reject_unused: false`
keeps unused records in the report without treating them as an error; other validation
still applies. An unused-permission failure retains the findings already computed.

Limits are 1,000 directives per source, 4,096 bytes per logical directive, 32 IDs per
directive, and 10,000 targets across its suppression plan. Structural resolution
and evidence matching also consume the configured `analysis.max_candidates` budget.
Exhausting a limit fails the analysis; it never turns into permission for the rest
of the file. A permission for a rule that abstained on the document, for example
with `budget_exhausted`, had no findings to cover. It keeps the status `unused`
in the report and is not an unused-permission error.

## Evidence and scores

A finding is suppressed only when permissions for its exact rule cover every
evidence segment. A repetition cluster spanning two paragraphs needs permission
for both occurrences. Separate directives may jointly cover it. Permission for
only one occurrence leaves the entire finding active and counts as unused unless
that directive covers another complete finding.

`findings` retains every raw finding. A permitted finding has `suppressed: true`
and `suppression_ids` linking it to records in `suppressions`. Each record stores
its reason, opening and optional closing locations, structural targets, permitted
finding IDs, used rule IDs, and status: `used`, `unused`, or `partially_used`.

Assessments retain raw `slop_score` and `contributions`. The engine recomputes
`effective_slop_score` from the remaining findings through the same correlation
and cap rules. When permissions affect a unit, `effective_contributions` explains
that calculation and marks removed evidence with `source-suppression`.
Subtracting a finding's old contribution would miss evidence that it previously
masked through correlation or caps.

Hard-rule and score gates use the effective findings and scores. Document summary
statistics remain raw. No suppression creates or adjusts a probability; this alpha
still reports `calibration_unavailable`. Baseline debt acceptance remains separate.

The CLI, public engine, and MCP return the same records. JSON and SARIF retain all
raw findings; SARIF marks permitted findings as accepted `inSource` suppressions
with their reasons. Text, Markdown, and HTML identify permitted findings and show
the audit records. Saved JSON renders without source files or another analysis.
Reasons remain in saved reports even when source inclusion is disabled; keep secrets
out of them. See [reports](reports.md) and [scoring](scoring.md).

An assistant using MCP should revise its draft and check again. A failed check
does not authorize adding a suppression merely to pass. Permissions in source are
policy-sensitive content; trusted-policy enforcement is tracked by issue #15.
