# Implementation-only amendment after the first replay

The first complete candidate replay found three new findings on the frozen
36-page set. The subsequent required `make check` passed Go tests but failed
lint: three matching functions exceeded the repository's complexity limits,
and three blackbox test lines were too long. The same run also identified
version-12 expectations and generated catalog/golden data requiring updates
after the rule became version 13.

The matcher was split into smaller helpers without changing its token
conditions, windows, suggestions, rule version, or policy. The test lines were
wrapped. Existing version expectations were updated, the rules catalog was
regenerated, and e2e feature goldens were regenerated. Review of the golden
diff found only the new ruleset hash. The final refactored binary was rebuilt
and replayed on all 36 pages; both compressed complete reports are byte-for-byte
identical to the first candidate reports:

| Profile | SHA-256, before and after refactor |
| --- | --- |
| Technical | `415639551d9ad7b923b5dcaf8b79c08e8ff4086d3e6fe2f4528579fe25faa459` |
| Strict | `b7574eaa20bcd6ead3ad3f354fb668965666292cf1086801de635095a510bc3a` |

The archived `runs.json` binds the final refactored binary. The original
failed `make check` output remains local at
`/private/tmp/unswell-hope-make-check.log`; it is not part of the frozen source
evidence.
