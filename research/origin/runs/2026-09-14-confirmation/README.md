# Confirmation run of the lexical origin fit, 2026-09-14

The confirmation partition separates. Recall 0.875 at a false-flag rate
of 0.052 on real code comments, with a Brier interval that excludes the
constant baseline. Both stopping conditions hold on held-out data, and
the development arm that admitted the partition is reproduced exactly.

This is the first confirmation result in this line of work.

## What was run

The [lexical fit](../2026-09-13-lexical/README.md) met both stopping
conditions on development, which is what admits the confirmation
partition. That partition had no controlled arm, because every long-form
task so far came from a repository pinned to training or development.

One was built: 200 tasks from the ten confirmation-pinned repositories
with an eligible unit at the fifty-word floor, both generator families,
1,600 responses, all complete, mean 75 and 72 words.
`confirmation-tasks.json` is the draw. The corpus was measured again with
them, 808 shards, none skipped.

The corpus holds both arms whole: 11,990 candidates over 4,366 sources.
The development partition scores 721 eligible units with 297 true and 12
false positives, the same counts the lexical run reported, so the
admission decision and the confirmation come from one corpus.

## Result

| Metric | Development | Confirmation |
| --- | --- | --- |
| False-flag rate on code comments | 0.040 | 0.052 |
| Recall | 0.744 | 0.875 |
| Precision | 0.961 | 0.973 |
| Brier | 0.144 | 0.083 |
| Constant baseline | 0.331 | 0.379 |
| Coverage | 0.968 | 0.924 |

Confirmation intervals: false-flag [0.028, 0.082], entirely below the
declared limit of 0.10; Brier [0.062, 0.130] against a constant of 0.379,
which it excludes; recall [0.756, 0.908].

## Two things still qualify it

**The languages differ.** The ten confirmation repositories carry no C#,
which is the model's worst case at 0.103 in development.

| Language | Comments | Flagged | Rate |
| --- | --- | --- | --- |
| Java | 114 | 6 | 0.053 |
| Rust | 75 | 1 | 0.013 |
| Go | 34 | 5 | 0.147 |
| JavaScript | 8 | 0 | 0.000 |

Go is the worst here at 0.147 on 34 comments, a thin count. A partition
without the language a model struggles on is an easier partition.

**Length tracking returns.** The development partition is flat across the
three sub-bands. The confirmation partition is not:

| Unit words | Generated | Code comments | Recall | False-flag rate | Partition |
| --- | --- | --- | --- | --- | --- |
| 25 to 39 | 178 | 158 | 0.264 | 0.025 | development |
| 40 to 59 | 328 | 89 | 0.744 | 0.056 | development |
| 60 to 89 | 177 | 52 | 0.893 | 0.058 | development |
| 25 to 39 | 127 | 129 | 0.307 | 0.016 | confirmation |
| 40 to 59 | 550 | 84 | 0.580 | 0.071 | confirmation |
| 60 to 89 | 399 | 18 | 0.677 | 0.222 | confirmation |

On held-out repositories the false-flag rate rises from 0.016 to 0.222
with unit length, where on development it barely moves. Eighteen comments
in the top band is thin, so the 0.222 is imprecise, but the direction is
the one the band was meant to remove. Recall inside each band is also
lower than on development, so the separation generalizes but less
sharply.

Coverage is 0.924 against 0.968: more units fall outside the calibration
range on unseen repositories.

## What this establishes

That the separation survives a partition the fit never saw, at a
false-flag rate a tool could carry. The headline numbers are not a
development artifact.

It does not establish a rate for C#, which the partition does not
contain, and it does not show the length independence the development
partition showed. A deployment reading would take the worse of the two
partitions per band, which is 0.222 above sixty words.

## What was not run, and one caveat the corpus carries

Contemporary text and mixed-provenance cases are untested. Neither entered
this fit and neither is scored anywhere, so no rate for them exists.

The `polish` responses are scored but excluded from the fit, because their
provenance is mixed by construction: a human wrote the paragraph and a
model rewrote it. On the confirmation partition the fit flags 0.342 of
them against 0.875 of the text written from a fact sheet. That number has
no single reading. As a recall on the hardest case it says a light editing
pass halves detection; as a false-flag rate on substantially human text it
says a third of lightly edited paragraphs are flagged. Deciding between
the two needs the human-labeled corpus of issue 22.

The confirmation partition carries no C#, the language the fit is worst on
at 0.103 in development, so no rate for it exists either.

By generator family and prompt on the confirmation partition, generate
responses only: anthropic-claude 0.877 over 367 units and openai 0.868
over 121; the neutral prompt 0.893 over 225 and the plain prompt 0.859
over 263. The operation matters more than either: families differ by
0.009, prompts by 0.034, and generating against polishing by 0.533.

**Contamination.** The generators may have seen the historical text. Every
repository in the corpus is public and predates the runs, so a model that
read it during training could reproduce its wording without that counting
as independent evidence. The generation records flag verbatim overlap with
the source, and the import reports it as `overlap_high`, but neither
establishes that a response is free of memorized text. A separation this
fit reports is therefore between a generated response and a snapshot the
generator may already know, not between generated and unseen prose.

## What this reports

It reports declared provenance, not human judgment. A false-flag rate on
comments a repository already contains is not a rate on comments a person
wrote today and would defend. That needs the human-labeled corpus of
issue 22 and the published evaluation of issue 25, both on hold.

## What the shipped pack is

`research/origin/packs/unswell-origin-lexical-v1.json` is the fit from
the lexical run. This run refits on a corpus that adds the confirmation
arm, so its model is not byte-identical to the packed one, but its
development partition reproduces the same counts.

## Engine notes

Two limits moved, both because this corpus is the first to need it.

A lexical prediction used to count every n-gram of every unit and retain
them all before scoring, which exhausted the 64 MiB retention budget.
Counting now drops the keys no column reads as it goes, since the
vocabulary is frozen by prediction time, so a unit retains at most 128
terms instead of thousands. Fitting, which has no vocabulary yet, is
unchanged.

The candidate ceiling was one constant, `MaxUnits`, used both for what a
single source may contribute and for what an artifact may hold. Those
protect different things: a corpus of many ordinary sources is not a
pathological file. The artifact bound is now `MaxCandidates` at 50,000
and the per-source bound stays at 10,000. Manifest sources keep their own
limit of 10,000, since a larger corpus is a sharded dataset rather than
one manifest.
