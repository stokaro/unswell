# ADR 0037: Concrete style diagnostics, not authorship

Status: accepted on September 11, 2026, after the maintainer restated the
product goal and an audit of the code, the documentation, the open issues,
and the recorded experiments against it.

## Context

Unswell exists to find formulaic constructions in technical prose and to
explain what to rewrite. The typical user already knows that a text came
from an AI agent. They want the stock phrases, rhetorical templates,
repetitions, and other traits of the wording that a reader would notice,
each pointed at with a reason. The author of the text is not the question.

The audit found the shipped linter aligned with that goal. Its 40 rules
name a construction, a scope, and an edit. The index is an editorial
policy index, and no report wording implies authorship. The origin channel
stays off by default and cannot decide a gate. Every catalog family carries
counter-examples of ordinary technical prose.

The research side had drifted. The E3 baseline runs of #176, #177, and
#178 fit classifiers of cohort membership, historical against contemporary
snapshot. The origin runs of #179 fit a classifier of generation
endpoints. Their results are negative: nothing separates the cohorts. As
authorship evidence they say little. The question was never whether a
paragraph reveals its snapshot date. As evidence about rules they say
nothing at all. A rule that fails to separate two cohorts can still name a
construction worth rewriting. A feature that separates them can mark a
period rather than a defect.

The historical corpus does one of the two jobs it exists for. It measures
the existing rules by cohort and period, with placebo boundaries and
paired changes. It does not yet propose anything. Six papers and the
hand-written catalog fixed every studied construction in advance. No
command reads the corpus and returns candidate constructions with their
frequencies by cohort and role. The prevalence tables admit roles as a
filter and carry no role stratum. The corpus itself holds GitHub
repositories only.

## Decision

The product goal is explainable, contextual diagnostics of formulaic
wording. The following statements bind every document and plan.

- A clean scan of an AI-written text is a correct result. A finding on a
  human-written text is not a false positive by that fact alone. An error
  is a diagnostic the stated rule does not justify.
- A diagnostic names the construction, its span, and the edit. It never
  attributes authorship, and no index, score, or estimate is presented as a
  probability that a tool wrote the text.
- Rarity in the historical corpus is not a defect. The corpus is the
  comparison base, not a norm; new terms and ordinary modern phrasing are
  not penalized for being new.
- Two observations stay apart: "unusual against the historical corpus" and
  "frequent in contemporary generations". Neither proves the other, and a
  useful pattern need not be exclusive to generated text.
- Rules need positive examples and counter-examples with similar words in
  ordinary technical prose. Density and repetition rules need checks over
  the window they judge, not over one sentence. A finding must go away
  after a real rewrite, and a score drop after a meaningless rewrite is not
  a success.
- Generation for research uses the agents of the tool already in use. No
  paid model API enters the research path, and no ordinary scan calls a
  model.

The evidence base keeps its meaning and its artifacts. Prevalence,
paired change, and warning load by cohort stay the measures of the pattern
protocol. The E3 and origin records stay as recorded negative results of
classification questions; they qualify no rule and disqualify none.

The following work leaves the mandatory path to a release and stays on
hold with its records intact.

- The origin channel experiments, #58 and #179.
- Calibrated revision probability and its held-out evaluation, #25.
- The human-labeled corpus and the rule qualification that needs it, #22
  and #26.
- The detector comparisons, #50, #51, and #52.

Each may return when a concrete diagnostic needs it and the need is
written down.

## Next stage

The next stage improves named diagnostics on real texts. It adds no
platform. Its items, in order:

1. Regression fixtures per rule: a matching example, a near miss from
   ordinary technical prose, and a rewritten version that must not fire.
   The protocol calls them E5; none exist today. They run in ordinary CI
   on frozen fixtures.
2. A role stratum in the prevalence tables, so a comment, a README, and a
   reference page are not compared against one pooled rate.
3. A frequency command over the existing candidates: word n-grams, sentence
   openers, and part-of-speech templates counted per cohort and role, with
   rates per 1,000 words and the ratio between cohorts. It proposes
   candidates; it decides nothing.
4. A review of the diagnostics on the texts this project already has: the
   800 controlled responses and the repository's own documentation, scanned
   with every catalog rule enabled. The record lists, per rule, the
   findings a reader judged justified under the rule's statement and the
   ones not, with the examples.
5. Source types beyond repositories for the historical corpus, starting
   with documentation sets and specifications whose historical version and
   license are verifiable. A registration date proves nothing about the
   text.

The stage is accepted when five things exist. The fixtures pass in CI.
The tables carry the stratum. The frequency command runs on the committed
candidates, with its output recorded. The review record exists with
examples. One source outside a repository enters the corpus under its
contract. No precision claim, no benchmark, and no probability is part of
it.

## Consequences

The [research plan](../research.md) and the [roadmap](../roadmap.md)
name the goal, the non-goals, and this stage. The README states the
non-goals. Existing tooling, records, and on-hold issues stay; their
labels change, not their content. A later product change that cites an
evidence card still needs its own decision record.
