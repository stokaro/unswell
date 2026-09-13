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

The two raw length features carry the smallest standardized weights of
the fourteen: 0.006 for prose-words and -0.018 for counted-characters.
That is not evidence that length does not drive the fit, and the section
below shows it does. The remaining twelve features all rise with the
length of the unit, and together they reproduce a length threshold.

## What the fit actually does, measured on code

Every unit in this trial carries the role `comment`. The negatives are
real comments in nine languages, and the positives are generated
replacements for comments of the same declarations, so the numbers above
are already a statement about code. Broken out, they say something the
headline hides.

The false-flag rate on real code comments is not one number. It ranges
over a factor of forty by language:

| Language | Comments | Flagged | Rate |
| --- | --- | --- | --- |
| Rust | 1,420 | 8 | 0.006 |
| JavaScript | 135 | 3 | 0.022 |
| Go | 241 | 6 | 0.025 |
| C | 122 | 4 | 0.033 |
| C++ | 332 | 11 | 0.033 |
| Java | 180 | 27 | 0.150 |
| C# | 245 | 66 | 0.269 |
| All | 2,675 | 125 | 0.047 |

The spread is not a property of the languages. Flagged comments average
54 to 68 words in every one of them, and clean comments average 7 to 15.
C# and Java sit at the top because they carry more long doc comments, not
because their prose differs.

Holding length fixed removes the separation entirely. Inside each band
the two arms are scored the same way:

| Unit words | Generated | Code comments | Recall | False-flag rate |
| --- | --- | --- | --- | --- |
| under 15 | 588 | 1,996 | 0.000 | 0.000 |
| 15 to 24 | 169 | 351 | 0.000 | 0.003 |
| 25 to 39 | 178 | 175 | 0.062 | 0.023 |
| 40 to 59 | 328 | 89 | 0.866 | 0.629 |
| 60 to 89 | 179 | 52 | 1.000 | 1.000 |
| 90 and over | 21 | 12 | 1.000 | 1.000 |

No unit under 23 words is ever flagged and no unit over 59 words is ever
clean, in the primary fit and in the ablation alike. The decision is a
threshold on unit length placed near 45 words. `length-diagnostic.json`
holds both tables.

The development recall of 0.372 therefore measures how much more of the
generated arm is long, not how much of it reads as generated. The
false-positive rate of 0.047 is low for the same reason in reverse: most
real comments are short enough to fall under the threshold. On the
comments that a writer would actually ask about, the ones long enough to
carry prose, the rate is 0.629.

This is a correction to the claim two sections above, and it is the
finding of the run. Length was the binding constraint on the earlier
nulls, and making the arms differ in length is what produced a non-null
number. It did not produce a detector.

A trial that can answer the original question has to match the two arms
on length before fitting, so that the classifier cannot buy recall by
counting words. That trial needs its own pre-registration and is not
reported here.

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

It establishes that unit length was the binding constraint on the earlier
nulls: making the controlled arm longer is what moved recall off zero. It
does not establish that the fit reads prose. Held at a fixed length, it
separates nothing, and what it learned is a threshold near 45 words.

It does not establish a false-positive rate for a product feature. That
needs the human-labeled corpus of issue 22 and the published evaluation
of issue 25, both on hold. It does not report a confirmation result,
because the stopping rule did not admit one.
