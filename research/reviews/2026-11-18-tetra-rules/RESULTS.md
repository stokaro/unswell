# Results

Written by `verify.py` from `data/sentences.jsonl.gz`. Do not edit by hand.

56 papers, 168 files of editors, 3,876 sentences; an edit touches 1,572 of 3,876 (40.6%).

## The bar

With the default rules the difference is -0.3 points; the bar asks for at least +5.0. The bar does not hold.

## Edited and unedited sentences

| | Edited, flagged | Share | Unedited, flagged | Share | Difference, points |
| --- | ---: | ---: | ---: | ---: | ---: |
| Default rules, every edit | 10 of 1,572 | 0.6% | 22 of 2,304 | 1.0% | -0.3 |
| Default rules, redundancy edits | 1 of 215 | 0.5% | 22 of 2,304 | 1.0% | -0.5 |
| Every rule, every edit | 585 of 1,572 | 37.2% | 618 of 2,304 | 26.8% | +10.4 |
| Every rule, redundancy edits | 77 of 215 | 35.8% | 618 of 2,304 | 26.8% | +9.0 |

Intervals of the difference at 95%, from resampled papers: default rules -0.9 to +0.3; every rule +6.8 to +14.1.

## Every rule, at equal length

The median edited sentence has 27 words, and the median unedited sentence has 19.

| | Edited, flagged | Share | Unedited, flagged | Share | Difference, points |
| --- | ---: | ---: | ---: | ---: | ---: |
| 1 to 10 words | 12 of 60 | 20.0% | 44 of 374 | 11.8% | +8.2 |
| 11 to 20 words | 105 of 418 | 25.1% | 206 of 869 | 23.7% | +1.4 |
| 21 to 30 words | 142 of 424 | 33.5% | 166 of 576 | 28.8% | +4.7 |
| 31 to 40 words | 93 of 259 | 35.9% | 68 of 225 | 30.2% | +5.7 |
| 41 and more words | 233 of 411 | 56.7% | 134 of 260 | 51.5% | +5.1 |

Weighted by the edited sentences of each band, the difference is +4.2 points.

## Every rule, rule by rule

20 rules fire; the table lists those that touch at least 20 sentences.

| | Edited, flagged | Share | Unedited, flagged | Share | Difference, points |
| --- | ---: | ---: | ---: | ---: | ---: |
| `readability.grade-metric` | 152 of 1,572 | 9.7% | 173 of 2,304 | 7.5% | +2.2 |
| `syntax.passive-candidate-density` | 164 of 1,572 | 10.4% | 154 of 2,304 | 6.7% | +3.8 |
| `gate.paragraph-score` | 121 of 1,572 | 7.7% | 159 of 2,304 | 6.9% | +0.8 |
| `syntax.long-sentence` | 152 of 1,572 | 9.7% | 82 of 2,304 | 3.6% | +6.1 |
| `readability.long-paragraph` | 102 of 1,572 | 6.5% | 115 of 2,304 | 5.0% | +1.5 |
| `syntax.parenthetical-load` | 78 of 1,572 | 5.0% | 83 of 2,304 | 3.6% | +1.4 |
| `filler.wordy-phrase` | 28 of 1,572 | 1.8% | 38 of 2,304 | 1.6% | +0.1 |
| `syntax.noun-stack` | 21 of 1,572 | 1.3% | 14 of 2,304 | 0.6% | +0.7 |
| `hype.vague-praise` | 15 of 1,572 | 1.0% | 15 of 2,304 | 0.7% | +0.3 |

## Two editors of one paper

2,896 pairs of sentences that both files hold: both editors edit 514, one edits 962, neither edits 1,420. Agreement 66.8%, kappa 0.26.
