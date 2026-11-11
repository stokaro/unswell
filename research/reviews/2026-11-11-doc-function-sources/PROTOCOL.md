# #349 document-function source acquisition

Status: frozen offline acquisition protocol. This packet selects public
technical documentation, not a detector or product change. It sends no source
content to a model and does not alter the original 123 #349 events or gates.

## Why this source mix

The preceding [README acquisition](../2026-11-11-evidence-aligned-pairs/RESULT.md)
stopped at its registered gate: nine clear opportunities in 12 development
pages, with 13 close controls. Eight pages described agent tools or their
workspaces. The new selection changes the source function and topic before
opening more source text. The older packet remains terminal and is not pooled
into this packet's feasibility denominator.

## Frozen metadata selection

On November 11, 2026, save the complete first page of 100 GitHub repository
search results for each exact query, `sort=stars`, `order=desc`, and
`per_page=100`, in this order:

1. `topic:database stars:500..5000 archived:false`
2. `topic:security stars:500..5000 archived:false`
3. `topic:observability stars:500..5000 archived:false`
4. `topic:testing stars:500..5000 archived:false`

The pre-freeze count-only feasibility probe returned 667, 1275, 157, and 467
results respectively; it did not inspect repository files or README text.
Preserve raw search JSON and SHA-256. Stop on an incomplete response. The
predeclared [exclusion ledger](exclusions.json) unions every prior #349 source
repository in the preceding packet with its 24 selected identities. Match
repository names case-insensitively. Exclude private, archived, fork, duplicate,
and previously used repositories.

Within each topic, inspect candidates in returned rank order using only the
default-branch commit and recursive Git tree metadata. Skip a missing or
truncated tree. Eligible Markdown files must be under a `docs`, `doc`, or
`documentation` path segment, have a `.md` or `.mdx` suffix, and have a Git
blob size of 8,192 through 32,768 bytes inclusive. Exclude files whose basename
starts with README, CHANGELOG, CONTRIBUTING, or LICENSE. These checks use path
and size only, never content, labels, current-engine findings, or an excerpt.

Assign files to two document functions using lowercase path components and
basename tokens:

- `guide`: guide, guides, tutorial, tutorials, how-to, howto, getting-started,
  quickstart, or manual;
- `concept_reference`: concept, concepts, explanation, explanations, reference,
  references, api, architecture, design, internals, or implementation.

If a repository has multiple eligible files in a function, select its largest
blob, breaking size ties by lexicographic path. If it has files in both
functions, assign it to the function with fewer selected repositories within
the current topic, breaking ties toward `guide`. Use each repository once
across all topics. Select the first four repositories in each topic/function
cell. Preserve every considered rank and reason. Stop after rank 100, without
replacement after source content is seen. A cell with fewer than four makes
the acquisition underpowered; do not repair it with a new query in this packet.

Pin each selected repository's default-branch commit, root tree SHA-1,
document path, Git blob SHA-1, and byte size. Fetch only the development
documents by blob SHA after the complete 32-source selection is sealed. Check
their Git blob SHA-1, size, UTF-8 validity, and SHA-256. Invalid sources remain
in selection accounting. Per topic/function cell, ordinal one and two are
development, three is calibration, and four is transfer. Calibration and
transfer content remain unopened until a later candidate protocol freezes its
algorithm, fitting inputs, scoring, and failure policy. Their source-only labels
must be sealed before candidate outputs are seen.

## Source-only review and stop rule

The root Codex assistant is the reviewer under ADR 0041, not a human annotator.
Read each complete valid development document. Record every clear #349
editorial opportunity with original UTF-8 byte span, exact quote, document
function, material proposition, nearby supporting or contradictory evidence
span, reader burden, retained facts, and edit-safety status. When local context
cannot establish the relation, record unresolved evidence rather than treating
absence of support as falsity. Review similar passages that must be retained,
including negation, quantities, versions, conditions, instructions, claims
with a real evidence trail, and protected Markdown/code boundaries. Unlabeled
passages remain unknown, not negative training examples.

The packet is informative only with at least 16 clear positives across eight
repositories and at least 32 close controls, including eight `retain`
controls. If any gate fails, report the complete counts and stop without a
model run, transfer claim, or revised cutoff. Do not borrow cases from the
previous README packet to make this screen pass.

If acquisition passes, a separate frozen protocol must define an automatic
candidate locator and evidence binder, not merely classify manually selected
pairs. Evaluate the full delivered stream and supplied edits against the
unchanged 57/46/20 #349 cohorts and the 80% FULL-event recall / 85% accepted
finding targets. Reserved prospective confirmation remains separate.
