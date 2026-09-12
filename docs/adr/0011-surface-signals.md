# ADR 0011: surface syntax and readability

Status: accepted for experimental implementation

Issue #19 adds eight opt-in rules to the existing engine. They use actual tokens,
POS tags, shallow NP chunks, and grammar-derived Markdown structure. Dependency
claims remain in #20. The research umbrella's feature registry and comparative
qualification remain in #56 and #57; no trained model or probability is introduced.

ARI uses a documented character/word/sentence formula with no syllable resources.
Its token-count protocol and limitations are explicit in the
[surface catalog](../surface-signals.md). Readability and formatting measurements
have no default gate, and ARI plus the two formatting rules have zero default weight.
The five other candidates have small initial weights. All eight remain disabled in
builtin profiles until qualification supplies evidence for a different decision.

The additive `Block.List` metadata records local grammar list/item identity, depth,
marker kind, and complex content. It is requested through `RequiresStructure` and
never restores excluded prose. The pinned grammar's missing list wrapper is handled
through contiguous sibling `list_item` nodes and grammar markers. Canonical context
labels remain identities; editorial heading checks use selected heading text.

New typed parameters cover word dictionaries, sentence counts, insertion depth, and
list sizes. They are optional in JSON and validated before analysis. API snapshots
and the report schema include them. Older strict readers may reject reports that
contain the new fields; this is an explicit additive alpha compatibility change.
Existing field names, schema identity, source permission syntax, and exit codes are
preserved. Saved results still render without reanalyzing source or loading NLP.

Shared counting and mapping functions serve the rules. Missing POS/chunk support,
invalid NP ranges, and excessive nesting fail explicitly. A rule that exhausts its
candidate budget abstains on that document with `budget_exhausted`.
Matched source ranges include conditions and negation; heuristics never supply an
automatic rewrite. Technical counterexamples and exact CLI goldens are required,
while corpus qualification and resource acceptance remain separate evidence.
