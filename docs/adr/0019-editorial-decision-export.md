# ADR 0019: Export editorial decisions from validated rounds

Status: accepted for implementation.

## Decision

Add `Round.Decisions(ctx)` and `annotate decisions` to the existing annotation
module. The output supplies resolved editorial labels and explicit unresolved
states for the later corpus/feature join in #22, #23, #56, and #57. It uses the
published rubric and original response records; it does not classify prose.

Every primary rater declared in a v1 round is assigned every unit. Missing answers
leave the unit unresolved, including when the round already contains an
adjudication based on two other responses. An assistant response cannot fill a
missing primary answer. After all primary answers exist, a recorded adjudication
takes precedence. Otherwise, both the quality label and the set of reason
categories must agree before selecting a label. Category disagreements require
review even when the quality labels agree. No majority vote or category union is
inferred. A selected `uncertain` remains unresolved with that explicit label;
missing answers and disagreements have no selected label.

The export records whether selection used unanimous original answers,
adjudication, or neither. It retains primary/auxiliary response counts and missing
rater IDs. The original round remains the authority for individual answers,
rationales, timestamps, and adjudication. Agreement statistics continue to use
original primary judgments only. Both operations share primary-response selection.

## Identity and scope

The version is `unswell-editorial-decisions-v1`. The output binds the round ID,
purpose, rubric, profile hash, frozen packet hash, and SHA-256 of the exact input
round bytes accepted by `Load`. Reformatting those bytes changes their audit hash;
it does not change the selected decisions. Unit order and rater lists in the
export are canonical. A separate export hash covers compact Go JSON with its own
hash field omitted, following the existing artifact convention.

Each target retains kind, role, source hash/length/format, declared prose language,
original source segments, extraction identity/policy, and hashes of target text
and context. Recorded allowed uses remain explicit. Source text, context text,
editorial rationales, author identities, and origin claims are not copied into
this export. None of the administrative fields becomes an editorial feature.
Hashes bind records; they do not authenticate human participation, permissions,
or source extraction.

Tutorial decisions identify their basis as simulation. Pilot/corpus decisions
identify declared human participation. Every export remains `not_qualified` for
human-corpus acceptance. A selected label is an annotation decision, not an
automatic authorization to train or a qualified model target. Consumers must
still verify source/candidate reproduction, matching units and context, grouped
splits, recorded training/evaluation permissions, annotation audit, and model
feature compatibility. This export does not assign partitions or map a block
feature vector onto sentence labels.

## Implementation and verification

Keep the shipping engine, report schemas, configuration, and gates unchanged.
The research library takes an already loaded round and performs no filesystem,
network, process, or model I/O. Input limits remain the annotation loader's limits;
compact decision output is bounded at 64 MiB. Cancellation or any export failure
returns no partial result. Returned slices and labels belong to the caller.

Test missing answers, auxiliary answers, unanimous labels, category disagreements,
adjudication precedence, final uncertainty, source/packet binding, permissions,
ownership, deterministic ordering, cancellation, and writer failures. A root CLI
golden uses the declared tutorial fixture and preserves its simulation status.
No tutorial output satisfies the human corpus, pilot, held-out, or model-quality
requirements. Those acceptance tasks remain open.
