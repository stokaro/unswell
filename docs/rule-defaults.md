# Which rules ship enabled

The catalog holds 60 rules. Thirty-three are enabled by their descriptors and 27
are opt-in; strict additionally enables em-dash density. This page says how that split was chosen and what each opt-in rule was
measured at. Every rule also carries its own reason: `unswell rules show ID`
prints it, the rules page shows it, and a test fails if an opt-in rule has
none.

Five local constructions now request an edit review: document self-justification,
evaluative closing clauses, circular definitions, bare verification slogans,
and unscoped assurances.
They have warning severity, weight 12, cap 24, and no rule prohibition in both
profiles. The [Ptah evaluation](research/ptah-rhetoric.md) records their source
cases and review load. The [framing recall follow-up](../research/reviews/2026-09-17-framing-recall/README.md)
records seven additional development detections and no confirmation gain. The weights are editorial policy, not measured
probabilities. The ordinary index gate can still pass a page with these warnings.

The complete-list-item duplicate rule also ships as an experimental warning,
weight 15, cap 30, with no unconditional gate. It preserves issue references and
identifier spelling within one list. [Repetition scope](repetition-signals.md#comparison-scope-and-short-complete-items)
documents its boundaries and evidence.

Three [local repetition checks](repetition-signals.md#short-local-repetitions)
also ship as experimental warnings. Their
[complete-page measurement](../research/reviews/2026-09-17-local-repetition/README.md)
records gains on known examples and no confirmation gain. These results do not
qualify their defaults or establish broad editorial recall.

## Qualification on the assistant-panel corpus

[ADR 0047](adr/0047-qualified-defaults-and-assistant-review.md) adds a second
kind of evidence: two assistants from different model families judge each
finding under the statement of its rule. The
[qualification record](research/rule-qualification.md) holds the tables.
Three defaults changed on that evidence.

- `hype.vague-praise` ships enabled. Its list was three phrases that no measured
  text contains. It now also holds words and phrases whose findings both
  reviewers accepted on development pages.
- `filler.wordy-phrase` and `hype.modifier-cluster` gained qualified entries.
- `repetition.sentence-openers` and `repetition.paragraph-openers` are opt-in.
  Both reviewers rejected most of their findings as deliberate parallel wording.

A phrase or rule without reviewed findings stays unlisted or opt-in.

[ADR 0048](adr/0048-severity-follows-review.md) ties default severity to the same
evidence. A rule reports warnings when both reviewers accepted at least 85% of 50
or more findings on real pages, and notes when it falls below that share.
`syntax.long-sentence`, `filler.instruction-scaffolding`, `hype.vague-praise`,
and `repetition.adjacent-word` report notes.

[ADR 0049](adr/0049-defaults-follow-reviewed-forms.md) reads the same reviews
inside a rule. A rule, a matcher form, or a block kind leaves the default when
both reviewers accepted fewer than half of 30 or more of its findings. A length
threshold sits at the lowest five-word step from which they accepted at least
75%.

- `syntax.long-sentence` starts above 45 words and does not measure table
  cells. Set `onset: 35` to restore the earlier threshold.
- `filler.instruction-scaffolding` no longer reports a conditional reader goal
  or a reminder that opens with `Be sure to`.
- `repetition.adjacent-word` requires both copies in the same case with no
  markup between them.
- `filler.document-metadiscourse` and `filler.unscoped-assurance` are opt-in.
  Both reviewers accepted 45% of 133 and 29% of 48 of their findings.

The share of ADR 0047 is read again whenever new pages are reviewed.
[ADR 0050](adr/0050-listed-phrases-leave-on-strong-evidence.md) says when a
listed family leaves: its share is below 80%, and the upper end of the 95%
interval is below 80% as well. Two families have left: `state-of-the-art` in
`hype.vague-praise`, and `foster` with its forms in `filler.wordy-phrase`.

## Tested on what people remove from AI-sounding documentation

[ADR 0056](adr/0056-rules-tested-on-de-slop-commits.md) adds a reference in the
tool's own domain: Markdown files before and after commits whose message says
that they remove AI-sounding wording. A rule enters the defaults there when
its findings are lower after the commit in at least 65% of the test pairs that
differ, with the lower end of the interval above 55%.

| Rule | What it reports | Lower after the commit | Default |
| --- | --- | --- | --- |
| `format.decorative-emoji` | Emoji in headings and list items | 77.7% (69.8% to 84.0%) | On |
| `format.em-dash-frequency` | A document with at least three em dashes and at least three per 1,000 words | 70.4% (64.1% to 76.0%) | On |
| `format.em-dash-density` | A paragraph dense with em dashes | 62.5% (51.0% to 72.8%) | Off outside the strict profile |

With the two new rules the default findings are lower after the commit in
72.7% (67.8% to 77.1%) of the pairs that differ, and the defaults
fire in 365 of 597 pairs. The defaults of v0.1.0-alpha.9 give
64.0% (56.1% to 71.2%) and fire in 158.

The em dash rule reports 24.1% of the pages of today's public repositories
and 1.7% of the pages that people wrote before November 2022. It has no gate.
Raise `onset` or `allowed_occurrences` to report fewer documents.

## Listing on what editors removed

[ADR 0055](adr/0055-phrase-listed-on-editor-reference.md) adds a second route
into a default list: what human editors did. `scaffold.chat-preamble` lists
`Sure,` and `Sure!` on that route, at the start of a document and in the first
sentence after its leading headings. On answers of chat models that experts
edited, the experts removed or replaced the flagged opener in 39 of 39 test
answers. The rule reports nothing on 3,302 pages of public repositories.
It does not look after later headings, where a person may answer a question
with `Sure,`. In source code the first comment or string of a file counts as
the start: the rule reports nothing on 265,025 source files of open-source
packages, and a file of chat replies that opens with a listed phrase belongs
in the extraction policy. The
[human-reference record](../research/reviews/2026-11-12-human-reference/RESULTS-words.md)
holds the counts and a table of every rule against the words editors changed.

## The decision rule

A rule ships enabled when it names a construction a reader can act on and its
warning load on ordinary technical prose is one a project can live with.

A rule stays opt-in for one of four recorded reasons.

1. It is a surface measurement. Such a rule describes a text; it does not
   name a defect, and its own limitations say so. `format.em-dash-density` is
   the one exception, in the strict profile alone, because it is the one
   measure elevated in generated prose.
2. It produced no finding on any measured corpus. Enabling it would add
   nothing and would suggest a coverage the measurement does not support.
3. Its warning load falls mainly on human technical prose. The rule is
   editorially sound; a project that wants that load asks for it.
4. It was measured against a frozen hypothesis and the association did not
   carry to unseen text.

A rule is not removed for failing to separate generated prose from human
prose. Separation is not what an editorial rule is for.

## Earlier measurements

The tables below preserve the earlier engine's measurements. They do not describe
the expanded version 2 contrast matcher, its corrected inline-code handling,
or the two construction rules added for #294.
The [contextual prose evaluation](research/contextual-prose.md) records that change
against pinned complete pages and a separate historical development sample.

Three corpora, every rule enabled through `research/acquisition/policy-e1.yaml`.

| Corpus | Words | What it is |
| --- | ---: | --- |
| Human | 3,603,135 | The historical cohort of the pattern protocol: code comments, documentation, specifications and release notes from pinned open-source snapshots |
| Generated | 426,460 | The controlled arm: model output under recorded prompts, two families |
| Documentation tree | 411,096 | `stokaro/ptah` at `cf75c79d`, stated by its author to be model-written and never proofread |

Rates below are findings per thousand prose words.

Nineteen of the 40 rules fire at all on the human and generated corpora, and
15 on the documentation tree. Twenty-one produce nothing anywhere. On the
documentation tree, 2,968 findings are warnings and none reaches forbid
severity, with every rule enabled.

## Opt-in rules

### Surface measurements

| Rule | Human | Generated | Note |
| --- | ---: | ---: | --- |
| `format.em-dash-density` | 0.001 | 0.070 | On in strict. The one measure elevated in generated prose |
| `readability.grade-metric` | 0.711 | 2.964 | Frozen hypothesis H2: +0.32 points on confirmation, p 0.747 |
| `format.list-fragmentation` | 0.000 | 0.000 | Never fires |

### No occurrence in any measured corpus

`filler.empty-transition`, `filler.section-announcement`,
`filler.stacked-hedging`, `hype.absolute-claim`, `hype.metaphor-cluster`,
`repetition.heading-echo`, `repetition.summary-echo`,
`syntax.rhetorical-question-density`, `syntax.triad-density`,
`syntax.whether-preface-density`, and `syntax.nominalization-chain` at 0.001
in human prose and none in the generated arm.

Several of these name a phrase family where a construction was meant, and
their phrases belong to a register the measured generators do not write.
Issue 268 records that finding. The rules stay in the catalog, off by
default, for a project that does hit them.

### Warning load falling mainly on human prose

| Rule | Human | Generated |
| --- | ---: | ---: |
| `repetition.ngram-density` | 0.603 | 0.056 |
| `repetition.paragraph-overlap` | 0.283 | 0.000 |
| `readability.long-paragraph` | 0.177 | 0.040 |
| `syntax.parenthetical-load` | 0.499 | 0.300 |
| `filler.weak-intensifiers` | 0.019 | 0.007 |
| `repetition.syntax-template` | 0.010 | 0.000 |

### Measured against a frozen hypothesis

| Rule | Human | Generated | Result |
| --- | ---: | ---: | --- |
| `syntax.noun-stack` | 0.095 | 0.213 | Frozen H1: +2.96 points on development, +0.41 on confirmation, p 0.727 |
| `syntax.passive-candidate-density` | 1.112 | 1.953 | Names candidates without a dependency parse |

### Repeated contrasts: opt-in surface measurement

Version 1 of `syntax.paired-contrast-density` fired at 0.007 in human prose, not at all in
the controlled arm, and at 0.161 on the documentation tree. One generator
under one set of prompts produced that; the measurement says nothing about
generated text in general. That matcher also discarded entire sentences containing
inline code. Version 2 preserves their surrounding prose and adds `X, not Y` to its
declared contrast frames. Version 3 retains that matcher and its allowance, but
leaves it disabled in every builtin profile. Explicit configuration still enables
it as a note with zero weight, zero cap, and `gate: none`.

On the [36 complete exposed development pages](../research/reviews/2026-10-04-contrast-defaults/README.md),
the unchanged frozen assistant judgments accept one of its 30 findings, reject
27, and leave two uncertain. Ordinary technical alternatives dominate the review
load. Frequency alone does not identify the wording to revise. Disabling the
measurement also removes the one accepted finding; this is a precision decision,
not a recall gain or product qualification. All original events and judgments
remain in the research scope. Projects can request the measurement explicitly;
the matcher, allowance, and gate thresholds are unchanged.

### Denial/redefinition and document self-description

`filler.document-metadiscourse` is opt-in since ADR 0049; the paragraph below
describes the rule when a project enables it.

`syntax.repeated-reframing` and `filler.document-metadiscourse` expose complete
clauses in technical and strict, with zero score and `gate: none`. They close the
[recognition gap](research/rhetorical-frames.md) demonstrated on a declared
constructed probe and a pinned Ptah page. One ordinary denial/redefinition stays
below the allowance. A document self-description is visible even when it occurs
once; it may provide useful navigation. These are unqualified advisory defaults,
not measured defect detectors. The historical tables above do not evaluate them.

## How to turn one on

```yaml
version: 1
extends: [builtin:strict-v1]
rules:
  repetition.ngram-density: {enabled: true}
```

`research/acquisition/policy-e1.yaml` turns on all 40 and is what produced
the numbers here.

## Reproducing the numbers

The human and generated rates come from `artifacts/measurement`, rebuilt by
`bash scripts/measure-corpus.sh`. The documentation tree is one scan:

```
git clone --depth 1 https://github.com/stokaro/ptah.git
cp research/acquisition/policy-e1.yaml ptah/.unswell-e1.yaml
cd ptah && unswell check --config .unswell-e1.yaml --report json:out.json docs
```
