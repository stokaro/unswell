# New historical technical source frame: offline preparation

Date: November 10, 2026. This is private source preparation for the origin
control track. **No EditLens scores or hosted model responses were produced.**
The score protocol remains a draft until its runner and verifier are pinned
before inference. The historical documents are controls by date, not verified
human-authorship labels.

## Selection and accounting

- The SHA-256-pinned DocMine 2021 CSV supplied repository discovery only.
  Exact content came from GitHub Git blobs at pinned pre-August-15-2021 commits.
- The source-disjoint [exclusion manifest](origin-new-corpus-exclusions-2026-11-10.json)
  removed 173 repositories used in prior scored technical and historical sets.
  A deterministic rank then selected up to eight first-party Markdown/RST
  blobs per repository before reading their content.
- Of 465 repositories, 340 contained at least one candidate, yielding 860
  selected blobs. Git SHA-1, raw SHA-256, UTF-8, NUL, preprocessing, tokenizer
  length, and duplicate checks left 796 preliminary eligible pages. The other
  64 were retained in the audit with reasons: eight blob/content errors, one
  duplicate cleaned source, four duplicate raw sources, three under the
  50-word window minimum, two previously scored raw sources, and 46 over the
  512-token window limit.
- A score-blind review queue flagged 93 preliminary eligible pages by path,
  script signature, language cue, table structure, or short prose. Assistant
  review of paths and flagged excerpts excluded 35 pages: 9 non-English, 12
  generated/template/report, 5 vendored, and 9 non-prose reference pages.
  All 58 remaining queued pages were retained. Unqueued pages received the
  automated checks only. This review does **not** verify the authorship of
  any document.
- Final source frame: **761 of 860** candidates. Repository-level hash split:
  **404 calibration** and **357 test** documents; 323 represented repositories
  split 160/163. Test-frame coverage before inference is 357/403 = **88.6%**.
  The 300/300 minimum is met. The score has not been computed, so the ≤1%
  test flag gate has no result.

The [frame](frame.json) SHA-256 is
`351add15fed1123dd1b312f9c371353de3ac8e06dc1c2e3745dfc1d190031d72`;
its identity digest is
`6364b4f7405796240bd751f593c91bc910f5c9aa6cadceef08275c5a7871d743`.
The [protocol draft](PROTOCOL.md) states the planned threshold and failure
policy. [Inventory](inventory.json), [blob audit](blob_audit.json),
[eligibility](eligibility.json), [review queue](review_queue.json), and
[source decisions](source_review.json) preserve all candidates and reasons.
The 8.3 MiB of exact source blobs is local at
`/private/tmp/unswell-origin-new-corpus-v1-sources`; it is not in the Git
tree or investor material.

## Verification

The independent `verify_inventory.py` confirmed 465 repository candidates,
2,010 metadata document candidates, and 16 repository-level errors, including
15 missing GitHub repositories and one missing historical commit. The
independent `verify_frame.py` read the exact local blobs and confirmed all 860
decisions, 35 source-review exclusions, 99 total exclusions, no repository
leakage between arms, unique selected raw/cleaned hashes, and the final frame
digest. Changing one selected source SHA-256 in a copied frame caused the
verifier to reject it. A separate eligibility replay produced the same SHA-256
`ded775eff6e65f56bab8edf9b2429ffb9eb93b5bd33a5e3c20353ea1d1dbdbd5`.

Next: pin the score runner, independent score verifier, modelkit, input frame,
threshold procedure, and failure rules in a local pre-inference commit. Run
calibration first, persist the resulting threshold, then score the untouched
test arm. Report the observed flag rate and source-cluster uncertainty without
calling it a human false-positive rate. A negative result retires this
configuration; it does not justify retuning on the test documents.
