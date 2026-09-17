# Rhetorical relations across variable clause forms

Continue #299 from PR #316, commit
`c86beb87c28b6f9e76ff771f5701eedc74a0cffa`. The 94 reviewed source identities
are exposed. Narrow examples have not established good complete-page recall.

## Development control and hypotheses

A temporary development-only ablation disabled both sentence/candidate scope
guards on the 94 exposed pages. It added only two findings: the whole-value-of-
the-verb closure on exposed-local c02 and the remaining document announcement on
exposed-construction c05. Every other finding stayed unchanged. Scope guards
alone therefore do not explain the missed constructions. Restore those guards;
do not ship the ablation or infer that conditions and measurements are irrelevant.

Test these relations with variable nominal, gerund and relative subjects:

- Cognitive-value announcements: information is worth knowing/stating, a decision
  worth stating, or a fact that matters to the explanation. The predicate must
  concern presenting or noticing information. Exclude monetary, computational,
  risk or cost-benefit comparisons and negative statements. Preserve the actual
  information and attached conditions; do not propose deleting the entire clause.
- Purpose clefts and restatement tails: an action is what a named operation is
  for, or a tail describes its premise as the review/question/purpose that a
  feature exists to express or prevent. Require a grammatical relation and an
  anaphoric or gerund action, not the isolated words purpose, feature or reason.
  Concrete component definitions, mechanisms and separately scoped alternatives
  remain controls. Ordinary why/how instructions do not establish redundancy.
- Abstract self-evaluation: a discourse subject, decision or explanatory entry
  is called the point/whole value, what matters most, honest or coherent; a result
  is described as measured rather than merely asserted without naming what was
  measured. Require the evaluative predicate and its argument relation. Retain
  test references, actual measured results and operational distinctions.
- Reader-centered tails: a clause portrays a result as what a generic reader or
  operator wants, understands or misses, without an explicit scoped requirement
  or evidence. Preserve actual permissions, declared requirements, measured
  surveys, conditional preferences and statements about machine actors.

Reuse the existing extraction, POS contract, rule engine and source mapping.
Allow relative and gerund subjects without claiming dependency parsing or semantic
equivalence. Candidate boundaries must follow the written clause, not arbitrary
word overlap. Guards apply to the proposed diagnostic; preserve quotations,
negation, protected operands, qualifiers, numbers and alternatives. Respect the
existing 48-token candidate and 96-token sentence bounds and work budget. No
blanket passive/contrast ban, lowered threshold, raised weight or authorship claim.

## Separate confirmation

Before reading prose, select one page per Ptah/historical and short/medium/long
cell from the existing pinned metadata frame. Exclude all 94 exposed references
and hashes and the original historical study. Rank by SHA-256 of
`rhetoric-relations-confirmation-v1\n` plus reference. Historical allocation is
long, medium, short with one page per repository; Ptah is short, medium, long.
Record shortages rather than selecting replacements after reading a page.

Freeze this protocol, selector, identities, source bytes and notices. Then read
all extracted prose across the seven rubric categories and record exact targets,
context, edits, uncertain events, controls and coverage. Freeze labels before
runtime edits or diagnostic inspection. The implementing Codex assistant is the
maintainer-accepted reviewer under ADR 0041; no human or independent qualification
is claimed. A cell with one page supports counts, not population estimates.

## Evaluation

Replay all exposed and new complete sources with both unchanged profiles. Preserve
full versus partial event credit, all changed/confirmation findings, every miss,
additional post-diagnostic observations, gates, applicability and resources.
The existing #312 abstention remains explicit unless its separate fix is merged.
Never give a length warning credit for an overlapping wording defect. No semantic
tuning after viewing confirmation output; another iteration needs new pages.
Keep the 80% recall / 85% soft-precision working objectives open unless measured.

Require public-API and CLI tests, close controls, mapping, configuration, bounds,
required product checks and CLI/MCP self-check. Preserve deferred race, active
fuzzing and coverage under #123. State which hypotheses remain unsupported.
