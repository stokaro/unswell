# Recognizing rhetorical constructions

Issue [#294](https://github.com/stokaro/unswell/issues/294) starts from a product
gap: the deployed engine could recognize `rather than`, but missed three complete
denial/redefinition pairs with different nouns. Lowering a score threshold cannot
recover a construction that produces no evidence.

Two rules now expose complete structures through the existing engine:

- `syntax.repeated-reframing`: groups nominal denial/redefinition pairs and points
  to both clauses of each pair. One ordinary contrast remains below the allowance.
- `filler.document-metadiscourse`: recognizes document self-description and
  announcements by their subject/verb structure, without storing full phrases.

Both are experimental notes in technical and strict, with zero weight, zero cap,
and `gate: none`. Recognition is separate from deciding whether to edit. Useful
technical definitions and navigation may use these structures intentionally.

## Inputs and before/after evidence

The baseline is Unswell `bcb2b6a7a38290b2e52e58102daf9a6a7115b21c`, also served by
play.unswell.dev during the reproduction. Its downloaded WASM had SHA-256
`c48572d8b5fddbc85ec89b84e71a091c7779dc4343a3ae86bb73e6481dfffbb0`.
The deployed runtime was exercised through its JavaScript host interface in Node;
this was not a browser UI test. The after measurements use the implementation and
fixtures in this change. Counts below describe these inputs only.

The [fixture provenance](../../e2e/testdata/rhetorical_frames/README.md) pins the
complete Ptah page to a commit and source hash. Constructed probes are declared
separately; neither set is a human-labeled qualification corpus.

| Input | Before | After | Evidence and editorial decision |
| --- | --- | --- | --- |
| Constructed paragraph: “Backups are not a checkbox. They are your last line of defense. Monitoring is not a dashboard. It is the foundation of operational confidence. Testing is not a phase. It is a commitment to quality.” | 0 findings in technical and strict | 1 grouped note, 3 pairs, 6 clause locations | Repeated denial/redefinition, despite different nouns. Review the three binary setups and support the affirmative claims with specific properties. |
| Complete Ptah database URLs page | 5 technical / 7 strict findings | 6 technical / 8 strict findings | The additional note spans “This page defines all four.” It identifies document self-description; the following navigation can justify keeping it. Existing safety conditions and URL literals remain intact. |
| Constructed technical definition: “A dev database is not the target. It is a disposable replay target.” | 0 findings | 0 findings | One ordinary distinction is allowed. Keep the distinction. |
| Two constructed dev/shadow definitions with the same shape | 0 findings | 1 unscored note; gate passes | The shape repeats, but the definitions express different roles. Keep their substance; the rule makes no semantic redundancy claim. |
| Constructed factual control: “Backups retain database contents. Monitoring reports replication lag. Testing checks retry behavior.” | 0 findings | 0 findings | Direct factual sentences lack the frame. This is a control, not a meaning-preserving rewrite of the metaphorical probe. |
| Direct navigation: “See the configuration reference for the complete list of environment variables.” | 0 findings | 0 findings | A link instruction is not a document self-description. |

The Ptah page still passes both profiles with maximum index 8. The change adds
structural evidence, not points selected to fail this page. It does not establish
that the rest of the page contains no formulaic prose.

A minimal structural revision of the first probe is: “Backups are your last line
of defense. Monitoring is the foundation of operational confidence. Testing is a
commitment to quality.” The denial setups disappear and the three affirmative
claims remain. The new rule no longer fires. Those claims are still vague; removal
of a construction is not proof of improved prose or preserved technical meaning.
No automated rewrite is applied to the Ptah document.

## Scope and limits

The implementation reuses source-preserving extraction, NLP tokens, terminology
exemptions, bounded prose runs, evidence locations, and the normal result/report
pipeline. Its internal frame holds clause token ranges. There is no parallel
extractor, model, network request, or general-purpose semantic classifier.

Copular number and tense agree across a pair. The second subject repeats the
first or uses an explicit anaphoric pronoun. Nominal complements use token/POS
cues, not dependency claims. Document self-description uses document nouns and
communicative verbs, including declared inflections. This vocabulary describes
roles rather than complete example sentences. The
[rule reference](../editorial-patterns.md#construction-frames) specifies bounds,
protected operands, and context exclusions.

These recognizers cover two construction families. They do not recognize every
rhetorical move, resolve arbitrary pronouns, establish empty meaning, or determine
AI authorship. The existing marker-density rule remains a separate signal;
about-pairs can satisfy both it and the more specific reframing rule. Their
shared correlation group and zero default weights avoid adding risk twice.

## Verification

Run the blackbox library and catalog tests:

```sh
go test . ./builtin -run 'Test(RhetoricalFrames|DocumentMetadiscourse|Frame)' -count=1
```

Run the compiled CLI fixture with exact diagnostic and SARIF locations:

```sh
go test ./e2e -run 'TestCLI/rhetorical_frames' -count=1
```

The cases include contractions, inflections, punctuation, Unicode, BOM/CRLF,
Markdown emphasis, opaque inline code, headings, fences, omitted comments, lists,
independent blocks, questions, negation, prohibitions, and window limits. Catalog
checks cover zero default score, observer failures, cancellation, and candidate
budget abstention. A default-profile test ensures the new notes are actually
visible without enabling experimental rules manually.

The playground consumes the same rule registry. Its WASM regression must check
the grouped pairs, complete source locations, Ptah self-description, negative
controls, and passing gates through the existing host API. A local WASM test and
a deployed-site update are separate pieces of evidence; neither is implied by
native tests alone.
