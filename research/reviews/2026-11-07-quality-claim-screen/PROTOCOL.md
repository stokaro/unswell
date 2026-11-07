# Source-local quality-claim screen for #349

Status: exposed development hypothesis, November 7, 2026. A Codex assistant
has already read the 36 pages, 123 reference events, current findings, and
several relevant target passages. This protocol cannot make the screen blind
or independent. It fixes one candidate family before the next CLI replay; no
reserved prospective page or model call is authorized by this record.

## Question

Can a bounded source-local rule find unsupported subjective capability claims
while excluding technical uses of the same adjectives? The methodological
source is the [Google developer documentation style guide](https://developers.google.com/style/excessive-claims):
the editorial concern is an unverified performance, cost, security, or
comparative claim, not the mere presence of a word. The
[Vale Google package](https://github.com/vale-cli/Google/blob/master/Google/ExcessiveClaims.yml)
is a useful negative control: its narrow `best/simplest/fastest/guarantee`
rule avoids `never/always/ensure` after those three terms produced 125 of 142
hits in a 950-file control corpus. That rule is not copied into Unswell.

The prior exploratory scan of the 36 exposed pages found 15 occurrences of
the Vale narrow terms with zero overlaps with the 17 currently non-FULL
`vague_claims` events. A broader, source-observed set of quality words had 21
occurrences and at most six overlapping events. These are unreviewed lexical
ceilings, not accepted-finding counts. The source-observed terms are explicitly
exposed development features; they cannot establish out-of-sample accuracy.

## Frozen inputs

Use the original 26 whole pages, six contextual pages, and four historical
confirmation pages, with their unchanged 57/46/20 event denominators and both
profiles. The input archive SHA-256 values are:

| Source | SHA-256 |
| --- | --- |
| `research/reviews/2026-09-17-full-page-recall/inputs.tar.gz` | `af567b371d3331ffdf1633947d89e20467a3a6bf38555a1cb9cc098199dae20c` |
| `research/reviews/2026-09-17-proposition-repetition/confirmation/inputs.tar.gz` | `f9f6590b0824c9762e7721caf2d8284f0c1445f42f5daa1cd8183744bd4d8a50` |
| `research/reviews/2026-09-26-outcome-announcements/confirmation-inputs.json` | `ef6ce4f996a499eced279cc36bb072a6054ee49f8981cbadff1be986102fdda4` |
| `research/reviews/2026-11-07-current-main-editorial-review/findings.json` | `7f1575602b2a7838e99fcd451d31b79654b331df3120e32c1e6d9e5154447f08` |
| `research/reviews/2026-11-07-current-main-editorial-review/events.json` | `c65f47f7b0a4566f5e42f94d80fe033624c9e54fd67c7bce58e58dc7d8c4b53e` |

The engine base is merged commit `d32e754cdfd876c71dae68025337ce8f35850eda`.
Prior source decisions are descriptive only. Any changed finding gets a new
source-context review; an overlapping byte span alone earns no event credit.

## Fixed candidate construction

One rule version may recognize only these bounded constructions in eligible
prose, with original UTF-8 byte spans:

1. A product or API is described with an unscoped pair of subjective quality
   predicates, including a future-proofing promise. A nearby concrete
   capability or criterion may narrow the claim; quoted claims and protected
   identifiers are excluded.
2. A software feature set is ranked with an unbounded superlative such as
   `most comprehensive` without a named comparison set or measured criterion.
3. An attributive pair of subjective qualities modifies a named software
   capability, such as a query language, without naming the operation or
   criterion that makes either quality checkable.
4. Defaults are described as `reasonable` without the actual default values
   or the criterion for the judgment. A sentence that supplies those values
   is a close legitimate control.
5. A generic `important effect/impact on quality` claim supplies neither the
   magnitude nor the affected behavior. Concrete consequence clauses and
   `important fixes` are close legitimate controls.

Do not add `safe` to this first candidate. It refers to state and operational
preconditions; a word-level match could wrongly flag a precise safety
condition. Do not expand the vocabulary after viewing replay outcomes. The
rule may request criteria or verification, but it must not propose deleting a
technical claim or rewriting a number, version, identifier, negation, or
condition.

## Decision and reporting

Run the complete CLI on all 36 unchanged sources in technical and strict
profiles. Bind the binary, source hashes, reports, and exact rule version.
Review **every** added, removed, and changed finding in full source context,
including any new warning outside frozen targets. Compare with the current
main baseline and the newer scoped-degree stream separately. Report FULL,
partial, missed, accepted, uncertain, nonactionable, and suggestion safety by
cohort; preserve the 80% FULL and 85% accepted-delivery gates exactly.

Reject the candidate if it removes an existing FULL witness, introduces an
unsafe edit suggestion, or if fewer than four newly FULL events are supported
with at least 85% accepted new findings on these exposed pages. Passing this
development screen would justify only an untouched prospective confirmation,
not a #349 completion or an AI-authorship verdict. The 18 previously frozen
constructed controls and additional close technical controls must remain in
the screen; no result can be called general precision from those small sets.
