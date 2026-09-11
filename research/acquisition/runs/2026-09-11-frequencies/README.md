# Frequency tables, September 11, 2026

This directory records the first run of `corpus frequencies` on the
current measurement. It is the frequency item of
[ADR 0037](../../../../docs/adr/0037-diagnostics-not-authorship.md): the
corpus counts constructions, so it can propose candidates that no rule
names yet, instead of measuring only the rules it already has. The inputs
are the 486 candidate artifacts of the measurement after the
[second generation run](../../../generation/runs/2026-09-11-run2/README.md)
joined the controlled cohort; `digests.json` names each one. Every number
is a count of one construction in the sentence units of one cohort and
role. A high ratio is a proposal to read the sentences behind it. It is not
a rule, not a defect, and not a claim about who wrote the text.

## What ran

```sh
corpus frequencies --target contemporary --target controlled --top 100 \
  --candidates artifacts/measurement/candidates/... > frequencies.json
```

The baseline is the historical cohort. A key enters the tables when one
stratum holds it five times, and a contrast when the target stratum holds
it in three provenance components. The command reads the sentence units
of the artifacts, the sentences of the paragraph units, and tags them
with the English provider. It skips fragments, which hold headings, list
items, table cells, and one-line comments. It counts five measures per
cohort and role: word unigrams, bigrams, and trigrams of lowercase
alphabetic words, the first three words of a sentence, and the
part-of-speech template of the sentence. The run took 41 seconds and
590 MB.

## Strata

The strata of the two cohorts that get contrasts, and of the baseline.

| Cohort | Role | Sentences | Words | Components |
| --- | --- | --- | --- | --- |
| historical | `comment` | 54,218 | 533,347 | 124 |
| historical | `documentation` | 5,051 | 65,599 | 42 |
| historical | `readme` | 1,149 | 13,892 | 35 |
| historical | `release_note` | 566 | 3,448 | 20 |
| contemporary | `comment` | 45,242 | 450,861 | 112 |
| contemporary | `documentation` | 6,390 | 75,676 | 51 |
| contemporary | `readme` | 768 | 8,936 | 38 |
| contemporary | `release_note` | 764 | 4,700 | 21 |
| controlled | `comment` | 701 | 10,010 | 32 |

The three dated periods are strata of the tables too. They get no
contrasts in this run, because the period record already compares them
on findings.

## What the contrasts show

Rates are sentences per 1,000 words of the stratum. The ratio divides the
target rate by the baseline rate.

The word contrasts of contemporary comments are the vocabulary of the
sampled repositories. The top unigrams are `cmd`, `derive`, `mypy`,
`worker`, and `clap`. A license notice supplies six of the top twenty
trigrams: `merchantability and fitness` holds 20 sentences in 3
components against 5 sentences in 2. A ratio between two cohorts of
different repositories carries the repository mix first. The component
support is the reading tool: a construction spread over many components
at a modest ratio says more than a high ratio from three.

Openers carry less vocabulary. The opener `this function is` holds 41
contemporary comment sentences in 11 components against 13 in 9, a ratio
of 3.7. At a ratio of 1.6, `we need to` holds 40 in 21 against 30 in
18. At a ratio of 1.8, `this is useful` holds 27 in 11 against 18 in 9.
Those
are the kind of proposal the command exists for: a construction spread
over many components at a modest ratio, with the sentences behind it one
lookup away.

The template contrasts show why every proposal needs reading. The
template `NN IN NN` holds 447 contemporary comment sentences in 40
components against 107 in 43, a ratio of 4.9. The template `NN JJ` holds
197 in 29 against 69 in 24, a ratio of 3.4. Tagging the short sentences
behind them gives a plain answer. The template `NN JJ` is mostly the
directive comment `mypy: allow-untyped-defs`. The template `NN IN NN` is
one-line labels such as `Usage with zsh:` and `Struct with primitives`.
Those are comment lines
the extractor admits as sentences, not prose. The documentation template
`NN NN NN NN`, 58 sentences in 15 components against 12, was not read.

The controlled stratum holds 10,010 words in 32 components, so its ratios
rest on small counts. Its top contrasts against the historical comment
stratum are the idioms of Go documentation comments. The unigram
`reports` holds 21 sentences in 5 components, 2.1 per 1,000 words
against 0.06. The bigram `it returns` holds 13 in 5, 1.3 against 0.07.
The keys `returning the`, `parses`, `prints`, and `runs` follow. Half of the tasks came from Go
sources, and the generated comments keep the "Name reports whether"
convention of those sources, while the historical comment stratum mixes
six languages. The like-for-like baseline for the controlled cohort is the
sentences of the same source paragraphs, which the paired tables use for
findings. The frequency command has no paired mode yet.

Two tokenizer effects show in the tables. The lone token `s` in the
controlled stratum is the parameter name in sentences such as "parses s
as a UUID", not a possessive; the tagger keeps `'s` as one token. The
opener `do n't use` is the tagger's split of "don't".

## What this changes

Nothing in the linter. None of the constructions above enters a rule from
this record. A rule needs positives, counter-examples, and a rewrite
fixture, and a contrast supplies none of them. The record shows what the
command can and cannot separate. It separates constructions by rate and
by component support. It cannot separate the era from the repository mix
or the language mix. A paired contrast of generated sentences against
the source sentences of the same tasks would remove the mix for the
controlled cohort. Directive comments and one-line labels reach the
sentence units. Whether the extractor should treat them as prose is a
question for the linter, recorded here and not acted on.

## What is in this directory

- `frequencies.json`: the tables, with contrasts for the contemporary and
  controlled strata, cut at 100 per measure and stratum.
- `digests.json`: SHA-256 of every input candidate artifact file and of
  the tables. The `inputs` field of the tables carries the digest each
  artifact declares over its compact JSON content, which differs from
  the digest of the indented file.
