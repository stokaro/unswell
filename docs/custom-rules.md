# Custom editorial rules

Use a local rule pack to enforce team wording in selected comments, strings, and
documents. Packs run in the existing Unswell engine and share its extraction,
English tokens, source coordinates, scoring, and gate policy. They cannot run
scripts or download resources. A rule match establishes your chosen policy's
condition; it does not establish authorship or a probability of poor writing.

## Load and test a pack

The [company example](../examples/rules/company.yaml) contains the full version-1
format, including executable examples. From the repository root:

```sh
unswell rules test examples/rules/company.yaml
unswell rules show company.no-dive-in --ruleset examples/rules/company.yaml
unswell check README.md --ruleset examples/rules/company.yaml
unswell config validate --ruleset examples/rules/company.yaml
unswell doctor --ruleset examples/rules/company.yaml
unswell explain README.md --at 1:1 --ruleset examples/rules/company.yaml
```

Repeat `--ruleset` for multiple files. A directory selects its regular `.yaml` and
`.yml` files in sorted order, without recursion or following directory symlinks.
An empty directory is an error. At most ten files can be loaded; each file is at
most 1 MiB. Directory enumeration stops with an error above 4,096 entries.

`rules test <path>` runs only that pack or directory's examples. Each fail example
must produce a finding from its rule; each pass example must produce none.
Malformed source and unavailable capabilities fail the command. A wrong expected
result names the rule and example number and exits with code 2. The command does
not rewrite expected results. With no path, `rules test` tests the selected catalog;
`--config` and `--ruleset` can select it.

The main configuration also accepts `rule_sets`, a sequence of complete inline
packs. These have the same schema as standalone files. This supports MCP clients
that provide a policy at server startup. File paths are a CLI concern; the library
does not read them from YAML. External and inline packs extend the selected
catalog, and duplicate IDs are errors, including collisions with builtins.

Rules are enabled by default except under `builtin:custom-v1`, which starts with
every rule disabled. The existing `rules.<id>` settings override severity, gate,
enabled state, and score. A disabled definition must still compile successfully.

## Pack and rule metadata

A pack requires `version: 1`, `namespace`, `release`, `license`, `provenance`, and
one to 100 `rules`. A namespace starts with a lowercase ASCII letter and contains
lowercase letters, digits, or hyphens. IDs start with that namespace and a dot;
each dot-separated component starts with a letter. IDs have at most 100 bytes.
Release, license, and provenance are author-supplied declarations, each limited
to 1,000 bytes. Unswell records them without verifying their legal or factual basis.

Each rule requires an ID, summary, message, scope, matcher, and nonempty pass and
fail examples. Its integer version defaults to 1. Severity defaults to `warning`,
gate to `none`, group to `custom`, and score weight/cap to zero. Use `gate: forbid`
for an explicit blocking policy. Existing scoring caps and correlation groups
continue to apply. A successful rule emits one finding per analysis unit, with
its other occurrences as related locations.

`contexts` selects extracted block kinds: `comment`, `string`, `paragraph`,
`heading`, `list-item`, and `table-cell`. Omitting it selects all these kinds.
It cannot restore text excluded by the extraction policy. The rule's `scope` is
`sentence`, `paragraph` (one extracted block), or `document` (selected prose blocks).
Scope controls aggregation, not extraction or permission to cross code boundaries.

Capabilities are inferred. An optional `requires` set must agree with the matcher:
all rules require `tokens` and `sentences`; POS sequences add `pos`, and chunk
features add `chunks`. An enabled rule with a missing capability fails engine
construction. No POS approximation substitutes for a dependency capability.

Examples accept a plain string or a mapping with `text` and `format`. Formats use
the same IDs as source inputs, such as `go`, `markdown`, or `yaml`. Set `format`
when testing comment-only or string-only policies. Each rule allows at most 100
examples, each with at most 10,000 bytes of source.

## Matcher semantics

| Type | Fields | Result |
| --- | --- | --- |
| `phrase` | `values`, optional `at`, `case_sensitive` | Consecutive English tokens matching a phrase |
| `token-set` | Same as phrase | Any listed single token |
| `regex` | `pattern`, optional `case_sensitive` | Nonoverlapping RE2 matches in a visible prose fragment |
| `sequence` | `tokens`, optional `at`, `case_sensitive` | Positioned tokens and bounded gaps |
| `all`, `any` | `items` | Boolean combination of two to 32 matchers |
| `not` | `match` | Negated condition on the current unit |
| `count` | `match`, `min` and/or `max` | Number of unique occurrence ranges |
| `density` | As count, plus `per` | Matches per 100 words or per 100 sentences |
| `feature` | `name`, `min` and/or `max` | Comparison with an existing shared measurement |

