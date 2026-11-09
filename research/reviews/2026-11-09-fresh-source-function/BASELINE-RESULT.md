# Current-main baseline on nine new pages

The sealed source-only v2 review has **17 clear positives and 30 controls**
across nine page-disjoint README files. The unchanged public CLI built from
main `14722f0` completed all nine technical-profile analyses with zero
operational errors. `baseline-audit.json` pins each complete JSON report by
SHA-256 and verifies that every finding span resolves to original UTF-8 bytes.
The complete reports remain under the local artifact directory named in
`RUN_LOG.md`.

The CLI delivered **11 findings**. Only one overlaps a clear positive:
Tencent/TNN's bare `It is very simple to use TNN` claim. Its diagnosis is
useful, but the generic suggestion does not itself supply the source-specific
safe deletion recorded by the reviewer. It located **1/17** clear positives;
this is a location count, not FULL recall or a qualification result. No
delivered finding overlaps a `retain` control. Two overlap `optional` controls.

Source-context review of **all 11 delivered findings**:

| Page | Rule | Disposition | Reason |
| --- | --- | --- | --- |
| 1 | announced importance | optional | The offset warning is useful; shortening its lead-in is a voice choice. |
| 2 | announced importance | optional | The security comparison is useful; shortening its lead-in is a voice choice. |
| 3 | long sentence, author block | reject | A name and HTML affiliation list is not a sentence to split. |
| 3 | long sentence, disease database | optional | Splitting could help, but the processing sequence and condition are substantive. |
| 4 | instruction scaffolding, width | reject | The conditional use case explains when to pass a number. |
| 4 | wordy phrase, `in order to` | optional | `To` is shorter, but the existing phrase is clear and harmless. |
| 5 | long sentence, apology | optional | An author's closing apology can be split, but its personal voice is intentional. |
| 6 | long sentence, history | optional | The historical chronology is dense but substantive. |
| 7 | instruction scaffolding, Electron | reject | The condition scopes the following installation steps. |
| 7 | instruction scaffolding, prompt | reject | The sentence states an additional prompt capability. |
| 8 | unscoped assurance | accept | A bare ease claim precedes the actual three-step instructions. |

Thus the strict accepted-delivery count on this deliberately cue-enriched
sample is **1/11**. Optional findings are not silently counted as accepted.
This sample does not estimate ordinary-page prevalence or replace the frozen
36-page/123-event #349 cohort. It demonstrates a concrete transfer gap:
current-main rules miss most newly reviewed praise, framing, and repetition
cases while issuing several source-role errors. No runtime detector changed.
