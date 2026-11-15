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
    Text A states every step in a sentence of the same shape.
    Text B has one cadence from the first paragraph to the last.
Facts: 41 of 41 statements kept.
Sentences: each of the 58 compared sentences of the page has its counterpart in the revision.
Numbers, code, and links: nothing lost.
Em dashes: 0 in the page, 0 in the revision.
Accepted.
Judge: example-model at api.example.com.
The judge did not see which text is the revision. It is a model: name one other than the model that made the revision.
```

## What is checked

A revision is accepted only if all five checks pass.

**Sound.** The judge gets the opening of both texts, whole blocks up to about
450 words, as text A and text B. It is asked which reads more like text
written by a language model. The question is asked twice: once with the page
as text A and once with the revision as text A. The question does not say
that one text is a revision of the other. The revision is better only when
the judge names the page both times. A judge that follows the order of the
texts names the same letter twice, which is a different text each time; that
counts as no difference, and the revision is rejected.

**Facts.** The model lists the statements of the page: its facts and its
instructions to the reader. It is then asked for each statement in the page
itself, and it has to quote the words
that state it. A statement that the model does not find in the page it took it
from is left out: the model worded it so that the page does not say it, and
it would be missing from any revision. The model is then asked for the other
statements in the revision, again with a quote for each. A statement counts as
kept only when the quoted words stand in the revision; the command checks that
itself. At least 98% of the statements have to be kept. Before a statement
counts as not kept, the command asks once more about it, in a list of those
statements only. A model errs on a long list: it paraphrases a quote, or
calls many statements at the end of the list changed.

**Sentences.** Without a model, the command compares each sentence of the
page with the revision. A sentence has its counterpart when one sentence of
the revision near its place, or two or three in a row, carry at least six
tenths of its content words. That share is counted twice: over all its
content words, and over those that the sentences around it in the page do not
have. The sentences around a removed sentence share its subject; they do not
have its own words. A sentence is a suspect when it has no counterpart, or
when its counterpart has fewer negations than it has. A sentence with fewer
than three content words is not compared.

Most suspects are harmless: a reviser may say "rather than" where the page
said "not", or put a sentence in words of its own. So the model is asked
about each suspect, with the sentence beside the sentences of the revision
from its place, and nothing else of either text. It clears a suspect only
with a quote that stands in that place. For a missing sentence the quote also
has to carry one of the sentence's own words: a model that quotes the
sentence beside the gap does not clear it. One suspect that is not cleared
rejects the revision, and the report shows the sentence with its line.

**Numbers, code, and links.** Without a model: every number of the page's
prose, every code block and inline code, and every link has to occur in the
revision. A number that the revision writes as a word counts as lost.

**Em dashes.** The revision may not have more em dashes in its prose than the
page. A model that rewrites a sentence often adds one, and the index counts
it.

The five parts stand in the report one by one, so a rejected revision says
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
`unswell-compare-v2`: the judge's answer to each order with its reason, every
statement with its outcome and the quoted words, the statements that were
left out, the number of sentences that were compared and each suspect with
the place it was looked for and what the judge said, what was lost, the em
dashes of both texts, and the verdict with its reasons. `--report text:PATH`
writes the text above.

The command makes five to seven calls: two for the sound, three for the
facts, one more when a statement was not found the first time, and one when
a sentence is a suspect. The three for the facts carry the whole text of the
page or of the revision. The call about suspects carries only those sentences
and their places.

## Measured behavior

The command was measured on posts of a product blog that a model wrote under
a style guide and on texts that two models wrote under a style guide. Two
models revised each page by the report of `unswell check`, and a judge of the
other model family ran the check. The numbers are on the measured half, with
the bars written on issue 583 before the run. The
[record](../research/reviews/2026-11-15-revision-check/RESULTS.md) recounts
them, and [ADR 0061](adr/0061-blind-check-of-a-revision.md) has the
decisions. The published binary of `v0.1.0-alpha.14` prints the report that
each reviser was given, and it reaches the recorded verdict on the 30 pairs
that could be replayed from kept answers of the judge; see the
[alpha.14 evidence](release/v0.1.0-alpha.14.md).

**It accepts few revisions.**

| Reviser | Judge of the check | Pairs | Accepted | Sound better, no difference, worse | Second text named in both orders |
| --- | --- | --- | --- | --- | --- |
| `gpt-6-astra` | `private-host/Qwen3.6-35B-A3B` | 86 | 15 | 27, 57, 2 | 55 |
| `private-host/Qwen3.6-35B-A3B` | `gpt-6-astra` | 57 | 15 | 19, 25, 13 | 21 |

The check reached a verdict on every pair. On most pairs the judge named the
second text in both orders, which is the page once and the revision once.
That is why the command asks twice: one answer to one order looks like a
preference and is mostly the order.

**What it accepts is more often better.** A third model, which revised no
page and judged in no check, read each page and its revision once. On the 30
accepted revisions it named the page as more like a model's text on 21 and
the revision on 9. On the 113 rejected ones it named the page on 57 and the
revision on 53.

**It can find a revision better.** On 30 pairs the difference is known to be
large: a text that a model wrote under a style guide stands as the page, and
the human post on the same subject stands in the place of the revision. With
`private-host/Qwen3.6-35B-A3B` as the judge, the comparison named the model's
text in both orders on 26 of them. On the other 4 it named the second text
both times.

**Em dashes.** With the line of advice that the report has now, none of 86
revisions of one reviser and 1 of 86 of the other added an em dash.

### The sentences and the statements

The build of `v0.1.0-alpha.14` had no comparison of sentences. Its search for
statements reported damage made on purpose on 39 and on 48 of 58 copies, and
the check accepted 3 and 2 damaged copies. One judge found 97.8% of the
statements kept, under 98%, because it quoted words that are not in the
revision. The comparison of sentences and the second question about a
statement answer that. A second measurement, with the bars written on issue
589 before the run, read them; its
[record](../research/reviews/2026-11-15-sentence-check/RESULTS.md) recounts
the numbers, and [ADR 0062](adr/0062-sentences-of-a-revision.md) has the
decisions. The published binary of `v0.1.0-alpha.15` names the same suspects
on all 421 copies of the record; see the
[alpha.15 evidence](release/v0.1.0-alpha.15.md).

The style-guide set has no post that no run had read: its 30 measured posts
were read once by the first measurement. So every bar was read on those posts
with new damaged copies, and on 30 texts that no run had read, which two
models wrote under a style guide. All 22 bars hold.

**Damage is reported.** Each page got a copy with one sentence removed and a
copy with the word "not" removed from one sentence.

| What | Posts | Texts that no run had read |
| --- | --- | --- |
| The damaged sentence is a suspect, without a model | 56 of 56 | 41 of 41 |
| The check reports it, with `gpt-6-astra` as the judge | 56 of 56 | 41 of 41 |
| The check reports it, with `private-host/Qwen3.6-35B-A3B` | 56 of 56 | 41 of 41 |
| Damaged copies accepted, either judge | 0 | 0 |

**Harmless copies are left alone.** A copy with two sentences joined, or with
its negations written as contractions, got a suspect on none of 56 and none
of 41.

**The statements are found kept.** Over the revisions of the first
measurement the check finds 99.8% of 3,039 statements kept with
`private-host/Qwen3.6-35B-A3B` as the judge and 99.5% of 3,216 with
`gpt-6-astra`. Over revisions of the texts that no run had read it finds
99.4% and 99.2%.

**A sentence alone rejects some ordinary revisions.** Of 86 and 57 revisions
by the report, 18 and 25 had a suspect. The judge cleared 18 of 25 suspects
and 49 of 68, and a sentence that it did not clear stood in 7 and 10
revisions. One of the two revisers drops or rewords a sentence more often
than the other.

## Limits

- The check is strict, and most revisions get "no difference". That is not a
  finding that a revision is no better. It says that this judge did not name
  the page in both orders.
- The judge reads the opening of both texts. A revision that changes only
  what comes later is judged on text that is the same in both.
- The search for statements depends on the judge, and it verifies a quote,
  not its meaning.
- The comparison of sentences reads words. A revision that keeps the words of
  a sentence and changes what it says, with "before" where the page said
  "after", has its counterpart. The damage that was measured is a removed
  sentence and a removed "not".
- A paraphrase that keeps none of a sentence's own words is not cleared,
  whatever the judge says: the revision is rejected, and the report names the
  sentence. Read it and decide.
- The words are English. The lists of function words, negations, and endings
  are English lists.
- An accepted revision more often reads less like a model's text than the
  page did. It is still told from text that a person wrote; see
  [a revised post against human text](ai-sounding-index.md#a-revised-post-against-human-text).
- The judges of the measurement are models. No person took part.

## From Go

The `review` package holds the prompts, the parsers, and the rule, and makes
no call itself. `review.NewRevision` pairs two prepared documents.
`SoundPrompt`, `FactsPrompt`, `ControlPrompt`, `KeptPrompt`, and
`SuspectsPrompt` return the text for a model; `ParseSound`, `ParseFacts`,
`ParseKept`, and `ParseSuspects` read its answers; `SoundVerdict`,
`Confirmed`, `BindFacts`, `Suspects`, `BindSuspects`, `Lost`, `EmDashes`, and
`Judge` decide. `Suspects` needs no model.
