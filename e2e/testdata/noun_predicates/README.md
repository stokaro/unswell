# Predicate boundaries in noun-stack candidates

The four paragraphs come from the saved version 3 Luna/Haiku outputs for fzf
and ripgrep. `provenance.json` identifies the complete source and paragraph
hashes. The same paragraphs are also checked inside Go comments.

The previous rule included `exists`, `resembles`, or `fix` in the reported noun
sequence. These cases now have no finding. The positive controls retain `fix`
as a noun modifier and `fixes` as a coordinated noun head. Inline and fenced
code remain protected. The harness checks BOM/CRLF input, Unicode coordinates,
JSON/SARIF spans, and the expected policy failure on the three positive cases.

The retained source-project licenses accompany the fixture. These generated
paragraphs are exposed regression inputs, not independent confirmation or
human quality labels. The archived version 3 measurements remain unchanged.

## Complete-document check

`comparison.json` records a separate noun-stack-only scan of the three complete
saved source files. The version 4 binary reports four findings (1, 1, and 2);
version 5 reports zero on the same SHA-256-verified bytes. The records retain all
old diagnostic spans and snippets. The normal E2E run stays compact and checks
noun-positive controls separately.

The source files can be unpacked from the study measurement archive under
`analysis/work/`. Build version 4 from commit
`966ffd841bcfa3dbbed2bc706ae742c18c2b569d` and version 5 from this change. Run each
binary with the policy below, an explicit `--project-root` containing the unpacked
files and policy, `--no-gate`, and `--report json:result.json`. Check the three
source paths listed in `comparison.json` individually.

```yaml
version: 1
extends: [builtin:custom]
rules:
  syntax.noun-stack:
    enabled: true
    gate: none
    score: {weight: 0, cap: 0}
```