Phrase, regex, and sequence matching stays within one sentence even when the rule
aggregates a paragraph or document. Inline code and other protected text split
lexical matching. Markdown syntax is already extracted; regex does not traverse
markup or reinterpret raw strings. Decoded strings retain original byte ranges.

Phrase patterns use the same Prose tokenizer and `document.Normalize` function as
the default English provider. Contractions and typographic apostrophes therefore
retain token boundaries. `case_sensitive: true` compares literal token text.
Custom NLP providers must supply compatible tokens and normalized text when using
these lexical matchers; capability names alone cannot establish compatibility.

`at` accepts `any` (default), `start`, or `end`. Start means the first token of the
sentence, including an opening quotation mark if tokenized. End permits trailing
unprotected punctuation, but no following word or protected token. Regex uses its
own anchors on each visible fragment. Patterns that can match without consuming
text are rejected, including standalone boundaries and empty alternatives.

A sequence token specifies `value`, a Penn Treebank `pos` tag, or both. Both
constraints must match. `not: true` negates that conjunction; it never accepts a
protected token. `target: true` limits reported segments to the marked tokens.
Without targets, evidence covers the consuming sequence, including gap tokens.
Gap entries use `gap: {min: 0, max: 3}`. Gaps occur between consuming tokens and
cannot be adjacent. Each bound is inclusive and cannot exceed 32 tokens.

Sequence search starts at each eligible token in source order. It tries shorter
gaps first and keeps the first complete match for that start. Overlapping starts
are allowed. It memoizes matcher states and checks a work budget; the number of
possible gap combinations does not cause unbounded backtracking.

`all` evaluates left to right and stops at the first false condition. On success,
it combines its children’s evidence. `any` stops at the first true condition and
returns only that branch's evidence. `not` emits no invented token location.
Conditions operate on the whole current unit; `all` is not a token adjacency test.

Count and density deduplicate occurrences with the same block, sentence, and
original bounding range. Their child must locate occurrences on every successful
branch. Pure features, negations, and nested counters cannot be counted directly.
An `all` can combine a locating matcher with predicate guards; every `any` branch
must locate matches. Comparisons are inclusive, finite, nonnegative, and at most
1 billion. Counts and feature bounds require integers. Density accepts decimals.
Protected code contributes neither words nor sentences to denominators.

Available shared features are `prose.words`, `prose.tokens`, `prose.sentences`,
`chunks.np`, `chunks.vp`, and `chunks.pp`. They use existing sentence word counts,
visible tokens, and the neutral NLP chunks. Chunks are shallow POS-based candidates,
not dependency parses. Features are counted across eligible sentences in the unit.
If a condition succeeds without a specific match location, the finding identifies
the unit's visible prose segments instead of blaming an invented word.

An `except` list contains up to 32 matchers. It runs before the main matcher in
source order. Any true exception suppresses this rule for the current unit.
Other rules still run. Unit exceptions differ from the separate suppression and
baseline policies; they are part of the rule's definition and content identity.

## Limits and result identity

Every definition rejects unknown fields, aliases, nulls, merges, and excessive
nesting. Matcher depth is limited to 16 with 512 nodes per rule, including exceptions
and sequence atoms. Sequences have at most 32 atoms; phrase lists have at most 100
values of 1,000 bytes each. Regex patterns have at most 4,096 bytes.

Each rule uses `analysis.max_candidates` as a per-document work budget, shared by
its units and exceptions. Token comparisons, sequence visits, Boolean operations,
regex chunks, and matches consume this budget. A unit can retain at most 10,000
occurrences. Existing document, token, timeout, and finding limits still apply.
Cancellation, exhausted budgets, and emitter errors make the analysis incomplete;
they cannot become a successful required gate.

The library compiles packs at construction, keeps them immutable, and owns each
returned descriptor. The source definition's serialized YAML hash is recorded
alongside namespace, release, license, and provenance. All definitions, including
disabled ones, participate in the policy and ruleset identities. Saved reports
contain the identities and findings needed by reporters, without reloading packs.

Use `ruleset.Load(bytes).Rules()` to extend a Go registry, or supply YAML byte
slices through `Options.RuleSets`. For inline configuration, the engine calls
`config.Compile` and registers its returned implementations. The
[external consumer test](../examples/consumer/ruleset_test.go) demonstrates public
imports and saved JSON. The [e2e cases](../e2e/README.md) retain exact diagnostics
and original coordinates through the CLI.
