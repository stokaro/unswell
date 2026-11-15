# Text that a model wrote under a style guide

The AI-sounding index of `v0.1.0-alpha.12` counts six marks. A page that a
model wrote under a style guide has almost none of them, and the index gives it
the level `low`. This record measures whether such a page still reads as a
model's, what in the text shows it, and what the index does with it after a
second reading was added.

After the release a control was added: pages that people wrote under a style
guide. It asks whether the second reading rises on such pages too, and what the
reading follows.

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
7. After the release, the published binary on pages that people wrote under a
   style guide: `scripts/fetch_styled.py` collects them and
   `scripts/control_export.py` writes what the build counts on each.

## The control: people who write under a style guide

The human pages above follow no style guide. A style guide asks for plain
statements, and the second reading counts runs of plain statements. A model's
text under a style guide has a higher index than a model's text without one.
So the pages above cannot tell whether the reading rises on a model's text or
on any text that follows such a guide. The control was run after the release,
with its bars written on issue 577 first.

**Pages.** Nine public repositories whose projects publish a style guide for
their writers, each read at its last commit before 2022-11-01. Four are sources
of articles and tutorials: `GoogleChrome/web.dev`,
`GoogleChrome/developer.chrome.com`, `GoogleCloudPlatform/community`, and
`istio/istio.io`. Five are documentation: `gitlabhq/gitlabhq`, `mdn/content`,
`kubernetes/website`, `dotnet/docs`, and `alphagov/gds-way`. Markdown pages are
tried in the order of a hash of the path, and the first ones with 500 prose
words or more in English are kept: up to 120 for a source of articles, up to 60
for a source of documentation, 701 pages in all. `data/control.jsonl.gz` names
the repository, the path, the commit, and the style guide of each page of the
measured half.

**Preparation.** The preparation of the record, with two steps around it.
Before it, a macro that stands inside a sentence is replaced by its visible
words: a glossary link, a cross reference, a tooltip. Template tags go. After
it, a shortcode that runs over several lines goes too.

**Halves.** The same hash as above assigns each page to a half. The build read
the measured half, 339 pages. The other half was not read.

**Matched pages.** A run of sentences forms only in paragraphs: a list, a
table, or a code block ends it. A page of lists and code is quiet whoever
wrote it, and documentation is mostly such pages. A page is matched when at
least 73.2% of its prose words stand in paragraphs. That is the tenth
percentile of the posts of the style-guide set. The bars are on the matched
articles, the genre of the style-guide set.

**Build.** The published binary of `v0.1.0-alpha.13`.

**The sentences.** After the run, and without a bar, the sentences of the same
pages were counted: how long they are, how many are plain statements, and how
many of the plain ones stand in runs. The record keeps each sentence of a
running paragraph as its stretch, its words, and whether it is plain. They are
rebuilt from the result of the build: its blocks, its sentences, and the spans
it protects. The rebuilt count of sentences in runs equals the build's own on
329 of the 339 pages.

What the control shows and what it does not:

- People who write under these nine style guides put about as many sentences
  in runs as people who follow none, and the second reading is `high` on one
  page of 339.
- They also write longer sentences than the style-guide set: 17 words against
  12.5, with about a third of their sentences over 20 words, against 8%. The
  reading is quiet on their pages for that reason.
- The runs follow from the share of plain sentences. A page has about as many
  sentences in runs as chance gives for that share. The second reading is in
  effect a measure of how many sentences are short plain statements.
- The control does not show that the reading tells a model from a person who
  writes in that manner. Such a person gets the same reading. The control
  holds few such pages.

## The build behind the numbers

The counts, the readings, and the index of every page in `data/pages.jsonl.gz`
come from the build of this change, and so does the reading of the earlier
record.

The revisers were given the report of a build one step earlier. It differs
from the final one in one line of one page: the earlier build named runs of
sentences in the index line also when the second reading was `low`, and the
final build names them only above `low`. For that page, a post of the
style-guide set with an index of 40 by its marks and a second reading of 29,
the reviser read `8 list items that start with a bold label, 13 runs of plain
sentences, 5 runs of sentences that open with their subject` where the final
build prints `8 list items that start with a bold label`. The lines of its
marks are the same. On the other 178 revised pages the final build prints the
report that the reviser was given, byte for byte, and on all 179 it gives the
page and its revision the index that the record holds.

The published binary of `v0.1.0-alpha.13` was run on the same texts after the
release. It writes `data/pages.jsonl.gz` and `data/earlier.jsonl.gz` byte for
byte, and it agrees with the final build on every report and every index
above; `scripts/reproduce.py` is that check.

## Data

| File | Rows |
| --- | --- |
| `data/pages.jsonl.gz` | every page of both halves: its set, its half, the count and the sentences of each signal, the prose words, the two readings and the index of the build, and the index of `v0.1.0-alpha.12` |
| `data/judgments.jsonl.gz` | one answer of one judge on one pair of the measured half |
| `data/revisions.jsonl.gz` | each page that a reviser was given: the index and the readings before and after, the words, and the em dashes that the revision adds |
| `data/earlier.jsonl.gz` | the pages of the earlier record as the build reads them |
| `data/control.jsonl.gz` | each page of the control in the measured half: its repository, path, commit, and style guide, the share of its prose in paragraphs, the count and the sentences of each signal, the prose words, the two readings and the index, and its sentences in running paragraphs as stretch, words, and whether plain |

The record holds no text of any page. The texts of the public posts can be read
again from the repository, the path, and the commit of each row. The written
texts, the revisions, and the judges' reasons stay with the maintainer.

## Not measured

People as judges. Genres other than product and engineering blog posts.
Whether a revision keeps every fact of its page: a reviser can change what a
sentence says, and one such change showed on the fitting half. A person who
writes in the manner of the style-guide set, with short plain statements and
few long sentences: the reading is counted from the sentences, so such a page
gets the same reading whoever wrote it. Whether each page of the control
follows its style guide: a project that publishes a guide does not enforce it
on every page.
