# Historical editorial PR screen: useful source-visible misses

Status: completed exposed acquisition and baseline replay, November 11, 2026.
This is development evidence for #349, not detector qualification. The root
Codex assistant is the reviewer under ADR 0041; no independent human or
human-only authorship claim is made.

The 13 selected public PRs were merged before 2023. Their complete pre-edit
files were fetched at the pinned base commits, checked against Git blob SHA-1,
and recorded with SHA-256 in `sources.json`. Two PR descriptions already
revealed proposed after text and were excluded from the primary source-only
yield. The remaining 11 source-only decisions and exact UTF-8 byte spans were
sealed in commit `a590687` before patch inspection. The search titles reveal
editorial intent, so the sample is purpose-enriched and exposed development,
not a representative or prospective test.

## Acquisition result

| Stage | Count | Interpretation |
| --- | ---: | --- |
| Selected merged PRs | 13 | Distinct repositories, frozen before source review |
| Primary source-only files | 11 | After text not inspected before labels were sealed |
| Necessary diagnoses in distinct primary files | 6 | Source-visible assistant judgments; passes the predeclared four-case acquisition floor |
| Diagnoses supported by the actual PR edit | 5 | Home Assistant, Umbraco, ReSpec, Coral, Rust |
| Necessary diagnosis unrelated to the PR edit | 1 | Full Stack Open; must not be counted as PR-supported |
| Primary Markdown files replayed through current CLI | 7 | Four source-only positives and three other files |
| Full diagnoses on the four Markdown positives | **0/4** | Source-only events; **0/3** among PR-supported Markdown positives |

The five PR-supported cases have distinct mechanisms. Home Assistant repeats
the same entity lifecycle condition within one sentence. Umbraco's `again`
implies a prior run of the same command, while the prior command creates a
different marker. ReSpec's abstract repeats a vague technical-document
category. Coral's troubleshooting instruction uses `try` twice for one
action. Rust repeats the ideographic-number caveat in adjacent paragraphs.
The full pre-edit source, exact source span, nearby retain span, and diagnosis
are pinned in `source_only_labels.json` and `sources.json`.

The Rust file is 1,741 lines of source code. Its source-only decision was
localized to the named `is_numeric` documentation and adjacent API comments;
the review did not exhaustively label every doc comment in that file. The
other primary files' document prose was reviewed as a whole. The title-guided
Rust case is useful for acquisition, not for whole-file recall measurement.

The Full Stack Open source-only diagnosis concerns stacked quantifiers about
event types, but its PR edited a different paragraph near the document end.
The Forem PR removed repeated `We use` starts from a parallel tech-stack
inventory; the source-only review judged that rewrite optional because each
bullet names a different service and purpose. The GraphQL PR fixed awkward
grammar outside the six-purpose contextual screen. The selected Elastic file
was not the one whose prose changed. These cases show why historical edit
occurrence alone cannot be treated as a necessary-edit label.
`patch_review.tsv` records the frozen-label-to-patch relation for all 11
primary PRs without changing the earlier labels.

## Current CLI replay

The `technical` profile was run on the seven primary Markdown base files with
the exact source bytes and no gate. All seven reports were `complete`, with no
operational errors. The compiled binary SHA-256, source tree, input hashes,
report hashes, and all finding spans are in `baseline-ledger.json`; full JSON
reports are local and can be regenerated with `run_baseline.py`.

The current CLI emitted zero findings on the Home Assistant page. It emitted
four generic `filler.wordy-phrase` findings on Umbraco; the one overlapping
the reviewed instruction flagged `in order to`, not the false sequence cue
`again`. On Coral, one `syntax.long-sentence` finding overlaps the reviewed
instruction but does not diagnose the repeated `try`. The Full Stack Open
findings were on unrelated long sentences. No overlap is upgraded to FULL.
Findings on the three other Markdown pages are retained in the ledger but
were not used to infer a complete-cohort accepted-delivery fraction.

This passes the acquisition floor and gives five concrete, historically
supported misses with nearby source controls. It does **not** meet or alter
the original #349 denominator, 80% FULL-event recall requirement, 85%
accepted-delivery requirement, or prospective confirmation requirement. A
next candidate should test whether it can compare local propositions and
instruction states across clauses while retaining the paired controls. A
new algorithm and evaluation protocol must be frozen before using these
cases for tuning or claiming a gain.
