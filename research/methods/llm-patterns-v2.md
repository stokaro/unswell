# Pattern protocol, version 2 (draft)

Identifier: `unswell-llm-patterns-v2`. State: draft. This document freezes
when the maintainer records its SHA-256 in `docs/acceptance.md` under stage
C; until then every value below may change, and no confirmation partition
is measured. Version 1 with its four amendments stays the record of the
pilot; this version restates only what the pilot changed.

## What the pilot fixed

| Field | Version 1 | Version 2 |
| --- | --- | --- |
| Families | at least two in the pilot; three required here, one held out | Claude (two models) and OpenAI (one model) select; Qwen (two models) is held out, with its pre-freeze exposure recorded in amendment 4 |
| Partitions | hash assignment per plan | pinned by `research/methods/partitions-v1.json` (SHA-256 in amendment 4); the confirmation partition is `final_test` |
| Confirmation sources | the confirmation partition of the corpus | the six pinned repositories plus new repositories acquired for confirmation and pinned to `final_test` before the confirmatory measurement, at least 20 components in every arm |
| Task type and operations | documentation units; `generate` and `polish` | unchanged |
| Prompt conditions | `neutral` and `plain` | unchanged; `plain` is exploratory |
| Decoding | provider default | the pilot's recorded settings per family; a family runs with one setting |
| Unit of independence | provenance component | unchanged |
| Intervals | cluster bootstrap, 10,000 replicates, seed 17, percentile | unchanged |
| MID | 3 percentage points | unchanged; the pilot's H0 prevalences of 0.1% to 6.4% per rule do not move it |
| Support | 5 components per arm | unchanged |
| Cluster minimum | 20 components per arm | unchanged; the confirmation partition must reach it before measurement |
| Screening | development partition, BH at 0.10 | run once, recorded below |
| Confirmatory list | at most twelve, one per construction | empty; no construction passed the development screening |
| Multiplicity | Holm at 0.05 across the list | unchanged |
| Stopping rule | none stated | one confirmatory measurement of the confirmation partition; no sample grows afterward |

## Exposure history

Every run record of stage C names the tables its maintainer produced and
read. The aggregate tables of every cohort and every run, including both
Qwen runs, were read before this draft. No per-partition table of the
development or confirmation partition was produced before `corpus screen`
ran. The screening output is the first per-partition table, and its hash is
recorded below.

## Screening record

`corpus screen` ran on September 12, 2026 on the development partition of
the pinned plan (`screening/2026-09-12-dataset-plan.json`, SHA-256
`c1a648fc11dfb1265c6177e56969111a6c230355ae4f809ced06d016bcca6c0c`) with
the pilot, run 2, and Haiku records of the Claude family and the luna record
of the OpenAI family. The record is `screening/2026-09-12-development.json`,
SHA-256 `dab04b57b48a177fd9ddfdd37cf35ef5d5b98a338157a00cc649bf99aa00fe97`.

The partition holds 10 components. The Claude arm holds 150 comment
paragraphs from 144 `generate`/`neutral` responses, the OpenAI arm 76 from
76, and the H0 arm 12,859 historical comment paragraphs. Every one of the
80 tests is `inconclusive`: 77 because the controlled arm counted no
finding, 3 because the partition holds fewer than 20 components. No test
passes the false discovery rate with a positive difference, so the
candidate list is empty. The three positive differences are the Claude arm
on `readability.grade-metric` at +1.9 points, `syntax.noun-stack` at +1.3
points, and `syntax.passive-candidate-density` at +0.5 points; each has a
q-value of 0.91 or above. The differences that pass the rate are negative,
led by `syntax.long-sentence` at 6.4 points below H0 in both arms. They say
that short generated paragraphs carry fewer of these findings than
historical comment paragraphs. They enter no list.

## Frozen confirmatory list

The list is empty. The screening named no construction, and the protocol
admits none that the screening did not name. Stage D under this version
measures the confirmation partition once and reports the null result with
the same tables. It tests no hypothesis, and Holm correction applies to
nothing. The negative differences of the screening come from shorter text;
they are not candidates and lead to no rule decision.

A later version may freeze a list only after a new screening on a
development partition that reaches the cluster minimum. The sample-size
record below says what that takes.

## Sample size and stopping rule

The screening record carries a sample-size record per test. For the three
rules that fired in a Claude response, the H0 prevalence, the design
effect from the bootstrap, and the paragraphs and components one arm needs
to detect three points at 80% power are:

| Rule | H0 prevalence | Design effect | Paragraphs per arm at alpha 0.05 | Controlled components | Paragraphs per arm at alpha 0.05/12 | Controlled components |
| --- | --- | --- | --- | --- | --- | --- |
| `readability.grade-metric` | 0.13% | 1.38 | 277 | 26 | 484 | 45 |
| `syntax.noun-stack` | 0.08% | 0.87 | 267 | 16 | 468 | 28 |
| `syntax.passive-candidate-density` | 0.87% | 1.50 | 400 | 40 | 700 | 70 |

Components follow from 15 paragraphs per component in the Claude arm. A
rule whose controlled arm counted nothing gets a design effect the
bootstrap cannot support, so its record is not read. The development
partition holds 10 components and the confirmation partition 6. No
screening or confirmation on the current corpus reaches the cluster
minimum of 20, and none reaches the 26 to 70 components above.

The stopping rule is fixed. A screening or a confirmation runs once per
protocol version on a partition the pins fix, and no sample grows after
that to reach a threshold. Before the next version, the corpus gains
repositories pinned to `development` and to `final_test` until each holds
at least 20 components. The development repositories then receive tasks
and responses of every family that enters selection. That means at least
10 new repositories for development and 14 for confirmation. The 26 to 70
components above are the power target, not the minimum.

## Holdout checks

Version 2 checks the four combinations version 1 named: new sources with
the studied families, new sources with the held-out family, known ecosystems
on new projects, and new roles or topics where support exists. The new
sources are the repositories acquired for the confirmation partition. The
held-out family is Qwen. Threshold tuning on a studied family never counts
as an unseen-family result.

## What must exist before the freeze

- The screening record above, produced by `corpus screen` on the
  development partition with the Claude and OpenAI runs.
- The sample-size record, restated from the screening output.
- The phrase list for E3 is the fifteen rules of the `filler`, `hype`, and
  `scaffold` families at the versions of the current engine; the SHA-256 of
  the sorted `id@version` list is
  `a38de2cc9112a4a80e7e91e33d19f7313427e03d744807a2feffd2af3c863b0f`. The
  Reinhart and Kobak word lists are not in the repository, as the E3 pilot
  record states; adding them is an E3 amendment, not a condition of this
  freeze.
- The papers of the sources record read in full, with the publication
  status of each preprint rechecked. Rechecked on September 12, 2026:
  `rallapalli-2026` (arXiv 2604.14111) is still version 1 of April 15,
  2026, with no journal reference or comments line; `liang-2023` keeps its
  published version in Patterns. The reading itself is not recorded here
  and stays a condition.
- At least 20 components in the confirmation partition, acquired and
  pinned.

## What must exist before the confirmatory measurement

- The frozen list and this document's hash in the acceptance table.
- The confirmation partition measured once, with the run hash recorded
  here; no second measurement under this version.
- Confirmatory tasks drawn from the confirmation partition by the seeded
  sampler, with the sampling record saved.
