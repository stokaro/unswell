# Two results of the check of a revision

November 17, 2026. `unswell compare` gave one verdict for two things: that a
revision loses nothing of its page, and that a judge hears it as less like a
model's text. The second does not repeat. On posts, where a judge mostly
hears no difference, a "better" from one run came again 3 times of 8 with one
judge and 0 of 3 with the other
([issue 596](https://github.com/stokaro/unswell/issues/596)). This record
measures the change that says the two apart and lets a caller require the
first alone.

[RESULTS.md](RESULTS.md) holds the numbers, and `make check-two-results`
recounts every one of them from the data here. The bars were written on
[issue 605](https://github.com/stokaro/unswell/issues/605) before the run on
the measured part. [ADR 0066](../../../docs/adr/0066-two-results-of-compare.md)
has the decision.

## Outcome

Eleven of the 14 bars can be read, and all eleven hold. Three are not read:
the run of the copies with one judge was stopped (see below).

- **The first result repeats.** The same revisions were checked twice with
  the same build. The result on nothing lost came out the same on 55 of 57
  with `gpt-6-astra` and on 49 of 57 with `private-host/Qwen3.6-35B-A3B`.
- **Harmless edits pass.** Of 155 copies that join two sentences, write
  negations as contractions, or leave out padding, 144 lose nothing with
  `private-host/Qwen3.6-35B-A3B`. The other 11 fail on a statement that the
  judge did not find, and on nothing else. With `gpt-6-astra`, all 93 that it
  judged lose nothing.
- **Damage does not pass.** No damaged copy loses nothing: 0 of 219 with one
  judge, and 0 of the 105 that the other judged, the 87 of the posts among
  them.
- **What the split buys.** With `kept` required the check accepts 19 and 17
  of 57 revisions in the two runs of one judge, and 41 and 43 in those of the
  other. With `all` required it accepts 6 and 5, and 12 and 13.
- **Verdicts.** Every pair that was run got a verdict on the first attempt.

One thing came out beside the bars. The two judges are far apart on how many
revisions lose nothing, 19 against 41 of 57, and each checked the revisions
of the other reviser. A control after the run tells the two apart: the judge
that found 41 then checked the 57 revisions that the other judge had checked,
and found nothing lost in 23 of them, where that judge found 19. So it is the
reviser. The revisions of `gpt-6-astra` lose nothing about twice as often as
those of `private-host/Qwen3.6-35B-A3B`, whichever judge reads them. The two
judges still differ on single revisions: they agree on 43 of the 57.

## The run that was stopped

The run of the copies with `gpt-6-astra` was stopped after 198 of 374
copies, when that judge was no longer available. Its two runs of the
revisions were complete by then. The copies were judged in the
order of the file, so the 198 are all 172 copies of the posts and 26 of the
202 copies of the fresh texts, and no sample of the rest.

So three bars of that judge are not read: harmless edits (93 of 155 judged),
damaged copies of the fresh texts (18 of 132), and verdicts on the copies.
The counts on what it did judge are in RESULTS.md beside those bars. The
issue keeps them open, and the remaining 176 copies can be run with the same
build.

## What the product gained

- **Two results.** The command says apart whether nothing is lost and whether
  the revision reads less like a model's text, each with its reasons.
- **`--require`.** With `all`, the default, a revision needs both, as before.
  With `kept` it needs the first alone. The judge is asked about the sound in
  either case, and the report says what it heard.
- **Quotes read contractions and capitals as the same words.** A statement is
  kept when the judge quotes words that stand in the revision. A quote with
  "does not" did not stand in a revision that says "doesn't", and one such
  statement rejects a short page. That came out of the fitting half.

## The fitting half

**The same revisions twice.** The 40 revisions of each reviser by the report,
in two runs. The result on nothing lost was the same on 40 of 40 with one
judge and on 36 of 40 with the other. No revision was accepted twice with all
required: 0 and 2 in the two runs of one judge, 2 and 1 in those of the
other, and none of them the same revision.

**Harmless copies.** 75 copies of the fitting posts with two sentences joined
or with contractions. With `gpt-6-astra` all 75 lose nothing. With
`private-host/Qwen3.6-35B-A3B`, 68 of 74 do, and 30 of the 35 with
contractions; with the quotes compared as this change says, 34 of 35.

The second runs of the revisions and the first two rows of harmless copies
come from the build of `v0.1.0-alpha.17`, whose checks are those of this
record apart from the quotes. The first runs of the revisions are the fitting
rows of the [record before this one](../2026-11-16-left-out-words/README.md).

## The measured part

**Pages.** As in the record before this one: the 30 posts of the measured
half of the style-guide set, and the 30 texts that two models wrote under a
style guide for that record. All of them have been read by earlier runs.

**Copies.** New ones, with a new salt, of sentences that no earlier run
changed: 219 damaged copies of six kinds, and 155 harmless ones of three
kinds. They were made and compared without a model before the bars were
written. That comparison found what the record before this one found: the
changed sentence is a suspect on every damaged copy, no harmless copy with
joined sentences or contractions has a suspect, and the command names the
padding on every trimmed copy.

**Revisions.** No new ones. The revisions of 27 posts and of 30 fresh texts
that each reviser made for the earlier records, each checked twice by the
judge of the other family.

## The build behind the numbers

The measured part comes from the build of commit `b4fcd8e6` of the branch of
this change. The final build differs from it in the version number and in
nothing that the command does.

The published binary of `v0.1.0-alpha.19` was run on the same texts after the
release. Without a model it finds the same suspects on all 374 copies. From
the kept answers of one judge it reaches the same results on the 57 revisions
of that judge's second run; `scripts/reproduce.py` is that check, and it
cannot reach a model. See the
[alpha.19 evidence](../../../docs/release/v0.1.0-alpha.19.md).

## Data

| File | Rows |
| --- | --- |
| `data/copies.jsonl.gz` | each copy that a judge was given: its kind, its pages, whether the check reached a verdict, whether nothing is lost and which checks failed, the verdict of the blind comparison, and whether the pair is accepted with all required |
| `data/revisions.jsonl.gz` | each revision with its judge and its run: the same parts, and how many statements were checked and found kept |
| `data/fitting.jsonl.gz` | each pair of the fitting half: the revisions in two runs, and the harmless copies with each judge and build |
| `data/same.jsonl.gz` | the control: each revision of one reviser under the judge of its own family, with the same parts |

The record holds no text of any page. A post of the style-guide set has the
number that it has in the earlier records. The copies, the revisions, the
statements, and the judges' quotes stay with the maintainer, and so do the
texts that the scripts in `scripts/` need.

## The control is no check

In the control a model judges revisions of its own family, which the command
warns against: a model tends to prefer its own text. That matters for the
blind comparison and little for a result that rests on quoted words, and the
control is read only on nothing lost. It was run after the bars had been
read, with the build of the measurement, and it moves no bar.

## Not measured

The copies that the stopped run did not reach. People as judges. Texts that
no run had read. Whether a revision that loses nothing is worth publishing.
A blind comparison that repeats: this change takes it out of the gate for
those who ask, and leaves it as it was.
