# Length-matched origin trial: what is declared before any label is opened

Written after the 2026-09-13 long-form run was measured and corrected, and
before any length-matched fit ran. Nothing below may change after a number is
seen.

## The question

The long-form run reached recall 0.372 at a false-positive rate of 0.047 on
declared provenance, and its own diagnostic shows why: the decision is a
threshold on unit length near 45 words. Below 25 words neither arm is ever
flagged. Above 60 words both arms always are. In between, recall of 0.866 comes
with a false-flag rate of 0.629 on real code comments.

So the question the earlier trials meant to ask is still open. Does generated
documentation prose differ from human comment prose when the two are the same
length?

## The design

One change from the long-form trial: every unit entering the fit, the
calibration and the scoring must fall in a fixed word band. Everything else is
held.

- **Band.** 25 to 89 words inclusive, fixed here. It is where both arms carry
  mass in the long-form run: 685 generated units and 316 code-comment units in
  the development partition alone. Below it the generated arm is fragments of
  its own responses; above it the historical arm nearly vanishes.
- **Features.** The same fourteen. No addition, no ablation.
- **Estimator and calibration.** Logistic with isotonic calibration, as before.
- **Threshold.** 0.5, frozen before scoring, as before.
- **Partitions.** The same frozen pins. Development is scored first and alone.

## The reference points, named in advance

1. A coin. With the arms matched on length and the band fixed, the honest null
   is the training prevalence, and the constant baseline is what it must beat.
2. The long-form run's in-band numbers, which are what a length threshold
   achieves inside the band: recall 0.866, false-flag rate 0.629 in 40 to 59
   words, and recall 1.000 with false-flag rate 1.000 above 60. A length-matched
   fit that reproduces those has learned nothing new.

## The metrics, in this order

1. False-flag rate on real code comments at the frozen threshold. This is first
   because it is the number that decides whether a tool can ship.
2. Recall.
3. Brier against the constant baseline, with its interval.
4. The in-band flag rate of each arm, reported separately by language.

## The stopping rule

The confirmation partition is opened only if the development fit beats the
constant baseline on Brier with an interval that excludes it, and its false-flag
rate on real code comments is below 0.10. Either one failing ends the trial and
it is reported as a null. The long-form run failed the first of these; this
trial adds the second because a rate above 0.10 is useless whatever the Brier
says.

## What a null would mean

That the fourteen prepared features do not separate generated documentation
from human comment prose at matched length. It would not mean no such signal
exists. It would mean this feature set does not carry it, and that the next step
is a different representation rather than a larger corpus.
