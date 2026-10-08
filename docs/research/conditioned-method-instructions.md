# Conditional action and method introductions

The existing `filler.instruction-scaffolding` rule recognizes an action
introduced by a temporal prerequisite and an impersonal safety or possibility
statement, followed immediately by an anaphoric passive invocation of a named
method or function. It asks the editor to combine the action and method under
the original prerequisite. Advice retains the safety qualifier, prior state,
actors, operands and limits. A capability remains a capability.

For example, a client-ready condition can introduce fetching records, followed
by a separate sentence announcing that the named fetch method is called. Both
sentences are source evidence for one construction. Neither a safety statement
alone nor a passive method description alone establishes this warning.

Each complete sentence must fit the existing 48-token candidate bound. The
prerequisite names a state after `After`, `Once` or `When`; the method sentence
starts with `To do this` or `To do that`. A named method or function and its
explicit invocation establish the relation. Negation, permission restrictions,
attribution, quotations, unnamed operations and automatic execution remain
excluded. An intervening sentence, heading, code block or other structural
boundary breaks the relation. Adjacent prose paragraphs can retain it.

## Measured change

The paired replay used all 36 original pages with their unchanged source bytes,
formats, 804 judgments and 123 reference events. The baseline is Unswell commit
`6b5c28ce7ccc043a9833aad85e8b9c38b7296b8f`.

| Complete original source set | Before | After | Added / removed criticisms |
| --- | ---: | ---: | ---: |
| Technical profile | 145 | 146 | 1 / 0 |
| Strict profile | 166 | 167 | 1 / 0 |
| Complete previously missed construction | 0 | 1 | One development event |

The recovered BuildKit example states a loaded-vertex prerequisite, a safe
result request, and the named scheduler method used for that request. The
diagnostic covers both complete sentences. Its advice preserves the loaded
job, previously loaded target vertex, edge, method signature and scheduler
association. Combining the two sentences with a calling-method complement
retains those facts and removes the separate method announcement. The root
Codex assistant accepted the diagnosis and advice under ADR 0041. This is
assistant review. Both profiles observe the same development event.

All preceding criticisms retain their messages, advice, source locations,
related evidence, metrics, policy and suppression fields. Rule version and
dependent identity hashes change. All scans completed without operational
errors or rule abstentions.

Six separate complete Ptah pages were selected by API, query and application
safety topics before candidate implementation or output inspection. Sources
are pinned to Ptah commit `a641f2dd20a9c8596629f43fae0cc2f811d7f5e0`:
public API, query builder, direct apply, versioned apply, migration generation,
and test-case reference. Source-hash duplicates of the original and preceding
control inputs were excluded. Their 154,614 source bytes retain all 51
technical and 57 strict findings. No new positive confirmation was observed.
Project-wide exposure history is unestablished; these are separate controls.

Public-engine blackbox tests cover five positive variations, technical controls,
structural boundaries, Markdown, Go and Python source mapping, opaque method
names, term exemptions, occurrence allowance and candidate bounds. CLI tests
verify both profiles, the preserving revision and related source locations.

## Limits

This bounded extension recovers one exposed development event. It does not
establish population precision, unseen positive recall or achievement of the
contextual quality goal. No model requests, probability calibration or
authorship claims are involved. Severity, score weight, gate policy and
ordinary offline execution remain unchanged. The separate model pilot still
fails its quality criteria. Remaining contextual misses and unnecessary
long-document remarks are tracked in
[#349](https://github.com/stokaro/unswell/issues/349).
