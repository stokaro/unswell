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

All 14 bars hold.

- **The first result repeats.** The same revisions were checked twice with
  the same build. The result on nothing lost came out the same on 55 of 57
  with `gpt-6-astra` and on 49 of 57 with `private-host/Qwen3.6-35B-A3B`.
- **Harmless edits pass.** Of 155 copies that join two sentences, write
  negations as contractions, or leave out padding, all 155 lose nothing with
  `gpt-6-astra`, and 144 do with `private-host/Qwen3.6-35B-A3B`. The other 11
  fail on a statement that the judge did not find, and on nothing else.
- **Damage does not pass.** No damaged copy loses nothing: 0 of 219 with
  either judge.
- **What the split buys.** With `kept` required the check accepts 19 and 17
  of 57 revisions in the two runs of one judge, and 41 and 43 in those of the
  other. With `all` required it accepts 6 and 5, and 12 and 13.
- **Verdicts.** Every pair has a verdict. With `gpt-6-astra`, 121 copies got
  none on the first attempt and got one on their one retry (see below).

One thing came out beside the bars. The two judges are far apart on how many
revisions lose nothing, 19 against 41 of 57, and each checked the revisions
of the other reviser. Two controls after the run tell the reviser from the
judge: each judge then checked the 57 revisions of its own family, which the
other judge had checked.

| Revisions of | Nothing lost with `gpt-6-astra` | With `private-host/Qwen3.6-35B-A3B` |
| --- | ---: | ---: |
| `private-host/Qwen3.6-35B-A3B` | 19 of 57 | 23 of 57 |
| `gpt-6-astra` | 49 of 57 | 41 of 57 |

So it is the reviser. The revisions of `gpt-6-astra` lose nothing about twice
as often as those of `private-host/Qwen3.6-35B-A3B`, whichever judge reads
them: 49 against 19 with one judge, and 41 against 23 with the other.

The judge matters less. Each judge finds nothing lost more often in the
revisions of its own family, by 8 and by 4 of 57, where two runs of one judge
differ by 2. On single revisions the two judges agree on 43 of the 57 of one
reviser and on 47 of the 57 of the other.

## The run in three parts

The run of the copies with `gpt-6-astra` was made in three parts, because
that judge was twice no longer available. Its two runs of the revisions were
complete before the first stop. Every part used the same build, the same file
of copies, and the same order.

- **Part 1.** 198 copies got a verdict: all 172 of the posts and 26 of the
  202 of the fresh texts. An earlier state of this record ended here: it read
  eleven bars and left three of that judge unread.
- **Part 2.** 55 more copies got a verdict. Then the judge stopped answering,
  and the runner did not notice. It started the other 121 checks, and the
  judge's program returned an error and no answer to each of them.
- **Part 3.** Those 121 were pairs without a verdict. The rule on the issue
  gives such a pair one retry, and each of them got a verdict on it.

The bar of verdicts is read after that retry, as the rule says: 374 of 374.
On the first attempt the number is 253 of 374, and RESULTS.md gives it under
the bars. What would be done with the second and the third part was written
on the issue before either was read.

The copies were judged in the order of the file, so the parts hold different
pages. RESULTS.md has the counts by part. In each part every harmless copy
loses nothing, and no damaged copy does.

The runner now starts no further pair once a check has failed and the judge
answers no question at all. A pair that was not started has had no attempt.

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
of that judge's second run, on the 57 of its control, and on the 120 copies
that the kept answers still cover; `scripts/reproduce.py` is that check, and
it cannot reach a model. See the
[alpha.19 evidence](../../../docs/release/v0.1.0-alpha.19.md).

## Data

| File | Rows |
| --- | --- |
| `data/copies.jsonl.gz` | each copy that a judge was given: its kind, its pages, the part of the run, whether the check reached a verdict and whether it needed its retry, whether nothing is lost and which checks failed, the verdict of the blind comparison, and whether the pair is accepted with all required |
| `data/revisions.jsonl.gz` | each revision with its judge and its run: the same parts, and how many statements were checked and found kept |
| `data/fitting.jsonl.gz` | each pair of the fitting half: the revisions in two runs, and the harmless copies with each judge and build |
| `data/same.jsonl.gz` | the controls: each revision of a reviser under the judge of its own family, with the same parts |

The record holds no text of any page. A post of the style-guide set has the
number that it has in the earlier records. The copies, the revisions, the
statements, and the judges' quotes stay with the maintainer, and so do the
texts that the scripts in `scripts/` need.

## The controls are no check

In a control a model judges revisions of its own family, which the command
warns against: a model tends to prefer its own text. That matters most for
the blind comparison, so the controls are read only on nothing lost, a result
that rests on quoted words. They show some of it there too: each judge finds
nothing lost more often in the revisions of its own family.

Both were run after the runs that the bars are read on, with the build of the
measurement, and they move no bar. How the second control would be read was
written on the issue before it was run.

## Not measured

People as judges. Texts that no run had read. Whether a revision that loses
nothing is worth publishing. A blind comparison that repeats: this change
takes it out of the gate for those who ask, and leaves it as it was.
