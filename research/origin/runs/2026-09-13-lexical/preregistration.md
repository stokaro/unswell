# Lexical n-gram origin trial: what is declared before any label is opened

Written after the matched-length fit was scored and recorded, and before this
one is scored. Nothing below may change after a number is seen.

## The question

The matched-length trial ended in a null. Its record names the reason and the
next step: the fourteen prepared features are fourteen aggregate statistics of a
unit, and aggregate statistics of a comment are close to a measure of how long
it is. What they cannot represent is the wording itself. The next step is a
different representation of the prose.

This trial takes the one the engine already carries. A lexical vocabulary is
fitted from the training targets alone: word n-grams of order 1 to 2 and
character n-grams of order 3 to 5, at most 128 retained columns. It represents
which strings occur, not how long the unit is.

## What is held and what changes

Held: the same corpus, the same word band of 25 to 89 words, the same frozen
pins, the same provenance labels, logistic with isotonic calibration, the
threshold frozen at 0.5, and the development partition scored alone.

Changed: the representation. No prepared feature enters. Columns are learned
from permitted, resolved training targets only, which is what the engine
enforces for a lexical fit.

## The reference point, named in advance

The matched-length trial on the same band and the same rows: false-flag rate
0.261 on real code comments, recall 0.655, Brier 0.235 against a constant of
0.324. This trial has to beat that to be worth anything, on the first metric
before the second.

## The metrics, in this order

1. False-flag rate on real code comments at the frozen threshold.
2. Recall.
3. Brier against the constant baseline, with its interval.
4. The flag rate inside each of the three length sub-bands, for both arms. A
   representation that has escaped the length rule shows a flat false-flag rate
   across the sub-bands; one that has not shows the rising pattern the prepared
   features show.

## The stopping rule

Unchanged from the matched-length trial. The confirmation partition is opened
only if the development fit beats the constant baseline on Brier with an
interval that excludes it, and its false-flag rate on real code comments is
below 0.10. Either one failing ends the trial and it is reported as a null.

## What a null would mean

That neither aggregate statistics nor surface n-grams, as this engine computes
them, separate generated documentation from human code comments at matched
length on this corpus. That would be worth recording plainly, because the two
are the only representations the engine has that need no external model, and it
would move the question to one that does.
