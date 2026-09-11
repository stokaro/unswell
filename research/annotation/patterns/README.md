# Pattern tables

This package builds the E1 baseline tables of the
[LLM-associated pattern protocol](../../methods/llm-patterns-v1.md) from
label-free finding artifacts. It supplies no label, no probability, and no
quality judgment; every number is a rule outcome under one pinned policy on
one cohort. See [ADR 0036](../../../docs/adr/0036-llm-pattern-evidence.md).

```sh
corpus analyze --classes ../methods/rule-classes-v1.json \
  --findings shard-a.findings.json --findings shard-b.findings.json > tables.json
```

Inputs are `unswell-corpus-findings-v1` artifacts from `corpus measure`, all
measured under one policy identity, and the committed rule classes, which must
cover the policy's catalog. Each artifact's digest is checked before use. The
output `unswell-pattern-tables-v1` keeps `human_corpus: not_qualified`.

For each rule and cohort the tables report the applicable documents, which
are those whose role the rule class admits, and the provenance components they
span. They report prose words, findings, and documents with at least one
finding. They report the document-level prevalence with its interval, the
findings per 1,000 words, and the share of units of each kind with a finding.
A cohort summary gives the whole-profile load and the failed documents,
whose policy run did not finish and which enter no table. A document without
a cohort is counted as unassigned and enters no table.

The unit of independence is the provenance component. Intervals come from a
cluster bootstrap with 10,000 replicates, PCG seed 17, and percentile bounds
at 2.5% and 97.5%. The bootstrap resamples components jointly across cohorts,
so a component's historical documents and all its responses enter or leave a
replicate together. Nothing counted, fewer than two components, or fewer
than two valid replicates leave the interval absent with a reason. When
a rule never fires in a cohort, the row carries the one-sided 97.5%
Clopper-Pearson upper bound `1 - 0.025^(1/n)` on the share of the `n`
components that could carry the construction.

One count is one document unless the command names a unit kind. The
`unit` field of the output says which.

- `--unit-kind paragraph` or `--unit-kind sentence` counts each retained
  unit of that kind inside the applicable documents.
- `counted` and `counted_with_finding` are then units, and the row's words
  and findings cover those units only.
- A document without a retained unit contributes nothing, not even to the
  component count.
- Paragraph units are the paragraph and comment blocks. List items, headings,
  table cells, and string literals are fragments and stay outside both unit
  analyses.
- A sentence inside a counted paragraph is a nested observation of it. The
  two unit analyses are two views of one text, never independent evidence.

```sh
corpus first-appearance --unit-kind paragraph \
  --earlier historical-a.json --later contemporary-a.json > filter.json
corpus analyze --classes ../methods/rule-classes-v1.json --findings ... \
  --unit-kind paragraph --first-appearance filter.json > tables.json
```

The filter compares the candidate artifacts of two cohorts. For the later
cohort it lists the units of the kind whose exact text the earlier snapshot
of the same repository does not hold. Units that repeat earlier text are
left out with their counts. So is every unit of a repository without an
earlier snapshot. The analysis then keeps only the listed units of the later
cohort. It checks that each was measured exactly once and reports the kept
and excluded counts under `first_appearance`. This is the protocol's
analysis of newly added or changed text, counted at its first evidenced
appearance. The filter compares snapshots of one repository. Copies of one
text across repositories, and repeated text inside one snapshot, are not
collapsed by it.

Each non-baseline cohort gets a contrast with the baseline, `historical`
unless the command says otherwise. The contrast holds the absolute difference
of document-level prevalence with a joint-resampling interval and the
prevalence ratio, which stays `undefined` when either cohort has no finding.
A difference is an association between cohorts under this policy. It is not
a false-positive rate, not recall, and not a decision. The protocol's
evidence-card rules decide a state from these numbers at a later stage.
