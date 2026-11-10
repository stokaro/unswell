# #349 atomic-claim acquisition and paired pilot

Status: source selection, source bytes, eligible-unit inventory, and source-only
labels are frozen locally; prompts, schemas, and model calls are pending. No
model call has been made for this packet. This is exposed development. It cannot
replace the 36-page/123-event development evaluation or the separately frozen
prospective confirmation. The earlier labels and denominators are unchanged.

## Question

The exhaustive eligible-unit inventory reached every source unit but still
produced only one diagnosis per page and 0/4 FULL events. A later source-role
candidate reached 15/17 FULL on exposed pages, then only 2/12 FULL and 2/22
accepted on other repositories; two proposed deletions damaged benchmark
evidence. The repair-first pilot stopped on a byte-length rule and missed its
one valid page's known event. A new test must address semantic omissions inside
units and preserve technical evidence, rather than retune a cue or score.

The atomic-claim hypothesis is that explicitly accounting for independent
claims inside each existing eligible unit will expose distinct obligations
that a one-diagnosis-per-page review omitted. A claim includes its source
role, stated scope, qualifiers, and any cited support. An arm using this
ledger will be paired with a free-form arm on the same complete pages. The
paired test measures whether this structure improves FULL event recall while
retaining accepted delivery and controls. It does not assume the hypothesis
is correct.

## Acquisition before content review

The three exact GitHub code-search queries are:

1. `"it is important to" filename:README.md size:4000..20000`
2. `"as a result" filename:README.md size:4000..20000`
3. `"we believe" filename:README.md size:4000..20000`

For each query, use the first 30 results sorted by indexed time descending
and save the complete JSON response before opening any matched page. A
conservative exclusion snapshot collects repository names from prior review
JSON values named `repo`, `repository`, or `full_name` at base commit
`9c4884ab44ef39b1f5dde62a63c5476ada02e7f1`. Select the first three
results per query whose repositories are absent from that snapshot and from
earlier selections. Keep all nine selected sources in the denominator; do not
replace a source after reading it because it is clean, ambiguous, unavailable,
or inconvenient. This cue-enriched sample cannot estimate prevalence.

`selection.json` records search hashes, ranks, repository, path, search Git
blob, and pinned commit before source content is fetched. Fetch each complete
original UTF-8 page at that commit, verify the Git blob SHA-1 using the Git
blob header, and record byte count and SHA-256. Reject a nonmatching,
unavailable, non-UTF-8, or incomplete source as invalid. Retain exact original
bytes; no rendered excerpt and no second Markdown parser may stand in for the
source. Use the current public source-inventory exporter to bind eligible
units, protected boundaries, and original byte spans. If the exporter reports
an operational error, keep the page and error in the denominator.

The metadata was committed as `50b4cfe` before source fetching. The nine
complete source files were then fetched at the selected commits, all matched
the search and GitHub API blob SHA-1, and were committed as `2fb7664` before
editorial review. `verify_selection.py` and `verify_sources.py` reject drift.
The current `reviewtext --source-inventory` exporter processed all nine pages
without error and returned 889 eligible units. Its input and output SHA-256
are `ef4b077536aa69bb16c0a851d144bb36e62e6780bb6fc32e8692d8d9ae298e30`
and `98b8a7bec295d16db53654cad67a3771060b8cc83b1766aab6a7a8caf426d87f`.

## Source-only reference review

The root Codex assistant reviews every complete page before seeing any new
candidate output. Record all clear editorial opportunities in the six #349
purposes: empty framing, wordiness, vague claims, unjustified intensifiers,
needless repetition, and needless complexity. Bind each event to an exact
original UTF-8 byte span and quote. Record the source role, independent
claims involved, reader burden, content to retain, and a concrete safe
revision or mark revision safety unresolved. A claim with a named successor,
explicit benchmark scope, caveat, code behavior, table cell, or cited support
must not be reduced to an unsupported promotional claim merely because it
contains a familiar cue.

