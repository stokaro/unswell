# Action grammar and support-clause projection

Base: PR #322, commit 0837c7dc357e1e030be3c6e2104fad5e5776c786.
Its fresh full-page result was 4/58 before and after. All 124 previously selected
source identities are exposed. The original objective remains 80% frozen-event
recall with 85% soft diagnostic precision, not a count of implemented patterns.

## Diagnosis and hypotheses

The current accumulated miss ledger contains 542 events without full coverage:
184 wordiness, 162 empty framing, 89 vague claims, 48 repetition, 47 intensifiers,
10 complexity and two transitions. This iteration addresses shared action
structure. It cannot by itself cover the other semantic families.

Existing instruction paths mix a short action-verb vocabulary with whole-clause
guards. They miss coordinated actions and support chains embedded in relative
clauses; the recent capability precision repair also loses real later support
layers. Specify the following hypotheses before reading fresh pages:

1. Reader-intention clauses can use an open grammatical action vocabulary,
   including coordinated infinitives, rather than just configuration verbs.
   Require a reader goal and a concrete following action. A world-state condition,
   another actor's need, permission or uncertain consequence alone is a control.
   Preserve the goal and conditions in any proposed direct instruction.
2. An affirmative support chain inside a relative clause can refer to a named
   operation or concrete nominal antecedent. Resolve that bounded antecedent
   rather than requiring the support verb to be the sentence's first predicate.
   A relative clause that describes a concrete capability without another support
   layer remains a control. Keep protected operands opaque and retain source spans.
3. Copular capable-of gerund actions and imperative be-sure-to actions can be
   recognized as grammatical support layers. Preserve capability, requiredness,
   optionality, prerequisite and negation; do not turn assurance about another
   actor or a factual state into the reader's action.
4. A purpose clause followed by a reader performing a list of steps is an indirect
   procedure announcement. A description of a scheduler's actual steps or an
   explicit measured process is a control. Identify the entire scaffold, not only
   in-order-to. Related method clauses may extend beyond the narrow this-is-done
   form when a shared action/object and reader identity establish the relation.

Use shared NLP and bounded surface roles in the existing instruction rule.
These roles are not dependency edges or semantic entailment. Do not add a second
extractor, a phrase blacklist, a lower gate, an online model or arbitrary scores.
Preserve negative controls from earlier iterations and report lost detections.
Do not infer a defect from text provenance.

## Confirmation freeze

Use seed `action-grammar-confirmation-v1`. Exclude all 124 exposed source
references and hashes and the original historical study. Select one short,
medium and long page per cohort from the pinned frame, with distinct historical
repositories. Freeze protocol, selector, input bytes, identities and notices
before reading prose. Read all extracted prose with source context, then freeze
all seven rubric categories, uncertainties and controls before implementation
or diagnostic output. Do not replace inconvenient pages after selection.

The implementing Codex assistant is the reviewer under ADR 0041, not a human or
independent annotator. Observed counts from six pages are not population recall.
Freeze semantic code and focused tests before opening confirmation output.
No later semantic tuning on those outputs; later work needs fresh inputs.

## Acceptance and reporting

Replay both profiles on all exposed and fresh complete pages. Retain reports,
identities, resource costs and the unchanged #312 abstention. Review every added,
removed or changed finding and all retained fresh warnings. A diagnosis must
identify the frozen defect at its source; partial matches and incidental length
or punctuation overlap cannot receive full-event credit. Post-diagnostic repairs
stay separate. Publish every miss, gains, losses, category/cohort counts and
review burden. A negative result leaves the product objective open.

Run public blackbox API and compiled CLI tests, source mapping, configuration,
resource and cancellation checks, required local checks and CLI/MCP dogfood.
Race, active fuzzing and coverage remain deferred to #123. No merge, release,
model qualification or playground deployment is implied by this experiment.
