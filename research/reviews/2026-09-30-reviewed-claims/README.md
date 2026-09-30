# Explicit migration of original editorial claims

This optional local research consumer replays caller-reviewed source roles and
claim boundaries through the existing Go accounting command. It preserves every
original parent, including parents whose old broad quotations contain protected
identifiers or supporting context. It does not narrow targets automatically to
make a candidate valid. [#367](https://github.com/stokaro/unswell/issues/367)
remains open for the complete migration and downstream use.

The full original pool remains 804 candidates on 36 exposed pages. A Codex
assistant inspected 13 complete sources and every original candidate on those
pages under [ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md).
The [aggregate record](summary.json) reports 254 migrated parents, one explicitly
declined parent, and 549 unreviewed parents. Existing Go binding and retention
checks preserve all 358 declared child claims. The pool is incomplete and the
command exits 2. No candidate disappears from its original denominator.

These are source-role and claim-boundary decisions. The declined parent remains
in the record with its original diagnostic, locations, hash, and migration reason.
Earlier editorial labels and event mappings remain unchanged. An observation
retained as a child is not thereby an accepted wording defect. No full-event
metric was recomputed, no model ran, no duplicate was collapsed, and no prospective
page was used. The 57 whole-page, 46 contextual, and 20 historical exposed events
remain the same 123-event development denominator. Neither this migration nor
successful source binding satisfies the 80% recall / 85% accepted-finding objective.

## Preserve the parent and separate its criticisms

[migrate.py](migrate.py) reads a frozen packet and a separate local review. Each
reviewed parent cites a digest of its complete original candidate, not only its
quote. The source and original request hashes must match the frozen manifest.
Every parent is recorded as migrated, rejected, uncertain, or unreviewed. Only
migrated parents declare children; missing reviews remain unreviewed rather than
silently accepted or omitted.

The caller declares each child key, category, diagnostic, explanation, guidance,
revision targets, and support references explicitly. A child target must remain
within an original parent location. Supporting evidence may name other exact
prose in the same original document; this does not expand the revision target.
The harness derives block ownership only from the existing Go extractor. It adds
no Markdown parser, token heuristic, label inference, or code-stripping fallback.

Go binding still rejects protected target or support bytes, invented quotations,
overlapping target/support roles, incorrect UTF-8 boundaries, ambiguous owners,
and undeclared or missing source identities. All inline identifiers stay in the
immutable original record; they do not become revision targets because neighboring
prose needs review. The initial candidate can have a broad original quotation,
while the caller-reviewed child refers only to the actual criticized wording.

An original criticism about a repeated conclusion can have earlier prose as
support. That earlier prose must not acquire a revision target just because it
was quoted alongside the conclusion. Separately declared grammar, rhetoric,
unsupported-assurance, and navigation criticisms retain separate IDs even when
their source ranges overlap. Removing or correcting one does not resolve another.

The replay calls the existing Go `reviewed-inventory-retention` stage. It verifies
that all supplied children, originals, roles, and IDs survive unchanged. This
stage establishes retention, not an editorial audit, selection decision, or
semantic acceptance. Later rewriting, audit, and selection still use the explicit
stage and approval contract of [ADR 0042](../../../docs/adr/0042-editorial-claim-accounting.md).

The claim run identity uses the original packet hash. Adding another reviewed
page or updating reviewer metadata cannot rename unchanged existing claims.
The exact review file has its own hash in the replay record. Changing a claim's
actual declaration produces a different claim ID; its unchanged parent survives.

## Trust and publication boundary

The supplied review is trusted application input, separate from model output.
The harness checks its structure and source evidence, but cannot authenticate a
reviewer or prove that a declaration is semantically atomic, correct, or complete.
The record names the actual assistant reviewer; no human annotation or independent
review is claimed.

Complete source text, original candidates, explicit declarations, and detailed
replay outputs stay local. They inherit the original packet's permissions and are
not published automatically. Public artifacts contain only this consumer, its
constructed tests, and aggregate counts/hashes. Without the intentionally excluded
packet and review, readers can reproduce the constructed contract checks but
cannot independently reproduce the source-role decisions on the original pool.

## Run the isolated checks

Build the existing consumer from `research/annotation`:

```sh
CGO_ENABLED=0 go build -o /tmp/reviewclaims ./cmd/reviewclaims
```

From this directory, run the optional standard-library Python checks against it:

```sh
UNSWELL_REVIEWCLAIMS_BINARY=/tmp/reviewclaims python3 -B -m unittest -v test_migrate
```

All 15 methods passed, including actual Go replay of a constructed compound
grammar/rhetoric parent with an inline identifier and a retained Unicode antecedent.
Negative cases reject changed parent criticisms, altered frozen sources, invented
quotes, incorrect byte ranges, repeated IDs, unordered references, protected targets, context relabeled
as a target, unknown fields, unsupported dispositions, and overwriting evidence.
Unreviewed and uncertain parents cannot produce complete migration results.

An initial local declaration on the twelfth page placed two explicitly selected
references after later source positions. Existing Go binding rejected it; that
attempt has no successful aggregate. The corrected review orders the same declared
locations without changing any source role. The failed local attempt is preserved
separately, and the constructed integration check covers the same rejection.

With the separately maintained local packet and review:

```sh
python3 -B migrate.py \
  --packet /path/to/frozen-packet \
  --binary /tmp/reviewclaims \
  --review /path/to/explicit-review.json \
  --output /path/to/new-local-evidence
```

The destination must be new. Input is bounded, duplicate JSON keys fail, and
every review is tied to its original packet/page/candidate identity. New detailed
records use private file permissions. Successful complete migration returns 0;
unreviewed or uncertain parents emit their incomplete aggregate and return 2.
Operational errors cannot emit a successful aggregate. Python is isolated
research tooling; the ordinary product, Go tests, and runtime remain independent
of a Python model environment or network access.
