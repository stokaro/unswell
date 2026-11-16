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
Sentences: 9 of 58 compared sentences of the page not in the revision word for word; the judge found each of them there.
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
page with the revision. It reads every line of the page outside fenced code
blocks and front matter: running prose, list items, rows of tables, headings,
the text of an image, an indented block. A sentence is kept when it stands in
the revision near its place with the same words in the same order. Case,
punctuation, markup, and contractions do not count, and the sentence may
stand inside a longer one, so two sentences joined with "and" are both kept.
The words of a code span are words of its sentence, so a term that the
revision sets as code changes nothing. Each sentence takes its own words in
the revision: a sentence that the page says twice needs two places there.
Every other sentence is a suspect: the revision changed at least one of its
words.

For a suspect the command looks for a counterpart: one sentence of the
revision near its place, or two or three in a row, with at least six tenths
of its content words. That share is counted twice, over all its content words
and over those that the sentences around it in the page do not have. A
sentence with fewer than three content words, such as "We did not run it.",
is compared by all its words. A suspect without a counterpart is missing. One
whose counterpart has fewer negations lost a negation. Any other is reworded.

A suspect is cut when a sentence of the revision has its words in their
order with some left out, and no other word: the revision cut a clause, a
phrase, or a word out of the sentence. The command then names the words that
were left out, in the report and to the model.

So nothing that a sentence says can change without the command seeing it.
The model gets each suspect beside the sentences that the revision wrote at
its place, and nothing else of either text; a long list goes in questions of
twenty. A sentence of the revision that is a kept sentence of the page is not
shown: it says what its own sentence says. When the revision wrote nothing at
the place of a sentence, the sentence is gone, and the command reports it
without asking a model.

Most suspects are harmless: a reviser says a sentence in another order or in
words of its own. The model clears a suspect only with a quote that stands in
that place. For a missing sentence the quote also has to carry one of the
sentence's own words: a model that quotes the sentence beside the gap does
not clear it. When the model finds a sentence and its quote is not in that
place, the command asks once more about that sentence; a model that says yes
may quote the sentence it was asked about.

A cut sentence has two ways to be cleared. The model calls the words that
were left out padding: they told the reader nothing. The command refuses that
answer when the words hold a negation. Or the model quotes words of the
passage that say what was left out, and that quote needs a word which the
rest of the sentence does not have. What is left of a sentence shows only
that the rest is kept, so a model that quotes it clears nothing.

One suspect that is not cleared rejects the revision, and the report shows
the sentence with its line, and for a cut sentence the words that were left
out.

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
dashes of both texts, and the verdict with its reasons. A suspect whose
sentence the revision cut has the words that were left out in `cut`, and
`padding` is true when the judge called them padding. `--report text:PATH`
writes the text above.

The command makes five calls or more: two for the sound, three for the
facts, one more when a statement was not found the first time, one for every
twenty sentences that the revision does not keep word for word and wrote
something in the place of, and one more when a sentence was found without the
words of its place. The three for the facts carry the whole text of the page
or of the revision. A call about
sentences carries only those sentences and their places.

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

### Every changed sentence

The build of `v0.1.0-alpha.15` named what makes a sentence a suspect: no
counterpart, or one negation less. A revision that kept the content words of
a sentence and changed what it says passed: "after" for "before", a "not" put
in. The rule is now the other way round: a sentence is kept only when it
stands in the revision word for word, and every other sentence goes to the
judge. A third measurement read that rule, with the bars written on issue 595
before the run; its
[record](../research/reviews/2026-11-16-changed-sentences/RESULTS.md) recounts
the numbers, and [ADR 0063](adr/0063-every-changed-sentence.md) has the
decisions. Of 32 bars, 31 hold. The published binary of `v0.1.0-alpha.16`
names the same suspects on all 553 copies of the record; see the
[alpha.16 evidence](release/v0.1.0-alpha.16.md).

Each page got copies with one sentence changed in one of five ways. The pages
are the 30 posts, which earlier measurements had read, and 30 texts that no
judge had read.

| What | Posts | Fresh texts |
| --- | --- | --- |
| The changed sentence is a suspect, without a model | 96 of 96 | 102 of 102 |
| A harmless copy has a suspect, without a model | 0 of 56 | 0 of 41 |
| The check reports the damage, judge `gpt-6-astra` | 96 of 96 | 101 of 102 |
| The check reports the damage, judge `private-host/Qwen3.6-35B-A3B` | 95 of 96 | 97 of 102 |
| Damaged copies accepted, the two judges | 0 and 0 | 0 and 2 |

| Judge | A sentence removed | "not" removed | "not" put in | A word turned around | A clause cut off |
| --- | --- | --- | --- | --- | --- |
| `gpt-6-astra` | 60 of 60 | 14 of 15 | 53 of 53 | 54 of 54 | 16 of 16 |
| `private-host/Qwen3.6-35B-A3B` | 60 of 60 | 14 of 15 | 53 of 53 | 54 of 54 | 11 of 16 |

