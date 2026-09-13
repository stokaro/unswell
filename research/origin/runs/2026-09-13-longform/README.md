# Origin fit on the long-form cohort, 2026-09-13

The origin task fitted again after the controlled cohort gained responses
that are four times longer than any earlier run's. It is the first origin
run whose development result is not a null.

`preregistration.md` was written before the corpus finished measuring and
before any fit ran. It names the reference points, the primary trial, the
one permitted ablation, the metric order, and the stopping rule. Its two
amendments record the corpus and calibration changes the run needed, and
both were written before a number was seen. `digests.json` names the tool
commit, `ac2716b`, and the digest and size of every artifact.
`union.json` and `union-plan.json` reproduce the corpus from the
acquisition work directory. Nothing here qualifies an origin model.

## Corpus

The controlled arm is the four long-form runs of 2026-09-13 only. The
eight short-form runs of September 11 and 12 are excluded by source ID,
listed in `excluded-short-form.txt`. Mixing them would have let a fit
separate the arms by counting words, and their union also exceeded the
16 MiB annotation input limit.

| Count | Value |
| --- | --- |
| Sources | 2,616 |
| Paragraph units | 6,981: 2,208 controlled, 4,773 historical |
| Groups | 12 connected components in the development partition |
| Training | 1,816 fitted rows: 144 `endpoint_generated`, 1,672 `human_snapshot` |
| Calibration | 497 rows: 102 endpoints, 395 snapshots |
| Development | 4,213 candidates, 3,467 eligible after polished responses are dropped |

The controlled responses average 103 words, against 24 in every earlier
run. Their paragraph units average 32 words, because a response of that
length breaks into several paragraphs. The historical units average 13.

## Result on the development partition

The primary trial is the fourteen features of `scripts/origin-experiment.sh`
unchanged, a logistic fit with isotonic calibration, and the threshold
frozen at 0.5 before scoring.

| Metric | 2026-09-11 enlarged | This run |
| --- | --- | --- |
| Recall | 0 | 0.372 |
| False-positive rate | 0 | 0.047 |
| Precision | undefined | 0.688 |
| Brier | 0.0451 | 0.148 |
| Constant baseline | 0.0490 | 0.189 |
| Largest calibrated response | 0.110 | 0.852 |

The Brier values are not comparable across the two runs: prevalence
differs, so each is only meaningful against its own constant. What
carries over is the ratio. The earlier run scored 0.92 of its constant;
this one scores 0.79.

Raw length does not drive the fit. On standardized inputs the two length
features carry the smallest weights of the fourteen: 0.006 for
prose-words and -0.018 for counted-characters. The largest are
mean-sentence-words at 0.119, hapax-token-ratio at 0.113 and
noun-token-ratio at 0.096. The separation comes from sentence shape and
vocabulary, not from the units being longer.

## The stopping rule held, so the confirmation partition stays sealed

The pre-registration admits the confirmation partition only when the
development Brier interval excludes the constant baseline. It does not.
The cluster bootstrap over 12 components gives [0.021, 0.353] for the
Brier and [0.029, 0.437] for the constant, and 0.189 falls inside the
first. The run therefore stops on the development partition and reports
what it found there. The harness gained `--score-partition` so that a run
can score one partition and stop; `plan-final_test.json` was never
written.

Twelve components is the floor the protocol sets, and an interval built
on twelve clusters is wide whatever the units say. Narrowing it needs
more pinned repositories in the development partition, not a different
estimator.

## The declared ablation

`ablation/` holds the same fit plus the three repetition features:
repeated-bigram-ratio, sentence-opener-repeat-ratio and
duplicate-sentence-ratio. It was named in the pre-registration and is the
only ablation reported.

| Metric | Primary | With repetition |
| --- | --- | --- |
| Recall | 0.372 | 0.331 |
| False-positive rate | 0.047 | 0.042 |
| Precision | 0.688 | 0.706 |
| Coverage | 0.985 | 0.847 |
| Brier against constant | 0.148 / 0.189 | 0.161 / 0.204 |
| Largest calibrated response | 0.852 | 0.986 |

The repetition family abstains on a paragraph of one sentence, which
drops coverage from 0.985 to 0.847 and costs 532 units. It buys a higher
maximum response and slightly better precision, and loses recall. It does
not replace the primary trial.

## What this does and does not establish

It establishes that generated documentation of about 100 words carries
structure a fourteen-feature fit can find on declared provenance, where
the same fit on 24-word responses found none. Unit length was the binding
constraint, as the pilot's record suspected.

It does not establish a false-positive rate for a product feature. That
needs the human-labeled corpus of issue 22 and the published evaluation
of issue 25, both on hold. It does not report a confirmation result,
because the stopping rule did not admit one.
