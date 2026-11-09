# Acquisition run log

- The search responses were complete. `selection.json` was committed as
  `7cb0a15` before any source was fetched or read.
- The first fetch attempt pinned all nine default-branch commit SHAs, but its
  base64 decoder rejected the line breaks in GitHub's response. No source
  bytes were decoded or reviewed. Its inventory is preserved in
  `artifacts/research/2026-11-09-fresh-source-function/sources-fetch-attempt-1/inventory.json`.
- The decoder was corrected to remove base64 whitespace before strict decode.
  The nine identities and source-match requirement are unchanged.
- All nine selected files were fetched successfully. Their Git blob SHA-1
  values match the search responses. Complete original bytes and raw search
  responses are retained locally in
  `artifacts/research/2026-11-09-fresh-source-function/`. The tracked
  `source-inventory.json` pins every repository, default-branch commit, path,
  byte length, Git blob SHA-1, and source SHA-256. The content is also
  retrievable at `https://github.com/<repo>/blob/<commit_sha>/<path>`.
- Source-only labels were sealed before candidate execution. The first seal
  and its pre-candidate consistency correction are both tracked; see
  `REVIEW-REVISION.md`. Use `labels-v2.json` for subsequent screens.
- The unchanged current-main CLI baseline completed on all nine whole pages.
  Complete reports are in the local `baseline/` artifact directory; report
  hashes and every delivered original-byte span are in `baseline-audit.json`.
  See `BASELINE-RESULT.md` for the manual source-context disposition.
- The exact offline source-function candidate and protocol were committed as
  `44732a7` before its first run. The output and full same-assistant review are
  tracked as `candidate-output.json` and `candidate-review.json`. The exposed
  screen located all 17 positives and credited 15 FULL, but one supplied
  deletion leaves a numbered list broken. `CANDIDATE-RESULT.md` retires v1;
  it was not integrated into the runtime or tested on prospective pages.
- V2 selection and 12 source identities were committed in `07cc47b` before
  transfer source content was fetched. All 12 source files were pinned and
  validated by Git blob SHA-1 and SHA-256; the tracked inventory was committed
  in `265a805` before full-page source review.
- Source-only transfer labels were sealed in `51b2907`: 12 clear positives and
  35 controls. The unchanged CLI then completed 11 pages and failed Markdown
  parsing on page 3. The pre-output operational-error amendment was committed
  in `0ee2469`; it leaves the diagnosis core untouched but makes this run
  exploratory rather than a clean preregistered transfer.
- The v2 candidate was run once. Its 22 diagnoses and complete dispositions
  are tracked in `transfer-v2-candidate-output.json` and
  `transfer-v2-audit.json`. `TRANSFER-V2-RESULT.md` records 2/12 FULL positive
  events, 2/22 accepted candidate diagnoses, three retained-control hits, two
  unsafe proposed deletions, and one operational error. V2 is retired; no
  runtime change or #349 qualification followed.
