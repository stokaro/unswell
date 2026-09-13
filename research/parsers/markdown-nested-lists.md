# Markdown nested-list parsing

Issue [#231](https://github.com/stokaro/unswell/issues/231) records a complete-scan
failure in FastAPI's `docs/en/docs/advanced/security/oauth2-scopes.md`. The source
is pinned to `77f7447ee8917697e702b7eca8d393e11dc1e365` in the #219 corpus.

## Cause and repair

The failure reproduces directly with gotreesitter v0.52.0, before Unswell's
framing or extraction. Four nested list items parse; five, eight, and sixteen
produce error nodes. Both two-space and four-space indentation reproduce it.

The external scanner emits `_block_close` once per open block when a list ends.
Several closes can share the same byte position and parser state. The runtime's
general guard permits four consecutive zero-width tokens. At EOF it replaces
the fifth with an end token; at an internal dedent it can skip a source rune.
`ParseStrict` can then return an error tree with a nil Go error and a complete
root byte range. Unswell must still reject error and missing nodes.

The patch classifies Markdown `_block_close` as a repeatable external token.
This uses the existing 4,096-token bound, alongside the existing parser time,
memory, and node limits. It does not change other grammars or accept error trees.
Unswell's 128-node syntax-depth guard remains in place.

The adapter also masks a leading BOM with blank lines, rather than three spaces.
The old mask changed the first list marker's indentation. Both masks retain the
original byte offsets; only the parser's owned input changes.

## Dependency record

- Upstream base: v0.52.0, `2295871057f860598a006d6068588a6303fefb02`.
- Parser patch: `c3dfaca` in [stokaro/gotreesitter](https://github.com/stokaro/gotreesitter/tree/unswell-v0.52.0).
- Published module: `github.com/stokaro/gotreesitter` at
  `v0.52.1-0.20260913084044-276f5cc0ec6f`.
- Fork revision: `276f5cc0ec6f9e8ca99e20b9fa6cea3b240309ca`.
- Upstream submission: [gotreesitter #1109](https://github.com/odvcencio/gotreesitter/pull/1109),
  based on upstream main, with only the parser change and regression test.
- The second fork commit changes module/import paths and adds a patch record.
  It does not change parser behavior. Original licenses and notices remain intact.

The fork has its own module path because a dependency's `replace` directives do
not apply to its consumers. This lets the CLI, MCP, and library use the same fix.
Return to the original module after an upstream release includes the patch and
passes the extraction and corpus checks. Do not change historical corpus files
to claim that they were measured with the new dependency.

## Regression checks

The upstream blackbox test checks depths 4, 5, 8, and 32, with two or four spaces
per level. Lists end at EOF, before a paragraph, or before an outer sibling.
The test checks error state, source coverage, item count, and nesting depth.
On diabolocom, the new cases fail against v0.52.0; the patched Markdown suite
passes. These are correctness tests, not measurements of the CLI resource budget.

Unswell's [blackbox tests](../../extract/markdown_nested_test.go) also check uneven
indentation, CRLF, BOM, Unicode, inline formatting, entity mapping, and code
exclusions. Excessive nesting and the configured block limit still fail.
The [root e2e case](../../e2e/testdata/markdown_nested_lists) checks expected
findings and exclusions through the compiled CLI, including JSON/SARIF parity.

The unchanged FastAPI selection now completes with 702 documents, 122,705 prose
words, 182 findings, and no operational errors. The previously failing document
adds 1,524 words and four findings. All 701 previously extracted source hashes,
policy hashes, block counts, and word counts remain unchanged. The command exits
1 because two paragraphs in `fastapi/security/oauth2.py` cross the configured
quality threshold. All five report formats are produced. The pull request records
the exact committed binary and its replay results.

The published-release performance claim in #219 remains open until a release
containing the parser repairs completes the repeated corpus measurements.
