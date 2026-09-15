# Ptah contrast examples

The three original passages come from `stokaro/ptah` at
`7b47e7cfb5d4ff32a38375345069f5533bff892f`. `provenance.json` records full-file
and excerpt hashes and the contiguous line selections. `LICENSE.ptah` preserves
the source's MIT notice. Expected-diagnostic annotations are masked by the e2e
harness before analysis; CRLF and BOM are materialized by the harness.

`rewritten.md.txt` contains agent-written alternatives, not independently rated
human improvements. The expected change is removal of the repeated contrast
frame. These obligations were checked against the source passages:

- Adoption keeps `--check`, unchanged unrelated bytes, complete refusal of
  `unsupported`, the same artifact, both URI schemes, and the missing-registry
  error. No successful partial conversion is introduced.
- Directory selection keeps all four overriding inputs and the failure for a
  missing explicit directory. Falling back to `./migrations` stays forbidden.
- Registry scope keeps every excluded hosted service and the supported
  publishing/promotion capability through any OCI registry.

The originals contain useful technical distinctions. A detection means their
contrast form repeats, not that the facts should be deleted. Both originals
and alternatives pass the configured gate because the rule is a zero-score note.
This fixture proves deterministic behavior and source coordinates. It supplies
no editorial accuracy estimate or authorship labels.
