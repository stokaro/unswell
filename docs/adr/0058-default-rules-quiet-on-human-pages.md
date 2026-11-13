# ADR 0058: A default rule stays quiet on pages that people wrote

Status: proposed on November 13, 2026. It takes effect when the maintainer
merges it.

## Context

The default rules of `unswell check` were chosen as rules of lean wording: a
wordy phrase, a long sentence, vague praise, a doubled word. The tool is for
documentation that reads as written by a model, and two measurements say that
most of those rules do not tell such wording from any other.

On 121 pages that people wrote before November 2022, 61 had a finding of a default
rule (50.4%). Six rules report on more than one page in twenty:

| Rule | Pages with a finding |
| --- | ---: |
| `filler.wordy-phrase` | 30 (24.8%) |
| `syntax.long-sentence` | 22 (18.2%) |
| `filler.instruction-scaffolding` | 11 (9.1%) |
| `hype.vague-praise` | 11 (9.1%) |
| `repetition.adjacent-word` | 9 (7.4%) |
| `format.decorative-emoji` | 7 (5.8%) |

The second measurement is a layer of edits that a person most likely typed:
153 Markdown files from 134 repositories, before and after a commit made
through GitHub's web flow that changes that one file and whose message is one
short line saying that it removes AI-sounding wording. On it the five rules
of lean wording above are lower after the edit in 23 pairs and higher in
25. `format.decorative-emoji` is lower in 20 and higher in 14, which does not
meet the reading registered on issue 525 (65%, lower end of the interval
above 50%). `format.em-dash-frequency` meets it: 70.0% (54.6% to 81.9%).

## Decision

1. **A rule ships on only if it reports on at most one in twenty of the pages
   that people wrote before November 2022.** The record recounts this for
   every rule, and its check fails when a rule over the limit ships on.

2. **Six rules ship off:** `filler.wordy-phrase`, `syntax.long-sentence`,
   `filler.instruction-scaffolding`, `hype.vague-praise`,
   `repetition.adjacent-word`, and `format.decorative-emoji`. Each records its
   reason in the catalog. This is the third reason of the decision rule in
   `docs/rule-defaults.md`: the warning load falls on human technical prose.

3. **The strict profile turns all six on.** They name real defects of lean
   wording, and a project that wants that reading asks for it with
   `extends: [builtin:strict-v1]` or by enabling a rule by name.

4. **The emoji rule goes with them.** It passed the bar of ADR 0056 on
   commits found by their message. It reports on 7 (5.8%) of the pages
   that people wrote, and it is not confirmed on the likely-manual edits.

## Consequences

With the rules that ship on, 11 of the 121 pages that people wrote have a finding
(9.1%):

| Rule that ships on | Pages with a finding |
| --- | ---: |
| `filler.nominal-support` | 3 (2.5%) |
| `format.em-dash-frequency` | 2 (1.7%) |
| `repetition.near-sentence` | 2 (1.7%) |
| `filler.announced-importance` | 1 (0.8%) |
| `filler.evaluative-closure` | 1 (0.8%) |
| `repetition.exact-sentence` | 1 (0.8%) |
| `repetition.repeated-claim` | 1 (0.8%) |

On the de-slop reference the rules that ship on are lower after the commit in
70.9% (64.2% to 76.7%) of the English test pairs that differ, and they
report in 210 of 500 pairs. The earlier defaults gave 72.9% (67.5% to 77.7%)
and reported in 300. The reach falls because the emoji rule is off; what remains
is carried by `format.em-dash-frequency`.

`unswell check` with no configuration now reports much less on ordinary
documentation. A project that relied on the six rules gets them back with the
strict profile, which also sets stricter gates, or by naming them.

The likely-manual layer has limits. A commit made through GitHub's API looks
the same as one made in the web editor, and a person can paste text into the
editor. The layer is small: the emoji rule reports in 36 pairs. Its result
for that rule is "not confirmed", not "wrong": the interval reaches
74%.

The [record](../../research/reviews/2026-11-12-human-reference/RESULTS-words.md)
holds the page counts and the pairs, and its check recounts these numbers.
