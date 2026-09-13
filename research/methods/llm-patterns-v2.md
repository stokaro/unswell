# Pattern protocol, version 2

Identifier: `unswell-llm-patterns-v2`. State: frozen on September 13, 2026.
The confirmatory list below is closed. No value in this document changes
from here, and no construction joins the list, whatever the confirmation
partition shows. The acceptance table records this document's SHA-256 under
stage C. Version 1 with its six amendments stays the record of the pilot;
this version restates only what the pilot changed.

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
| Screening | development partition, BH at 0.10 | three exploratory runs before the freeze, all recorded below |
| Confirmatory list | at most twelve, one per construction | two hypotheses, frozen below from the third screening |
| Multiplicity | Holm at 0.05 across the list | unchanged |
| Stopping rule | none stated | one confirmatory measurement of the confirmation partition; no sample grows afterward |

## Exposure history

Every run record of stage C names the tables its maintainer produced and
read. Before this draft the maintainer read the aggregate tables of every
cohort and every run, both Qwen runs included. Nothing produced a
per-partition table of the development or the confirmation partition before
`corpus screen` ran. That screening output is the first such table, and its
hash sits below.

Then the second acquisition round added repositories to both partitions.
Two round-two generation runs added responses to the development partition,
and `corpus screen` ran again on the grown partition. A recount by role then
showed both partitions still short, so a third round added six more
repositories and two more runs, and the screening ran a third time. Each
record sits below. The maintainer read all three before this draft. No table
of the confirmation partition exists yet.

## First screening record

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

## Second screening record

The second acquisition round added 24 repositories, ten pinned to
`development` and fourteen to `final_test`, and two round-two generation
runs answered 600 requests each on 150 new tasks from seven of the new
development repositories. `corpus screen` ran again on September 12, 2026
against the grown plan (`screening/2026-09-12-dataset-plan-2.json`, SHA-256
`8b51bd3cbbaa81031e763c6ee6ed5b8cfa0549d9de89543a57f6d3bca54630f6`) with
four Claude records and two OpenAI records. The record is
`screening/2026-09-12-development-2.json`, SHA-256
`2b97a8a98ab49c7172c4bde45c5dfd7719ef68894b0839ed1b613b1b757037c5`.

The partition now holds 20 pinned repositories. Nineteen carry a historical
comment paragraph and seventeen carry a controlled response. The Claude arm
holds 311 comment paragraphs from 277 `generate`/`neutral` responses, the
OpenAI arm 181 from 177, and the H0 arm 19,692 historical comment
paragraphs. The first screening read 150, 76, and 12,859.

Every one of the 80 tests is still `inconclusive`: 74 because the
controlled arm counted no finding, 6 because an arm holds fewer than 20
components. Two tests now pass the false discovery rate with a positive
difference, so the candidate list is no longer empty:

| Rule | Family | Controlled | H0 | Difference | q-value | State |
| --- | --- | --- | --- | --- | --- | --- |
| `syntax.noun-stack` | Claude | 9 of 311 | 14 of 19,692 | +2.8 points | 0.009 | inconclusive, cluster minimum |
| `readability.grade-metric` | Claude | 14 of 311 | 306 of 19,692 | +2.9 points | 0.095 | inconclusive, cluster minimum |

Both differences fall short of the minimum useful difference of three
points, and both arise in the Haiku model alone: the Opus responses of the
partition carry neither finding. `syntax.passive-candidate-density` is the
third positive Claude difference at +1.9 points with a q-value of 0.82. No
OpenAI test has a positive difference; the OpenAI arm counted no finding
for either candidate rule. A candidate is a rule to write a hypothesis
template for. It is not a confirmed construction and not a rule decision.

The negative differences repeat the first screening.
`syntax.long-sentence` leads them. It sits 4.2 points below H0 in the
Claude arm and 5.8 points below it in the OpenAI arm. Short generated
paragraphs carry fewer of these findings than historical comment
paragraphs. They enter no list.

## Sample size after the second screening

The record carries a sample-size record per test. For the five rules that
fired in a Claude response, the paragraphs and components one arm needs to
detect three points at 80% power are:

