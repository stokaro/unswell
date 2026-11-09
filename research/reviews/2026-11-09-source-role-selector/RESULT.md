# Source-role selector: exposed replay and page-disjoint check

Status: research only; #349 remains open. The selector and the page-disjoint
source list were signed before transfer sources were fetched. All 19 transfer
labels were signed before the selector was run. The reviewer was the assistant.

## Archived replay

The previously published operator filter retained 2311 of 2847 archived
proposals. Of its 45 excluded authorial-narration aliases, the source-role
selector admitted 17, withheld 16, and abstained on 12. All 17 admissions have
accepted grades in the exposed archive. Whole-page FULL increased from 45/57
to 46/57, and accepted delivered diagnoses from 1443/1700 (84.88%) to
1460/1717 (85.03%). Context stayed at 38/46 FULL and 389/413 accepted;
historical confirmation stayed at 17/20 FULL and 190/198 accepted.

The assistant had seen this bank and its grades before implementing the
selector. The 17 admitted items correspond to known old positives. These
numbers show only exposed replay arithmetic. Five archived supplied edits
still need safety resolution, and the candidate has no complete standalone
generation path. See [`archive_replay_result.json`](archive_replay_result.json).

## Page-disjoint transfer

Eight exact commit-pinned README pages from other repositories produced 19
lines under the frozen inventory rule. The assistant labeled all 19
`withhold` after reading the full pages: the lines contained technical facts,
reader steps, safety advice, comparisons, or deliberate attribution. The
selector withheld 10 and abstained on nine; it admitted zero. Every source's
local Git blob SHA-1 matched GitHub's contents API. The original source SHA-256,
byte spans, labels, and matrix are in `transfer_inventory.json`,
`transfer_labels.json`, and `transfer_result.json`.

This test found no false admission on its 19 controls. It has **zero positive
cases**, below the frozen minimum of five across three repositories, so its
status is **underpowered**. It provides no evidence of useful recall on new
pages. The archived numerical threshold crossing is therefore not a verified
product improvement or #349 qualification.

## Next decision

The current selector is a conservative admission filter over cached proposals,
not an end-to-end detector. A new, independently frozen page-disjoint set must
contain positive empty-narration cases and complete-page negative controls.
After any transfer success, evaluate generated findings in their actual
diagnosis-only delivery form, resolve supplied-edit safety, and measure the
unchanged FULL and accepted-diagnosis gates in every cohort. Do not tune this
selector on the exposed replay or present the 85.03% as deployed quality.
