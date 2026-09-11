# Paired frequency tables, September 12, 2026

This directory records the paired run of `corpus frequencies`, the
increment the [first frequency run](../2026-09-11-frequencies/README.md)
named. That run contrasted the controlled cohort with the pooled
historical comments, and its top contrasts were Go documentation idioms,
because half of the tasks came from Go sources. Here the responses of
both generation runs stand against the sentences of the very paragraphs
they answer, so the repository and language mix of the tasks drops out.
Every number is a count of one construction in the sentence units of one
stratum. A high ratio is a proposal to read the sentences behind it. It
is not a rule, not a defect, and not a claim about who wrote the text.

## What ran

```sh
corpus frequencies \
  --pair-tasks research/generation/runs/2026-09-11-pilot/tasks.json \
  --pair-records research/generation/runs/2026-09-11-pilot/records.json \
  --pair-tasks research/generation/runs/2026-09-11-run2/tasks.json \
  --pair-records research/generation/runs/2026-09-11-run2/records.json \
  --baseline historical-paired --target controlled-generate \
  --target controlled-polish --top 100 --candidates ... > frequencies.json
```

The pairing registers 200 tasks and 800 responses. A response counts
under `controlled-generate` or `controlled-polish`, and the sentences of
the paragraph a task names count under `historical-paired`. The minimum
count of five and the minimum support of three components are the
defaults. `frequencies-reverse.json` is the same run with
`controlled-polish` as the baseline, so the constructions the originals
hold more often show too.

| Stratum | Sentences | Words | Components |
| --- | --- | --- | --- |
| `historical-paired` | 342 | 4,735 | 42 |
| `controlled-generate` | 295 | 4,173 | 25 |
| `controlled-polish` | 406 | 5,837 | 29 |

The strata are small, so only unigrams and bigrams reach the minimum
count. No opener and no part-of-speech template does.

## What the responses hold more often

Rates are per 1,000 words of the stratum. The ratio divides the target
rate by the baseline rate. The generate arm, which writes from the code
alone, moves furthest from the originals.

| Key | Generate | Originals | Ratio |
| --- | --- | --- | --- |
| `the given` | 37 in 11 components | 6 in 4 | 7.0 |
| `whether` | 23 in 10 | 6 in 3 | 4.4 |
| `for the` | 20 in 9 | 6 in 4 | 3.8 |
| `given` | 39 in 11 | 13 in 7 | 3.4 |
| `an error` | 11 in 5 | 5 in 3 | 2.5 |
| `returns a` | 18 in 7 | 9 in 4 | 2.3 |
| `is not` | 14 in 5 | 7 in 6 | 2.3 |
| `returns` | 64 in 15 | 34 in 10 | 2.1 |
| `and` | 96 in 16 | 61 in 21 | 1.8 |
| `that` | 97 in 11 | 65 in 16 | 1.7 |

The polish arm, which rewrites the original, moves the same way but less
far.

| Key | Polish | Originals | Ratio |
| --- | --- | --- | --- |
| `the given` | 20 in 3 components | 6 in 4 | 2.7 |
| `whether` | 20 in 5 | 6 in 3 | 2.7 |
| `than` | 19 in 8 | 6 in 4 | 2.6 |
| `is not` | 18 in 9 | 7 in 6 | 2.1 |
| `when` | 30 in 10 | 12 in 9 | 2.0 |
| `otherwise` | 13 in 5 | 7 in 4 | 1.5 |
| `does not` | 19 in 8 | 11 in 5 | 1.4 |

## What the originals hold more often

From `frequencies-reverse.json`, with the polish arm as the baseline:

| Key | Originals | Polish | Ratio |
| --- | --- | --- | --- |
| `will` | 25 in 14 components | 9 | 3.4 |
| `on the` | 20 in 4 | 8 | 3.1 |
| `used` | 13 in 8 | 7 | 2.3 |
| `as` | 24 in 8 | 13 | 2.3 |
| `be` | 56 in 19 | 39 | 1.8 |
| `we` | 19 in 11 | 15 | 1.6 |
| `should` | 10 in 7 | 8 | 1.5 |
| `this` | 35 in 18 | 28 | 1.5 |
| `you` | 13 in 7 | 11 | 1.5 |

## What it shows

Against the sentences they answer, the responses reach for `the given`,
`whether`, `returns a`, `is not`, and `does not`, and they drop `will`,
`we`, `you`, and `should`. The first group is the shape of a definition:
"Returns whether the given X is not Y". The second is the voice of a
maintainer speaking to a reader about what the code will do. `the given`
is the strongest proposal in the tables: eleven components carry it in
the generate arm against four in the originals, and the polish arm keeps
half of that gain.

Every one of these is a proposal to read the sentences behind it. None
is a rule, and a rule would need positives, counter examples, and a
rewrite fixture. The strata hold four to six thousand words each, so a
ratio below two rests on a handful of sentences, and no opener or
template reaches the minimum count.

## What is in this directory

- `frequencies.json`: the paired tables with the originals as the
  baseline.
- `frequencies-reverse.json`: the same run with the polish arm as the
  baseline.
- `digests.json`: SHA-256 of the two task sets, the two generation
  records, every candidate artifact, and the two outputs.
