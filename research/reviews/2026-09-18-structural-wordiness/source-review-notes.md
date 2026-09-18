# Source-only review

The input protocol, metadata pool, selector, twelve source identities and archived
bytes were frozen before reading the selected prose. The review used extracted
prose and original source context; it did not use findings from either runtime.
No runtime change has been made in this iteration at the annotation freeze.

The reviewer is the Codex assistant. This is maintainer-accepted development
review under ADR 0041, not human annotation or an independent qualification set.
No origin label was inferred from a publication date or from the wording.

All seven editorial categories were considered on every page. The labels contain
63 definite editing events, 16 uncertain events and 53 controls. Eighteen definite
events occur on the six Ptah pages. Two Ptah pages have no definite event; neither
was replaced to obtain more positives. These counts describe the review, not
detector recall or precision. Diagnostic matching and review are still pending.

Controls preserve the technical distinctions that a shorter rewrite must keep:
transaction atomicity, tombstones versus deletion, corpus key encodings, hard
expectations versus ranked scores, seed checksum and duplicate-key behavior,
per-request schema reads, Boolean build constraints, version-gated transition
steps, test-runner modes, cardinality and refutability. A control targets the
necessary clause when another part of its paragraph has a definite defect.
Defect targets do not overlap uncertain or control targets.

Potentially useful explanation remains uncertain rather than becoming a forced
positive. Examples include restating the consequence of transaction atomicity,
the link between a corpus digest and `--baseline`, and a capability consequence
of optional ER-diagram fields. Grammar mistakes and factual inconsistencies were
not relabeled as editorial defects to expand the denominator.

## Source availability and expansion

The prior 19 sets contain 160 exposed source identities. The original pinned
pools have no unused long pages. The protocol therefore fixes four short and two
medium Ptah pages, plus two historical pages per length stratum from six distinct
repositories. New long historical pages cannot substitute for new long Ptah
evidence. All earlier long Ptah pages remain required exposed regressions.

Four historical repository snapshots were added at the newest available commit
before January 1, 2022: grpc/grpc, golang/proposal, rust-lang/book and
kubernetes/community. Eligibility roots, size limits, exclusions and deterministic
ranking were fixed before source review. All four contribute to the selected
sample. Root notices and exact source hashes are retained. These are dated
snapshots with unknown unit-level origin, not certified human negatives.

The source-only overlap check found no normalized complete prose block of at
least twelve whitespace-delimited tokens shared with the 160 prior identities.
This is a limited exact-block check; it does not rule out shorter, paraphrased or
semantically related material. Selection word counts are extraction-based
whitespace counts, not the runtime's NLP word counts.

The complete new pages are sealed for confirmation. Their labels do not become
training examples or runtime fixtures. Runtime development uses the exposed
sets and independently written construction tests. Candidate semantics must be
frozen before viewing diagnostics on these pages, and a failed confirmation
does not permit tuning against the same pages and calling them fresh again.
