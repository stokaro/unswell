# Post-run verification amendment

Recorded November 11, 2026, after the model run completed and before any gate
counts were inspected. The frozen `verify.py` stopped at its first response row:
`ValueError: Identity changed in response`. The frozen `score.py` used the key
`family` for both the saved input generator family and MELD's predicted family.
Its `item.update(...)` replaced the former in each scored response row. The
raw score files, calibration threshold, frozen input manifest, and model inputs
were not changed or rerun.

The completion marker has SHA-256
`b73a53b874d1aa6df4a9ca28fcff5d6b6321074e5a4c06aad1eef88a07408657`.
The immutable raw score digests are recorded in that marker. The threshold
file has SHA-256
`bc56770fe561ee42b4e5a319a2df8bbeccac5c154dc4ff82f30ae95dabe7f8b6`.

`verify_v2.py` is an analysis-only amendment. For each response, it requires
every sealed field other than `family` to match the pre-inference manifest in
the same position, including task ID, operation, prompt, source-text digest,
prose digest, length, and eligibility. It verifies that the stored `family`
is one of the checkpoint's declared prediction classes, then restores the
input generator family from the sealed manifest for descriptive grouping.
It independently recomputes the calibration threshold and all three gates
from the unchanged raw numeric scores. The original failed `verify.py` is
retained as evidence. This amendment does not change an input, model,
threshold, label, denominator, score, or gate; no new inference is permitted
under this run.
