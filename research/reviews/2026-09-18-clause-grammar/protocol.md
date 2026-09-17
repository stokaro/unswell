# Clause grammar, capability support and attention frames

Base: draft PR #325, commit `2f29e957b2e52d14300e561aae937dcf16327005`.
The previous boundary repair improved exposed examples but reduced new-page
recall from 4/56 to 2/56. All 142 source identities are exposed. Do not treat
those regression tests as a new quality measurement.

## Hypotheses before fresh source review

1. Subject structure must distinguish a gerund plus its particle/object, a
   nominal way/how construction, and an introductory goal before a separate
   main clause. Move goal-boundary disambiguation out of the blanket nominal
   determiner restriction. Retain the timeout-relative and conjunction controls.
2. A nominal capability can wrap the same action as a support verb: supports,
   provides, offers or gives an ability/capability to perform a named action.
   Resolve the generic reader role and bounded nested support chain rather than
   matching ability alone. Preserve modals, conditions, permission constraints,
   actor identity and opaque operands; the diagnostic does not require a duty.
3. An impersonal attention notice has a quality/importance complement and a
   cognitive infinitive before a that-clause. An information subject can also
   select what a generic reader has to read/know. Require explicit presenting
   or cognitive roles; actual verification, prerequisites and permission checks
   remain controls. Avoid duplicate ownership with existing phrase rules.
4. A whole-point/value judgment about reading or knowing an explanation can
   close a discussion without adding behavior. Keep concrete component purposes,
   reader ordering conditions and technical consequences outside this frame.

Use existing source-preserving extraction, POS enrichment, budgets and rules.
No dependency-parse claim, second engine, new model, online inference, blanket
keyword ban, lower gate or weight change. Unsupported semantic cases remain
misses. The objective is useful wording diagnoses, not increased finding count.

## Freeze and confirmation

Before source reading, select six complete pages with seed
`clause-grammar-confirmation-v1`: one short, medium and long per cohort,
distinct historical repositories, excluding the 142 prior identities/hashes
and the original historical frame. Freeze sources and selection. Review all
extracted blocks under all seven categories before runtime edits or diagnostics,
recording exact spans, repairs, uncertainty and acceptable technical controls.
Check whole-block source overlap, retaining and disclosing any partial exposure.

The implementing Codex assistant is the reviewer accepted by ADR 0041, not a
human or independent annotator. Freeze code and focused public tests before
opening confirmation outputs. No semantic tuning afterward in this iteration.
Replay all 18 sets in both profiles, reviewing every changed and fresh finding.
Count only actual diagnoses of frozen defects; keep incidental overlap,
additional post-diagnostic repairs, losses and abstentions separate. Six pages
cannot establish population recall or a within-cell bootstrap interval.

The targets remain 80% recall and 85% soft precision, not measured properties.
Run required ordinary tests, strict lint, policy/schema/catalog checks,
resource probes and CLI/MCP dogfood before publishing a draft. Race, active
fuzzing and coverage remain deferred to #123. No merge, release or deployment.
