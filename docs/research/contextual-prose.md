# Contextual diagnostics on technical pages

The September 15 development evaluation found a real coverage defect: a windowed
rule discarded an entire sentence whenever it mentioned protected inline code.
Repeated constructions around command names disappeared from consideration.
The corrected window keeps surrounding prose and the original sentence distance.
It still excludes code contents and refuses to join a candidate's words across code.

The contrast matcher also recognizes `X, not Y`, alongside `rather than`,
`instead of`, and its existing paired about-sentence form. These structures can
express necessary distinctions. The technical and strict profiles therefore show
repeated contrasts as notes with zero weight, zero cap, and no rule gate. A single
contrast stays below the allowance. Explicit project policy can override this.
Unswell's own check uses the same advisory policy. Its earlier explicit contrast
ban would reject necessary contract distinctions and frozen research records under
the expanded matcher. The rule remains enabled, with every finding visible in CLI
and MCP self-check reports. Other project prohibitions are unchanged.

## Complete Ptah documentation tree

The source is `stokaro/ptah` at
`7b47e7cfb5d4ff32a38375345069f5533bff892f`, restricted to all Markdown and MDX
under `docs/site/src/content/docs`: 133 pages and 203,146 prose words. Every scan
completed. The baseline engine is `f3d09fd3aeaeb355d6519b9cacb98598f149ef6d`.
The user identified the repository as substantially AI-generated. It is an exposed
development case study, with no passage-level authorship or quality labels.

| Profile | Before | After | Repeated contrast findings after |
| --- | ---: | ---: | ---: |
| technical | 570 | 720 | 150 |
| strict | 726 | 876 | 150 |
| All 40 rules | 1,357 | 1,829 | 150 |

The technical scan previously found 523 long sentences and 47 other findings.
Those counts are unchanged. Strict adds the same 156 em-dash findings before and
after. The all-rule scan exposes the shared correction's other effect:
passive candidates increase from 112 to 453. That rule remains opt-in.

The change has two observable parts. Correcting inline context alone increases
paired-contrast findings from 19 to 101. Adding comma-not alternatives increases
them to 150. The earlier scan discarded 3,968 of 8,963 paragraph sentences because
they contained a protected token. Among 608 sentences with unprotected lexical
contrast markers, 314 were discarded. These are lost candidate opportunities,
not 314 missing diagnostics: a window still needs repeated patterns.

## Historical development comparison

Before scanning the correction, a seeded manifest selected historical documentation
and README pages from the development partition, with 400–6,000 previously measured
words, at most two pages per global provenance group, and a cap of 40 pages.
The eligible selection contains 27 pages, 31,488 words, and 17 independent groups.
No diagnostic selected or excluded a page. An initial unmeasured draw used shard
groups; it was replaced before measurement with global dataset-plan groups.

With all rules enabled, findings increase from 140 to 174: passive candidates
increase from 25 to 58 and contrast findings from zero to one. The latter is a
Logrus README passage with repeated `instead of` constructions. Every other count
is unchanged. This measures review load on different technical material. Historical
provenance does not certify clean prose, so these counts are not false-positive
rates. The sample does not meet the frozen confirmatory study's requirements.

## Inspectable examples and limits

The [Ptah e2e fixture](../../e2e/testdata/contextual_ptah/README.md) preserves three
contiguous source passages, hashes, the MIT notice, exact expected detections,
and agent-written alternatives. The review checks failure conditions, negations,
URI schemes, command names, and configuration inputs. Removing a technical
restriction is not a successful rewrite, even if the diagnostic disappears.

The [inline-context fixture](../../e2e/testdata/contextual_inline) checks Markdown,
MDX, and Go through the actual CLI with CRLF and a BOM. It includes protected
keywords, interrupted markers, fenced boundaries, independent comments, a valid
single contrast, and clean rewrites. JSON and SARIF retain original coordinates.
Library tests also verify bounded windows and feature availability.

These results justify the coverage repair and an advisory inspection channel.
They establish neither universal AI-text recognition nor improved editorial
quality. Necessary technical contrasts remain legitimate; the note asks for review.
The [run record](../../research/reviews/2026-09-15-contextual/README.md) contains
source selection, saved summary artifacts, and reproduction commands.

## Complete-document study

The [version 3 study](../../research/generation/studies/long-prose-v3/README.md)
now supplies the longer-prose experiment: 38 whole documents from 20 fresh
repositories, three model families, and four operations/prompt arms. All 456
requests have a saved outcome; 455 are complete and one Qwen response is truncated.
Every measured arm retains 20 global components. The original freeze and the
dispatch-reservation amendment are both retained.

The [results](../../research/generation/studies/long-prose-v3/results.md) include
all 12 primary comparisons, uncertainty, applicability, length sensitivity,
context opportunities, ablations, and all 40 rule dispositions. Only ARI meets
the frozen cross-family association criterion, and it exceeds the historical
review-load budget. The other three primary constructions do not establish that
association. These outcomes do not qualify a blocking default.

The older confirmation remains underpowered at 16 components. This development
sample and Ptah remain exposed data; neither becomes a fresh holdout. The new
[construction review](../../research/generation/studies/long-prose-v3/inspection.md)
and offline fixtures preserve removable prefaces and nominalizations alongside
necessary contrasts, repeated terms, and tagging counterexamples. Human quality
qualification remains separate in #22/#26.
