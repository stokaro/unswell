# Scores and gate decisions

The alpha index is an editorial-policy index from 0 to 100, not a percentage.
An exact phrase match establishes a configured rule condition; it does not prove
that an editor would reject the text. Every builtin rule is experimental.

Each rule emits evidence, original source locations and an activation in integer
thousandths. Density and length rules ramp linearly from onset to saturation;
exact phrase matches have full activation. The engine multiplies activation by
weight and retains integer thousandths until report serialization.

For each sentence and paragraph, the engine sorts contributions by descending
raw value with stable identifiers as tie breakers. Same-group identical or nested
evidence contributes only the strongest signal. Rule caps and correlated-group
caps limit the remaining contributions. The total is capped at 100. The trace
records raw and effective points and why a contribution was reduced.

Sentence and paragraph assessments are independent; sentence totals are not
added again to paragraph evidence. Document statistics report the paragraph
maximum, nearest-rank median and P90, and eligible flagged fraction. Appending
clean paragraphs cannot lower the index or gate decision for an existing one.

Severity controls presentation. `gate: forbid` is a separate hard-policy setting.
Local score gates trigger at **greater than or equal to** the configured threshold
when the unit meets `min_words`. They create derived diagnostics that are never
fed back into scoring. Display limits do not remove evidence from the gate.

Probability is `null` with `calibration_unavailable`. Suppressions and baseline
acceptance are not implemented in this alpha, so raw and effective scores agree
after normal evidence deduplication and caps. Stage 2 must preserve raw evidence
when those mechanisms are added.
