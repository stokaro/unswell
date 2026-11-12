# Findings against what people edited

November 12, 2026. Every earlier number in this repository compares a model
with a model: two assistant annotators, two assistant reviewers, an assistant
auditor. This record measures the tool against what people did. It uses three
public sources of human edits and one public study of human preferences, and
it keeps coordinates and choices only. No text of any source is stored here.

[RESULTS.md](RESULTS.md) holds every number. It is generated from
`results.json`, and `verify.py` recounts `results.json` from the data files.

## Sources

| Source | What people did | Genre | Use here |
| --- | --- | --- | --- |
| Documentation pull requests | Maintainers reworded pages of their own project | Technical documentation, written by people | Does a finding fall on a sentence that a maintainer changed? |
| [Beemo](https://huggingface.co/datasets/toloka/beemo) (MIT) | Experts edited answers that models gave to instructions | Explanations, summaries, answers | The same question, on model-written text |
| [LAMP](https://github.com/salesforce/creativity_eval/tree/main/Writing_Alignment) at `3d02987` (BSD-3-Clause) | Professional writers edited paragraphs that models wrote; three writers edited the same 54 paragraphs; writers ranked three versions of 201 paragraphs | Fiction and creative non-fiction | Agreement between editors; whether a judge model shares the writers' preferences |

The pull requests come from GitHub search: merged before November 2022, with
an editorial word in the title (`copy edit`, `copyedit`, `reword`, `rephrase`,
`wording`, `concise`, `wordy`, `verbose`, `redundant`, `tighten`,
`readability`, `simplify docs`). The searches returned 3,392 pull requests. The
first 1,200 by hash were examined. 238 change at most three files, all
Markdown, and rewrite 1 to 40 lines of a file of at most 40,000 bytes; 184
remain with one per repository; 121 of those change at least one prose
sentence as the engine extracts it. Repositories are split by hash into 60 for
development and 61 for the test.

The study that is closest to this problem,
[Measuring AI "Slop" in Text](https://arxiv.org/abs/2509.19163), has span
annotations by professional editors. Its data was not released when this
record was made.

## Method

The plan, the measures, and the decision rules were written on
[issue 525](https://github.com/stokaro/unswell/issues/525) before the
corresponding results were read. Two amendments were posted there before any
result they affect.

A text is split into sentences: by one fixed splitter for LAMP and Beemo, and
by the engine's own prose sentences for documentation, so that code and tables
are outside. A sentence is *edited* when a person changed any part of it and
*flagged* when a finding touches it. TPR is the share of edited sentences that
are flagged, FPR the share of unedited sentences that are flagged, and J is
TPR minus FPR. A system that flags at random has J = 0 whatever its rate.

Systems: `unswell check` with the default rules and with every rule,
`unswell review` as merged in #544 with two reviewer models, alone and with
the other model confirming, and the span-detection prompt that the LAMP
authors published, with the same two models.

For preferences, each judge model ranked the three versions of every LAMP
triplet with labels hidden, and chose between the raw and the expert-edited
answer of 300 Beemo test rows. A model was to be admitted as a judge of
writing only if it agreed with the writers' majority.

## What the results say

1. **Editors do not agree on which places to change.** One writer's edited
   sentences predict another's barely above chance. At the word level Cohen's
   kappa between writers is 0.11, and 11.9% of words are marked by all three
   (`ceiling_words.py`). The LAMP paper reports the same data as a span
   precision of 0.57, which does not correct for writers marking a third to a
   half of the text. "Finds what an editor would mark" is therefore not a
   stable target, for a tool or for a second editor.

2. **Editors do agree on which version is better.** Two writers give the same
   answer for four pairs in five. They prefer text that a model pipeline
   edited to the raw model text in 71.6% of judgments. That is the published
   human evidence that finding and rewriting helps, in fiction.

3. **Judge models do not share the writers' preferences.** On the pair
   "writer-edited or raw model text" the three judge models agree with the
   writers' majority about half the time, and they prefer the raw model text
   to the writer's edit in about two thirds of the triplets. Their reasons
   show why: they penalize the typos of a hand edit and reward smooth stock
   phrasing. No model was admitted as a judge. This matches published work on
   model evaluators favoring model text
   ([Laurito et al.](https://arxiv.org/abs/2407.12856)).

4. **The fixed rules do not find what people edit.** On documentation that
   maintainers reworded, the default rules flag edited and unedited sentences
   at the same rate, and enabling every rule does not help.

5. **Revising what the review flags did not help the first reader.** In the
   blind comparison of 40 documentation paragraphs, the revised paragraph was
   preferred in 8 of 15 decisive answers where the revision followed the
   review's findings, and in 6 of 13 where the model rewrote random other
   sentences. The reader chose the revision where it trimmed a stock phrase or
   an intensifier, and the original where the revision deleted a whole
   sentence or changed what a sentence claims. That pattern is an observation
   on 15 answers, not a result.

6. **Review by a model does a little.** It flags edited sentences more often
   than unedited ones in all three genres, by a margin that is small on
   documentation and fiction and larger on expert-edited answers. Confirmation
   by a second model shortens the list. J stays about the same on
   documentation and falls or stays flat on the other two.

## Decisions

Each rule was registered on issue 525 before its result was read.

1. *Report the test numbers against people next to the numbers against model
   annotation.* Done in `docs/review.md`, the README, and ADR 0052.
2. *Revise the rubric if the published span prompt beats `unswell review` by 5
   points of J with both reviewer models on both development sets.* The gaps
   are +3.5, -0.4, +14.7, +3.6 points (fiction and answers for each model). Not all four reach 5 points, so the rubric stays.
   The published prompt flags most sentences of a text, between
   56.8% and 82.9%.
3. *Read a model-free linter's rules as candidates if it beats the default
   rules on both development sets.* Not decided: the linters were not run.
4. *State that no available method finds what editors change if no system
   reaches half of the human ceiling on the test.* Half of the ceiling is
   3.5 points. A system reaches it on the fiction test, so the rule does not apply as written.
   The ceiling itself is the finding: at +7.1 points, editors do not find what
   another editor changes either, and ADR 0052 says so in place of this rule.
5. *Admit a model as a judge of writing only if it agrees with the writers.*
   None was admitted: agreement with the writers' majority on "writer-edited
   or raw" is 50.7%, 49.2%, 51.7%, against a floor of 70%. The comparison
   of revisions by model judges was therefore not run. A blind comparison for
   readers was registered in its place.
6. *On documentation, present review as the way to find what maintainers
   tighten if it reaches a J of 10 points on the test and the default rules do
   not.* The review reaches +7.2 and +4.7 with the two reviewer models, the
   default rules -0.3. Neither reaches 10 points, so the documentation says that neither finds what maintainers tighten.

7. *Add a revision step if readers prefer the revised paragraph in at least
   70% of decisive answers, with at least 15 such answers and the lower end of
   the interval above 50%; say that the review's selection adds nothing if it
   does not lead the control arm by 15 points.* One reader answered all 40 pairs.
   The revised paragraph was preferred in 8 of 15 decisive answers in the reviewed
   arm and in 6 of 13 in the control arm. No revision step is added, and the
   documentation says that the selection added nothing a plain rewrite did not.

## Limits

- An unedited sentence is not a clean sentence. A person changed what they
  came to change. FPR therefore overstates false alarms, and J is a lower
  bound on how well a system separates wording that needs work.
- The pull requests are purpose-enriched by title, and contributors are
  maintainers, not editors. Their source text was written by people, so this
  set says nothing about wording typical of models.
- LAMP is fiction, and Beemo's answers come from older and smaller models than
  those that write documentation today.
- Agreement between editors was measured on fiction only. No set gives it for
  documentation.
- Open-source linters for model-style wording were registered as a comparison
  and were not run: installing and executing third-party packages was not
  permitted in the session that made this record. One of them publishes an
  audit of its own in which its phrase sets fired more often on text written
  by people than on model text.
- Each configuration ran once. A model's answers differ between runs.
- The reader comparison has one reader, the maintainer of this project, who
  knew its purpose but not which version was which. Readers of technical
  documentation are engineers, and this is one of them. The intervals are
  wide: 30% to 75% for the reviewed arm.
- In that comparison a model rewrote the flagged sentences and deleted
  some of them. The result is about that way of revising. It does not
  show that a careful human revision of the same sentences would not help.

## Files

| Path | Content |
| --- | --- |
| `data/sentences.jsonl.gz` | Per document: sentence spans, the spans a person changed, each system's finding spans, and a digest of the text |
| `data/preferences.jsonl.gz` | Per LAMP triplet: the three writers' orders and each judge's order |
| `data/expert-pairs.jsonl.gz` | Per Beemo pair: each judge's choice and the position it was shown in |
| `data/readers.jsonl.gz` | Per reader and pair of the blind comparison: the arm, the reviewer model, and which version was preferred |
| `results.json`, `RESULTS.md` | Every published number, recounted and rendered |
| `verify.py`, `test_verify.py`, `render.py` | The check, its negative tests, and the renderer |
| `scripts/` | The scripts that fetched, prepared, ran, and measured, as they ran |

```sh
make check-human-reference
```

The check runs offline. To rebuild the data, the scripts need the two public
datasets, the GitHub command-line client, and access to the models.
