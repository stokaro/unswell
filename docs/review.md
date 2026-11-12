# Review in context

`unswell check` applies fixed rules. A rule names a construction. It cannot
tell that a sentence repeats an earlier one, that an opening adds nothing, or
that a claim has no scope. A reader can, and `unswell review` uses a model as
that reader.

The command gives each document to a model you name. The model reads the whole
document and flags items under a fixed rubric. Unswell binds every flag to
source bytes and drops a flag that does not quote its item. A second model can
then judge the bound flags one by one.

The findings are a model's judgment. They are advisory. They never enter
`unswell check`, its scores, or its gate, and the same document can get
different flags on a second run.

## What leaves the machine

`unswell check`, the Go library, the GitHub Action, and the MCP server send
nothing anywhere. `unswell review` is the exception, and only on your
instruction. It sends the complete text of every selected document, with the
rubric, to the endpoint or the local program you name. An API key goes to that
endpoint in the `Authorization` header and nowhere else. It is not written to
a report or to an error message.

Without a named model the command fails before it reads a document. Review a
document only where your policy lets that text reach that model.

## Run it

With an endpoint that implements the OpenAI chat completions API:

```sh
export REVIEW_API_KEY=...
unswell review README.md docs \
  --endpoint https://api.example.com/v1/chat/completions \
  --model example-model \
  --api-key-env REVIEW_API_KEY
```

`--api-key-env` names the environment variable that holds the key, so the key
never appears in a command line. A model served on this machine needs no key:

```sh
unswell review README.md \
  --endpoint http://localhost:8080/v1/chat/completions --model local-model
```

Plain `http` is accepted for this machine only. For another host use `https`,
or pass `--insecure-endpoint` when the network between the two is yours.

A local program can be the model. It must read the prompt on standard input
and print the answer on standard output, as the command-line client of an
assistant does:

```sh
unswell review README.md --command assistant --command-arg=--print
```

The program runs by its arguments, without a shell, in an empty temporary
directory, so it finds no project file there. Repeat `--command-arg` for each
argument, in order. Give an assistant's client the arguments that make it
answer from the prompt alone: no tools, no saved session, and no instructions
of its own.

Paths, `--stdin`, `--filename`, `--format`, `--config`, `--profile`, and
`--project-root` select and extract documents exactly as they do for
[`unswell check`](inputs.md). Extraction decides what counts as prose, so the
model is asked about the same text that the rules see: code blocks, protected
identifiers, and excluded regions are context, not items.

## A second model

Name a confirming model with the same flags under the `confirm-` prefix:

```sh
unswell review docs \
  --endpoint https://api.example.com/v1/chat/completions \
  --model example-model --api-key-env REVIEW_API_KEY \
  --confirm-model another-model
```

`--confirm-model` alone uses the reviewer's endpoint and key. A different
service takes `--confirm-endpoint` and `--confirm-api-key-env`; a local program
takes `--confirm-command` and `--confirm-command-arg`.

The confirming model sees the document, each bound flag's category, its quote,
and its item. It does not see the first model's reasons. It answers `yes`,
`no`, or `uncertain` for every flag.

