# Review on real text: Ptah, September 11, 2026

This directory records the first review of the kind that
[ADR 0037](../../../docs/adr/0037-diagnostics-not-authorship.md) asks
for. Every catalog rule ran on a real repository. A reader judged a
sample of the findings under each rule's own statement. The subject is
[stokaro/ptah](https://github.com/stokaro/ptah) at commit `7b47e7cfb` of
September 8, 2026, MIT licensed, the development case study of the
research plan. Its maintainer states heavy AI use in its history. That
statement labels no sentence, and nothing here is an authorship claim.

One reader judged the sample: the maintainer's agent session, on the day
of the run. The maintainer has not confirmed the judgments. A judgment
says whether the construction the rule names is present in prose a reader
could act on. It is not a quality verdict on Ptah.

## What ran

The tool was the merged commit `d3c9f23` of this repository. Three runs
scanned the 4,110 tracked files that carry prose, 2.56 million prose
words, with `--no-gate --include-source --timeout 20m --jobs 4`. The
configuration file has to sit inside the scanned project root, so each
run copied its file into the checkout and removed it afterwards.

| Run | Configuration | Status | Findings |
| --- | --- | --- | --- |
| Default | the `technical` profile, 16 rules | incomplete: one test file holds a non-Latin block | 11,067 |
| All rules | [config-all.yaml](config-all.yaml): every catalog rule, a larger candidate budget, that file excluded | complete | 24,227 |
| Comments only | [config-comments.yaml](config-comments.yaml): the same without string literals | complete | 22,160 |

Two settings were needed to complete a run. `analysis.max_candidates`
rose from 100,000 to 2,000,000, because one file of 16,355 prose words
exhausts the default budget of three rules and the exhaustion fails the
whole document by design. One test file with Russian test strings had to
leave the scan through `files.exclude`, because a non-Latin block is a
document error and makes the run incomplete. The default timeout of 30
seconds also stops a repository of this size; the runs took 22 to 35
seconds with four workers.

[summary.json](summary.json) holds the counts of every run by rule and by
file kind: Go, Go tests, documentation, other.

## Findings by rule

Counts of the all-rules run, with the comments-only count beside it. The
difference is what string literals contribute: SQL, error texts, help
texts, regular expressions, and test names.

| Rule | All contexts | Comments and Markdown | Go tests, all contexts |
| --- | --- | --- | --- |
| `syntax.long-sentence` | 10,279 | 9,810 | 3,852 |
| `readability.grade-metric` | 8,331 | 8,204 | 2,934 |
| `syntax.passive-candidate-density` | 2,163 | 2,072 | 647 |
| `syntax.parenthetical-load` | 1,042 | 651 | 409 |
| `format.em-dash-density` | 613 | 609 | 174 |
| `repetition.exact-sentence` | 564 | 225 | 318 |
| `repetition.ngram-density` | 287 | 154 | |
| `syntax.noun-stack` | 269 | 88 | 123 |
| `repetition.paragraph-overlap` | 231 | 68 | |
| `readability.long-paragraph` | 216 | 153 | |
| `repetition.near-sentence` | 148 | 46 | 80 |
| `repetition.sentence-openers` | 39 | 39 | 0 |
| `filler.wordy-phrase` | 27 | 26 | |
| `repetition.paragraph-openers` | 8 | 8 | 0 |
| `repetition.syntax-template` | 4 | 2 | |
| `filler.weak-intensifiers` | 3 | 3 | |

The other 23 rules fired nowhere in 2.56 million words. Among them are
every scaffold rule, every hype rule, the section and transition fillers,
the rhetorical patterns, and the heading and summary echoes.

## The judged sample

[sample.json](sample.json) holds a seeded sample of ten findings per rule
from the all-rules run, or every finding where a rule has fewer, with the
judgment and its reason for each. The table condenses it.

| Rule | Justified | Not | Why not |
| --- | --- | --- | --- |
| `syntax.long-sentence` | 8 | 2 | SQL in a string literal; a bulleted list inside a comment read as one sentence |
| `readability.grade-metric` | 3 | 7 | identifiers, keywords, and references drive the formula, not the prose |
| `syntax.passive-candidate-density` | 10 | 0 | the form is present; changing the voice is a separate call in reference prose |
| `syntax.parenthetical-load` | 2 | 8 | code, SQL, citation keys, and enumeration markers in parentheses |
| `format.em-dash-density` | 10 | 0 | |
| `repetition.exact-sentence` | 2 | 8 | repeated string literals: SQL, help texts, regular expressions, expected errors |
| `syntax.noun-stack` | 0 | 10 | test names and established technical terms |
| `readability.long-paragraph` | 8 | 2 | a shell comment in a workflow; a one-line help text |
| `repetition.near-sentence` | 5 | 5 | near-duplicate string literals |
| `repetition.sentence-openers` | 10 | 0 | |
| `filler.wordy-phrase` | 5 | 5 | the phrase mentioned by a style guide and a style-check script, not used |
| `repetition.paragraph-openers` | 8 | 0 | |
| `repetition.syntax-template` | 1 | 3 | error strings and test names of one shape |
| `filler.weak-intensifiers` | 3 | 0 | |

## What the review says about Ptah

Three traits of the prose stand out, all in comments and documentation
rather than in code strings.

- Sentences over 35 words: about 9,800 in 1.9 million words of comments
  and documentation, five per 1,000 words. The comments explain a decision
  and its alternatives in one sentence with several subordinate clauses.
- Em dashes: 609 paragraphs at or above two per 100 words, 231 of them in
  the documentation.
- Repeated openings and near-duplicate doc comments: the reference pages
  open many sentences and paragraphs with the same words, and similar
  declarations carry near-identical comments.

Also worth an edit: 26 uses of a wordy phrase in comments and scripts, and
a few intensifiers. None of this is a share of generated text.

## What the review says about the diagnostics

- String literals are the main source of unjustified findings for the
  repetition rules, the parenthetical rule, and the noun-stack rule. A
  review of prose should exclude the `string` context, as the second
  configuration does, or scope it to string owners that hold prose.
- The readability formula is not a useful signal on code comments:
  identifiers and keywords inflate it. It stays off by default, and this
  run gives no reason to change that.
- The noun-stack rule fired ten times in the sample on test names and
  established terms and never on a stack a reader would unwind. Its
  message already tells the reader to keep technical terms; on this
  corpus that is every finding.
- Mention is not use: a style guide and a style-check script name the
  wordy phrase they forbid. A reasoned suppression covers such files.
- Three operational limits shaped the run: the candidate budget, the
  non-Latin block, and the 30-second default timeout. Each is documented
  behavior. Whether a budget exhaustion or a non-Latin block should fail a
  document or abstain on it is a product decision this record only raises.

## Limits

- One reader, one day, ten findings per rule. The counts are exact; the
  judgments are one person's reading of short snippets, without the full
  paragraph in most cases.
- Ptah is one repository with one dominant author and workflow. Nothing
  here transfers to other projects by itself.
- The two derived gate diagnostics in the sample are index outcomes, not
  constructions, and count as justified only in that sense.
