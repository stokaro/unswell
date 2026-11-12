# ADR 0054: Trimming is measured against what editors removed

Status: proposed on November 12, 2026. It takes effect when the maintainer
merges it. It adds a human reference to
[ADR 0053](0053-trim-admitted-by-model-consensus.md) and applies decision 6 of
[ADR 0052](0052-claims-follow-human-references.md) to a command that changes
text.

## Context

ADR 0053 admitted `unswell trim` on a weighted consensus of models. That
consensus splits by model family, and ADR 0052 found that the same families do
not share professional writers' preferences. Nothing compared the command with
people.

The human-reference record holds three kinds of text that people edited:
model answers that experts edited, model fiction that writers edited, and
documentation pages that maintainers reworded. For the first two it holds the
text before and after the edit.

## Decision

1. **The measure.** A command that changes text is compared with the editor
   word by word. Precision is the share of removed words that the editor
   removed or replaced. It is compared with two chance values: random removal
   of the same number of words from the same sentences, among the words the
   check allows to remove, and from the same documents. Where the editor's
   text exists, the net movement is the change of the word-level edit distance
   to that text per removed word, compared with the same random removal.

2. **The rule.** On a kind of text, trimming agrees with editors above chance
   when, on the test part and with both trimming models, precision exceeds
   both chance values and the net movement is below that of random removal,
   each with a 95% interval that excludes zero. It moves text toward the
   editor's version when the net movement itself is below zero in the same
   way. The rule was registered on issue 525 before the command ran on these
   texts.

3. **The result.** The rule is met on model answers that experts edited. It
   is not met on fiction or on documentation written by people.

4. **What the documentation states.** The documentation of the command and
   the README state the result for model answers that experts edited, with
   that scope. The documentation of the command also states that the rule is
   not met on the other two kinds of text and that documentation written by
   models and edited by people is not measured.

5. **Later changes.** The test parts have now been read for trimming. A later
   change of the prompt or the check is run on them to show that it does not
   lose this result. A claim that a change improves the command needs texts
   that this record has not used.

6. **The parser.** On the development parts 5 of 277 documents failed with one
   model, because the model corrected itself and wrote its JSON object twice.
   The parser now reads the last object. The test parts ran with that build
   and no document failed.

## Measurement

`unswell trim` ran on every development and test document with
`claude-sonnet-5-5` and `gpt-6-astra`. The
[record](../../research/reviews/2026-11-12-human-reference/RESULTS.md) has every
number, and its check recounts them from the coordinates of every removed
word. Test parts:

| Text | Trimming model | Removed words | Editor changed them | By chance in the same sentences | Difference, points |
| --- | --- | ---: | ---: | ---: | --- |
| Model answers that experts edited | `claude-sonnet-5-5` | 1,617 | 71.2% | 62.7% | +8.5 (+6.3 to +10.6) |
| Model answers that experts edited | `gpt-6-astra` | 2,110 | 68.0% | 59.3% | +8.7 (+6.7 to +10.7) |
| Model fiction that writers edited | `claude-sonnet-5-5` | 1,037 | 51.9% | 49.3% | +2.6 (+0.0 to +5.2) |
| Model fiction that writers edited | `gpt-6-astra` | 853 | 49.6% | 46.5% | +3.1 (-0.1 to +6.4) |
| Documentation that maintainers reworded | `claude-sonnet-5-5` | 375 | 8.5% | 6.6% | +2.0 (-0.7 to +5.7) |
| Documentation that maintainers reworded | `gpt-6-astra` | 942 | 6.6% | 4.4% | +2.2 (+0.6 to +4.4) |

On model answers that experts edited, the trimmed text is also closer to the expert's text than the
original was. The word-level edit distance to the expert's text falls by 0.28 and 0.23 words per
removed word with the two models. Random removal from the same sentences gives 0.15 and 0.09,
because the experts removed or replaced more than half of the words. On fiction the distance does not fall
(+0.05 and +0.10), though it rises less than under random removal (+0.14 and +0.21).

The development parts were read first and agree in direction on all three
kinds of text; on fiction they showed a larger difference than the test part
did.

## Consequences

`unswell trim` now has evidence against people on one kind of text: where
experts edited what models wrote, it removes words that the experts also
changed, more often than chance, and the result is closer to what the experts
made. The size is modest: +8.5 and +8.7 points over chance in the
same sentences.

The consensus of models of ADR 0053 stays what admitted the command. This
record does not validate that consensus; it measures the command by another
route.

The fixed rules of `unswell check` remain where ADR 0052 left them. Under the
same word-level comparison the words under their findings are not words that
editors changed more often than chance on any of the three kinds of text.

No reader has judged a trimmed text, and the case the tool is built for,
documentation that models wrote, has no human reference.
