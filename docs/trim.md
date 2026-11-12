# Trim

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
model call.

## Measured behavior

No person judged these trims. Four models did, by a rule about information: a
version is worse if it lacks a fact, condition, qualification, or instruction
that the other has; otherwise the shorter version is better when the extra
words carry no information. `claude-fable-5-2` has two votes,
`claude-sonnet-5-5`, `gpt-6-astra`, and `private-host/Qwen3.8-27B` one each,
and a tie goes to the first. [ADR 0053](adr/0053-trim-admitted-by-model-consensus.md)
records why the project accepts this judge.

| Trimmed paragraphs of generated documentation | Trimmed preferred | Original preferred | Trimmed among decisive answers |
| --- | ---: | ---: | ---: |
| Sentences that `unswell review` flagged | 16 | 4 | 80.0% of 20 |
| Random other sentences | 17 | 3 | 85.0% of 20 |
| `unswell trim` as built, on other documents | 33 | 7 | 82.5% of 40 |

Where the original is preferred, the consensus found that the trim lost
information. That is why the command proposes and does not decide.

The judges do not agree with each other. On the first row the single models
prefer the trimmed paragraph in 16, 18, 9, and 6 of 20 pairs, in the order above.
The numbers in the table depend on the two votes of the first model. The
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
