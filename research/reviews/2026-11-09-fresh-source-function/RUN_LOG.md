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
