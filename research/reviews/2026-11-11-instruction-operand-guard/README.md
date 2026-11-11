# Configuration operands are not instruction actors

**Status: exposed development precision correction.** Rule version 18 of
`filler.instruction-scaffolding` excludes a simple modal `used to` or `used for`
capability when its subject is an opaque code identifier alone or an explicit
environment variable. These subjects identify a configuration operand. The
change preserves warnings about a named function's action and about a separate
relative helpfulness clause: the latter still contains two support layers.
It does not interpret the contents of protected identifiers.

Two previously reviewed bat release-note warnings illustrated the error:
`BAT_CACHE_PATH` can be used to place cached assets, and `PAGER` / `BAT_PAGER`
environment variables can be used to control the pager. The November 7
complete-source review marked both nonactionable. The condition and optional
capability remain meaningful; neither sentence asks the reader to perform an
unnecessary setup step. The old rule treated modal `can` plus passive `used`
as two support layers even when the subject was an operand.

The [paired reports](record.json) replay the same frozen 36 complete pages on
base main `7f86b91` and the candidate working tree with Go 1.27.3. Both
technical and strict reports are complete, cover the same original source
bytes, and have no skipped rules or operational errors. The verifier first
checks that the base report has exactly the same normalized findings as the
archived November 10 stream; it then checks that the two bat findings are the
only removals and that all retained diagnoses are unchanged. The reports are
exposed development evidence, not prospective confirmation.

| Complete cohort | Before | After | Accepted after | FULL events after |
| --- | ---: | ---: | ---: | ---: |
| Technical whole page | 126 findings | 124 findings | 82/124 | 25/57 |
| Strict whole page | 145 findings | 143 findings | 82/143 | 25/57 |
| Technical contextual | 27 findings | 27 findings | 21/27 | 16/46 |
| Technical historical confirmation | 9 findings | 9 findings | 6/9 | 4/20 |

The table uses the later, conservative assistant review for event credit. The
earlier assistant review credited three additional whole-page events on the
unchanged stream; neither interpretation reaches the required 80% FULL-event
recall. Whole-page accepted delivery rises only from 82/126 to 82/124 for the
technical profile, and from 82/145 to 82/143 for strict. Both remain below the
85% requirement. No event label, source, denominator, suggestion, or unrelated
rule was changed. This correction is not origin detection or proof of a safe
edit. Issue #349 remains open.

`record.json` pins the source archives, archived review and reports, local
binary digests, and the code-patch digest. The local candidate binary used
`BuildCommit=local-operand-guard-v4`; that marker denotes a working-tree
replay, not an immutable commit or CI result. The published PR and main CI
must provide separate exact-tree acceptance. To check the committed evidence:

```sh
make check-instruction-operand-guard
```
