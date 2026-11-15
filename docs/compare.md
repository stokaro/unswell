# Compare

Experimental. `unswell compare` checks a revision of a page. A model you name
reads the page and the revision without a hint of which is which, and names
the one that reads more like text written by a language model. The command
then accepts or rejects the revision.

The [AI-sounding index](ai-sounding-index.md) takes no part in this. The
index falls after almost any edit of a run of sentences, also after one that
makes the page read more like a model's; see
[Measured behavior](#measured-behavior).

```text
docs/guide.md against its revision docs/guide.revised.md
Sound: the revision reads less like a model's text. The judge named the page in both orders.
    Version A states every step in a sentence of the same shape.
    Version B has one cadence from the first paragraph to the last.
Facts: 41 of 41 statements kept.
Numbers, code, and links: nothing lost.
Em dashes: 0 in the page, 0 in the revision.
Accepted.
Judge: example-model at api.example.com.
The judge did not see which text is the revision. It is a model: name one other than the model that made the revision.
```

## What is checked

A revision is accepted only if all four checks pass.

**Sound.** The judge gets the opening of both texts, whole blocks up to about
450 words, as version A and version B. It is asked which reads more like text
written by a language model. The question is asked twice: once with the page
as version A and once with the revision as version A. The revision is better only when the judge names the page both
times. A judge that follows the order of the texts names the same letter
twice, which is a different text each time; that counts as no difference, and
the revision is rejected.

**Facts.** The model lists the statements of fact of the page. It is then
asked, for each statement, whether the revision still states it, and it has to
quote the words of the revision that do. A statement counts as kept only when
those words stand in the revision; the command checks that itself. At least
98% of the statements have to be kept.

**Numbers, code, and links.** Without a model: every number of the page's
prose, every code block and inline code, and every link has to occur in the
revision. A number that the revision writes as a word counts as lost.

**Em dashes.** The revision may not have more em dashes in its prose than the
page. A model that rewrites a sentence often adds one, and the index counts
it.

The four parts stand in the report one by one, so a rejected revision says
which check it failed.

## Exit code

| Code | Meaning |
| --- | --- |
| `0` | The revision is accepted. |
| `1` | The revision is rejected. |
| `2` | The comparison did not complete, or the options were invalid. There is no verdict. |
| `130` | The run was interrupted. |

A comparison that did not complete is never a pass. A judge that does not
answer in the requested form is asked again, up to `--attempts` times.

## What leaves the machine

The complete text of the page and of the revision goes to the endpoint or the
local program you name, with the instructions. An API key goes to that
endpoint in the `Authorization` header and nowhere else. Without a named model
the command fails before it reads a file.

## Run it

The model is named as for [`unswell review`](review.md#run-it):

```sh
unswell compare docs/guide.md docs/guide.revised.md \
  --endpoint https://api.example.com/v1/chat/completions \
  --model example-model --api-key-env REVIEW_API_KEY
```

or with a local program that reads the prompt on standard input:

```sh
unswell compare docs/guide.md docs/guide.revised.md --command assistant --command-arg=--print
```

The first path is the page and the second its revision. To check a change in
your working tree against the committed page, write the committed page to a
file first:

```sh
git show HEAD:docs/guide.md > /tmp/guide.before.md
unswell compare /tmp/guide.before.md docs/guide.md --command assistant --command-arg=--print
```

Name a judge other than the model that made the revision. A model tends to
prefer its own text.

`--format`, `--config`, `--profile`, and `--project-root` set how both texts
are read, as they do for [`unswell check`](inputs.md). A text larger than
`--max-document-bytes` fails before anything is sent.

`--report json:PATH` writes the complete result with schema
`unswell-compare-v1`: the judge's answer to each order with its reason, every
statement with its outcome and the quoted words, what was lost, the em dashes
of both texts, and the verdict with its reasons. `--report text:PATH` writes
the text above.

The command makes four calls: two for the sound and two for the facts. The
two for the facts carry the whole text of the page or of the revision.

## From Go

The `review` package holds the prompts, the parsers, and the rule, and makes
no call itself. `review.NewRevision` pairs two prepared documents.
`SoundPrompt`, `FactsPrompt`, and `KeptPrompt` return the text for a model;
`ParseSound`, `ParseFacts`, and `ParseKept` read its answers; `SoundVerdict`,
`BindFacts`, `Lost`, `EmDashes`, and `Judge` decide.
