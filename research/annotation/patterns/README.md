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

Every row also carries `roles`: the same counts and the same interval for
each document role the rule admits, so a comment, a README, and a
reference page are not compared against one pooled rate. A stratum's
components are those that hold documents of the role in that cohort. The
strata add up to the row.

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

`corpus dedupe --unit-kind paragraph --order historical,contemporary
--candidates ...` writes an `unswell-unit-selection-v1` selection instead.
It keeps the first occurrence of each unit text across every cohort in the
stated order, whatever repository a copy sits in. Within a cohort the first
occurrence is the first by repository, source, and unit ID. Each cohort
reports its kept units, the units that repeat text kept earlier in the same
cohort, and the units that repeat text kept in an earlier cohort. `corpus
analyze --unit-kind paragraph --selection selection.json` then keeps the
listed units in every cohort, which must be exactly the cohorts of the
inputs, and reports the kept and excluded counts under `selection`. That is
the protocol's rule of one count per content hash across snapshots, copies,
and identical document versions. A selection and a first-appearance filter
cannot apply together.

Each non-baseline cohort gets a contrast with the baseline, `historical`
unless the command says otherwise. The contrast holds the absolute difference
of document-level prevalence with a joint-resampling interval and the
prevalence ratio, which stays `undefined` when either cohort has no finding.
A difference is an association between cohorts under this policy. It is not
a false-positive rate, not recall, and not a decision. The protocol's
evidence-card rules decide a state from these numbers at a later stage.

## Frequency tables

`corpus frequencies --candidates ...` counts constructions instead of
findings. The corpus can then propose candidates that no rule names yet.
The command reads the sentence units of the candidate artifacts and tags
them with the English provider. It counts five measures per cohort and
role: word unigrams, bigrams, and trigrams of lowercase alphabetic words,
the first three words of a sentence, and the part-of-speech template of
the sentence. A number, an identifier, or a punctuation mark breaks the
n-gram window. The template holds the coarse tags of the words, cut at
twelve. The output is `unswell-frequency-tables-v1`.

A stratum is one cohort and role with its sentences, words, and
provenance components. Each measure lists contrasts for every target
stratum against the same role of the baseline. The targets are the
cohorts named by `--target`, or every cohort but the baseline. A key
enters a contrast when it reaches `--min-count` in the target stratum
and appears in `--min-components` of its components. A contrast holds
both counts, both rates per 1,000 words, and both component supports. It
holds the ratio of the rates only when the baseline also reaches the
minimum count. Otherwise its status says whether the baseline holds the
key below the minimum or not at all.

Defined ratios sort first, descending. The other statuses follow, sorted
by target count, and `--top` cuts each list. The items of a measure are
the keys that enter a contrast, with their cell in every stratum where
they reach the minimum.

`--pair-tasks` and `--pair-records`, given once per generation run in the
same order, relabel the strata of that run. A response whose task is in
the set counts under `controlled-<operation>`, and the sentences of the
paragraph a task names count under `<cohort>-paired`. Every other unit
keeps its cohort. With `--baseline historical-paired` the contrasts put
the responses against the sentences they answer, so the repository and
language mix of the tasks drops out. The output names the pairing.

The command reads the artifacts twice: once for the counts, once for the
component support of the kept keys. It holds one artifact and the counts
in memory. A high ratio is a proposal to read the sentences behind it. It
is not a rule, not a defect, and not a claim about who wrote the text.

## Screening

`corpus screen` runs the exploratory screening the pattern protocol names
for its selection partition. It takes the rule classes, the dataset plan,
the finding artifacts, and the records and task sets of the generation
runs that enter selection, one `--records` and one `--tasks` per run in
the same order. The partition of every source comes from the plan, never
from a finding artifact, so a plan made with partition pins decides what
the screening reads. The output is `unswell-screening-v1`.

Every rule whose class admits the role stratum meets every family the
runs name. The controlled arm holds the units of the chosen kind in the
complete `generate`/`neutral` responses of that family. The H0 arm holds
the units of the same kind and role in the tasks' cohort. Both arms keep
the units the plan puts in the screened partition. The difference
`D` is the unit prevalence of the controlled arm minus the H0 arm, with
the cluster bootstrap interval of the tables. The p-value is two-sided
from the same replicates: twice the smaller share on either side of zero,
never below one replicate's share. The Benjamini-Hochberg step runs across
every test with a p-value at `--fdr`, and a candidate is a test that
passes it with a positive difference.

Each test carries a card state. The protocol defines it from the
interval, the point estimate, the support in components, and the cluster
minimum. The baseline load is not checked here, and the output says so.

The sample-size record starts from the observed H0 prevalence and the
bootstrap dependence. It gives the units and components per arm that
detect `--mid` at two-sided alpha 0.05 with power 0.80. It gives them
again at alpha 0.05 over twelve, the Holm worst case. The design effect is
the bootstrap variance of `D` over the binomial variance of the two unit
counts.

A candidate is a rule to write a hypothesis template for. No candidate is
a confirmed construction, a rule decision, or a claim about who wrote the
text.

