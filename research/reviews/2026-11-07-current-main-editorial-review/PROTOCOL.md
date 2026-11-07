# Exposed complete-stream editorial review for #434

Status: one Codex assistant's source-context review, November 7, 2026, under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md). The reviewer
is an AI assistant, not a human annotator or an independent rater. The 36 source
pages, 123 frozen events, September reviews, current findings, and the
maintainer's earlier comments on selected pairs were available before this
review. This is an exposed development assessment; it is not prospective
confirmation, inter-rater agreement, or a population estimate.

The public-engine reports were recorded from commit
`5dd76191740c3d2f003b6ec90a9b381ccf1b772a` under the preceding
[baseline protocol](../2026-11-07-current-main-baseline/PROTOCOL.md) and merged
without engine or policy changes as `b98cb51e4155213a208bdc8b1a044805c7c1875d`.
The reports retain all 155 technical and 176 strict findings. No runtime rule,
threshold, source, reference label, or cohort denominator is changed here.

Every unique strict finding was read with the surrounding original page and
related locations. A finding is `actionable` only when it identifies a wording
change required by the [editorial rubric](../../../docs/editorial-annotation.md)
and a plausible revision can preserve the technical claims. `nonactionable`
means the warned construction is acceptable in its context; `uncertain` keeps
borderline cases in the delivered denominator. A long sentence or em dash is
not automatically a defect. The 21 strict-only findings are reviewed separately
from the shared stream. Each finding has an exact ID, original byte span,
profile membership, decision, and rationale in `findings.json`.
Actionable findings also include a proposed repair action. These are scoped
editing directions, not applied sentence replacements; the per-finding rationale
names the relevant cases or clauses.

The complete 123-event ledger in `events.json` retains the frozen categories,
source paths, and denominators. `full` requires an actionable delivered finding
that identifies the same defect and overlaps every target location. `partial`
means useful but incomplete diagnosis of a multi-layer event. `uncertain`
remains unresolved. All other events are `missed`; no generic length or
punctuation warning earns a rhetorical event merely by touching its bytes.
The verifier checks these necessary span and actionable-witness conditions.
Semantic judgments remain the reviewer's, not a claim the verifier can prove.

Advice is assessed separately. A missing suggestion is `not_supplied`.
`conditional_guidance` means the finding is actionable and its text offers only
general editing guidance, not a concrete replacement whose meaning was tested.
`not_applicable` marks advice attached to a nonactionable finding; `unresolved`
marks advice on an uncertain finding. None of these statuses means an edit was
applied or proved safe. No automatic rewrite was introduced.

The first scratch pass was revised after comparing its rationales with the
unchanged frozen annotation explanations. The final review was frozen only
after that reconciliation; it is not a blind confirmation. The maintainer's
earlier judgment that the entire offline-HTML address sentence has an awkward
rhythm remains visible as an unresolved event when the delivered warning
isolates only one clause. The original annotation is not silently changed.

The experiment selected no new operator after the 18-control screen. Its
operator-specific before and after streams are therefore identical to the
current 36-page baseline; gained and lost events from an admitted #434 rule are
both zero. The September-versus-current finding-location changes are reported
separately in the preceding baseline record and are not credited to this audit.
The unchanged #349 80% FULL-event and 85% accepted-delivery thresholds are
applied to every complete cohort. Failure to meet either is a negative result,
not a reason to change the denominator or open the reserved prospective pages.

`freeze.json` binds the original archives and labels, both raw reports, the
baseline analysis, review decisions, this protocol, and verifier code by SHA-256.
`verify.py` checks source quotes, identities, complete finding/event coverage,
profile membership, actionable witnesses, advice dispositions, and exact
`summary.json` reproduction. Its negative tests reject missing review entries,
unsupported FULL credit, and an inapplicable advice promotion. This audit uses
Python only for offline evidence checking; the product runtime is unchanged.
