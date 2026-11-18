# The rules on sentences that the editors of TETRA changed

[Issue 365](https://github.com/stokaro/unswell/issues/365) asked whether the revisions of professional editors
in TETRA can serve the engine. Its plan was written for a screen of 123 events and for a fitted model.
[ADR 0046](../../../docs/adr/0046-assistant-panel-corpus.md) retired the screen, and
[ADR 0047](../../../docs/adr/0047-qualified-defaults-and-assistant-review.md) stopped the model. One question of
the issue can still be answered with the product as it is: do the rules of `unswell check` find the sentences
that these editors changed? The measurement and its bar were written on the issue before the rules were run.

## Source

[TETRA](https://aclanthology.org/2024.bea-1.21/) by Mita, Sakaguchi, Hagiwara, Mizumoto, Suzuki, and Inui holds
academic papers that three professional editors revised: the title, the abstract, and the introduction of each
paper. The [pinned repository](https://github.com/chemicaltree/tetra/tree/054ba8b308b7ed5a3a2aca9619427bc8f8fbd405)
declares CC BY 4.0. The [import record](../2026-09-30-tetra-import/README.md) verifies every source file and
leaves 56 eligible groups of papers. This record reads those groups.

## Method

| Part | Choice |
| --- | --- |
| Texts | The original sections of the 56 groups, one document for each section of each editor's file: 168 files, 504 sections. |
| Reference | A sentence is edited when an edit of that editor with a span in the original overlaps it. An insertion has no such span and is left out. The one edit that the source marks as unchanged is left out too. |
| Systems | `unswell check` of release `v0.1.0-alpha.20`, once with the default rules and once with all 62 rules on. A sentence is flagged when a finding overlaps it. |
| Sentences | Split with the rule of the [record of human references](../2026-11-12-human-reference/README.md). |
| Bar | With the default rules, the flagged share of edited sentences is at least 5 points above that of unedited sentences. |

The run asks no model. Nothing is fitted on these texts, so every sentence is measured.

## Result

The data hold 3,876 sentences, and an edit touches 1,572 of 3,876.

| Run | Edited sentences, flagged | Unedited sentences, flagged | Difference |
| --- | ---: | ---: | ---: |
| Default rules | 10 of 1,572 (0.6%) | 22 of 2,304 (1.0%) | -0.3 points |
| Every rule | 585 of 1,572 (37.2%) | 618 of 2,304 (26.8%) | +10.4 points |

1. The bar fails. The default rules are nearly silent on this text. They flag 32 sentences, and an edited
   sentence is flagged no more often than an unedited one. The interval from resampled papers runs from -0.9 to
   +0.3 points.
2. With every rule on, most of the difference is length. The editors change long sentences more often: the
   median edited sentence has 27 words against 19. The rules that fire most on this text read length, passive
   verbs, and reading grade. Compared inside bands of equal length, the difference falls to +4.2 points.
3. One rule follows the editors. `syntax.long-sentence` touches 9.7% of the edited sentences and 3.6% of the
   unedited ones. It is off by default. Every other rule of the table in [RESULTS.md](RESULTS.md) stays within
   4 points.
4. Redundancy edits change nothing. The corpus groups conciseness and repetition edits under redundancy. The
   default rules flag 1 of the 215 sentences with such an edit.
5. Two editors of one paper agree weakly. On 2,896 pairs of sentences that both files hold, the agreement on
   which sentence to edit has kappa 0.26. The writers of LAMP in the record of human references reached 0.11.

## Reading

The answer matches the one on documentation, where the default rules flag 4.0% of the sentences that maintainers
reworded and 4.2% of the sentences that they kept
([ADR 0052](../../../docs/adr/0052-claims-follow-human-references.md)). The default rules read wording that
models favor, and editors of academic papers change other things: grammar, word choice, clarity.

The rules about sentence length and structure do follow these editors a little. They also flag 26.8% of the
sentences that no editor touched, which is why they stay opt-in.

The texts are academic papers. A result on them carries no weight for documentation, and it stands here as the
negative result that the issue asks to keep.

## Files

| Path | Content |
| --- | --- |
| `data/sentences.jsonl.gz` | One row for each sentence of each editor's file. It names the paper, editor, section, and number, and gives the length in words and a hash of the words. It says whether an edit or a redundancy edit touches the sentence, whether the default rules flag it, and which rules touch it with every rule on. The record holds no text. |
| `scripts/run.py` | Runs both configurations on the import output and writes the data. |
| `verify.py` | Recounts `results.json` and `RESULTS.md` from the data and checks the counts that this page states. |
| `results.json`, `RESULTS.md` | Every number of the record. |
| `files.sha256` | The digest of the data file. |

## Reproduce

`make check-tetra-rules` recounts the record from its data. To produce the data again, replay the import as its
record describes, then run:

```sh
python3 research/reviews/2026-11-18-tetra-rules/scripts/run.py <import output.json> <unswell binary> \
  docs/site/src/data/rules-catalog.json research/reviews/2026-11-18-tetra-rules
```
