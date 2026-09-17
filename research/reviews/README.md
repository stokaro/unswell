# Reviews on real texts

A review record runs every catalog rule on a real text. A reader then
judges a sample of the findings under each rule's own statement. The
question is one: is the construction the rule names present in prose a
reader could act on? The record names the reader, the tool commit, the
configuration, and the counts, and it says what the judgments do not
establish. It is the review item of
[ADR 0037](../../docs/adr/0037-diagnostics-not-authorship.md).

Under [ADR 0041](../../docs/adr/0041-assistant-review-acceptance.md), the
maintainer accepts assistant review for diagnostic development and rule
acceptance. The review identifies its actual author and measures behavior
against that author's judgments on the stated sample. It shows where a
diagnostic finds the intended construction and where it fires without a
useful reason. It also records missed defects when whole pages are annotated.
Independent human review is not a prerequisite for this work.

- [2026-09-11-ptah](2026-09-11-ptah/README.md): every rule on the Ptah
  repository, one reader, ten findings per rule.
- [2026-09-11-responses](2026-09-11-responses/README.md): every rule on
  the 800 controlled responses, one reader, every finding judged.
- [2026-09-11-unswell](2026-09-11-unswell/README.md): every rule on this
  repository under its own file selection, one reader, ten findings per
  rule.

- [2026-09-13-replay](2026-09-13-replay/README.md): the same pinned inputs
  under old and repaired engines, source-based alignment, revisited samples,
  retained positive controls, and bounded follow-ups.

- [2026-09-17-full-page-recall](2026-09-17-full-page-recall/README.md): complete
  sources annotated before new detector-output review, missed-event recall,
  all-finding dispositions and historical controls. The maintainer accepted these
  single-assistant judgments as diagnostic development evidence.

- [2026-09-17-local-repetition](2026-09-17-local-repetition/README.md): short
  repeated assertions and adjacent-word fixes, with all diagnostic deltas and
  12 newly annotated complete pages. The new pages show no detection gain;
  broad recall and default qualification remain open.

- [2026-09-17-context-recall](2026-09-17-context-recall/README.md): five contextual
  wording warnings recover six more exposed defects. The eleven new pages stay
  at 1/42 detected events, with all diagnostics reviewed. The full-page objective
  remains unmet.

- [2026-09-17-instruction-recall](2026-09-17-instruction-recall/README.md):
  indirect instruction and redundant-predicate rules raise new complete-page
  detections from 4/99 to 12/99. All eight gains occur on one historical Mermaid
  page; Ptah remains at 2/35. The records retain partial matches, review burden
  and the unchanged candidate-budget abstention on a long reference page.

- [2026-09-17-construction-recall](2026-09-17-construction-recall/README.md):
  local grammatical scaffolding raises exposed instruction-page recall from
  12/99 to 27/99; six new complete pages move from 0/32 to 1/32. Existing
  review burden and embedded-program false alarms remain visible.

- [2026-09-17-purpose-recall](2026-09-17-purpose-recall/README.md): contextual
  purpose and reader-claim rules recover ten exposed defects. Six new complete
  pages remain at 1/28, with no new detection; the record retains all 27 misses
  and the distinction between useful examples and broad recall.

- [2026-09-17-verb-scaffolding](2026-09-17-verb-scaffolding/README.md): nested
  action and passive method diagnostics recover three exposed historical defects.
  Six new pages remain at 0/43. Additional post-diagnostic editing observations
  stay separate from frozen recall; Ptah has no measured gain in this iteration.

- [2026-09-17-instruction-clauses](2026-09-17-instruction-clauses/README.md):
  narrated actions and named operations gain four exposed detections but lose
  two under a stricter capability guard. Six fresh pages remain at 4/58, while
  three capability false positives disappear. The complete report retains the
  losses, partial matches and 54 missed events.
