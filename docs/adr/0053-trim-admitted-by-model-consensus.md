# ADR 0053: A weighted consensus of models admits a trim-only fix

Status: proposed on November 12, 2026. It takes effect when the maintainer
merges it. It replaces decision 5 of
[ADR 0052](0052-claims-follow-human-references.md) and that ADR's sentence
about the README.

## Context

ADR 0052 said that the tool reports findings and does not rewrite. That
decision rested on one reader's answers to a blind comparison. The maintainer
has withdrawn those answers as a measure. No decision rests on them now, and
they are kept in the record marked as withdrawn.

The maintainer decided, under
[ADR 0041](0041-assistant-review-acceptance.md), that a consensus of models
judges such comparisons, and that a stronger model weighs more where models
disagree. The judge validation of ADR 0052 stands as measured: on fiction the
same model families did not share professional writers' preferences. This ADR
does not claim otherwise. It records which judge the project accepts.

The maintainer also decided that the README describes what the tool does and
links to the measurements. Limits and negative results belong in the
documentation of each command and in the research records.

## Decision

1. **The judge.** Four models answer a rule about information, not taste: a
   version is worse if it lacks a fact, condition, qualification, or
   instruction that the other has; otherwise the shorter version is better when
   the extra words carry no information. `claude-fable-5-2` has three votes,
   `gpt-6-astra` two, and `claude-sonnet-5-5` and `private-host/Qwen3.8-27B`
   one each. The answer with the most votes is the consensus, and a tie goes
   to `claude-fable-5-2`. The maintainer set these weights after the answers
   were read. The weights registered first, two votes for `claude-fable-5-2`
   and one for each other model, give the same consensus on every pair.

2. **The rule.** A kind of revision enters the product as an option when the
   consensus prefers the revised paragraph in at least 70% of at least 15
   decisive answers, with the lower end of the 95% interval above 50%.

3. **`unswell trim`.** A model proposes sentences with words removed. A
   deterministic check keeps a proposal only if no word is added, replaced, or
   moved, at least half of the sentence remains, and no negation, number, modal
   or hedging word, code, link, quotation mark, or emphasis marker is removed.
   Only sentences that lie in one piece of source are offered. Files change
   only with `--write`. The public package `review` holds the prompt, the
   check, and the application; the command only calls the model.

4. **Not tied to review findings.** Trimming random sentences was preferred as
   often as trimming the sentences that `unswell review` flagged. The command
   therefore offers every eligible sentence to the model and does not need a
   review first.

5. **Free revision is admitted and not built.** It meets the rule only where
   the revision followed review findings, by a small margin over a control arm
   with few decisive answers, and it has no deterministic guarantee. It can be
   built later under the same rule.

6. **The README.** It describes the three commands and links to how each was
   measured. It carries no list of limits.

## Measurement

Paragraphs of generated documentation were compared before and after a
model's revision, by the judge of decision 1. The
[record](../../research/reviews/2026-11-12-human-reference/RESULTS.md) has
every answer.

| Comparison | Arm | Revised preferred | Original preferred | No real difference | Revised among decisive answers |
| --- | --- | ---: | ---: | ---: | ---: |
| Free revision | after review findings | 15 | 5 | 0 | 75.0% of 20 |
| Free revision | random sentences | 6 | 3 | 11 | 66.7% of 9 |
| Trim-only | after review findings | 16 | 4 | 0 | 80.0% of 20 |
| Trim-only | random sentences | 17 | 3 | 0 | 85.0% of 20 |
| `unswell trim` as built | every eligible sentence | 33 | 7 | 0 | 82.5% of 40 |

Trim-only meets the rule in both arms, and trimming after review findings is
not ahead of trimming random sentences.

The consensus hides a split. On the trim-only pairs after review findings the
single models prefer the trimmed paragraph in 16, 18, 9, and 6 of 20 pairs, in the
order `claude-fable-5-2`, `claude-sonnet-5-5`, `gpt-6-astra`,
`private-host/Qwen3.8-27B`. The two models of one family approve. The other two
find that about half of the trims lose information, and the two that approve
hold four of the seven votes.

With one equal vote per model family, for `claude-fable-5-2`, `gpt-6-astra`, and
`private-host/Qwen3.8-27B`, the rule is not met: the trimmed paragraph is preferred in
11 of 20, 13 of 20, and 21 of 40 decisive answers for flagged sentences, random sentences, and the command as built.
The admission therefore rests on counting both models of one family.

The command as built was then run on 60 documents outside both comparisons, with two trimming models. The consensus prefers the trimmed paragraph in 33 of 40 decisive answers of 40 pairs (82.5%, 68.0% to 91.3%); single models: `claude-fable-5-2` 31, `claude-sonnet-5-5` 34, `gpt-6-astra` 21, `private-host/Qwen3.8-27B` 16. That meets the rule, so the command ships as built.

## Consequences

Unswell now changes text when asked. The check bounds what a change can be:
words go, nothing arrives. It cannot know what a word means, and the
documentation of the command says so.

The outcome depends on which models count. One vote per model family gives a
different result on the same answers, and the record keeps every single
answer so that it can be recounted.

Acceptance by a consensus of models is agreement between models. Where the
documentation reports it, it names the judges.
