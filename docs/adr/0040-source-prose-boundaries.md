# Source prose boundaries

Status: accepted for implementation in #237.

Go comments use the block grammar documented in [Go Doc Comments](https://go.dev/doc/comment).
The default `extraction.go_comments: godoc` applies to comment prose, including
ordinary comments. `plain` retains the previous comment extraction behavior.
The adapter preserves byte positions while recognizing paragraphs, headings,
list items, and indented examples. Its block scanner follows Go 1.25.0,
commit `6e676ab2b809d46623acb5988248d95d1eb7939c`, under the Go BSD license.
It does not interpret Go comments as Markdown. Go's list grammar does not
support nested lists; nesting is available in explicitly selected Markdown.

`extraction.markdown_strings` selects static string literals by project-relative
paths, optional source formats, and optional enclosing symbols. Every nonempty
selector must match. Values within each selector are alternatives. Matching
strings use the existing Markdown grammar after language-specific decoding.
Unselected strings retain their current interpretation. Selection does not
override disabled contexts, extraction exceptions, technical-literal exclusions,
or GitHub Actions shell routing. A selected string containing protected dynamic
interpolation or control characters fails explicitly; it cannot be parsed as a
complete static Markdown document.

Inner Markdown contexts are selected using the Markdown context set, after the
outer source `string` context. Blocks retain the outer `comment` or `string`
kind, and structure records their inner kind. Existing source mapping composes
the decoded text and grammar positions. Code examples remain exclusions;
prose beside examples remains eligible. Each paragraph or list-item paragraph
becomes an independent analysis block. Ordinary Go URLs retain their existing
protection; Markdown inline protection uses the existing Markdown extractor.

These settings participate in policy and preparation identities. Corpus plans
freeze the default Go mode. Saved results and model bindings from earlier
extraction policies require explicit remeasurement; they are not rebound by
changing a stored hash. Report schemas and the NLP pipeline remain shared.

Acceptance includes blackbox extraction tests, root CLI golden cases, original
Ptah source replay, configuration rejection tests, and the existing repository
checks. Race, active fuzzing, and coverage remain deferred to #123.
