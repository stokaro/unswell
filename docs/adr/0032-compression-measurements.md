# ADR 0032: Measure compression against an explicit reference

Status: accepted for a numerical primitive; corpus selection and qualification remain open.

## Decision

Add #52's compression primitive to the shared `feature` package. A caller supplies
an explicit reference and measures existing `nlp.PreparedUnit` targets. The
primitive does not extract text, choose corpus rows, infer origin, or apply a gate.
Reference acquisition, licenses, ordering, and training-partition verification
belong to the research adapter before any model learns these features.

The reviewed ZipPy revision is
[`37fe2b2`](https://github.com/thinkst/zippy/blob/37fe2b2424fa9c24174283fe8827fee35ce7e91a/zippy/zippy.py).
Its zlib variant compares mean raw-DEFLATE compression ratios across reference
chunks and assigns an origin label from their difference. Its normalization,
compressor, chunking, and classification are not adopted here. No upstream code
or reference corpus is imported. This experiment does not reproduce ZipPy.

Let `C(x)` be the byte size of a complete Go `compress/zlib` stream, including its
header and trailer, at an explicitly recorded level. It has no preset dictionary,
intermediate flush, normalization, or concatenated streams. `S` is the exact
UTF-8 reference; `T` is one prepared target; `P = S + LF`; `Q = LF`.

```text
seeded_increment = C(P + T) - C(P)
control_increment = C(Q + T) - C(Q)
incremental_bytes = seeded_increment / len_utf8_bytes(T)
reference_gain = (control_increment - seeded_increment) / len_utf8_bytes(T)
```

Both baselines are nonempty. This avoids comparing an empty stream's block framing
with a nonempty stream in the unconditioned control. A separator prevents direct
word concatenation with the reference. It is part of the formula and hash.
Neither ratio is perplexity, semantic similarity, or an origin probability.
Negative gain is a valid observation. Nonpositive increments retain raw counts
but produce absent feature values with `compression_boundary`.

The reference prefix is at most 32 KiB; a target is at most 64 KiB. A separate
input-byte budget bounds repeated compression work. The implementation records
the exact Go version because the standard library does not promise stable
compressed bytes between releases. Changing compiler or compressor settings
requires new compatible research artifacts; no compatibility shim is provided.

## Acceptance and follow-up

Blackbox tests must cover the stored-block control, shared prepared targets,
Unicode, protected boundaries, source/context identities, repeated identifiers,
short fragments, cancellation, limits, and concurrent calls. Freeze any observed
compression-boundary case with its toolchain identity, not a universal authorship
threshold. No source text is added to normal reports.

The next adapter must build explicitly ordered reference packs from permitted
training targets, compare matched reference cohorts through #57, retain cohort
provenance without feeding target provenance into the editorial model, and measure
cold/warm time, memory, size, coverage, and added predictive value. API templates,
short comments, code-like strings, and out-of-domain text remain required controls.
This primitive cannot complete #52 without those data and experiments.

Reserve reference groups inside the training partition before classifier fitting.
Keep every source, rewrite, and template connected to those references out of the
classifier's fitted rows. Otherwise a training target can compress against itself,
creating an advantage unavailable on held-out targets. The reference bank remains
fixed for classifier training, calibration, and prediction. Freeze this additional
group assignment in the manifest and report the resulting sample counts.
