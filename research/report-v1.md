# Which constructions mark generated technical English, and what a linter can do about it

Research report, version 1, September 15, 2026. The evidence release of
[issue 154](https://github.com/stokaro/unswell/issues/154).

The short answer is that the two constructions this project froze as
hypotheses did not replicate, and that the signal which did survive a
held-out partition does not live in any one rule's count.

## Questions

1. Which named constructions occur at a different rate in formulaic
   LLM-generated technical English than in human technical English?
2. How stable is that difference across model families, prompts and roles?
3. Which differences can an explainable rule detect at a warning load a
   project can carry?

The third question is about a linter. The first two are about language.
Neither is a question about the authorship of a given document, and nothing
here answers one.

## Related work

`research/methods/llm-patterns-sources-v1.json` pins six studies with their
data terms and the decision taken on each. Two shaped this design. Reinhart
and colleagues compare corpora on grammatical and rhetorical features, not on
detector accuracy. That is the comparison this work needed. Kobak and colleagues measure excess vocabulary over time and
state plainly that a population shift does not identify one document's
authorship, which is the limit this work adopted. The RAID and DetectRL-X
benchmarks were read and deferred: their published accuracies are for
detectors on essays and abstracts, and transferring them to code comments
would be unsupported.

## Design

`research/methods/llm-patterns-v2.md` is the protocol. It fixes everything
below before any confirmatory data was opened.

- The hypotheses and the unit of analysis.
- A minimum useful difference of three percentage points.
- A cluster minimum of twenty components per arm and a support minimum of
  five.
- Ten thousand bootstrap replicates at seed 17.
- Holm correction at 0.05 for confirmation, Benjamini-Hochberg at 0.10 for
  screening.
- The stopping rule.

Three development screenings preceded the freeze. The confirmation partition
was opened once.

## The corpus

`research/methods/data-card-v1.md` describes it: 28,740 documents, 5,186,062
prose words, six cohorts, 76 pinned repository snapshots, 73 independence
groups, and a controlled arm of 11,856 model responses from five models in
three families under recorded prompts.

## Main result: the frozen hypotheses did not replicate

| Hypothesis | Development | Confirmation | p | Holm |
| --- | --- | --- | ---: | --- |
| H1 `syntax.noun-stack` | +2.96 points | +0.41 points | 0.727 | not rejected |
| H2 `readability.grade-metric` | +2.43 points | +0.32 points | 0.747 | not rejected |

Both confirmation estimates sit an order of magnitude below the minimum
useful difference, both intervals contain zero, and Holm rejects neither.
The controlled arm also held 16 components against the cluster minimum of
20, so both cards read `inconclusive` rather than `unsupported`, and the
frozen list is untouched.

This is the headline result and it is negative. A development estimate of
+2.96 points became +0.41 on text the screening had not seen.

## Second result: the catalog looks for a register current models do not write

Measured on a documentation tree its author states was model-written and
never proofread, 411,096 prose words with all forty rules enabled: fifteen
rules fire, 2,968 findings are warnings, and none reaches forbid severity.
Twenty-five rules produce nothing.

Across the whole corpus, twelve of the twenty-four opt-in rules produce no
finding anywhere. They name phrase families belonging to an assistant
register: a self-reference, a "fast-paced world" opening, an offer of
follow-up questions. `docs/rule-defaults.md` records each one.

That silence is not evidence the prose is clean.

## Third result: length, role and model family outrank origin

| Factor | Range observed |
| --- | --- |
| Length | 1.0% of documents under 50 words carry a finding; 91.5% above 1200 do |
| Role | Human specification prose 19.5 findings per thousand words; human release notes 0.7 |
| Model family | 8.9 per thousand words for one family against 1.4 for another, same tasks, same prompts |
| Origin | Human comments 7.1 per thousand words; generated comments 6.4 |

`research/diagnostics/2026-09-15-document-level-behavior.md` carries the
tables. The ordering is the practical finding of this work. What a reader
receives from this catalog follows the length of the text, then its kind,
then the model that wrote it. Whether a model wrote it comes last.

## Fourth result: the signal that survived is in the wording

The origin channel separates generated documentation from human code
comments on a held-out partition at recall 0.875 and a false-flag rate of
0.052, with a Brier interval excluding its constant baseline. It reached
that on surface n-grams, after three nulls on fourteen prepared features
that turned out to measure length. `research/origin/` holds the runs and the
shipped pack.

That is a group-level separation on declared provenance. It is not an
authorship claim about any document, and ADR 0037 keeps it out of the
editorial gate.

## Negative and null results, stated

- Both frozen hypotheses: not supported.
- The fourteen prepared features: three consecutive nulls; the fit was a
  length threshold.
- `syntax.long-sentence`: measured at a third of the human specification
  rate while producing 449 of 489 findings a shipped profile gave on one
  generated tree. Its score weight fell from 30 to 20 as a result.
- Twelve catalog rules: no occurrence in 5.6 million measured words.
- Three section-scope rules: no occurrence even on a million words of
  heading-bearing prose.

## Sensitivity, holdout and ablation

The run measures the placebo boundaries at 2012, 2016 and 2018 alongside the
main historical boundary. A first-appearance filter drops a later cohort's
units whose text an earlier snapshot of the same repository already holds.
The Qwen family entered no feature selection and has its own numbers. The origin work ran a declared ablation on the repetition family
and a matched-band diagnostic that exposed the length threshold.

## Limitations

- One language, one broad domain, sources chosen by license and history
  availability. Not a cross-section of English.
- `contemporary` origin is unknown by construction.
- No per-passage labels of pattern presence, so no precision or recall of
  unwanted findings and no editorial false-positive rate.
- The confirmatory measurement ran with 16 components against a minimum of
  20. Whether it spent the one permitted measurement is recorded as a
  maintainer's decision rather than settled here.
- Every measurement before September 15, 2026 was taken with an extraction
  that read example code inside documentation markup as prose.

## What changed in the product

| Change | Evidence |
| --- | --- |
| `format.em-dash-density` on in the strict profile | The one measure elevated in generated prose, 11.3 em dashes per thousand words in one generated tree against 0.0 in human specifications |
| `syntax.long-sentence` weight 30 to 20 | It decided gates while running below the human rate |
| `syntax.paired-contrast-density` widened from one phrase pair to its construction family | Two contrasts in a paragraph appeared in none of 29,063 human paragraphs |
| `corpus propose` added | A fixed catalog cannot anticipate a generator's habits; this learns a tree's own |
| Every opt-in rule carries its reason | The default split was inherited rather than chosen |

## Reproduction

```
bash scripts/measure-corpus.sh          # rebuild the measurement
python3 scripts/reproduce-tables.py     # rebuild the tables of this report
```

The evidence cards and the confirmatory result rebuild from committed
records alone. The corpus tables need the measurement, which is not
committed for its size.

## Conclusion

Prevalence differences of individual named rules between generated and human
technical English are weak, and the two this project froze did not survive a
held-out partition. A group-level separation in the wording did survive, and
so did a per-project method for naming one generator's habits. For a linter
the useful result is the ordering. Length, role and model family govern what
a reader receives. A catalog built around a 2023-era assistant register finds
nothing in what current models write.