With a confirming model the text report lists only the flags it answered
`yes` to. The others are counted there and kept in the JSON report with the
verdict. `--show-unconfirmed` lists them in the text report as well, marked
`review, not confirmed`. [Measured behavior](#measured-behavior) shows what
confirmation keeps and what it costs.

## What is reported

```text
docs/guide.md:3:1: review [empty_framing] The opening only announces the statement.
    "It is important to note that"
Reviewed 1 document: 1 finding.
Reviewer: example-model at api.example.com. No second model confirmed these findings.
Findings are a model's judgment, bound to source bytes. They are advisory and never gate.
```

Each finding names the position of the quoted text, the rubric categories, the
model's reason, and the quote as it occurs in the source. The seven categories
are `empty_framing`, `needless_repetition`, `wordiness`,
`unjustified_intensifiers`, `vague_claims`, `formulaic_transitions`, and
`needless_complexity`.

A flag is dropped, and counted, when it names an item that does not exist or
was already flagged, has no category or reason, or quotes text that does not
occur inside one segment of its item. A quote never crosses protected inline
code. Runs of white space compare equal, so a quote copied across a wrapped
line still binds.

`--report json:PATH` writes the complete result with schema
`unswell-review-v1`: the reviewer and the confirming model by name and host,
and for every document its findings with byte spans, the flags the confirming
model did not accept with its verdict, and the dropped flags with the reason.
`--report text:PATH` writes the text above. Repeat `--report` for several
outputs; `-` is standard output. SARIF and HTML are not produced.

Progress goes to standard error, one line per finished document. `--quiet`
turns it off.

## Exit codes

| Code | Meaning |
| ---: | --- |
| 0 | Every selected document was reviewed, with or without findings |
| 2 | A document could not be reviewed, no selected document has prose, or the options were invalid |
| 130 | The run was interrupted |

There is no exit code for findings. A model's judgment is not reproducible and
is right only part of the time, so it cannot fail a build. When one document
fails, the others are still reported and the run ends with code 2: an
incomplete review never looks complete. A run that reaches `--timeout` reports
the documents it finished and counts the rest as failed.

## Limits

| Flag | Default | Meaning |
| --- | ---: | --- |
| `--max-documents` | 50 | The run refuses a larger selection before any call |
| `--max-document-bytes` | 200000 | A larger document fails instead of being sent |
| `--jobs` | 2 | Documents reviewed at the same time, 1 to 16 |
| `--attempts` | 3 | Calls per question, 1 to 5. A rejected key or request is not repeated |
| `--retry-delay` | 2s | Delay before a repeated call; it grows with each attempt |
| `--request-timeout` | 5m | Limit for one model call |
| `--timeout` | 2h | Limit for the whole run |
| `--max-tokens` | 0 | Sent as `max_tokens` when positive. A model that reasons before it answers may need it |

One document costs one call to the reviewer and, when it has bound flags, one
call to the confirming model. Each call carries the whole document. A model
returns at most 256 flags for one document; more is a failed document.

A document without prose is counted and not sent.

## Measured behavior

The measurement ran the command on 60 documents that models wrote for the first
corpus, one process per document. Two annotators had labeled these documents
earlier. The comparison for each run uses the marks of the annotator that took
no part in it. The
[qualification record](research/rule-qualification.md#the-review-command-on-generated-documents)
has every run.

| | Findings per 1,000 words | Marked items reached | Findings on a marked item |
| --- | ---: | ---: | ---: |
| Default rules of `unswell check` | 0.49 | 7.3% | 42.6% |
| Review by one model | 4.04 | 57.0% | 38.2% |
| Review confirmed by a second model | 2.56 | 45.3% | 47.9% |

A marked item is a sentence or fragment that the reference annotator labeled
for revision. The third column is the share of marked items that a finding
touches. The last column is the share of findings that lie on a marked item.

A second model drops 285 of 778 flags and a fifth of the reach. A flag it
drops lies on a marked item 21.4% of the time, a flag it keeps 47.9%. Use it
when a shorter list matters more than reach.

Read these numbers as agreement between models, not as recall and precision
against an editor. The reference is a model's annotation under the same
rubric, and no human reviewed these flags. A finding on an unmarked item is not
shown to be wrong: on real documentation pages a second model accepted 70.1% of
the flags it was asked about directly, and both reviewers accepted 77.3% of the
default rules' findings. Each configuration ran once, and a model's answers
differ between runs.

## From Go

The package `github.com/stokaro/unswell/review` holds everything except the
model call. It reads no file and uses no network.

```go
reviewer, err := review.New(review.Options{})
prepared, err := reviewer.Prepare(ctx, document.Source{Name: "guide.md", Bytes: text})
answer := askYourModel(prepared.Prompt())
flags, err := review.ParseFlags(answer)
findings, dropped, err := prepared.Bind(flags)
```

`Document.ConfirmationPrompt` and `review.ParseVerdicts` do the same for a
second model. The MCP tools `unswell_review_request` and
`unswell_review_submit` use the same package; there the calling assistant is
the reviewer. See [MCP](mcp.md#contextual-review-by-the-calling-assistant).
