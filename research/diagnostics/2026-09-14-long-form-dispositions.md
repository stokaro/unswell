# Diagnostics on longer technical prose, September 14, 2026

A disposition for every rule that fires on the long-form controlled arm,
with the passage that produced it. The stratum is the September 13 runs:
200 tasks per family, responses averaging 103 and 107 words against 22 in
the earlier runs, same pipeline, same partitions, same prompts.

This studies what the rules do on longer prose. It qualifies nothing and
decides nothing, and it is not a screening: the version 2 list is frozen
and closed.

## What fires, and how often

| Rule | Findings | Disposition |
| --- | --- | --- |
| `readability.grade-metric` | 461 | measures length |
| `syntax.passive-candidate-density` | 340 | measures length |
| `syntax.long-sentence` | 109 | measures length |
| `syntax.parenthetical-load` | 56 | false positive, code in comments |
| `syntax.noun-stack` | 41 | positive |
| `format.em-dash-density` | 5 | positive |
| `readability.long-paragraph` | 5 | positive |
| `repetition.sentence-openers` | 5 | inconclusive, parallel documentation |
| `repetition.ngram-density` | 4 | inconclusive, parallel documentation |

Thirty-one rules produce nothing on 267 long paragraphs, among them every
rule this stratum was built to exercise: section announcements, summary
echoes, repeated transitions and rhetorical templates.

## Positives

`syntax.noun-stack` on a generated Java comment:

> ...clearly demonstrate both possible outcome scenarios using float array
> containment verification checks with specific index-based positioning.

Four nouns in a row carrying one idea. A reader has to unpack
"containment verification checks" before the sentence resolves. The rule
names what it claims to name.

`format.em-dash-density`:

> ...object traversal (A→B→C→A). When starting with any object—A, B, or
> C—it yields the exact same hashCode value.

A parenthetical pair inside a sentence that already carries a
parenthesis. The rule counts density and not presence, and this passage
earns the count.

`readability.long-paragraph` fires on genuinely long paragraphs. Five
findings on 267 units is a low rate, and the rule is doing what it says.

## Measures length rather than habits

`readability.grade-metric` is the loudest rule on this stratum at 461
findings, eight times its rate on short responses. `syntax.long-sentence`
rises fourfold and `syntax.passive-candidate-density` nearly fourfold.

These three scale with the size of the unit. Their rise on longer prose is
arithmetic, not evidence that longer generated prose is worse. A passage
they flag:

> The describedAs method assigns a Description to the current SELF
> instance and returns that same SELF reference, enabling fluent method
> chaining throughout the entire object construction.

That is long and nominal, and a reader is entitled to want it split. It is
also exactly what a human reference manual sounds like: measured per
thousand words, human specification prose carries three times more
`syntax.long-sentence` findings than this generated tree does.

## One false positive, with a cause

`syntax.parenthetical-load` fires 56 times, and the passages are Java:

> As often, an example helps: `<pre><code class='java'>` Employee yoda =
> new Employee(1L, new Name("Yoda"), 800); ...

The flagged parenthetical is a constructor call. A Javadoc comment may
embed example code in `<pre><code>` markup, and the extraction does not
protect those regions, so an example's punctuation reaches the rules as
prose. Issue 279 carries the measurement: 18% of findings in two of the
corpus's Java repositories sit inside such markup.

Every number in this record that touches a Java source carries that, and
so does every H0 prevalence computed over a Java-heavy stratum.

## Two the corpus cannot settle

`repetition.sentence-openers` and `repetition.ngram-density` both fire on
parallel parameter documentation:

> A flag indicating whether dictionary keys should be processed. A flag
> indicating whether explicitly specified property names should be
> processed.

> onBreak (action called when transitioning to Open state), onReset
> (action called when transitioning to Closed state)

The rules are right that the wording repeats. Whether that is a defect is
a different question: parameter descriptions are conventionally parallel,
and rewriting each one differently would cost a reader more than it
returns. Four and five findings is too thin a base to choose, and choosing
would need the human judgments of issue 22.

They stay inconclusive rather than being counted as either positives or
noise.

## What this does not do

It inspects the rules that fire. It cannot inspect the thirty-one that do
not, beyond recording that a stratum built to give them opportunity gave
them opportunity and they produced nothing.
