# Editing opportunities in Ptah documentation

The published playground's rules passed all 136 Ptah pages in the September 16
snapshot. That result did not establish that the wording needed no edits. Most
findings described sentence length, paired contrasts, or punctuation. The new
local constructions identify nine additional clauses on seven pages, with a
specific editing suggestion for each.

## What changed

| Rule | Example from Ptah | Review action |
| --- | --- | --- |
| `filler.document-justification` | `this page exists so the operator is reachable from here` | State the destination or scope directly |
| `filler.evaluative-closure` | `which is the point of scoping it at all` | Keep the consequence; remove a purpose restatement if it adds nothing |
| `repetition.definition-echo` | `The default is a default, not a fallback` | Explain when the default applies and when an explicit input fails |
| `syntax.slogan-contrast` | `So Ptah checks, rather than trusts` | State the actual check and its failure condition |

The evaluative-closure rule also recognizes deictic judgments such as
`which is the honest answer, not a gap`. A refusal to roll back a drifted
generation is useful behavior to document. Calling that refusal honest adds no
condition, evidence, or recovery instruction in the reviewed passages.

These match bounded constructions, with their operands and clause ranges.
An isolated `rather than` is insufficient. Concrete alternatives such as checking
a fencing token instead of trusting a previously read lease remain unflagged by
these rules. A purpose clause naming the `EXEC` wrapper also remains unflagged:
it explains how SQL Server's separate-batch requirement is satisfied.

The rules run in the existing engine after source-preserving extraction and NLP.
They share its observations, exemptions, candidate limits, cancellation, scoring,
reports, and MCP result. They do not add a model, external service, second parser,
or automatic rewrite. Paragraphs, list items, comments, and strings are supported;
headings, table cells, excluded regions, and questions are outside these matchers.

## Defaults and limits

Both profiles enable the four rules as warnings, with weight 12 and cap 24 per
rule. One occurrence reaches full activation; occurrences are grouped within an
eight-sentence window in a single block. The parameters `window_sentences`,
`allowed_occurrences`, and `saturation_occurrences` remain configurable. Clause
length is bounded at 48 tokens. Protected code can supply context but never a
construction keyword. Definition echoes exclude protected nominal operands.

The warnings suggest a review, not an unconditional deletion. Their weights are
alpha editorial policy proposals. No independent human precision, authorship
association, or improvement in reader comprehension has been established.
The rules have no individual prohibition. Existing index thresholds are unchanged.

Consequently, all 136 pages still pass both default gates. A warning and a failed
gate answer different questions: the warning identifies a particular review
opportunity; the gate applies the configured local thresholds and prohibitions.
Forcing a page to fail would not demonstrate wider pattern recognition.

## Complete-page measurement

Ptah is pinned to `654eae5591392278e6c8bce8e54737f780766f19`. Every Markdown and
MDX page under `docs/site/src/content/docs` is included: 103 Markdown pages and
33 MDX pages. MDX is parsed as MDX, not passed to the playground's Markdown parser.
The reference engine is `fd88e26b7afa8fd8e7543502d87df24914aeaba8`, whose rule
implementation was merged as `17971f5a54e10928dbe507eac23f0b87dbf44bc3`.

| Profile | Before | After | New local warnings |
| --- | ---: | ---: | ---: |
| technical | 772 | 781 | 9 |
| strict | 930 | 939 | 9 |

All scans complete. Existing finding identities and locations remain identical.
The new counts are two document justifications, five evaluative closures, one
definition echo, and one bare slogan contrast. The original 136-page pass result
was reproduced with the native CLI; the published WASM had separately checked
the 103 Markdown pages. These native measurements do not certify deployment of
the new rules to the playground.

The historical development comparison reuses the previously pinned 27 pages,
31,488 prose words, and 17 provenance groups. All pages complete and the four
rules produce zero findings in both profiles. Historical provenance does not
make every passage editorially clean; this is an observed review load, not a
false-positive estimate.

## Inspecting the evidence

The [casebook](../../e2e/rhetoricdata/README.md) preserves nine positive excerpts,
twelve useful technical controls, and nine proposed revisions. It records exact
source bytes, hashes, expected clause spans, and the reason for each judgment.
The CLI checks every original, revision, and control in both profiles. The
originals were used during rule development; they are not a held-out success rate.
The revisions retain conditions and operational details. They are authored review
examples, not proof from independent annotators.

The [run record](../../research/reviews/2026-09-16-ptah-rhetoric/README.md) contains
the complete source manifest, comparison summaries, every new finding, historical
source bindings, engine hashes, and reproduction steps. No source or generated
response was added to the frozen training or confirmation corpora.

This work establishes specific missed editing opportunities. It does not show
that every Ptah page needs editing, or that all unwanted rhetoric is covered.
General metaphor, factual relevance, and whether two differently worded claims
add meaning still require a reader or a separately qualified model. A passing
page must not be described as proven high-quality or proven human-written.

## Follow-up on missed constructions

The [whole-page framing comparison](../../research/reviews/2026-09-17-framing-recall/README.md)
adds version 2 evaluation and document-justification constructions plus the
experimental `filler.unscoped-assurance` warning. It detects seven more frozen
development events, but no additional event on nine separate confirmation pages.
The earlier measurements above remain tied to their original engine.
