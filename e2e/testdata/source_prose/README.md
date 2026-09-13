# Source prose boundaries

The negative cases retain the function-attribute comment and Cobra `Long` help
text from Ptah commit `7b47e7cfb5d4ff32a38375345069f5533bff892f`:

- [Function attributes](https://github.com/stokaro/ptah/blob/7b47e7cfb5d4ff32a38375345069f5533bff892f/core/schemamodel/types.go#L947)
- [Verification help](https://github.com/stokaro/ptah/blob/7b47e7cfb5d4ff32a38375345069f5533bff892f/internal/cli/oci/verify.go#L31)

The surrounding declarations are reduced to valid Go fixtures. Added long prose
has explicit `want` annotations. The comment list must not trigger
`syntax.long-sentence`; the separate help paragraphs and indented YAML example
must not trigger `readability.long-paragraph`. The added prose must trigger both
configured checks in their respective files. The runner uses CRLF and BOM and
verifies JSON and SARIF locations against the unchanged fixture bytes.

The help selector is explicit. Its paragraphs are independent analysis blocks;
the YAML example remains excluded. The cases exercise extraction and diagnostics,
without assigning human editorial or authorship labels to the source.