**The bar that does not hold** is the last cell: a sentence that lost its
last clause, with one judge. The command showed the judge the sentence beside
what was left of it, and on 5 of 16 copies the judge found the sentence
there. Both damaged copies that the check accepted are of this kind. So the
choice of the judge matters: the command decides which sentences the judge
is shown, and the judge decides what it reads in them.

**Ordinary revisions pay for it.** Almost every revision by the report has a
sentence that is not kept word for word, and the judge clears most of them:
260 of 265 and 471 of 507 on the posts. A sentence that it does not clear
stands in 5 and in 13 of 27 revisions of posts. On the fitting half the
earlier rule left 5 and 6 of 40 revisions with such a sentence, and this one
leaves 9 and 14. The report names each such sentence with its line.

### The words that a revision leaves out

That measurement left a bar open, and two things behind it. A judge that was
shown "A when B." beside "A." answered that the sentence is there and quoted
"A". And when a page said a sentence twice, the judge was shown the kept
copy. The command now names the words that a revision leaves out of a
sentence, takes no quote of what is left of it, and shows the judge only what
the revision wrote. A fourth measurement read that, with the bars written on
issue 599 before the run; its
[record](../research/reviews/2026-11-16-left-out-words/RESULTS.md) recounts
the numbers, and [ADR 0064](adr/0064-words-left-out-of-a-sentence.md) has the
decisions. All 38 bars hold.

The pages are the 30 posts and 30 texts that no run had read. The copies have
one more kind of damage, the last words of a sentence cut off from a
preposition, and one more harmless kind: a revision that leaves out a phrase
of padding, such as "It is worth noting that".

| What | Posts | New texts |
| --- | --- | --- |
| The changed sentence is a suspect, without a model | 112 of 112 | 148 of 148 |
| Of them, reported without a question to a model | 30 | 30 |
| A harmless copy has a suspect, without a model | 0 of 56 | 0 of 40 |
| The check reports the damage, judge `gpt-6-astra` | 111 of 112 | 148 of 148 |
| The check reports the damage, judge `private-host/Qwen3.6-35B-A3B` | 112 of 112 | 147 of 148 |
| Damaged copies accepted, the two judges | 0 and 0 | 0 and 0 |
| The check reports a sentence that only lost padding, the two judges | 0 and 0 of 30 | 0 and 0 of 30 |

| Judge | A sentence removed | "not" removed | "not" put in | A word turned around | A clause cut off | Last words cut off |
| --- | --- | --- | --- | --- | --- | --- |
| `gpt-6-astra` | 60 of 60 | 18 of 18 | 47 of 47 | 52 of 52 | 23 of 23 | 59 of 60 |
| `private-host/Qwen3.6-35B-A3B` | 60 of 60 | 18 of 18 | 47 of 47 | 52 of 52 | 23 of 23 | 59 of 60 |

A removed sentence no longer depends on a judge: the revision wrote nothing
at its place, so the command reports it. A cut clause is reported on every
copy with each judge. Each judge called one cut of last words padding.

**Ordinary revisions pay again.** A sentence that the check does not clear
stands in 4 and in 15 of 27 revisions of posts, where it stood in 5 and in
13, and in 8 and in 24 of 30 revisions of fresh texts, where it stood in 7
and in 19.

**The blind comparison is not stable.** The fitting half showed it by the
way. The check had accepted 4 of 40 revisions of one reviser; asked again
with the same prompt, the judge named the page in both orders for 1 revision,
where it had done so for 5, and the check accepted none.

## Limits

- The check is strict, and most revisions get "no difference". That is not a
  finding that a revision is no better. It says that this judge did not name
  the page in both orders.
- The judge reads the opening of both texts. A revision that changes only
  what comes later is judged on text that is the same in both.
- The search for statements depends on the judge, and it verifies a quote,
  not its meaning.
- The command shows the judge every sentence that the revision changed. What
  the judge reads in it is the judge's. For a sentence that only lost words,
  the command names the words and takes no quote of the remains; whether the
  words were padding is still the judge's call. In the measurement each judge
  called one of 60 cuts of last words padding.
- A sentence that a revision cuts and also rewords is not a cut sentence. It
  goes to the judge as a reworded one, without a line for what was left out.
- Words that a revision puts in around a kept sentence are not shown to the
  judge: a sentence is kept when it stands in the revision word for word,
  also inside a longer one.
- The answer of a judge to the blind comparison of the sound can differ from
  one run to the next.
- The check rejects many ordinary revisions for one sentence. A reviser that
  rewords freely drops a detail somewhere, and one sentence that the judge
  does not find rejects the revision. The report names the sentence with its
  line; read it and decide.
- A paraphrase that keeps none of a sentence's own words is not cleared,
  whatever the judge says.
- What a page says across several sentences, such as the order of steps, is
  compared sentence by sentence. A sentence that the revision keeps word for
  word and moves a few sentences away is kept.
- Fenced code is compared as code: a block with one changed line is reported
  as lost and is not shown to the judge.
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
