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

# November 7 prospective disposition

The original exposed-development protocol, candidate code at `35e3350`, raw
reports, and assistant decisions remain unchanged. A second protocol was
committed at `53b8578` before external source search, and the nine selected
source identities plus 13 passage judgments were committed at `c5c67c1`
before candidate replay. The first search query was repeated solely to archive
the full metadata page after its first response was inspected without a saved
copy; its first three hit identities and order matched. The second and third
queries ran once.

Both exact binaries completed on every selected page in technical and strict
profiles. The candidate added one warning to an acceptable unchanged council
request and supplied inapplicable advice; no prior finding changed. The
prospective set had no definite actionable hope event, so it offers no recall
estimate. This is a negative default-promotion result. The final PR removes
the candidate runtime change while retaining the historical, source-bound
research record. It does not revise the 123 frozen #349 event labels or any
original exposed-development result.