Separately label at least one plausible `retain`, `optional`, `out_of_scope`,
or `technical_uncertain` control per valid page where possible. Include table
data, table-scoping prose, numbered steps, comparisons, safety warnings, and
purposeful repetition when present. Unlabeled text is unknown, not a clean
negative. The assistant is the actual reviewer under ADR 0041 and must not be
recorded as a human annotator. An amendment to a source-only judgment keeps
both versions and its date; no silent relabeling.

Freeze the complete sources, inventory, reference labels, and hashes before
running either model arm. Proceed to model packet preparation only if there
are at least 12 clear positive events and 12 explicit controls across at
least four repositories, with all nine source identities accounted for.
Otherwise report the shortfall and stop without model calls. Do not choose a
favorable six-page subset after reading labels.

The source-only review records **15 clear events** on seven repositories and
**21 controls** covering all nine pages. Pages 6 and 9 have no clear positive
event and remain in the denominator. An initial draft contained three more
possible events; before candidate output, the reviewer moved those personal
or navigational phrases to `optional` controls. Both draft and final records
are retained. All final quotes are unique original byte slices inside one
eligible unit. `reference.json` SHA-256 is
`9550fd43d25463c5e1d24d8e7d4806a5698321d40d59fdd4501053bef4258d39`;
`verify_reference.py` checks its source, inventory, dispositions, and spans.
The minimum source-only feasibility gate is met. This is the named assistant's
judgment on a purposive sample, not a precision or recall result.

## Model packet to freeze after the source-only review

Use the same complete source, eligible-unit inventory, original byte spans,
protected regions, model, effort, output-token ceiling, and per-page
diagnosis limit in both arms. The free-form arm requests source-specific
diagnoses without a claim ledger. The atomic arm must return one accounting
record for every eligible unit; each record contains zero or more independent
claims with exact in-unit quotes, source role, scope, qualifiers, support, and
`retain`, `diagnose`, or `uncertain` disposition. A diagnosis names the exact
editorial burden and useful meaning to preserve. A unit with no independent
claim has an explicit empty record and reason. Do not request replacement
text in either arm. No arm receives reference labels, reviewer grades, or the
other arm's output. Use separate ephemeral tool-free sessions.

The packet must pin exact prompts, response schemas, binder, request bytes,
source and inventory hashes, arm order, model, effort, time and token limits,
failure rule, and scorer in a committed Git tree before the first call. The
maximum is 18 calls, one per arm on each of nine pages, sequentially, with
one attempt each, ten minutes per call and three hours total. Stop on the
first CLI, trace, schema, binding, source, resource, or timeout error; never
retry or repair a response under the same packet. Preserve raw request,
response, trace, stderr, and usage. A partial run is terminal evidence, not a
pilot pass. No model calls are authorized by this acquisition artifact alone.

## Scoring and decision

Bind every diagnosis to one eligible original span. A repeated quote must be
disambiguated by unit and exact byte location. Reject protected overlap,
source mismatch, missing unit accounting, fabricated quote, or changed
response shape. Preserve all delivered diagnoses, including duplicates, in
the accepted-finding denominator. `FULL` requires the diagnosis to identify
and resolve the complete frozen source-specific burden; source overlap or a
shared category alone is partial. One diagnosis may cover more than one
frozen event only if it explicitly identifies each independent burden. A
generic warning about a sentence with several distinct claims earns no
automatic credit for all claims. Review every diagnosis and every frozen
event and control against the complete page. Report each arm's FULL, PARTIAL,
MISSED, accepted/uncertain/rejected delivery, invalid and abstained pages,
control hits, and binding errors. Supplied-edit safety is not applicable to
diagnosis-only arms; a later edit test must measure it separately.

Advance the atomic arm only if every page and call is valid, it gains at least
two FULL events over the paired free-form arm on the same frozen positives,
at least 85% of its delivered diagnoses are accepted, and it hits no explicit
retain control or protected source span. This is a directional pilot gate,
not #349 qualification. Even a pass requires the unchanged main engine and
candidate's entire combined finding stream to be reviewed on all original
123 exposed events, with at least 80% FULL recall and 85% accepted delivery
in each complete cohort, followed by separate prospective confirmation and
suggestion-safety evidence before any product promotion. A failed pilot is
published with frozen denominators; the exact candidate is retired.
