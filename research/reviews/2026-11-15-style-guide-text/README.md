# Text that a model wrote under a style guide

The AI-sounding index of `v0.1.0-alpha.12` counts six marks. A page that a
model wrote under a style guide has almost none of them, and the index gives it
the level `low`. This record measures whether such a page still reads as a
model's, what in the text shows it, and what the index does with it after a
second reading was added.

[RESULTS.md](RESULTS.md) holds the numbers. `make check-style-guide-text`
recounts every one of them from the data here. The bars were written on
[issue 577](https://github.com/stokaro/unswell/issues/577) before each run.

## Texts

**Human blog posts.** 151 posts from the blogs of 11 public repositories:
`ariga/atlas`, `ent/ent`, `supabase/supabase`, `vitessio/website`,
`tikv/website`, `prometheus/docs`, `grpc/grpc.io`, `golang/website`,
`linkerd/website`, `kubernetes/website`, and `nats-io/nats-site`. Each
repository is read at its last commit before 2022-11-01, so a post is the text
of that day. Kept are English posts of 500 prose words or more, up to 14 for
each source, in the order of a hash of the path. `data/pages.jsonl.gz` names
the repository, the path, and the commit of each.

**README files.** The 566 README files written before November 2022 of the
[earlier record](../2026-11-12-human-reference/README.md), with their ids.

**The style-guide set.** 70 posts of one product blog, 500 prose words or more
each, that a model wrote under the blog's style guide and that were published
without edits by a person. The source is not named. The record holds the
readings of each post under a number and nothing else about it.

**Written texts.** For each human blog post one model reduces the post to an
outline of its facts. Two writers, `private-host/Qwen3.6-35B-A3B` and
`gpt-6-astra` at low reasoning effort, each write a post of about 700 words
from the title and the outline, without the human post: once with no
instruction about style (plain) and once under a style guide (guided). The
style guide of a post is one of two by a hash of its id: A, the sections on
voice, plain English, and anti-slop rules of the documentation style guide of
an open-source project; B, a digest of the Google developer documentation
style guide. Two guided texts of the first writer stayed under the floor of 450
words after two requests and are left out.

**Preparation.** Every post goes through the same steps before a judge or the
build reads it: front matter, components, images, bylines, and typographic
quotes go, a code block becomes one placeholder line, and each paragraph
becomes one line. A judge reads the opening of a post, whole blocks up to about
450 words. The README files go to the build as they are.

**Halves.** A hash assigns each human page and each post of the style-guide
set to the fitting half or to the measured half. A written text is in the half
of its human post. Signals, weights, and the wording of the report were chosen
on the fitting half. Nothing of the measured half was read before its bars were
written, the judges' reasons included.

## Judges

Three judges, blind to which text is which: `private-host/Qwen3.8-27B`,
`private-host/Qwen3.6-35B-A3B`, and `gpt-6-astra` at low reasoning effort.
Each sees two openings as A and B in an order fixed by a hash, and names the
one that reads more like text written by a language model, or calls them the
same. The panel sums the three votes of a unit. A unit is one pair, or one post
of the style-guide set with its three pairs against human posts.

## What was run

1. The judges on the measured half: plain texts against human posts (the
   admission), guided texts against human posts, and the style-guide set
   against human posts.
2. The judges on the fitting half, to read their reasons. They name one thing
   above all: uniform sentences.
3. A build that counts runs of sentences on the fitting half, to fit the second
   reading and to keep the human pages quiet.
4. That build on the measured half and on the pages of the earlier record.
5. A model that revises pages of the fitting half by the report, with several
   wordings of the remedy, and two judges on the result.
6. A model that revises the pages of the measured half by the final report, and
   the three judges on the result.

## Data

| File | Rows |
| --- | --- |
| `data/pages.jsonl.gz` | every page of both halves: its set, its half, the count and the sentences of each signal, the prose words, the two readings and the index of the build, and the index of `v0.1.0-alpha.12` |
| `data/judgments.jsonl.gz` | one answer of one judge on one pair of the measured half |
| `data/revisions.jsonl.gz` | each page that a reviser was given: the index and the readings before and after, the words, and the em dashes that the revision adds |
| `data/earlier.jsonl.gz` | the pages of the earlier record as the build reads them |

The record holds no text of any page. The texts of the public posts can be read
again from the repository, the path, and the commit of each row. The written
texts, the revisions, and the judges' reasons stay with the maintainer.

## Not measured

People as judges. Genres other than product and engineering blog posts.
Whether a revision keeps every fact of its page: a reviser can change what a
sentence says, and one such change showed on the fitting half. A person who
writes under the same style guide: the human pages here are from before
November 2022 and mostly follow no such guide, so the second reading may also
rise on a careful person's page of short plain statements.
