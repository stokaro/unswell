# Compact models do not solve whole-page recall

The strongest of four fixed Go comparisons retrieved 74 of 872 positive
prepared units (8.5%) on page groups excluded from its training and threshold
selection. It selected 12 of 1,234 explicitly clean units and another 100 units
whose quality was not labeled. This is too little coverage to become a useful
general editing diagnostic. No product rule, model pack, profile or gate changes.

These are exposed assistant judgments used for development under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md). They are not
human annotations, new confirmation or a probability qualification set.
The [protocol](protocol.md) and [split plan](split-plan.json) were fixed before
the first fit. The [representation audit](../2026-09-19-representation-audit/README.md)
binds every label to existing archived source bytes.

## Results

All models use normalized regularized logistic regression implemented in Go.
Vocabulary selection and normalization use only training groups. A separate
development fold chooses the threshold with the greatest recall subject to
explicit-control FPR at most 1% and precision at least 85%. The remaining fold
is evaluated once. This five-fold comparison is internal development evidence.

| Inputs | Positive units selected / 872 | Clean units selected / 1,234 | Labeled precision | Unlabeled units selected / 27,644 | Retrieved event targets / 994 |
| --- | ---: | ---: | ---: | ---: | ---: |
| L: length | 23 (2.6%) | 7 (0.57%) | 76.7% | 44 | 27 |
| S: shared structural, POS and repetition features | 7 (0.8%) | 5 (0.41%) | 58.3% | 17 | 8 |
| W: length and word n-grams | 65 (7.5%) | 11 (0.89%) | 85.5% | 99 | 77 |
| SW: shared features and word n-grams | 74 (8.5%) | 12 (0.97%) | 86.0% | 100 | 81 |

S had no qualifying development threshold in three of five folds. Those folds
remain in the denominator and make no selections. The operating constraints
are not guaranteed on the evaluation fold: SW exceeds 1% control FPR in two
folds. Its historical cohort has 11/533 false selections (2.06%) and 35/525
positive selections (6.7%); Ptah has 1/701 and 39/347 (11.2%), respectively.
Aggregate rates hide this difference. The [summary](summary.json) retains every
fold, cohort and length band, including all unavailable operating points.

Event retrieval requires selection of every prepared piece touched by an
event's eligible target prose. It only measures whether an editor would be
shown that target. The model has not identified its defect, offered a supported
explanation or distinguished the edit from neighboring technical content.
Therefore 81/994 is **not diagnostic recall**. Unlabeled selections are neither
false positives nor verified useful findings. Explicit-control FPR is not the
FPR of a complete production document stream.

The retained [rule evaluation](../2026-09-18-structural-wordiness/summary.json)
uses source-bound full-event judgments. Its length warnings are not credited
for other defects merely because they overlap. Those diagnostic measurements
and the retrieval counts above have different meanings and cannot be ranked
as interchangeable recall estimates. That earlier rule run used engine
`d15cc2055753c58d534072ec4e71e81e5de0b899`; this numerical experiment uses the
shared features at `fa1743891a3e82d2d5361f41762604767b497b8f`.

The [bound comparison](rule-reference.json) preserves 144/994 full diagnostic
credits from that rule review. Of SW's 81 retrieved targets, 16 were already
diagnosed and 65 had been missed. Those 65 are candidates for further analysis;
this generic selection adds zero supported diagnoses to the measured total.

## Limits and decision

The 157 pages form 146 groups after merging long exact repeated passages.
There are still 588 shorter repeated text hashes; grouping does not establish
independence of short templates or paraphrases. The deliberately reviewed
positive and control units also have different length distributions. L exposes
one consequence of that selection, but it cannot remove the bias. These rates
describe these finite development labels; they are not population estimates.

The representation audit found that 920/994 targets fit an existing prepared
piece. Here SW retrieved only 81 event targets. Accessible source bytes and
generic numerical features do not by themselves produce good diagnostics.
Pieces separated by protected code also lose surrounding context in these
models, and a lexical vector cannot establish whether an assertion contributes
information or repeats an earlier promise.

Retain all four models as research baselines. Do not promote SW, relabel misses
or relax the product gate. The next candidate needs context connecting prose
pieces and sentences, with source-specific explanations that distinguish
technical consequences from redundant framing. Any useful candidate still
needs fresh complete-page confirmation frozen before inspecting its outputs.
The broad recall objective remains open.

## Reproduction

The research command fits in Go and never downloads resources. It is separate
from the product runtime and does not emit an installable model pack.
Python assembles already reviewed inputs and verifies saved outputs.

```sh
cd research/annotation
CGO_ENABLED=0 go build -o /tmp/unswell-reviewbaseline ./cmd/reviewbaseline
cd ../..
PYTHONDONTWRITEBYTECODE=1 python3 research/reviews/2026-09-19-learned-proposals/tools/check_evidence.py \
  --rerun /tmp/unswell-reviewbaseline
```

Run on the recorded runtime base with the archived trainer sources to reproduce
the recorded experiment. Later engine or feature changes must not overwrite
these artifacts. The verifier rebuilds input and grouping from the audit,
checks source identities and model outputs, recalculates the summary, and
compares a new Go fit byte for byte with all retained predictions and parameters.
The original run completed in 8.40 seconds on the recorded macOS/arm64 host;
this is an observation, not a product performance qualification.

[`measurement-record.json`](measurement-record.json) records the base commit,
compiler, executable, source archive, input, protocol and prediction hashes.
[`trainer-source.tar.gz`](trainer-source.tar.gz) retains the exact Go sources.
[`predictions.json.gz`](predictions.json.gz) contains all 20 fitted models and
all 119,000 evaluation rows, including unlabeled units. Raw linear scores are
not percentages. [`validation.json`](validation.json) records the independent
artifact checks and byte-identical repeated fit.

Blackbox Go tests cover evaluation-label and vocabulary isolation, page-order
invariance, tied scores, cancellation and invalid source bindings. Python
evidence tests reject changed groups, labels, thresholds, missing predictions
and nonfinite values, and require every event target piece for retrieval credit.
