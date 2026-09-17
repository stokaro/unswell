# Clause boundaries and relative instruction roles

Base: draft PR #324, commit `60d211726d5e9db6c8069d88214e04fadc420667`.
The preceding candidate gained 11 full detections on exposed pages but lost a
fresh configure-options diagnosis and added a false alarm on timeout advice.
Its fresh result fell from 2/44 to 1/44. All 136 source identities are exposed.

## Hypotheses before new source review

1. A copular quality judgment must have a main-clause subject. An imperative
   with a relative restriction (raise a timeout for databases that are slow)
   states a symptom/action relationship, not an unscoped endorsement.
2. Local quality scope can be supplied by the surrounding paragraph, including
   a named cost/tradeoff and its opt-out. Require a linked subject and concrete
   scope; unrelated technical context must not hide a judgment. Preserve
   conditions, negation, measurements and source locations.
3. A relative action antecedent may end in a protected identifier after its
   common-noun head. Named functions and configuration options are operations;
   checksums, tokens, filenames and labels are supplied operands. Recover
   useful instruction diagnostics without returning the caching false positive.
4. Goal and topic prefixes must not conceal the main quality predicate. A goal
   does not establish that a process is easy or simple; an explicit mechanism
   or condition remains a control.

These changes use existing surface clauses and POS tokens, not dependency
parsing or semantic entailment. No gate, weight, threshold or rule count changes.
No model or network inference. Preserve every old annotation and missed event.

## Freeze and evaluation

Select six complete pages with seed `clause-boundaries-confirmation-v1`, one
short, medium and long page per cohort, distinct historical repositories,
excluding all 136 source references/hashes and the original historical frame.
Freeze protocol and sources before reading. Review every extracted block against
all seven categories and record defects, uncertainty and technical controls.
Audit complete-block overlap against earlier sources; retain and disclose any
reuse instead of replacing a page. The implementing Codex assistant is the
reviewer accepted in ADR 0041, not a human or independent annotator.

Freeze labels before runtime edits; freeze semantic code and focused public
tests before confirmation diagnostics. No tuning after opening that output.
Replay all 17 sets in both profiles, reviewing every changed finding and fresh
warning. Report losses, additional post-diagnostic repairs, misses, review burden
and the known #312 abstention separately. Six pages do not establish population
recall or support a within-cell bootstrap. The 80% recall and 85% soft precision
objectives remain targets, not claims.

Run ordinary module tests, strict lint, policy/schema/catalog/resource checks
and CLI/MCP dogfood before publishing a draft. Race, active fuzzing and coverage
remain deferred to #123. No merge, release or deployment is authorized here.
