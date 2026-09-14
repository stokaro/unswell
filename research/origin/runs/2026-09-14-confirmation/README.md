# Confirmation run of the lexical origin fit, 2026-09-14

The confirmation partition is scored. It separates: recall 0.907 at a
false-flag rate of 0.071 on real code comments, with a Brier interval
that excludes the constant baseline.

Three things qualify that number, and all three are stated before it is
used for anything. They are not footnotes; they decide how much the
result is worth.

## What was run

The [lexical fit](../2026-09-13-lexical/README.md) reached both stopping
conditions on the development partition, which is what admits the
confirmation partition. That partition had no controlled arm, because
every long-form task so far came from a repository pinned to training or
development. One was built: 200 tasks from the ten confirmation-pinned
repositories with an eligible unit at the fifty-word floor, both
generator families, 1,600 responses, all complete, mean 75 and 72 words.

`confirmation-tasks.json` is the draw. The corpus was measured again with
them, 808 shards, none skipped.

## Result

| Metric | Development | Confirmation |
| --- | --- | --- |
| False-flag rate on code comments | 0.136 | 0.071 |
| Recall | 0.783 | 0.907 |
| Precision | 0.908 | 0.965 |
| Brier | 0.184 | 0.084 |
| Constant baseline | 0.304 | 0.320 |
| Coverage | 0.994 | 0.981 |

The confirmation Brier interval is [0.063, 0.132] and excludes its
constant of 0.320. Its false-flag interval is [0.031, 0.116].

## The first qualification: the corpus is not the one that admitted it

A corpus holding a full development arm and a full confirmation arm does
not fit in one artifact. `corpus.MaxUnits` is 10,000, a resource bound on
a single manifest that the module's own comment says does not define
sample adequacy, and the two arms together exceed it. Keeping both meant
lowering the per-checkout cap from fifteen to ten, which changes the
development arm as well.

Nothing derives the value. It arrived with the first corpus commit,
paired with an identical bound on sources, and neither that commit's
message nor the architecture record that accompanied it mentions the
number or argues for it. It is round, it bounds memory and serialization,
and it has stood unexamined since. That is worth knowing before treating
it as a finding about the corpus rather than about the tool.

The bound has an escape hatch and this harness does not use it. A dataset
of shards carries up to 200,000 sources across 1,024 manifests, which is
how acquisition holds a corpus this size. `dataset union` then collapses
the selection back into one manifest so a fit can read it whole, and that
one manifest is what the ceiling applies to. Teaching the fit to read a
sharded candidate set, or raising the per-artifact bound, would remove
this constraint. Neither is done here.

So the development column above is 0.136, not the 0.040 the lexical run
reported. The decision to open the confirmation partition was made on a
corpus this run could not reproduce. Several attempts to preserve the
development arm exactly and trim only the confirmation side all exceeded
the same ceiling; they are not reported because none of them ran.

A result whose admission was decided on a different corpus is weaker than
one whose admission was decided on its own. This one is that.

## The second qualification: the languages are not the same

The model's worst case is C#, at 0.431 in this run's development
partition. The ten confirmation repositories contain no C# at all:

| Language | Comments | Flagged | Rate |
| --- | --- | --- | --- |
| Java | 111 | 8 | 0.072 |
| Rust | 73 | 1 | 0.014 |
| Go | 37 | 7 | 0.189 |
| JavaScript | 3 | 0 | 0.000 |

Part of the gap between 0.136 and 0.071 is that the partition the model
does worse on holds the language it does worse on. The confirmation rate
is not a like-for-like improvement.

## The third qualification: length is back

The banded fit was supposed to stop the false-flag rate tracking unit
length. In the confirmation partition it tracks it again:

| Unit words | Generated | Code comments | Recall | False-flag rate |
| --- | --- | --- | --- | --- |
| 25 to 39 | 128 | 137 | 0.398 | 0.036 |
| 40 to 59 | 551 | 70 | 0.637 | 0.100 |
| 60 to 89 | 402 | 17 | 0.756 | 0.235 |

The lexical run's development partition was flat across these bands, 0.025
to 0.058. Here the rate rises sixfold from the first band to the third.
Seventeen comments in the top band is a thin count, so the 0.235 is not
precise, but the direction holds across all three.

## What this establishes

That the separation is real and survives a held-out partition. It is the
first confirmation result in this line of work, and it is not a null.

It does not establish the rate a tool would carry into a repository. The
admission decision, the language mix, and the length behavior all point
the same way: the honest reading is the development column, not the
confirmation one, and the development column on this corpus is 0.136.

## What the shipped pack is

`research/origin/packs/unswell-origin-lexical-v1.json` is the fit from
the lexical run, not this one. This run refits on its own corpus, so its
numbers describe a different model. The pack's record stays the lexical
run's development result, and this run does not transfer to it.

## Engine note

A lexical prediction used to count every n-gram of every unit and hold
them all before scoring. The confirmation partition exhausted the 64 MiB
retention budget that guards it. Counting now drops the keys no column
reads as it goes, because the vocabulary is frozen by then, so a unit
retains at most 128 terms instead of thousands. Fitting, which has no
vocabulary yet, is unchanged.
