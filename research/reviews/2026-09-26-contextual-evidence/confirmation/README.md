# Prospective Ptah confirmation pages

Three complete pages are pinned at Ptah commit
`eabfc8197bcfc118ee1f48a4f957315c7606b677`. They were selected before reading their
contents: all documentation paths absent from the preceding conservative
exclusion inventory. There were 147 documentation files in the pinned tree and
three eligible new paths. No length, keyword, diagnostic, or outcome filter was
used. The earlier local Ptah checkout was stale and had no eligible new pages.

| Page | Definite defects | Acceptable controls | Uncertain alternatives |
| --- | ---: | ---: | ---: |
| Verify a release against the database | 8 | 4 | 2 |
| The migration log | 2 | 4 | 3 |
| The online mode | 4 | 4 | 2 |

The reviewer is the Codex assistant under ADR 0041, not a human or an independent
annotator. Every body section was read, including tables and technical context.
The 14 definite defects, 12 controls, and seven uncertain cases were recorded
before running any detector or model on these pages. No prediction or measured
recall is available. Source checks prove integrity, not the editorial judgments.

Before any output, a second reading by the same assistant moved two initial
defects to uncertain: an Atlas user's expectation explains a real compatibility
concern, and the server-enforcement comparison supplies a concrete technical
basis. The initial 16-defect review remains in `annotations-initial.json`;
`review-amendment.json` records both changes and the resulting 14-event denominator.
This is a correction to one assistant's judgment, not independent adjudication.

The review distinguishes functional contrasts and safety explanations from
avoidable rhetorical framing, vague responsibility, and categorical claims
contradicted or unsupported by the surrounding text. For example, the release
verification page makes an unqualified claim that nothing changes, then explains
that an Oracle autonomous routine can insert and commit. Both broad promises are
one multi-target defect. The page's concrete explanation of why empty checks
cannot pass is a clean control, not a defect merely because it uses a contrast.
Code and front matter are not diagnostic targets. The review does not establish
implementation correctness or independently validate database behavior.

The known-exposure search covered ordinary research/documentation files in the
main-derived study worktree and recovered September 21 worktree, plus 2,026 gzip
artifacts. No selected path was found. This does not prove absence from every
conversation or model's training. All pages belong to one project and share
editorial conventions; this small cohort cannot alone qualify broad contextual
recall or measure cross-project transfer.

`selection.json` records the selection and pinned Git blob identities;
`source-freeze.json` binds the retrieved full sources and retained MIT notice.
`annotations.json` contains exact UTF-8 ranges, rationales, suggested repairs,
meaning-preservation constraints, and complete section coverage.
`review-freeze.json` binds the pre-output judgments. Revisions require an explicit
amendment; do not silently relabel after seeing output.

These pages are **not added to the pending 36-call experiment**. Their original
purpose is separate prospective confirmation after a successful development
screen. Keep the model, prompt, extraction, and thresholds fixed before output
inspection; future calls need their own authorized scope. Retain every selected
page and defect if execution fails, and judge every emitted finding, including
those outside labeled defects. Keep uncertain diagnoses in the finding precision
denominator and assess suggestion safety separately. Same-project confirmation
must be reported separately from the original 36 exposed pages and any later
cross-project confirmation.

Validate without running a model or the detector:

```sh
python3 -B research/reviews/2026-09-26-contextual-evidence/verify_confirmation.py
```
