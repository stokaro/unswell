# Document-level and window behavior, September 15, 2026

Issue 218 asks for two things. One is the unit-level rates. The other is how
the rules behave on whole documents and windows, stratified by role, length,
model family and operation. It also asks that differences coming from context
or opportunity counts be explained rather than left in the numbers.

The unit-level rates are in `research/methods/llm-patterns-v2.md`. This record
is the document-level half. It measures what a reader actually receives: how
many warnings a document carries and whether it carries any at all.

The corpus is `artifacts/measurement` after the extraction fixes of issue 279,
28,740 documents and 5.19 million prose words under
`research/acquisition/policy-e1.yaml`, which enables all forty rules. The
generated arm is 11,856 documents of model output under recorded prompts.

## Length decides the document-level answer

| Arm | Length band | Documents | Findings per document | Per 1000 words | Share with any |
| --- | --- | ---: | ---: | ---: | ---: |
| Human | under 50 | 5,253 | 0.01 | 1.051 | 1.3% |
| Human | 50-149 | 3,045 | 0.19 | 2.024 | 13.0% |
| Human | 150-399 | 3,275 | 1.53 | 6.745 | 67.9% |
| Human | 400-1199 | 1,329 | 4.25 | 6.370 | 80.4% |
| Human | 1200+ | 530 | 25.19 | 8.305 | 91.5% |
| Contemporary | under 50 | 1,157 | 0.01 | 0.602 | 1.0% |
| Contemporary | 150-399 | 973 | 1.49 | 6.643 | 69.6% |
| Contemporary | 1200+ | 194 | 14.95 | 4.913 | 94.3% |
| Generated | under 50 | 8,799 | 0.03 | 1.514 | 2.6% |
| Generated | 50-149 | 2,787 | 0.78 | 10.927 | 57.2% |
| Generated | 150-399 | 266 | 1.06 | 5.245 | 56.4% |

The share of documents carrying at least one finding runs from about one
percent under fifty words to over ninety percent above twelve hundred. The
curve has the same shape in human and contemporary text. Findings per thousand
words is flat by comparison.

Three quarters of the generated arm sits under fifty words and four documents
of 11,856 exceed four hundred. Any document-level comparison of the arms that
does not hold length fixed is therefore a comparison of length. That is the
same trap the origin work fell into with the fourteen prepared features, and
it is recorded here so the document-level numbers are not read the same way.

## Held at one length band, the model matters more than the origin

The one band where the arms diverge is 50 to 149 words: 10.927 findings per
thousand words in the generated arm against 2.024 in human prose. The
division by family, on the same tasks and the same prompts, is where that
comes from.

| Family | Operation | Documents | Per 1000 words | Share with any |
| --- | --- | ---: | ---: | ---: |
| Claude | long-form | 1,008 | 8.904 | 56.2% |
| Claude | short | 3,720 | 9.082 | 24.2% |
| OpenAI | long-form | 1,008 | 1.429 | 4.9% |
| OpenAI | short | 3,720 | 4.816 | 10.2% |
| Qwen 32B | short | 800 | 2.090 | 3.0% |
| Qwen 27B | short | 800 | 2.134 | 3.1% |

One family draws six times the warnings of another on the same work. The
spread between families is larger than the spread between the generated arm
and human prose. A reader who wants to know how loud this catalog will be on
their documentation learns more from which model wrote it than from whether a
model wrote it.

## Role separates further than origin does

| Arm | Role | Documents | Per 1000 words | Share with any |
| --- | --- | ---: | ---: | ---: |
| Human | specification | 8 | 19.465 | 100% |
| Human | comment | 12,220 | 7.063 | 29.8% |
| Human | documentation | 850 | 5.531 | 52.2% |
| Human | readme | 271 | 3.241 | 38.0% |
| Human | release_note | 83 | 0.712 | 61.4% |
| Generated | comment | 11,856 | 6.430 | 16.7% |

Human specification prose carries 19.5 findings per thousand words and every
one of its eight documents carries some. Human release notes carry 0.7. The
range across roles inside human prose is a factor of twenty-seven; the
difference between human and generated comments is a factor of 1.1.

## Window scope and opportunity

| Arm | Document scope | Paragraph scope | Sentence scope | Section scope |
| --- | ---: | ---: | ---: | ---: |
| Human | 3.057 | 1.407 | 2.387 | 0.000 |
| Generated | 2.061 | 3.381 | 0.987 | 0.000 |
| Contemporary | 2.271 | 0.982 | 1.864 | 0.000 |

Findings per thousand words by the scope of the rule that produced them.

Document-scope rules, which look across a whole prose run, produce the largest
share in human text. Paragraph-scope rules produce the largest share in
generated text. Sentence-scope rules run more than twice as high in human
prose.

The section-scope rules produce nothing anywhere. They are
`repetition.heading-echo`, `repetition.summary-echo` and
`format.list-fragmentation`. Opportunity explains part of it: a doc comment
has no heading structure, so a section rule has nothing to bound. It does not
explain all of it, because the corpus also holds 671 thousand words of
documentation and 181 thousand of readme prose, both of which carry headings,
and those produce no section-scope finding either. The three rules are
recorded as opt-in in `docs/rule-defaults.md` for that reason.

## What this settles and what it does not

It settles the box that asked for document and window behavior stratified by
role, length, family and operation. The answer is that the catalog's load is
governed by length first, by role second, by model family third, and by
whether a model wrote the text last.

It does not establish a false-positive rate. Every number here is a firing
rate on declared provenance. Whether a finding is a defect needs the human
judgments of issue 22, which stays on hold.

## Reproducing

```
bash scripts/measure-corpus.sh
```

then read `artifacts/measurement/findings/*.json`. Each document record
carries its cohort, role, partition, prose word count, finding count and the
per-rule breakdown. Family and operation come from the shard name.
