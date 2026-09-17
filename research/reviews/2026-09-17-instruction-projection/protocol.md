# Project the action from layered capability instructions

Continue #299/#309 from PR #318, commit
`d87199b852126bb3a28fa22aedba09efb77953e2`. All 100 reviewed identities are
exposed. Previous isolated construction additions have not established useful
complete-page recall. Default gates, weights and thresholds remain unchanged.

## Hypothesis and preservation contract

Parse a bounded instruction into its grammatical actor, capability/possibility
layer, action and operands, with optional method and conditions. Reuse the source
NLP tokens; do not claim dependency parsing or general semantic equivalence.
Recognize the composition of support predicates rather than complete literal
sentences. Candidate forms include:

- A tool or API allows/enables a generic reader to perform a transitive action.
  Reader intent and the support layer can be replaced by a direct instruction
  naming the same tool, action and operands. Named actor permissions, access
  grants, uncertain outcomes and concrete capability limits remain controls.
- A tool has the ability/capability to perform an action. Preserve the ability
  qualifier while removing the nominal ability wrapper. A bare can + action
  remains a control. Do not infer that the operation actually occurs.
- An object can/may be used to perform a transitive action, including an optional
  information/configuration prefix. Preserve the object as the method or actor,
  the modal meaning and any condition. Descriptions of hazards and authorization
  must retain their warning or permission meaning and are not instruction edits.
- A nominal action is carried out/performed/accomplished using a stated method,
  or a chain repeats more than one capability layer around the same action.
  An object that merely contains an action word is not a nominal action.
- A capability or action announcement followed by an anaphoric method clause may
  span adjacent prose blocks. Relate the actual source ranges only when local
  structure gives an unambiguous antecedent; do not cross headings, lists,
  excluded code or unrelated actors. Keep partial detections distinct.

No finding follows from can, passive voice, a long sentence, a technical noun or
an allow/enable verb alone. Require the support relation, projected action and
its operand. Keep code operands opaque. Infinitive/imperative syntax can establish
the verb role where the POS model mislabels an uncommon action, but an arbitrary
word ending is not proof of an action. Negative, quoted, reported, conditional
permission and failure statements are counterexamples. Ordinary conditions on
an instruction remain visible, not silently deleted. Prefer abstention over
turning an actor's authorization into a reader command.

Use existing rule IDs, source maps, windows, budgets, configuration, reporters and
MCP. Bounds are at most 48 tokens per candidate and 96 per sentence. No new runtime
process, model call, download or Python dependency. The suggested review names
what to preserve; this does not implement automatic rewriting.

## Frozen confirmation

Select one whole page in each Ptah/historical and short/medium/long cell from the
existing pinned metadata. Exclude all 100 reviewed references and hashes and the
original historical study. Rank by SHA-256 of
`instruction-projection-confirmation-v1\n` plus reference. Historical selection
uses long/medium/short order and one repository per page; Ptah uses
short/medium/long. Record a shortage rather than refill after inspection.

Freeze selection, source bytes, license notices and this protocol before reading
new prose. Review every extracted block across all seven editorial categories,
with exact source targets, context, proposed repairs, controls and uncertainties.
Freeze labels before implementation or diagnostic output. The implementing
assistant is accepted under ADR 0041; it is not a human or independent reviewer.
One page per cell supports counts, not population precision or recall.

Replay all exposed and new pages with both unchanged profiles and pinned before
and after binaries. Review every changed finding and every confirmation finding.
A length warning does not earn credit for overlapping a wording issue. Preserve
full, partial, disputed and additional post-diagnostic judgments separately.
Existing #312 abstention remains explicit unless its separate fix is integrated.
After confirmation output is opened, do not tune semantics on these pages.

## Acceptance

Report frozen event recall, all finding dispositions, applicability, operational
errors, gates and resource costs. Retain all misses and unsupported hypotheses.
The working 80% recall and 85% soft-precision objectives remain unmet until actual
evidence supports them. A few exposed fixture gains do not satisfy this goal.

Add blackbox public API and real CLI tests with close controls, mapping,
configuration, bounds and cancellation as applicable. Run required local checks,
self-check through CLI/MCP, and evidence-validator negative tests. Race, active
fuzzing and coverage stay deferred under #123. Record the result without forcing
AI-assisted pages to fail or treating their origin as an editorial label.
