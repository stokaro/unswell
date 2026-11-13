# Trim

Experimental. Against human editors the command was measured on three kinds
of text and meets its rule on one of them; see
[Measured behavior](#measured-behavior).

`unswell trim` asks a model you name for sentences with empty words removed:
announcements, stock phrases, intensifiers, repeated qualifiers. The model
proposes. A deterministic check decides which proposals you see.

```text
docs/guide.md:3:1: trim
    - These features significantly improve the ability to type-check your code.
    + These features improve the ability to type-check your code.
1 trim in 1 document.
Model: example-model at api.example.com.
Nothing was written. Use --write to apply these trims.
A trim only deletes words. A deleted word can still carry meaning that the check cannot see.
```

## What the check guarantees

A proposal is kept only if all of this holds:

- Every word of the trimmed sentence occurs in the original, in the same
  order. Nothing is added, replaced, or moved. Case and the punctuation next
  to a word may change.
- At least one word is removed and at least half of the words remain, so a
  sentence is never deleted.
- No negation is removed (`not`, `no`, `never`, `without`, `unless`, `only`,
  words ending in `n't`), no number, and no modal or hedging word (`may`,
  `must`, `should`, `usually`, `often`, `likely`, `approximately`).
- No code or link is removed, and a kept word keeps its quotation marks and
  its emphasis markers.

Only sentences that lie in one piece of source are offered. A sentence with
inline code, a link, or emphasis inside is left alone, as are headings, list
markers, tables, and code. So a trim cannot touch a command or an identifier.

The check is exact about words. It does not know what a word means. A removed
adjective can be the point of a sentence, so read each trim before you keep
it.

## What leaves the machine

The complete text of every selected document goes to the endpoint or the
local program you name, with the instruction. An API key goes to that endpoint
in the `Authorization` header and nowhere else. Without a named model the
command fails before it reads a document.

## Run it

The model is named as for [`unswell review`](review.md#run-it):

```sh
unswell trim README.md docs \
  --endpoint https://api.example.com/v1/chat/completions \
  --model example-model --api-key-env REVIEW_API_KEY
```

or with a local program that reads the prompt on standard input:

```sh
unswell trim docs/guide.md --command assistant --command-arg=--print
```

Paths, `--stdin`, `--filename`, `--format`, `--config`, `--profile`, and
`--project-root` select and extract documents as they do for
[`unswell check`](inputs.md).

`--write` puts the accepted trims into the files. The command reads each file
again first and leaves it alone if it changed during the run or is read-only;
that document then counts as failed. A written file keeps its mode. Review the
change in your version control before you keep it.
`--write` cannot be combined with `--stdin`.

`--report json:PATH` writes the complete result with schema `unswell-trim-v1`:
for every document the accepted trims with byte spans, the original and the
trimmed sentence, the removed words, and the rejected proposals with the
reason. `--report text:PATH` writes the text above.

## Exit codes

| Code | Meaning |
| ---: | --- |
| 0 | Every selected document was processed, with or without trims |
| 2 | A document could not be processed, no selected document has a sentence to trim, a file could not be written, or the options were invalid |
| 130 | The run was interrupted |

The limits are those of `unswell review`: `--max-documents`,
`--max-document-bytes`, `--jobs`, `--attempts`, `--retry-delay`,
`--request-timeout`, `--timeout`, and `--max-tokens`. One document costs one
model call. An answer in which the model corrects itself is read as for the
review: the last object counts.

## Measured behavior

### Against what editors removed

The [human-reference record](../research/reviews/2026-11-12-human-reference/README.md)
has texts that people edited. The command ran on the versions before the edit,
with two models. A removed word is a hit when the editor removed or replaced
that word. Chance is what random removal of the same number of words from the
same sentences gives. The numbers are from the test parts, and the rule was
registered before the run. Brackets hold 95% intervals.

| Text | Trimming model | Removed words | Editor changed them | By chance in the same sentences | Difference, points |
| --- | --- | ---: | ---: | ---: | --- |
| Answers of chat models that experts edited | `claude-sonnet-5-5` | 1,617 | 71.2% | 62.7% | +8.5 (+6.3 to +10.6) |
| Answers of chat models that experts edited | `gpt-6-astra` | 2,110 | 68.0% | 59.3% | +8.7 (+6.7 to +10.7) |
| Model fiction that writers edited | `claude-sonnet-5-5` | 1,037 | 51.9% | 49.3% | +2.6 (+0.0 to +5.2) |
| Model fiction that writers edited | `gpt-6-astra` | 853 | 49.6% | 46.5% | +3.1 (-0.1 to +6.4) |
| Documentation that maintainers reworded | `claude-sonnet-5-5` | 375 | 8.5% | 6.6% | +2.0 (-0.7 to +5.7) |
| Documentation that maintainers reworded | `gpt-6-astra` | 942 | 6.6% | 4.4% | +2.2 (+0.6 to +4.4) |

On answers of chat models that experts edited, the trimmed text is also closer to the expert's text than the
original was. The word-level edit distance to the expert's text falls by 0.28 and 0.23 words per
removed word with the two models. Random removal from the same sentences gives 0.15 and 0.09,
because the experts removed or replaced more than half of the words. On fiction the distance does not fall
(+0.05 and +0.10), though it rises less than under random removal (+0.14 and +0.21).

By the registered rule the command agrees with editors above chance on answers
of chat models that experts edited, and there it moves the text toward the
editor's version. On fiction and on documentation the differences are positive and the
rule is not met: an interval reaches zero. The documentation set is pages that
people wrote, where a maintainer changed about 4% of the words.

Read the first two rows as a result on chat answers, not on documentation.
The answers come from ten open chat models of 2023 and early 2024, among them
Llama 2, Mistral 7B, Zephyr, and Gemma, and the experts removed or replaced
more than half of their words. No public set has documentation that models
wrote and people then edited, so that case is not measured. Editors also disagree with each other about which
words to change, so a word an editor kept is not shown to be needed.

### By a consensus of models

No reader judged the trimmed text. Four models did, by a rule about information: a
version is worse if it lacks a fact, condition, qualification, or instruction
that the other has; otherwise the shorter version is better when the extra
words carry no information. `claude-fable-5-2` has three votes, `gpt-6-astra`
two, `claude-sonnet-5-5` and `private-host/Qwen3.8-27B` one each, and a tie
goes to the first. [ADR 0053](adr/0053-trim-admitted-by-model-consensus.md)
records why the project accepts this judge.

| Trimmed paragraphs of generated documentation | Trimmed preferred | Original preferred | Trimmed among decisive answers |
| --- | ---: | ---: | ---: |
| Sentences that `unswell review` flagged | 16 | 4 | 80.0% of 20 |
| Random other sentences | 17 | 3 | 85.0% of 20 |
| `unswell trim` as built, on other documents | 33 | 7 | 82.5% of 40 |

Where the original is preferred, the consensus found that the trim lost
information. That is why the command proposes and does not decide.

The judges do not agree with each other. On the first row the single models
prefer the trimmed paragraph in 16, 9, 18, and 6 of 20 pairs, in the order above.
The two models of one family approve and hold four of the seven votes.
With one equal vote per model family, for `claude-fable-5-2`, `gpt-6-astra`, and
`private-host/Qwen3.8-27B`, the rule is not met: the trimmed paragraph is preferred in
11 of 20, 13 of 20, and 21 of 40 decisive answers on the three rows. The
[record](../research/reviews/2026-11-12-human-reference/RESULTS.md) keeps every
single answer.

## From Go

The package `github.com/stokaro/unswell/review` holds the prompt, the reading
of the answer, the check, and the application. It makes no model call.

```go
reviewer, err := review.New(review.Options{})
prepared, err := reviewer.Prepare(ctx, document.Source{Name: "guide.md", Bytes: text})
answer := askYourModel(prepared.TrimPrompt())
proposals, err := review.ParseTrims(answer)
trims, rejected, err := prepared.BindTrims(proposals)
revised := prepared.ApplyTrims(trims)
```
