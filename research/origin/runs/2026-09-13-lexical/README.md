# Origin fit on surface n-grams at matched length, 2026-09-13

The first origin fit that separates generated documentation from human
code comments without counting words. On the development partition it
flags 4.0 percent of real code comments and recovers 74.4 percent of
generated units, against 26.1 and 65.5 for the prepared features on the
same rows.

`preregistration.md` was written after the matched-length trial was
recorded and before this one was scored. It fixes the representation, the
reference point, the metric order and the stopping rule.
`digests.json` names the tool commit and every artifact. Nothing here
qualifies an origin model.

## What changed from the matched-length trial

Only the representation. The corpus, the word band of 25 to 89 words, the
frozen pins, the provenance labels, the logistic estimator, the isotonic
calibration and the threshold of 0.5 are all held.

The prepared features are fourteen aggregate statistics of a unit, and
aggregate statistics of a comment are close to a measure of how long it
is. This fit uses none of them. A vocabulary is learned from the training
targets alone: word n-grams of order 1 to 2, character n-grams of order 3
to 5, at most 128 retained columns. It represents which strings occur.

## Result on the development partition

Metrics in the order the pre-registration fixed, with the matched-length
trial beside them on the same rows.

| Metric | Prepared features | Surface n-grams |
| --- | --- | --- |
| False-flag rate on code comments | 0.261 | 0.040 |
| Recall | 0.655 | 0.744 |
| Precision | 0.758 | 0.961 |
| Brier | 0.235 | 0.144 |
| Constant baseline | 0.324 | 0.331 |

The Brier interval is [0.014, 0.236] and excludes the constant. The
false-flag interval is [0.004, 0.087] and lies below the declared limit
of 0.10.

## It is no longer a length rule

The matched-length trial still tracked length inside its own band: the
false-flag rate rose from 0.116 to 0.371 to 0.583 across the three
sub-bands. This fit does not.

| Unit words | Generated | Code comments | Recall | False-flag rate |
| --- | --- | --- | --- | --- |
| 25 to 39 | 178 | 158 | 0.264 | 0.025 |
| 40 to 59 | 328 | 89 | 0.744 | 0.056 |
| 60 to 89 | 177 | 52 | 0.893 | 0.058 |

The false-flag rate is flat across the band, between 0.025 and 0.058.
Recall still rises with length, which is what more evidence per unit
should do, but it no longer comes at the cost of flagging longer human
comments.

By language, on real comments inside the band:

| Language | Comments | Flagged | Rate |
| --- | --- | --- | --- |
| C# | 87 | 9 | 0.103 |
| C++ | 61 | 1 | 0.016 |
| Java | 42 | 0 | 0.000 |
| Rust | 38 | 1 | 0.026 |
| Go | 29 | 0 | 0.000 |
| JavaScript | 24 | 0 | 0.000 |
| C | 18 | 1 | 0.056 |

C# falls from 0.586 to 0.103 and stays the worst case. Java, Go and
JavaScript flag nothing.

## What the vocabulary keyed on

The fit retained 128 columns: 117 character n-grams and 11 word
unigrams, all of the latter function words. The heaviest positive
columns are fragments of written-out nouns and of the connective tissue
around them, among them `met`, `ame`, `ter`, `ati`, `ation`, `tio`, and
the articles `a` and `an`.

Read together they describe fuller sentences: generated documentation
spells out `parameter`, `name`, `returns`, and `-ation` nominalizations,
and carries articles, where a human comment is terser. The signal is the
wording, not the size of the unit.

The 128-column cap is the engine's maximum, so the representation is
coarse by construction. It is also a lever: a larger vocabulary is the
obvious next thing to try, and this run does not try it.

## What this establishes and what it does not

It establishes that on this corpus, at matched length, surface n-grams
separate generated documentation from human code comments where fourteen
aggregate statistics do not, and that the separation survives the
false-flag limit a tool would have to meet.

It reports declared provenance, not human judgment. A false-flag rate of
0.040 on comments a repository already contains is not the same as a
false-flag rate on comments a person wrote today and would defend. That
needs the human-labeled corpus of issue 22 and the published evaluation
of issue 25, both on hold.

## The confirmation partition is generated but not yet scored

Both stopping conditions hold on development, so the pre-registration
admits the confirmation partition. It had no controlled arm: every
long-form task so far came from repositories pinned to training or
development. One was built for it.

Two hundred tasks were drawn from the ten confirmation-pinned
repositories that carry an eligible unit at the fifty-word floor, and
both generator families answered all four prompts and operations for each.
The arm is 1,600 responses, mean 75 and 72 words, all complete, recorded
under `research/generation/runs/2026-09-13-confirm2-haiku` and
`-luna`. `confirmation-tasks.json` is the draw. The corpus was measured
again with them, 808 shards, none skipped.

Scoring it stops at an engine limit rather than a research question. A
lexical prediction counts n-gram terms for every unit it will score, and
holds them under a 64 MiB retention budget. The banded confirmation
partition needs more than that. Two ways forward, neither taken here:
raise the budget, which is a memory guard and not a research parameter,
or make the prediction stream its terms instead of retaining them.

Nothing about the development result depends on this. The confirmation
numbers are simply not in hand, and this record does not report any.

One deviation is recorded here because it touched a checkout. The notice
file of FasterXML/jackson-databind is a symbolic link inside its own
repository, and the importer requires a regular file. It was materialized
from its in-repository target in the local acquisition cache. No measured
source changed.
