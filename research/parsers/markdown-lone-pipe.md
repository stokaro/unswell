# Markdown lone-pipe boundaries

Issue [#69](https://github.com/stokaro/unswell/issues/69) describes lost prose after
a line containing one pipe at the end of a Markdown table. The pinned Go parser
can return anonymous delimiters without the intervening words, despite reporting
no error. Unswell previously rejected that tree as incomplete.

## Boundary comparison

The [recorded comparison](markdown-lone-pipe.json) contains 13 structural probes.
They cover a lone pipe, surrounding spaces and tabs, multiple-pipe empty rows,
blank lines, an ordinary one-cell row, quotes, nested quotes, lists, CRLF, and EOF.
These inputs test parsing and source mapping; they have no editorial or authorship
labels. GitHub rendered the same inputs through its Markdown API in GFM mode on
September 9, 2026. The saved HTML is the observed reference for block boundaries.

GitHub ends a table at a lone pipe. The following line belongs to a paragraph,
including inside quotes and lists. A tab-indented pipe may form a protected code
block before that paragraph. Rows such as `||`, `|||`, and `| | |` remain table
rows. Ordinary text without a pipe can also remain a one-cell data row.

The comparison checks gotreesitter v0.52.0 and upstream revision
`851945aa6fec19e1808acaf181334084127ec6be`. Both lose the following prose in the
minimal case. The C grammar at gotreesitter's locked revision
`f969cd3ae3f9fbd4e43205431d0ae286014c05b5` also mishandles this boundary: it retains
the following line as a table row, sometimes with an error node. C compatibility
alone would therefore preserve the wrong block context.

The C reference used Python tree-sitter 0.25.2 in an isolated research environment.
Neither C nor Python is needed by the delivered runtime or normal tests.

## Scanner repair

The block grammar offers both an ordinary line ending and a table continuation
at the affected boundary. Unswell checks whether the next line contains only one
pipe after indentation and quote prefixes. When it does, the adapter removes the
table-continuation option and delegates to the original scanner. The existing
grammar then parses the following block normally.

The adapter owns a separately decoded copy of the embedded grammar and resolves
external token IDs from that grammar. It does not change shared cached grammars,
input bytes, inline parsing, or the existing empty-row normalization. Missing or
orphaned nodes remain errors. There is no raw-text extraction fallback.

Lookahead uses a value copy of the pinned external lexer, whose source slice is
read-only and whose cursor state consists of scalar fields. Incremental reuse is
disabled because this probe does not update the original lookahead frontier.
Unswell uses full parses. Scanner checkpoints and the original error-reuse policy
remain available through the embedded scanner.

The research scanner runs on inputs with Unswell's leading paragraph and final
newline framing. It does not apply empty-row markers. Its multiple-pipe controls
therefore retain the separate upstream empty-row limitation. The public extractor
comparison uses the complete production path, including that existing repair.

## Product evidence

Blackbox tests check table-cell and paragraph selection, quote exclusion, lists,
Unicode byte spans, LF/CRLF, EOF, protected code, and unchanged shared grammars.
The root CLI fixtures require the following paragraph to produce exact findings
before replacing the old operational-error expectation with a complete scan.
Their goldens include exclusion ranges, line and column positions, and exit codes.

One `*_internal_test.go` file supplies an orphaned token from the original grammar
directly to the private visitor. Its explanation follows `package`: public
extraction applies the repair before the visitor and cannot inject that old tree.
This test preserves rejection of incomplete parser output.

Run the product regressions without network access after dependencies are present:

```sh
go test -count=1 ./extract ./e2e
```

The JSON records source and output hashes, parser error flags, target block kinds,
and the original GitHub HTML. Parser projections include named node kinds, UTF-8
byte ranges, missing flags, and ordered named children. They do not compare field
labels or anonymous tokens. Local research retains the raw projections and the
compiled C reference. The comparison does not claim general GFM conformance.
