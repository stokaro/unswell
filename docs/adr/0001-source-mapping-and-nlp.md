# ADR 0001: own source maps and adapt the English backend

Status: accepted for the experimental alpha.

Unswell needs original UTF-8 byte ranges after removing Markdown syntax, comment
markers and string delimiters. The public document model must also work with another NLP
provider without exposing vendor types.

## Extraction decision

Use gotreesitter's block and inline grammars for Markdown/GFM and its source
grammars for comments and strings. This replaces the initial Goldmark adapter.
Own the normalized-text to original-byte map in Unswell. Each byte maps to a rune or
escaped entity; a finding may therefore contain several discontiguous segments.
Inline code, URLs and other protected atoms insert a boundary marker with spacing.
They do not join surrounding words or count as prose in density denominators.

Use Go's parser and AST for comment selection. Read the selected comment bytes
from the original source. `ast.CommentGroup.Text()` is unsuitable because it
normalizes comment markers, directives and whitespace. Even `ast.Comment.End()`
can undercount CRLF bytes because the AST comment text omits carriage returns;
the extractor finds the original terminator from the AST's starting offset.

Use grammar DFAs with their attached scanners. The registry's C-family token
factory misclassifies C++ raw strings; the DFA regression retains the literal's
content and exact range. Strict parsing, error-node checks and a Python orphaned
string-delimiter check reject known partial-tree cases. The adapter bounds parse
time and nesting and propagates context cancellation.

An alternative of stripping Markdown markers or using normalized Go comment text
would be smaller but loses entity, emphasis and CRLF coordinates. The extraction
tests retain exact expected original ranges rather than searching the normalized
text back into the source. This also handles repeated words without ambiguous
reverse lookup.

## NLP decision and comparison

Use Prose v3.2.1's separate `tokenize`, `segment` and `tag` packages. The root
Prose package would load an unnecessary entity model. Add an explicitly shallow
NP/VP/PP tag-run chunker. Do not advertise lemmas or dependency parsing.

`TestSegmentationBackendComparison` runs both Prose's segmenter and the generic
neurosnap Punkt tokenizer with the same bundled English training data. On the
four retained abbreviation, version, quotation and Unicode cases, both returned
the expected two sentences. Prose also passed exact byte-slice assertions. This
small comparison establishes adapter compatibility, not superior accuracy.

Choose Prose for the integrated offset-aware token and POS interfaces, immutable
shared model loading and its sentence-specific customizations. A direct neurosnap
adapter remains viable if the Prose boundary contract changes. The provider
interface keeps that replacement separate from rules and reports.

Reproduce with:

```sh
go test ./nlp/english -run TestSegmentationBackendComparison -v
go test ./nlp/english -run '^$' -bench BenchmarkProseSegmentation -benchmem
go test ./extract
```

One development measurement on an Apple M3 Pro, macOS arm64, Go 1.27.0 measured
5,298 ns/op, 1,595 B/op and 56 allocations for the retained short segmentation
benchmark. This is a component measurement, not the full-document performance
acceptance test or a cross-platform guarantee.

## Limits

POS tags and chunks are statistical or shallow features. They do not establish
grammatical correctness. Domain terms, abbreviations, markup edge cases and long
technical documents need broader evaluation. Model identities and upstream
distribution notices are recorded in `THIRD_PARTY_NOTICES.md` and the result
manifest. Editorial probability remains unavailable without an independently
validated revision model.
