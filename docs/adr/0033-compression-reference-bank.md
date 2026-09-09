# ADR 0033: Reserve compression references before classifier fitting

Status: accepted for explicit reference-bank assembly; model integration and qualification remain open.

## Decision

Build #52's reference bank through the existing corpus verification and prepared
targets. The caller supplies a frozen candidate artifact, exact source and notice
bytes, an independently curated annotation round, and explicit ordered cohort
selections. The library neither downloads a corpus nor infers origin from prose.

Export origin claims from the validated round separately from editorial decisions.
Retain target bindings, round identity, and simulation status. Reference selection
uses no editorial judgments. Require unit-scoped `human` or `generated` claims;
unknown, mixed-origin, and edited-origin targets are not confirmed endpoints.
These are checked declarations, not independent proof of their truth.

Every selected target must belong to the training partition, have the requested
kind, and declare training permission. Reproduce its prose and all source ranges
before assembly. Preserve the explicit unit order and insert LF between targets.
The final LF added by the shared compression primitive counts toward its 32 KiB
prefix limit. Never truncate a target or silently select fewer references.

A bank can contain human, generated, and mixed reference cohorts. A mixed cohort
contains both endpoint classes; it is not a claim that its examples were matched
by topic, length, or source. Such matching belongs to the frozen experiment
protocol and still requires verification. All cohorts use the same target kind
and compressor settings.

Reserve the union of connected source groups across every cohort in the bank.
Retain all affected source and target IDs, including related targets that were
not themselves selected. Future classifier fitting must exclude this common union
for every compared cohort and baseline. A bank must remain fixed for fitting,
calibration, and prediction. Different per-cohort exclusions would change the
training population and invalidate the intended paired comparison.

The bank artifact contains reference prose and provenance intentionally. It is a
developer artifact, not a normal scan report or blinded annotation packet.
Training permission does not imply permission to redistribute it. Its digest
identifies bytes and declarations; it does not establish rights or authenticity.

## Acceptance and follow-up

Use blackbox tests for source/notice reproduction, exact order and separators,
cohort membership, missing claims, partition and rights rejection, transitive
group reservation, size limits, cancellation, and detached results. Tutorial
fixtures remain explicitly simulated and cannot qualify a human corpus.

Model fitting and prediction must consume the bank and its reservation together,
check its identities, retain unavailable measurements, and reuse the shared
compression formula. Do not add these features to the default product gate.
Acquiring real cohorts, matching them, measuring predictive value, and qualifying
a model remain requirements of #52 and #57.
