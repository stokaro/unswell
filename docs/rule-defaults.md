# Which rules ship enabled

The catalog holds 46 rules. Twenty-three are on in the shipped profiles and 23 are
opt-in. This page says how that split was chosen and what each opt-in rule was
measured at. Every rule also carries its own reason: `unswell rules show ID`
prints it, the rules page shows it, and a test fails if an opt-in rule has
none.

Four local constructions now request an edit review: document self-justification,
evaluative closing clauses, circular definitions, and bare verification slogans.
They have warning severity, weight 12, cap 24, and no rule prohibition in both
profiles. The [Ptah evaluation](research/ptah-rhetoric.md) records their source
cases and review load. The weights are editorial policy, not measured
probabilities. The ordinary index gate can still pass a page with these warnings.

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
`hype.vague-praise`, `repetition.heading-echo`, `repetition.summary-echo`,
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

### Repeated contrasts: advisory by default

Version 1 of `syntax.paired-contrast-density` fired at 0.007 in human prose, not at all in
the controlled arm, and at 0.161 on the documentation tree. One generator
under one set of prompts produced that; the measurement says nothing about
generated text in general. That matcher also discarded entire sentences containing
inline code. Version 2 preserves their surrounding prose and adds `X, not Y` to its
declared contrast frames. It ships in technical and strict as a note with zero
weight, zero cap, and `gate: none`. One contrast remains below the allowance.
The note exposes repetition without claiming redundant facts or established
association with a generator. Minimal and custom leave it disabled.

### Denial/redefinition and document self-description

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
