# Ptah rhetoric cases

`ptah.json` contains 21 contiguous excerpts from `stokaro/ptah` at
`654eae5591392278e6c8bce8e54737f780766f19`. Each entry binds the full source
file's SHA-256, an original UTF-8 byte range, the excerpt's SHA-256, and its text.
The source license is retained in `LICENSE.ptah`.

Nine entries have exact expected diagnostics and proposed revisions. Twelve
controls preserve useful explanations with similar vocabulary. These are
assistant-authored development judgments, exposed while designing the rules.
They are not independent human labels, a test of authorship, or a precision
estimate. The revisions are examples for review; Unswell does not apply them.

Read an entry's `rationale`, then compare `text` and `revision`. A useful revision
must preserve the failure conditions, commands, precedence, and safety constraints,
not just remove the match. The retained controls cover lease fencing, exit-status
verification, byte-order ambiguity, batch boundaries, privacy, and absent state.

`go test ./e2e -run '^TestPtahRhetoric$' -count=1` builds the real CLI and checks
both shipped profiles on every original and revision. Expected diagnostics name
the complete clause and its byte range. All four new rules must stay silent on
the controls and revisions; other rules remain enabled and may report independent
concerns. Runs must complete. No network or Ptah checkout is needed.

Synthetic grammar, exclusion, mapping, and scope tests live in
`../../local_rhetoric_test.go`. They exercise vocabulary changes, tense, plurals,
negation, questions, opaque code, lists, comments, strings, BOM/CRLF, Unicode,
emphasis, term exemptions, and block boundaries.
