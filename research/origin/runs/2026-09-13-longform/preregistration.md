# Long-form origin trial: what is declared before any label is opened

Written before the enlarged corpus finished measuring and before any fit ran.
Nothing below may change after a number is seen.

## The question

The origin run of 2026-09-11 fitted a classifier on 5,608 paragraph units and
separated nothing: recall 0 on both partitions, maximum calibrated response
0.110, Brier 0.0210 against a constant baseline of 0.0261. Its own record notes
that of the 58 missed endpoints in the confirmation partition, 44 held fewer
than 20 words and 14 held 20 to 49.

This trial changes one thing: the controlled arm now also holds 1,600 responses
whose mean length is 107 words (Haiku) and 103 words (luna), against 24 words in
every earlier run. Everything else is held fixed.

## The reference points, named in advance

1. The 2026-09-11 enlarged run. Same features, same estimator, same calibration,
   same threshold. It is the benchmark this trial must beat to mean anything.
2. The constant baseline: predicting the training prevalence for every unit.
   The prior run did not beat it; its Brier interval contained the constant.
3. The frozen threshold of 0.5, fixed before scoring, as the prior run fixed it.

## The trials, in this order

- **Primary.** The fourteen features of `scripts/origin-experiment.sh`
  unchanged: prose-words, counted-characters, prose-sentences,
  mean-sentence-words, sentence-word-stddev, shortest-sentence,
  longest-sentence, type-token-ratio, hapax-token-ratio, noun-token-ratio,
  verb-token-ratio, adjective-token-ratio, adverb-token-ratio,
  automated-readability-index. Logistic estimator, isotonic calibration.
  Identical to the prior run so that only the corpus differs.
- **One declared ablation.** The same fourteen plus the repetition family. The
  harness excludes repetition today because it abstains on a paragraph of one
  sentence, "which most comments are". A 107-word paragraph carries several
  sentences, so those features become computable. This is the only ablation
  that will be reported, and it is declared here rather than chosen later.

No third trial. No feature added after seeing a result.

## The metrics, in this order

1. Recall at the frozen threshold.
2. False-positive rate.
3. Brier against the constant baseline, with its interval.
4. Maximum calibrated response.

## The stopping rule

The development partition is scored first. The confirmation partition is scored
only if the development result beats the constant baseline on Brier with an
interval that excludes it. A null on development ends the trial and is reported
as a null; spending the held-out partition on a result that already failed
would buy nothing and cost the partition.

## What a positive result would and would not mean

A model that separates the long-form controlled arm from historical prose shows
that generated documentation of that length carries measurable structure. It
would not establish a false-positive rate for a product feature. That needs the
human-labeled corpus of issue 22 and the published evaluation of issue 25, both
on hold. This trial reports separation on declared provenance, nothing more.

## Amendment, written before any fit ran

The union of every controlled source with the historical shards produces an
annotation input of 17,083,266 bytes. The engine refuses an input above
16,777,216 bytes, so the corpus as first described cannot be measured at all.

The controlled arm now holds ten runs. Eight of them are the short-form runs of
2026-09-11 and 2026-09-12, whose responses average 24 words; two are the
long-form runs of 2026-09-13. Keeping all ten would confound length with every
other property under test: a fit could separate the arms by counting words and
still say nothing about the prose. Restricting the controlled arm to the two
long-form runs removes that confound and brings the input under the limit.

So the controlled arm of this trial is the long-form runs only:
2026-09-13-long-haiku, 2026-09-13-long-luna, 2026-09-13-longtrain-haiku and
2026-09-13-longtrain-luna. The 6,400 short-form controlled sources are passed
as explicit exclusions, listed in `excluded-short-form.txt` beside this file.
The historical arm, the features, the estimator, the calibration, the threshold
and the metric order are unchanged.

This narrows what a positive result would mean. It would say that generated
documentation of about 100 words separates from historical prose. It would say
nothing about the 24-word case, which the 2026-09-11 run already reported as a
null.

## Second amendment, also written before any fit ran

The first attempt at the primary trial failed on a second count: isotonic
calibration had no samples. The fit uses four frozen partitions, and the
long-form controlled arm covered only two of them. Every long-form task came
from a repository pinned to development or training, so the calibration
partition held historical prose and nothing generated.

Thirty calibration-arm tasks were drawn to fill it, from the six repositories
pinned to calibration, at the same fifty-word floor the evaluation arm used and
under the same two prompts and two operations. Both generator families answered
all of them, so the calibration arm holds 240 responses.

Two facts about that arm limit what it can carry, and both are stated here
rather than after a result:

1. The pool is small. Only four of the six pinned repositories yielded an
   eligible unit, and one of them, PowerShell/PSScriptAnalyzer, contributes
   several copies of the same comment template. The calibration arm is
   therefore narrower in subject than the evaluation arm.
2. Its units are shorter. Calibration tasks average 60 words against 103 in the
   evaluation arm, because the eligible pool at the fifty-word floor held
   nothing longer.

Isotonic calibration fits a monotone map from score to probability, so a
narrower calibration arm moves the reported probabilities without changing the
ranking the classifier produces. Recall and the false-positive rate at the
frozen threshold do depend on that map, so this is a limitation of the trial,
not a detail of its bookkeeping.

The harness now takes `--score-partition`, so the development partition is
scored alone. The confirmation partition stays sealed until the stopping rule
above admits it.
