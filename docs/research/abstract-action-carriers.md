# Unnamed action carriers in technical prose

The existing `filler.instruction-scaffolding` rule now recognizes a staged,
unnamed piece, bit, or part of logic or code that introduces a transitive
action through `allows`, `enables`, or `supports`. It asks for the component
and supported action directly, retaining capability, quantities, operands,
conditions, and limits. Named components, numbered architectural parts,
unmarked descriptions, attribution, protected construction words, and
negation or restriction remain outside this construction.

One exposed BuildKit development example previously received no finding:
an unnamed final piece of solver logic introduces merging two edges into one
when both returned the same cache key. The revised rule identifies the
complete carrier and action, rather than commenting on the closing word
alone. Its original UTF-8 span retains the entire condition.

## Measured change

The replay used all 36 existing source pages in their original Markdown or
MDX format. Original sources, 804 judgments, and 123 reference events were
unchanged. The baseline is Unswell commit
`ecb1e75e63dee980d33754507872d4afcf457ad0`.

| Same complete source set | Before | After | Added / removed criticisms |
| --- | ---: | ---: | ---: |
| Technical profile | 143 | 144 | 1 / 0 |
| Strict profile | 164 | 165 | 1 / 0 |
| Complete previously missed construction | 0 | 1 | One development event |

Every old criticism retains its message, advice, source spans, related
evidence, policy, metrics, and suppression fields. The revised rule's version
and dependent identity hashes change. The root Codex assistant accepted the
one new diagnostic under ADR 0041: it identifies the unnamed support layer
around the action and explicitly preserves the conditional capability.
This is assistant review, not independent human annotation.

Six separate complete Ptah pages were selected by topic before inspecting
their candidate outputs. Sources are pinned to Ptah commit
`a641f2dd20a9c8596629f43fae0cc2f811d7f5e0`: inference consistency, provider
capacity, recovery, Kubernetes deployment, PostgreSQL, and distributed
databases. Their 132,155 source bytes differ from the original 36-page inputs.
The recognition treatment and advice stayed fixed after selection. A later
lint-required function decomposition was checked for identical results.
These pages retain all 66
technical and 75 strict findings, with no added, removed, or changed
criticisms. They supply negative controls; no new positive confirmation
was observed.

Blackbox tests exercise five positive variations and 17 technical controls,
plus original source mapping for Markdown, Go comments, and Python strings,
and term exemptions, occurrence policy, and candidate length bounds.
Existing instruction and CLI regressions remain required. The ten feature
report goldens change only catalog identity hashes.

## Limits

This recovers one exposed development event. It does not establish broad
precision, unseen positive recall, an authorship claim, or achievement of the
contextual quality goal. No model calls or probability calibration were
performed. Severity, score weight, and gate policy are unchanged.

The latest standalone generation pilot still fails its separate acceptance
gates. Adding this runtime construction does not promote a partial model
diagnosis or combine answers to manufacture a complete witness. The remaining
misses and long-document false alarms remain open in
[#349](https://github.com/stokaro/unswell/issues/349).
