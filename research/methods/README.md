# Method registry

This directory implements the design record for
[#55](https://github.com/stokaro/unswell/issues/55), under
[#59](https://github.com/stokaro/unswell/issues/59). It contains source review and
a fixed [comparison protocol](protocol-v1.md), with links to subsequent evidence.
The [ADR](../../docs/adr/0015-research-methodology.md) defines the product boundary.

[`registry-v1.json`](registry-v1.json) uses format `unswell-method-registry-v1`.
Its `revision` increases when an observation or decision changes; a shape or
meaning change needs a new format version. Each entry identifies its publication,
unit, computation, code revision, components, limits, decision, and follow-up.
Unknown component terms, resource hashes, and measured costs remain null with a
reason. A repository license is not permission for its models or source datasets.

[`reviewed-resources-v1.json`](reviewed-resources-v1.json) contains measured SHA-256
digests and byte sizes of the small upstream files read on September 8-9, 2026.
Each URL includes the reviewed Git commit. These are source-review identities,
not hashes of downloaded model packs or datasets. No upstream code, weights,
tables, or corpus data is bundled in this directory. The original September 8
review did not execute upstream code. The later
[LLMDet component experiment](../llmdet/README.md) runs its pinned proxy function
and classifier on authored numerical controls; it does not reproduce text detection.
Downloading and hashing a README is not reproduction of its scientific results.

For each method, track these evidence fields independently:

| Field | Evidence needed for `complete` |
| --- | --- |
| `reviewed` | Named sources and a bounded account of what was inspected |
| `reproduced` | Pinned original implementation/resources, executable run, saved outputs, and measured costs |
| `ported` | Limited supported pack, Go/reference token and numeric parity, declared tolerances, and failure cases |
| `qualified` | Human-labeled evaluation on the declared product scope meeting the locked protocol |

Other states are `not_run`, `not_applicable`, and `blocked`, each with a reason.
An upstream publication's evaluation cannot populate Unswell's `reproduced` or
`qualified` evidence. A benchmark dataset has no Go-port requirement, but its
evaluation and resource review still need evidence. A new revision can revoke
compatibility without erasing the older evidence record.

Decisions are `adopt`, `experimental`, `defer`, or `reject`, always with an explicit
scope. Adopting a feature hypothesis or benchmark design does not adopt a trained
detector or authorize data redistribution. The current decisions are:

| Method | Decision and scope | Next evidence |
| --- | --- | --- |
| StyloAI / stylometry | Adopt compact feature comparison; defer upstream reproduction | #50 and #56; locate an authoritative code/model release |
| LLMDet | Experimental numerical component parity | #51; resolve pack terms, reproduce tokenization and tables, then measure prose coverage |
| Fast-DetectGPT | Experimental reference only | #53/#57; isolated fixed-budget model run |
| Binoculars | Defer optional reference execution | #53; justify the two-model budget and tokenizer/model terms |
| DivEye | Defer code use and execution | #53; resolve component terms and the model dependency |
| ZipPy / compression | Experimental shared Go measurements; no upstream reproduction | #52; build independent reference groups and measure added value |
| RAID | Adopt robustness dimensions; defer dataset execution | #57; review source-data permissions and pin a permitted subset |
| DetectRL-X | Adopt editing scenarios; defer dataset/code use | #57; resolve conflicting stated usage terms and pin resources |

The [DetectRL-X README](https://github.com/ATH-MaaS/Marco-LLM/blob/f291775fc1f78d00e52ebbf390d5a1445e05f0af/DetectRL-X/README.md#-license)
names Apache 2.0 and also restricts use to noncommercial academic research.
Record that unresolved conflict before importing any data. The
[DivEye license](https://github.com/IBM/diveye/blob/ee7d53abfb56ea4dc4cac2cb24e17beec0146e5e/LICENSE)
is CC BY-NC-SA 4.0; its code is not a permissive runtime dependency.

To extend the registry, inspect pinned primary sources, record separate terms for
code, tokenizers, weights/classifiers, tables, and data, and add hashes only for
bytes actually obtained. Preserve prior run references when updating decisions.
Use `not_run` for an experiment that has not happened, even when a paper supplies
favorable numbers. #57 will add executable comparisons; this directory does not
provide a benchmark runner or change public configuration/report schemas.
