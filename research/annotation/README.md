# Annotation protocol tools

For source acquisition, grouped splits, and reproducible unlabeled units, use the
[corpus preparation tool](corpus/README.md). It feeds this protocol and assigns no
human responses or quality labels.

This developer module implements the format and measurement portion of
[#21](https://github.com/stokaro/unswell/issues/21). Read the
[editorial rubric](../../docs/editorial-annotation.md) before preparing a round.
It supplies no trained model, human-labeled corpus, or automatic editorial verdict.

The shipped CLI and MCP server do not import this module. It uses Go and local
JSON only. Loading a round does not read references, execute source files, download
resources, invoke another program, or evaluate prose through a second engine.

## Run the commands

From this directory, using the repository's supported Go compiler:

```sh
mkdir -p ../../artifacts/annotation
go run ./cmd/annotate validate < testdata/tutorial.json
go run ./cmd/annotate packet < testdata/tutorial.json > ../../artifacts/annotation/packet.json
go run ./cmd/annotate agreement < testdata/tutorial.json > ../../artifacts/annotation/agreement.json
go run ./cmd/annotate decisions < testdata/tutorial.json > ../../artifacts/annotation/decisions.json
```

`validate` reports structural validity, not corpus acceptance. `packet` emits the
reviewer's target/context view. `agreement` measures original primary responses.
`decisions` exports selected labels and unresolved states for later corpus work.
Successful commands exit 0; input, permission, and output errors exit 2.
Interrupting a command exits 130. Output goes to stdout and errors to stderr.
Redirect output to a new file and publish it only after the command succeeds.

Use a real round file for annotation. Keep the administrative input separate from
the packet. It contains source references, provenance, permissions, actor
declarations, and answers that annotators must not see during independent review.

## Versioned format

[schema.json](schema.json) defines the strict JSON shape. Unknown properties, bad
enums, missing fields, and duplicate keys are rejected. Semantic checks enforce
record references, one response per rater/unit, profile hashes, source-range
bounds, context adequacy, and the ordering of adjudication after original answers.

| Record | Purpose |
| --- | --- |
| `version`, `rubric`, `purpose`, `round_id` | Contract identity and explicit tutorial, pilot, or corpus use |
| `profile` | ID, exact instructions, and SHA-256 of their UTF-8 bytes |
| `units` | Opaque ID, kind, role, target, context, source segments, extraction identity, provenance, and rights |
| `actors` | Opaque ID, human/assistant/simulation declaration, and rater/adjudicator role |
| `judgments` | Original label, categories, context adequacy, explanation, and RFC 3339 submission time |
| `adjudications` | Later final label, categories, reviewers, explanation, and time; original answers remain present |

Unit IDs follow `u000001`; actor IDs follow `a001`. Assign them without encoding
origin, names, models, or quality. Keep the private identity key outside the packet.
All primary raters in a round are assigned every unit. Omit missing responses
instead of inserting an empty label. Use another round for a different assignment.
String roles distinguish error messages, log messages, UI text, and other strings;
the generic `string` role means the subtype is unknown. This retains the corpus
sampling and analysis distinctions required by #22.

Start a round with empty `judgments` and `adjudications` arrays. `packet` supplies
the `sha256` that each response and adjudication must echo as `packet_sha256`.
Changing a target, context, kind, role, purpose, rubric, or profile invalidates old
responses. Administrative provenance changes do not alter the packet. The digest
covers compact Go JSON of the packet with `sha256` omitted and units sorted by ID;
field order and encoding belong to packet version 1. Consumers should echo the
supplied hash, not invent their own JSON canonicalization. It binds a response to
the shown task, not to a verified human identity.

`tutorial` accepts simulated primary raters and rejects human claims. `pilot` and
`corpus` require at least two declared human raters and reject simulated actors.
Assistant ratings are auxiliary and never enter primary agreement. A declaration
is not proof that independent humans participated; collection audit is required.

The schema caps input at 10,000 units, 32 actors, 320,000 judgments, and 10,000
adjudications. The loader limits JSON to 32 MiB and nesting to 32 levels, rejects
invalid UTF-8 and unpaired surrogate escapes, and checks cancellation. Input bounds
do not define statistically reliable fragment lengths.

Original source segments reuse `document.Span`. The loader checks ordering and
bounds against the recorded byte length. It does **not** fetch the referenced
source or verify that an asserted source hash and extraction map are true. Corpus
acquisition must perform those checks using the existing engine and retain the
evidence. The tutorial's actual source files and hashes are verified in tests.

Version changes are explicit. This format does not change `unswell.RunResult`,
product configuration, model compatibility, scoring, or CI gate semantics.
Source/provenance fields never become editorial model inputs by implication.

## Agreement statistics

We implement nominal alpha from Klaus Krippendorff's
[computational note](https://www.asc.upenn.edu/sites/default/files/2021-03/Computing%20Krippendorff%27s%20Alpha-Reliability.pdf),
dated January 25, 2011 with references updated September 13, 2013. Its examples
A-C provide numerical reference tests; no source code is copied.

For a unit with `m >= 2` ratings and `n[c]` occurrences of category `c`, accumulate
`(m*(m-1) - sum(n[c]*(n[c]-1))) / (m-1)` observed disagreement. Let `N` be the
total pairable ratings and `N[c]` their category counts. Alpha is:

```text
1 - observed_disagreement * (N - 1) / sum(N[c] * (N - N[c]))
```

The denominator uses only pairable ratings. A singleton does not change its
marginals. Quality labels are nominal, including `uncertain`; their order conveys
no distance. Negative values are retained, and zero denominators yield null with
a reason. Tests cover the published binary, five-category, and missing-data
examples, plus degenerate cases.

Raw pair agreement counts matching ordered pairs divided by all ordered pairs.
It has different weighting from alpha when units have different rating counts.
The uncertain share uses all primary responses, including singletons. Each reason
category is a separate present/absent comparison excluding uncertain responses.
Output retains counts, exclusions, unit kinds, and roles. Original labels alone
enter these calculations; adjudication and provenance do not change them.

These point estimates do not include sampling intervals. The
[rubric's uncertainty protocol](../../docs/editorial-annotation.md#agreement-and-uncertainty)
requires grouped intervals for corpus publication. No universal acceptance
threshold is built into this command.

## Editorial decision export

`Round.Decisions(ctx)` and `annotate decisions` produce
`unswell-editorial-decisions-v1`. Every assigned primary rater must answer a unit
before a label can be selected. Missing answers produce `missing_judgments`, even
when an adjudication based on two other responses is recorded. Auxiliary assistant
responses are counted separately and cannot fill that gap.

Once primary responses are complete, a recorded adjudication takes precedence.
Otherwise the quality label and the set of reason categories must agree.
Disagreement produces `judgment_disagreement` and a null label; no majority vote
or union of categories is inferred. A selected `uncertain` retains that label and
an unresolved status. Acceptable or revision-required selections have `resolved`
status and record the `unanimous` or `adjudication` basis.

The export binds the exact input round bytes, frozen packet, rubric, and profile.
It retains target/context hashes, original source segments, source identity,
extraction policy, and recorded allowed uses. It omits prose, rationales, author
identities, and origin claims. Archive the original round for individual answers,
timestamps, and explanations. Reformatting its JSON changes `round_sha256`, even
when the selected decisions remain identical. The export's own hash covers compact
Go JSON with its `sha256` field omitted.

Tutorial output has `basis: simulation`; pilot/corpus output records
`basis: declared_human`. Both retain `human_corpus: not_qualified`. These records
alone cannot authenticate participants, verify source extraction or permissions,
establish independent splits, or qualify a training corpus. A resolved label does
not override the recorded allowed uses. The later corpus/feature join must verify
the actual target and context; block measurements cannot be assigned to a shorter
sentence merely because their ranges overlap.

The library performs no I/O and returns detached values. Cancellation or failure
returns no partial export. Compact output is limited to 64 MiB. See
[ADR 0019](../../docs/adr/0019-editorial-decision-export.md) for the contract and
remaining training/evaluation requirements.

## Teaching fixtures

`testdata/tutorial.json` contains 17 teaching units, with source files under
`testdata/sources/`. Seven pairs illustrate the rubric categories. Other units
exercise insufficient context and the two reported noun-stack false positives.
An assistant prepared the teaching text. The reused project comments
have unknown unit-level origin. Scripted rater and adjudication records demonstrate
the format; their timestamps are illustrative, not submission evidence.

The assistant's exact model revision and generation parameters were not recorded.
The task context was the maintainer's annotation-protocol request; the seven pairs
were written as teaching material and encoded by a local fixture-building script.
No human reviewer participated. The two reused comments have pinned source links
in their origin-evidence records. Preserve these limitations when reusing the data.

These fixtures are not the human-labeled corpus required by #22. They do not
fulfill the pilot, 5,000-unit minimum, independent-rater requirement, or final test.
Permission to reuse an example does not change its status as a teaching fixture.

Root [e2e tests](../../e2e/annotation_test.go) build the command with cgo disabled,
verify the blinded packet and decision export against goldens, and check the
agreement summary and rejection exits. Module tests cover the schema, provenance boundaries,
missingness, source references, cancellation, writer/read failures, and published
numeric examples. Ordinary CI tests, lint, race, fuzz, and coverage include this
module through `.gomodules`; no separate runtime exemption is added.
