# Instruction clauses, prerequisites and concrete capabilities

Start from PR #321, commit 50f25c744b36be044f816ce267cbfaecc6a83748.
The 118 exposed source identities remain development material. The broad product
objectives remain 80% frozen-event recall and 85% soft diagnostic precision;
a few detected constructions or familiar pages do not satisfy them.

## Hypotheses before confirmation

Extend the existing instruction rule and clause representation, not the gate.
The exposed misses include narrated tutorial setup, nested action descriptions
and nominal methods. Prior confirmation also shows false positives in concrete
capability explanations. Treat both recall and precision as required:

1. A narrated prerequisite followed by a first-person need/action can obscure a
   direct instruction. Identify the prerequisite separately from the action and
   any support layers (need to, make sure that, can be used for). Preserve ordering,
   actor identity, requiredness, optionality, negation and exceptions. Do not flag
   first-person statements, can, need, or temporal clauses alone. Actual operational
   conditions are not removable framing; show them as context, not useless words.
2. A named operation described through a used-to/used-for wrapper may restate its
   action indirectly. Require the operation role or an additional support layer;
   a component's ordinary use or intended purpose alone is a negative control.
   Preserve intent versus achieved behavior and useful method/agent distinctions.
3. Project actions through infinitive and gerund complements using their syntax
   and shared NLP, with bounded work and original spans. A named failure, condition
   or protected operand elsewhere must not automatically erase an independently
   established wrapper. Scope guards to the candidate and retain dependencies;
   do not infer a full dependency parse from POS tags.
4. A standalone allows-user statement can explain a concrete capability. Require
   a redundant support layer or an explicit adjacent method before calling that
   construction instructional scaffolding. Permission, access control, a measured
   feature boundary, a simultaneous benefit and genuine uncertainty are controls.

These hypotheses come from the exposed #309 examples and the 0/46 proposition-
repetition confirmation. Do not reduce the task to one listed phrase or add words
from new confirmation pages to the matcher after output. Unsupported semantic
cases stay explicit. No blanket word blacklist, lower gate or online model.

## Fresh source review

Use seed `instruction-clauses-confirmation-v1` to rank the pinned source frame.
Exclude all 118 exposed references and hashes and the original historical study.
Select one short, medium and long page from each cohort, with distinct historical
repositories. Freeze this protocol, selector, source identities, bytes and notices
before reading prose. Then read all extracted prose in native formats, with source
context, and freeze defects, uncertainties and controls for all seven editorial
categories before runtime changes or diagnostic output.

The implementing Codex assistant is the reviewer accepted under ADR 0041. This is
assistant development review, not human labeling or independent qualification.
Unknown historical provenance remains unknown. One page per cohort/length cell
supports observed counts, not population recall or a bootstrap interval.

## Measurement and acceptance

Freeze semantic changes and focused tests before opening new confirmation output.
Replay both profiles on all exposed and fresh complete pages. Retain source/policy
identities, actual spans, artifacts, resource costs and abstentions. Review every
added, removed or changed finding and all retained confirmation warnings. Partial
matches and incidental overlap with sentence length never earn full event credit.
Additional post-diagnostic defects cannot increase frozen recall.

Report all misses, per-cohort/category counts and review burden, including lost
useful findings if precision changes suppress them. Preserve the existing #312
abstention. Later semantic tuning requires a new source freeze. Do not promote
experimental defaults or claim completion from an exposed-only improvement.

Verify public blackbox API and CLI behavior, source mapping across comments and
strings, policy, cancellation and resource failures. Run required checks and
CLI/MCP self-checks. Race, active fuzzing and coverage remain deferred to #123.
No merge, release or playground deployment is authorized by this protocol.