| Rule | H0 prevalence | Design effect | Paragraphs per arm at alpha 0.05 | Controlled components | Paragraphs per arm at alpha 0.05/12 | Controlled components |
| --- | --- | --- | --- | --- | --- | --- |
| `syntax.noun-stack` | 0.07% | 0.94 | 266 | 14 | 466 | 24 |
| `syntax.parenthetical-load` | 0.42% | 1.09 | 324 | 20 | 568 | 34 |
| `syntax.passive-candidate-density` | 1.32% | 1.65 | 475 | 43 | 831 | 75 |
| `readability.grade-metric` | 1.55% | 1.16 | 513 | 33 | 898 | 57 |
| `syntax.long-sentence` | 5.83% | 23.02 | 1,182 | 1,487 | 2,069 | 2,603 |

Each component gives 18.3 paragraphs in the Claude arm, and the counts
above follow from that. The arm holds 311 paragraphs in 17 components. So
`syntax.noun-stack` already meets its unadjusted target. It misses the Holm
target and the cluster minimum.

`syntax.long-sentence` carries a design effect of 23. Long sentences
cluster inside a few components, which is what such an effect means. No
corpus this project will hold can test that rule at this MID. Its
difference is negative in both arms in any case.


## Third screening record

Amendment 6 added six repositories, two pinned to `development` and four to
`final_test`, and two round-three generation runs answered 600 requests each
on 150 new tasks. `corpus screen` ran on September 13, 2026 against the
grown plan (`screening/2026-09-13-dataset-plan-3.json`, SHA-256
`2cd58567e18b47970bfd275aefd89ae6389178fd4c7ccaa5dea90402256398e3`) with
five Claude records and three OpenAI records. The record is
`screening/2026-09-13-development-3.json`, SHA-256
`2005d1ae0e83d3638d3d961177302f5507db530e9db28e6edee206346903c8a2`.

This is the first screening whose arms meet the cluster minimum. Of the 22
pinned repositories in the partition, 21 carry a historical comment
paragraph and 20 carry a controlled response. Claude contributes 458 comment
paragraphs from 424 `generate`/`neutral` responses across 20 components.
OpenAI contributes 281 from 277, also across 20. H0 contributes 25,462
historical comment paragraphs across 21. No test carries the reason
`cluster_minimum`.

Evidence now divides the 80 tests, where sample size used to. Seventy-four
stay `inconclusive` on a zero count. Three more stay `inconclusive` because
the interval spans zero or the estimate falls short of the MID. Three are
`unsupported`, their upper bound below the MID. Two pass the false discovery
rate with a positive difference:

| Rule | Family | Controlled | H0 | Difference | 95% interval | q-value | Support | State |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `syntax.noun-stack` | Claude | 14 of 458 | 24 of 25,462 | +2.96 points | +1.18 to +4.17 | 0.004 | 7 components | inconclusive |
| `readability.grade-metric` | Claude | 18 of 458 | 383 of 25,462 | +2.43 points | +0.65 to +4.40 | 0.021 | 6 components | inconclusive |

Both intervals sit entirely above zero and both clear the support minimum of
five components. Neither reaches the minimum useful difference of three
points, and `syntax.noun-stack` misses it by four hundredths of a point. The
protocol's card rule is written for exactly this case: a difference the
interval distinguishes from zero is still `inconclusive` when the point
estimate falls below the MID the pilot fixed in advance. Lowering the MID now
would be tuning a threshold on an outcome, which the protocol forbids.

The three `unsupported` tests are `syntax.long-sentence` in the Claude arm
at 4.3 points below H0, `syntax.parenthetical-load` at 1.1 points below, and
`syntax.passive-candidate-density` in the OpenAI arm at 0.6 points below.
Shorter generated text explains all three, as in the earlier rounds, and
none of them enters a list. No OpenAI test has shown a positive difference
in any screening so far.

## Sample size after the third screening

| Rule | H0 prevalence | Design effect | Paragraphs per arm at alpha 0.05 | Controlled components | Paragraphs per arm at alpha 0.05/12 | Controlled components |
| --- | --- | --- | --- | --- | --- | --- |
| `syntax.noun-stack` | 0.09% | 0.86 | 270 | 11 | 473 | 18 |
| `syntax.parenthetical-load` | 1.28% | 11.68 | 467 | 239 | 818 | 417 |
| `readability.grade-metric` | 1.50% | 1.11 | 505 | 25 | 883 | 43 |
| `syntax.passive-candidate-density` | 1.35% | 2.77 | 480 | 58 | 839 | 102 |
| `syntax.long-sentence` | 5.57% | 26.44 | 1,142 | 1,318 | 1,999 | 2,308 |

