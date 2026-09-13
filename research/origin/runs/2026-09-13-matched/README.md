# Origin fit at matched length, 2026-09-13

A null under its own stopping rule, and the number that matters is the
false-flag rate on real code comments: 0.261.

`preregistration.md` was written after the long-form run was corrected
and before this fit ran. It fixes the band, the features, the estimator,
the threshold, the metric order and a stopping rule with two conditions.
`digests.json` names the tool commit and every artifact. Nothing here
qualifies an origin model.

## Why the band

The [long-form run](../2026-09-13-longform/README.md) reached recall
0.372 by learning a threshold on unit length near 45 words. Its two arms
differed in length as well as in origin, so counting words bought recall.
This trial removes that route: no unit outside 25 to 89 words enters the
fit, the calibration or the scoring. The band is where both arms carry
mass. Everything else is held: the same fourteen features, logistic with
isotonic calibration, the threshold frozen at 0.5, the same frozen pins.

The band costs most of the corpus. Training keeps 378 rows of 1,824,
calibration 124 of 863, and the development partition offers 721 eligible
units against 3,467 before.

## Result on the development partition

Metrics in the order the pre-registration fixed.

| Metric | Value | Interval |
| --- | --- | --- |
| False-flag rate on code comments | 0.261 | [0.115, 0.415] |
| Recall | 0.655 | [0.531, 0.964] |
| Brier | 0.235 | [0.081, 0.335] |
| Constant baseline | 0.324 | [0.174, 0.394] |

Precision is 0.758 and coverage 0.968.

## The stopping rule ends the trial

Both conditions fail. The false-flag rate of 0.261 is well above the
declared limit of 0.10, and the Brier interval [0.081, 0.335] still
contains the constant baseline of 0.324. The confirmation partition stays
sealed and `plan-final_test.json` was never written.

## The band narrowed the fit's reliance on length without removing it

Inside the band the fit still tracks length, in both arms together:

| Unit words | Generated | Code comments | Recall | False-flag rate |
| --- | --- | --- | --- | --- |
| 25 to 39 | 178 | 173 | 0.129 | 0.116 |
| 40 to 59 | 327 | 89 | 0.826 | 0.371 |
| 60 to 89 | 165 | 48 | 0.982 | 0.583 |

Recall and the false-flag rate rise together across the three bands,
which is the signature of a length rule, not of a prose rule.

There is a real gain over the unbanded fit, and it is small. In the 40 to
59 band the long-form fit paired recall 0.866 with a false-flag rate of
0.629; this one pairs 0.826 with 0.371. Fitting inside the band cuts the
false flags at that length by two fifths at nearly the same recall. That
is the size of the prose signal the fourteen features can reach.

By language, on real comments inside the band:

| Language | Comments | Flagged | Rate |
| --- | --- | --- | --- |
| C# | 87 | 51 | 0.586 |
| C++ | 62 | 9 | 0.145 |
| Java | 53 | 9 | 0.170 |
| Rust | 37 | 2 | 0.054 |
| Go | 29 | 3 | 0.103 |
| JavaScript | 24 | 4 | 0.167 |
| C | 18 | 3 | 0.167 |

C# stays the worst case at 0.586, and the band does not fix it. Its long
doc comments are the ones the fit is least able to tell from generated
documentation.

## What this establishes

That the fourteen prepared features do not separate generated
documentation from human code comments at matched length. They carry a
signal, measurable as the drop from 0.629 to 0.371 false flags at equal
recall, and it is far too weak to act on. One comment in four would be
flagged wrongly, and in C# nearer three in five.

It does not mean no such signal exists. It means this feature set does
not carry enough of it, and the next step is a different representation
of the prose rather than a larger corpus or another estimator.
