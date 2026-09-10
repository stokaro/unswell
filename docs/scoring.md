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

A source permission does not adjust a probability. Baseline debt acceptance
remains a separate roadmap item.

## Revision probability

The index is not a probability. Without a model, `slop_probability` is `null`
with `probability_status: calibration_unavailable`, and nothing else changes.

A run estimates revision probability only from an explicitly supplied pack:

```yaml
calibration:
  model: pack
  accept_experimental: false
  on_incompatible: unavailable
```

```sh
unswell check --model ./editorial-pack.json docs
```

Both halves are required: a pack without `model: pack` is refused, and analysis
under `model: pack` without a pack fails instead of quietly abstaining.
`unswell doctor` still reports such a configuration so it can be diagnosed. A pack that declares
experimental status is refused unless `accept_experimental` is true, which keeps
an unqualified research model out of a default run.

Each pack qualifies one unit kind, declares the columns it needs, and declares
the minimum words below which it abstains. A unit therefore reports one status:

| `probability_status` | Meaning |
| --- | --- |
| `available` | `slop_probability` holds the calibrated estimate for that unit |
| `unsupported_unit` | The pack does not qualify this unit kind, or no target was prepared |
| `insufficient_evidence` | The unit has fewer words than the pack's declared minimum |
| `missing_feature` | A required measurement is unavailable; `probability_detail` names it |
| `calibration_range` | The score falls outside the fitted calibration range |
| `incompatible_model` | This source's extraction, quoting, or provider differs from the pack |
| `calibration_unavailable` | No pack is configured |

A missing measurement is never treated as zero, and an abstention never carries a
value. `on_incompatible: fail` turns an incompatible source into an operational
failure, which cannot pass a gate. Estimates do not change findings, the index,
or the gate decision; probability gating is not implemented yet.

`unswell-mcp --model` accepts the same pack, so the server and the CLI agree.
The text, Markdown, and HTML reports name the configured pack and how many units
of its kind were estimated; JSON and SARIF carry the per-unit statuses.

### Gating on a probability

A probability can fail a run only when the pack declares acceptance:

```yaml
gate:
  probability:
    fail_at: 0.85
    require: false
```

`fail_at` is a probability above 0 and at most 1. A unit whose estimate reaches
it fails the gate with a `gate.<kind>-probability` diagnostic. The gate requires
`calibration.model: pack` and refuses `calibration.accept_experimental`, so an
experimental research pack cannot decide a build.

`require: true` also fails on an absent estimate the pack did not predict: an
unavailable column, a score outside the fitted calibration range, or a source
whose measurement inputs differ from the pack. It does not fail on the pack's own
declared limits, so a unit below the declared minimum words, or of another kind,
still passes. Those abstentions are the policy the pack states in advance.

Exit codes keep their meanings. A gate failure exits 1, an operational failure
such as an incompatible pack under `on_incompatible: fail` exits 2, and
`--no-gate` records the reasons without changing the decision.

The manifest records the pack's own declarations, including whether its author
declared it accepted and whether it names a qualified corpus. Unswell checks that
these declarations are consistent and that the pack matches the run; it cannot
verify that a corpus was qualified or that an evaluation exists.
