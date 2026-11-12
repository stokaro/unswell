# ADR 0051: A review command reads documents with a model the user names

Status: proposed on November 12, 2026. It takes effect when the maintainer
merges it.

## Context

Unswell exists to reduce AI-sounding prose. Four rounds of review made the
default rules more often right:
[77.3% and 79.4%](../research/rule-qualification.md) of their findings are
accepted by both reviewers on new pages. They did not make the rules reach
more. On the 455 generated documents of the first corpus the default rules
touch 125 of the 909 items that both annotators marked for revision, 13.8%.
They report 0.65 findings per 1,000 words there and 0.84 on the 38
human-written originals. A fixed rule names a construction. Most of what the
annotators marked is not a construction: a sentence that repeats an earlier
one, an opening that adds nothing, a claim with no scope.

[ADR 0047](0047-qualified-defaults-and-assistant-review.md) answered this with
review by the calling assistant through the MCP server. On 120 real pages its
flags reach 47.5% of the items the other annotator marked, where the rules
reach 4.2%. That path needs an assistant session with an MCP client. A person
at a terminal and a CI job have no access to it.

Until now every Unswell program ran without a network, and the README said so.

## Decision

1. A public package, `review`, holds the review that is not a model call: the
   items of a document, the rubric, the prompts for a reviewer and for a
   confirming reviewer, the reading of their answers, and the binding of a flag
   to source bytes. It reads no file, starts no process, and uses no network.
   The MCP tools and the new command both use it.

2. `unswell review` calls a model the user names, either an endpoint that
   implements the OpenAI chat completions API or a local program that reads a
   prompt on standard input. There is no default model, no default endpoint,
   and no vendor SDK. A key is read from an environment variable the user
   names. Plain `http` is accepted for the local machine only, unless a flag
   permits it.

3. Nothing else changes its boundary. `unswell check`, the Go library, the
   GitHub Action, and the MCP server send nothing anywhere. `net/http` and
   `os/exec` stay confined to `internal/cli` by the repository policy.

4. Review findings are advisory. The command exits 0 when every selected
   document was reviewed, whatever was found. The findings do not enter a
   check result, a score, a baseline, or a gate, and they qualify no rule. A
   model's answer is not reproducible and is right only part of the time, which
   is the opposite of what a gate needs.

5. The command fails closed on everything except findings. A flag that does not
   quote its item is dropped and counted. A document that cannot be reviewed
   is reported as failed, the other documents are still reported, and the run
   exits 2. A refusal, an empty answer, and an answer without the requested
   object are failures, never a clean document.

6. A second model can judge the bound flags. It sees the document, each flag's
   category, quote, and item, and not the first model's reasons. With a
   confirming model the text report shows the flags it accepted; the JSON
   report keeps the others with the verdict, and `--show-unconfirmed` lists
   them. The measurement below decided this by a rule registered before the
   confirmed runs were read.

## Measurement

The command ran on 60 of the 455 generated documents, one process per
document, in four configurations. The
[qualification record](../research/rule-qualification.md#the-review-command-on-generated-documents)
reports all of them. Two runs have a reference annotator that is neither the
reviewer nor the confirming model. Pooled over those two:

| | Findings per 1,000 words | Marked items reached | Findings on a marked item |
| --- | ---: | ---: | ---: |
| Default rules | 0.49 | 38 of 521, 7.3% | 40 of 94, 42.6% |
| Reviewer's bound flags | 4.04 | 297 of 521, 57.0% | 297 of 778, 38.2% |
| Flags the second model kept | 2.56 | 236 of 521, 45.3% | 236 of 493, 47.9% |

The registered rule had two conditions. Kept flags had to lie on a marked item
at least 10 points more often than dropped flags: they do by 26.5 points, 47.9%
against 21.4%. Coverage after confirmation had to be at least two thirds of
coverage before: it is 79.5%. Both hold, so the text report shows confirmed
flags only.

The reference is another model's annotation under the same rubric, so these
numbers measure agreement between models. They are not recall and precision
against an editor, and no human reviewed the flags.

## Consequences

The README no longer says that Unswell as a whole sends no text to an AI
service. It says which programs send nothing and that `unswell review` sends
the selected documents to the model the user names.

Each run costs model calls and discloses text. Both are the user's decision per
run: without a named model the command fails before it reads a document, and
`--max-documents` refuses a large selection before any call.

The rubric and the prompts are now part of a public package. Changing them
changes what users' models are asked, so a change needs a new protocol name
and a new measurement.

Not done in this decision: SARIF output for review findings, splitting a
document that exceeds `--max-document-bytes`, a configuration section that
names the model, and reuse of answers between runs.
