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
added again to paragraph evidence. Document statistics report the raw paragraph
maximum, nearest-rank median and P90, and eligible flagged fraction. Appending
clean paragraphs cannot lower the index or gate decision for an existing one.

Severity controls presentation. `gate: forbid` is a separate hard-policy setting.
Local score gates trigger at **greater than or equal to** the configured threshold
when the unit meets `min_words`. They create derived diagnostics that are never
fed back into scoring. Display limits do not remove evidence from the gate.

Source [suppressions](suppressions.md) retain raw findings and `contributions`.
The engine recomputes `effective_slop_score` using only active findings through
the same correlation and cap rules. `effective_contributions` explains affected
units, including permitted findings with zero points and `source-suppression`.
Gates use effective values; document summary statistics remain raw. Derived gate
diagnostics are generated after suppressions and cannot themselves be suppressed.

Probability is `null` with `calibration_unavailable`; a source permission does not
adjust a probability. Baseline debt acceptance remains a separate roadmap item.
