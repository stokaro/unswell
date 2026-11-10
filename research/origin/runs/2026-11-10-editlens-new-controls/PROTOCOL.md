# New historical technical-document origin control: preparation v1

Status: pre-inference protocol, November 10, 2026. No model has scored this
frame. The final source, eligibility, split, preprocessing, model, runner, and
verification hashes must be sealed in `freeze.json` and a local Git commit
before any new inference. Later changes require a new version and cannot alter
this inventory or earlier scored results.

## Discovery frame

Use DocMine v2's 2021 `Repositories.csv` only to discover repositories; its
SHA-256 is `5da66ce7e15c41af2df8d6883e8d1ccef79b2157afd0b665f4157420df087710`
and MD5 is `035a49eded3e1cbe18f71ecc766f59ba`. Include all C#, C++, Java, and
Python rows with at least 100 recorded stars and at least three textual files,
then exclude the 173 repositories in
`origin-new-corpus-exclusions-2026-11-10.json` (SHA-256
`f3dcfc6021254fdee28012b4ffc563f2b83a71c6d23c5cc3dcab6e3170220f79`).
The resulting 465 repositories are ordered by SHA-256 of lowercase
`owner/repo`, NUL, and `unswell-techdoc-new-controls-v1`.

For each repository, inspect its latest GitHub commit with committer date
strictly before August 15, 2021 UTC. Exclude a missing commit or truncated
recursive tree, recording the reason. From the exact tree, consider first-party
Markdown/RST blobs of 2,000–200,000 bytes. Exclude vendored/dependency paths
and administrative files as implemented in `inventory.py` and checked by
`verify_inventory.py`. For each repository, take at most eight candidates in
SHA-256 order of lowercase repository, NUL, exact path, NUL, and
`unswell-techdoc-new-document-v1`. This order is fixed before reading any text
or score. Store the complete repository and candidate inventory, including
errors; do not silently substitute a failed repository or blob.

Fetch exact Git blobs, verify Git blob SHA-1 and raw SHA-256, decode UTF-8,
reject NUL and known fixture/vendored/non-English pages by a recorded source
review before model inference. Preserve every attempted source, exclusion
reason, and digest. Keep source text local; the report uses hashes and source
identities. Reject a selected page if its raw SHA-256 matches a previously
scored page or another candidate. Check cleaned-text hashes and possible
clones separately before freezing. No source is selected or dropped based on
model output.

## Planned score protocol, pending source audit

Reuse the previously pinned EditLens ONNX FP32 model (SHA-256
`ddd1173f2ef517ad499965e5029fae8099a8054a2bc76d8134e5889cc4ed1b3e`),
tokenizer, and text
cleaning. Split all cleaned prose into nonoverlapping 100-word windows,
folding a final tail under 50 words into its predecessor. A page is eligible
only if it has at least 50 cleaned words and each resulting window has 1–512
tokenizer tokens. Treat malformed
source text as an ineligible document, and missing model capability or
inference failure as an operational error that stops the run. The denominator
for coverage includes all source-audited, first-party candidate documents,
including ineligible ones. An origin score is an index, never a probability.

Assign whole repositories to calibration or test by the low bit of the first
digest byte of SHA-256(lowercase repository, NUL, `unswell-techdoc-new-split-v1`), with zero
for calibration and one for test. Do not move repositories between arms after
observing counts or scores. Require at least 300 eligible documents in each
arm and no repository or exact source overlap with exposed prior sets. If the
frame falls short, stop before inference and report the shortfall.

The runner scores **only calibration windows first**, computes each document's
arithmetic mean of its full cleaned-prose windows, and writes a durable
`calibration.json` containing every calibration score and the strict `>`
threshold equal to the maximum calibration-document mean. Only after that
file exists does it score the untouched test windows. A test document is
flagged when its mean is strictly greater than the persisted threshold. The
score is an index, never an authorship probability. The control feasibility
gate requires at
least 70% eligible-document coverage on test-frame candidates and no more
than 1% observed flags among eligible test documents. Report all document
and window scores, abstentions, language/repository/length breakdowns, and a
descriptive 95% Wilson interval for the fraction of test repositories with at
least one flag. Repository clustering limits a document-level independence
interpretation; the Wilson interval is not used to move the gate. Historical
commit dates are not verified human-authorship
labels; use the term *historical-control flag rate*, never human false-positive
rate. Possible training overlap with EditLens remains unknown.

The full frame has 860 selected metadata candidates, 761 selected after
source-only checks, 404 calibration documents, 357 test documents, and 403
test-frame candidates. The exact `frame.json` identity SHA-256 is
`351add15fed1123dd1b312f9c371353de3ac8e06dc1c2e3745dfc1d190031d72`.
Verify all source blobs and the frame immediately before inference. Use the
pinned CPU provider, four threads, batch size eight, and the same weighted
four-class score index as the earlier full-window study. An input hash,
provider, shape, nonfinite score, or inference error stops the run. Do not
replace a source, rerun a failed window, resume a partial packet, or recalibrate
after seeing test scores. Keep every raw score and error. A separate verifier
recomputes document means, threshold, flags, group counts, and gate from the
saved window scores.

This gate alone cannot qualify a detector. The matched generation/editing
screen needs a separately frozen packet and source/generator holdouts. If the
control gate fails, preserve the negative result and retire the configuration
without tuning on the test scores.