Each component gives 22.9 paragraphs in the Claude arm. The arm holds 458
paragraphs in 20 components, so `syntax.noun-stack` meets both its
unadjusted and its Holm target. `readability.grade-metric` meets the
unadjusted paragraph target and misses the Holm one. The two rules with a
design effect above ten cluster inside a few components, and no corpus this
project will hold can test them at this MID.

## Frozen confirmatory list

Two hypotheses are frozen here. Both come from the third screening and both
sit in the Claude family. The content closed in commit 4f8794b. That was
before any task was drawn from the confirmation partition, and before any
table of that partition existed. Wording changed once after the draw, to
pass the repository's own prose gate. No hypothesis, threshold, design or
estimate changed with it. The acceptance table records the hash of the final
text.

### H1. `syntax.noun-stack`

The prevalence of `syntax.noun-stack` in Claude `generate`/`neutral` comment
paragraphs of the `final_test` partition exceeds its prevalence in the
historical comment paragraphs of the same partition by at least three
percentage points.

Primary comparison: `D` = controlled prevalence minus H0 prevalence, unit
kind `paragraph`, role `comment`, partition `final_test`, family
`anthropic-claude`, cluster bootstrap with 10,000 replicates and seed 17,
percentile interval. Development estimate: +2.96 points, interval +1.18 to
+4.17, q-value 0.004, support 7 components.

### H2. `readability.grade-metric`

The prevalence of `readability.grade-metric` in Claude `generate`/`neutral`
comment paragraphs of the `final_test` partition exceeds its prevalence in
the historical comment paragraphs of the same partition by at least three
percentage points.

Primary comparison: the same design as H1. Development estimate: +2.43
points, interval +0.65 to +4.40, q-value 0.021, support 6 components.

### Decision rule

Holm correction at 0.05 covers the pair. Four conditions together make a
card `supported-within-scope`. Its interval's lower bound exceeds zero. Its
point estimate reaches the MID of three points. Support holds at five
components per arm. The baseline load is acceptable.

A development estimate below the MID does not decide the confirmation. The
confirmation partition is other text, and its estimate may land on either
side of the MID.

Nothing else is tested. The OpenAI arm of the confirmation partition is
measured and reported, because the holdout checks name it, and it carries no
hypothesis. Three screenings produced no positive OpenAI difference at all.
Negative differences are not hypotheses either.

## Stopping rule

The stopping rule is fixed. Under a frozen protocol version a screening
runs once, and so does a confirmation, each on a partition the pins fix. No
sample then grows to reach a threshold. The three screenings above ran while
this document was a draft, which is where exploratory screening belongs; the
freeze keeps the third and closes the series. No fourth screening runs under
this version.

Before the freeze the corpus gained repositories until each of those two
partitions held at least 20 components. Each pin lands before any of that
repository's text is measured. The development repositories then receive
tasks and responses of every selecting family. The second acquisition round
added ten to `development` and fourteen to `final_test`, and the third
added two and four. `development` now holds 22 pinned repositories and
`final_test` 24, of which 21 each carry a historical comment paragraph.

The paragraph and component counts of the sample-size record are power
targets, not thresholds. The threshold is the cluster minimum: 20
components per arm.

## Holdout checks

Version 2 checks the four combinations version 1 named: new sources with
the studied families, new sources with the held-out family, known ecosystems
on new projects, and new roles or topics where support exists. The new
sources are the repositories acquired for the confirmation partition. The
held-out family is Qwen. Threshold tuning on a studied family never counts
as an unseen-family result.

## What the freeze required

- A screening record produced by `corpus screen` on the development
  partition with the Claude and OpenAI runs, on a controlled arm of at
  least 20 components. The third record meets that condition; the first two
  do not, and the list comes from the third.
- The sample-size record, restated from that screening output. Done above.
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
- At least 20 components in the confirmation partition that carry a
  historical comment paragraph. Amendment 6 met this: of its 24 pinned
  repositories 21 do.
- At least 20 components in the H0 arm and in the controlled arm of the
  development partition. Amendment 6 met this too: of its 22 pinned
  repositories 21 carry a comment paragraph and 20 carry a controlled
  response. The third screening reports no test blocked by the cluster
  minimum.

## What must exist before the confirmatory measurement

- The frozen list and this document's hash in the acceptance table.
- The confirmation partition measured once, with the run hash recorded
  here; no second measurement under this version.
- Confirmatory tasks drawn from the confirmation partition by the seeded
  sampler, with the sampling record saved.
