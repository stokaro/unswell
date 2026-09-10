# ADR 0034: Load explicit probability packs with declared applicability

Status: accepted for the pack contract and per-unit applicability; engine
integration, calibrated gating, and qualification remain open.

## Decision

Add #24's revision-probability contract as the public `probability` package. A
caller supplies one pack as bytes; the package performs no I/O, fits nothing, and
never discovers a model. A research training artifact is not a pack: it carries
corpus rows, partitions, and exclusions that must not ship, and it declares no
applicability limits. Producing a pack from an experiment stays a separate,
explicit step.

A pack declares the run inputs that change measured column values: the feature
and unit contracts, complete column descriptors with their digest, the NLP
provider identity, the exact requested capabilities, the preparation hash, and
the quote and structure switches. `Compatible` compares those inputs only. Gate
thresholds, severities, baselines, and report selection do not change prepared
measurements, so they do not invalidate a pack. Columns must equal this build's
`feature.UnitCatalog` definitions for the pack's kind, so a renamed, reversioned,
or reformulated measurement is rejected instead of silently reused. Rule
activations and dependency-backed columns are outside this contract until their
ruleset and backend identities are bound as well.

The first supported estimator is one regularized logistic model with separate
isotonic calibration over its linear score, matching the specification's primary
shipping baseline. Other estimators are rejected rather than approximated. A
calibration mapping needs at least two knots; a single knot maps one score, not
an interval.

`declared_status`, `human_corpus`, and `evaluation` are author statements. The
package checks them for internal consistency: an accepted pack must name a
qualified corpus and a published evaluation. It cannot verify that any of this is
true, and loading a pack is not acceptance evidence for #22, #25, or #26.

Applicability is decided per unit, before any number is reported:

| Status | Condition |
| --- | --- |
| `available` | The unit matches the pack's kind, meets its declared minimum words, has every column measured, and its score lies inside the calibrated range |
| `unsupported_unit` | The unit kind differs from the pack's single qualified kind |
| `insufficient_evidence` | The unit has fewer words than the pack's declared minimum |
| `missing_feature` | A required column has no value; the detail names the column and the measurement's own reason |
| `calibration_range` | The linear score lies outside the fitted knots |
| `incompatible_model` | The source's effective measurement inputs differ from the pack |
| `calibration_unavailable` | No pack applies to the run; this is the existing model-free status |

The minimum word count comes from the pack, not from a constant in this
repository: a universal threshold would be an unverified claim. A missing
measurement is never imputed as zero, and an abstention never carries a value.
Every abstention keeps its machine-readable status; a document estimate is never
painted onto its sentences, because a pack estimates exactly one unit kind.

A vector that does not match the pack's columns is a caller error, not an
abstention: it means the run measured something other than what the pack
requires. Corrupt bytes, an altered digest, an unsupported contract, and an
inconsistent declaration are errors for the same reason. Errors and cancellation
return no partial estimate.

## Engine and command integration

`Options.Model` carries one pack as bytes, and `check --model` reads the named
file: the library still opens nothing. A pack requires `calibration.model: pack`,
and analysis under that policy requires a pack, so neither is silently ignored.
Construction still succeeds without the pack, so configuration inspection can
diagnose the policy; the run itself fails instead of abstaining silently. A pack that
declares experimental status needs `calibration.accept_experimental`, which keeps
an unqualified research model out of a default run. The policy records
`calibration` only when a model is requested, so model-free runs keep their
existing identities and accepted debt. Accepted debt records the pack identity,
because a different model changes the estimates behind a scored unit.

Preparation for the probability channel runs separately from optional feature
collection, with exactly the pack's declared capabilities and its single kind,
so measured columns match the pack contract instead of whatever a run happens to
collect. Compatibility is decided per source, because per-file overrides can
change extraction and quote handling. `on_incompatible: unavailable` reports
`incompatible_model` for that source; `on_incompatible: fail` makes it an
operational failure, which cannot pass a gate.

Estimates never change findings, the index, or the gate decision in this
increment. A unit whose kind the pack does not qualify, and a unit for which this
build prepared no target, both report `unsupported_unit` rather than an invented
value.

Every output carries the same decisions. JSON and SARIF already serialize the
assessments and the manifest. The text, Markdown, and HTML writers replace their
fixed "no calibrated model" sentence with the configured pack, its declared
status and corpus, and how many units of its kind were estimated; they keep the
existing sentence when no pack applies. Declared pack values are data, so each
format escapes them. `unswell-mcp --model` mirrors `check --model`, so the server
and the CLI agree.

## Acceptance and follow-up

Blackbox tests must cover strict decoding, digest verification, contract and
column rejection, capability requirements, declaration consistency, each
applicability status, calibrated values with their uncalibrated score, mismatched
vectors, cancellation, detached snapshots, and concurrent estimates.

The remaining work under #24 is calibrated gating: a gate that requires a
probability cannot pass while that estimate is absent, and exit codes 0, 1, 2,
and 130 keep their meanings. Reserved reasons such as unsupported
domain and dictionary coverage need inputs the engine does not yet carry. A
command that builds a pack from a research training artifact is separate work.

A loadable pack is not a qualified model. Held-out evaluation (#25), corpus
qualification (#22), and the product decision to gate on a probability remain
open.
