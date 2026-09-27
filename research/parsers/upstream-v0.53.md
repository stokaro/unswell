# Return to the upstream parser

Issue [#234](https://github.com/stokaro/unswell/issues/234) replaces the temporary
`stokaro/gotreesitter` fork with `github.com/odvcencio/gotreesitter` v0.53.0.
This is the first upstream release after the required Markdown repair merged.
The newer v0.55.1 release is outside this migration.

## Dependency evidence

[Upstream PR #1109](https://github.com/odvcencio/gotreesitter/pull/1109) merged as
`0b2d728d210b814c5b6c8aa6639f7ec4e1db6360` on September 16, 2026. The v0.53.0
module includes `grammars/markdown_nested_lists_test.go` and its
`TestMarkdownNestedListClosures` regression. The
[v0.52.0-to-v0.53.0 comparison](https://github.com/odvcencio/gotreesitter/compare/v0.52.0...v0.53.0)
contains 142 commits, so this update includes changes beyond the original fix.
The release also changes lexical dependency checks for incremental parsing and
caches scanner probes. The Bash change below is another upstream repair carried
by this update.

The upstream MIT license is byte-identical to the retained license, SHA-256
`b174fbe1e1cffafb096528de3e8361c90a0d97e7e3f1a32061aae51e983a7cfc`.
All runtime and consumer modules use the upstream path. New corpus manifests
record that path and version; historical corpus records retain their original
fork identity.

## Observable grammar changes

The upstream Bash parser now accepts the empty assignment in `! A=x B= command`.
The shared-grammar regression checks successful parsing before and after the
Unswell adapter, while retaining its tree-identity check.

The upstream inline Markdown grammar treats comments containing double hyphens
as literal prose instead of returning an error tree. Unswell's reason separator
uses those hyphens. A bounded trial parse therefore checks whether normalizing
only a candidate directive's hyphens produces an HTML-comment boundary at the
same byte range. Only confirmed boundaries are used for the final parse.
Quoted attributes, escaped markup, code examples, and images cannot supply
active directives. The original source is unchanged. A scan considers at most
1,000 candidates per inline input and fails explicitly above that limit.

Ordinary non-directive text with double hyphens follows the upstream literal
interpretation. It is not normalized into a comment or silently discarded.

## Pinned corpus replay

The matched replay uses the rule catalog from main commit
`bef6148d8ab2c2deed054756ad6ad19f414fdada`, including adjacent-word rule version 3.
Both builds analyzed FastAPI commit
`77f7447ee8917697e702b7eca8d393e11dc1e365`, using the existing strict profile and
unchanged selection from the performance corpus: default exclusions plus
`docs/ja/**` and `docs/zh/**`.

| Result | Fork build | Upstream build |
| --- | ---: | ---: |
| Documents | 702 | 702 |
| Prose words | 122,532 | 122,532 |
| Findings | 167 | 167 |
| Operational errors | 0 | 0 |
| Exit code | 1 | 1 |

Both scans are complete. Their document identities, source ranges, exclusions,
block and sentence counts, findings, and gate outcomes match. All five report
formats were produced. The protected-code, nested-list, BOM, CRLF, Unicode,
and directive extraction regressions also pass.

These are source-build correctness observations with the current rule catalog.
They do not replace the historical counts from #231 or establish a new
published-release speed, memory, or editorial-recall result.
