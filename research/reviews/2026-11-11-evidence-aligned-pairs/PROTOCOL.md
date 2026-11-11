# #349 evidence-aligned source-pair acquisition

Status: offline acquisition protocol. This packet does not run a detector,
external model, or training fit, and does not qualify #349.

## Hypothesis and prior evidence

The current source-function v2 candidate failed on page-disjoint documents:
2/12 FULL events, 2/22 accepted diagnoses, three retained-control hits, and
two unsafe deletions. The explicit-proposition NLI screen received manually
chosen premise and hypothesis spans and failed its contradiction gate. A new
candidate must discover material claims and connect each one to the exact
source context that supports, contradicts, or makes it necessary. A heading or
the whole page repeated as generic context is insufficient.

[Akhter et al. (IJCAI 2026)](https://ijcai.org/proceedings/2026/608)
found that decomposed-claim verification improved only with subclaim-aligned
evidence; repeated claim-level evidence could make performance worse. That is
a design hypothesis here, not a transferred editorial-accuracy result. This
acquisition will record source-aligned positive and retain controls before
building a candidate. It will not treat a discourse-relation label, a high
embedding similarity, or an unsupported factual claim as an editorial defect.

## Source selection before content review

Use GitHub repository search on November 11, 2026. Save the complete first
page of 100 results for each exact query below, with `sort=updated` and
`order=desc`. Preserve the raw JSON and its SHA-256 before opening any README.

1. `topic:developer-tools language:Go stars:500..5000 archived:false`
2. `topic:developer-tools language:Rust stars:500..5000 archived:false`
3. `topic:developer-tools language:Python stars:500..5000 archived:false`

Within each stratum, inspect only repository metadata and the default branch's
root Git tree, in search rank order. Select the first eight distinct public,
non-fork repositories with a root `README.md` Git blob between 4,096 and
30,720 bytes inclusive. Before selection, exclude repository identities used
by any #349 original, reserved prospective, or previously published fresh
source packet at this main commit. The exclusion audit must name the exact
source-inventory files and record their hashes; uncertain overlap is an
unavailable candidate, not a discretionary substitution. Preserve every
skipped rank and reason. Stop after rank 100 even if a stratum has fewer than
eight eligible repositories. Do not replace a source after seeing its text,
labels, baseline findings, or candidate behavior.

Pin each selected repository's default branch commit and root README blob SHA-1.
Fetch the README at that commit. Require the Git blob SHA-1 and recorded size
to match the tree entry; record the original UTF-8 SHA-256 and bytes. Invalid,
missing, or over-limit sources remain in the 24-selection accounting. No
rendered GitHub excerpt or second Markdown parser supplies source coordinates.

The first four selected repositories in each stratum are development pages;
the next two are calibration pages; the final two are transfer pages. This
rank-based split is fixed before source content is read. The transfer pages
remain unopened until the candidate algorithm, training inputs, and scoring
policy are frozen. Their source-only labels are sealed before candidate output.

## Source-only review and paired controls

The root Codex assistant is the actual reviewer under ADR 0041, not an
independent human annotator. Read each complete valid development page and
mark every clear opportunity in the six #349 purposes. Each positive record
must include the exact original UTF-8 byte span and quote, the material
proposition or document function at issue, the exact nearby source span that
supports or contradicts it when present, the reader burden, facts to retain,
and whether a specific edit is established as safe. When no local evidence is
sufficient, record `evidence_state: unresolved`; do not infer that absence of
local support makes a statement false.

For each positive, seek a close control from the same page or another selected
development page: a similar cue whose claim, scope, instruction, limitation,
number, or context is necessary. Record its exact bytes, evidence span, and
`retain`, `optional`, or `technical_uncertain` disposition with a reason. Also
review cue-bearing passages outside the paired controls; an unlabeled passage
is unknown rather than clean. Preserve negation, numbers, versions,
identifiers, code, table cells, and protected regions.

No source-only label may be changed after a candidate output is viewed. Keep
original #349 labels and the 57/46/20 denominators unchanged. The first packet
is informative only if it has at least 12 clear positives across six
repositories and at least 24 close controls, including six `retain` controls.
If this acquisition is underpowered, report the count and stop without a
model run or a passing transfer claim. Do not inspect additional repositories
outside the registered selection to repair that result.

## Candidate and evaluation boundary

This protocol freezes acquisition, not a model. A later protocol must name
the candidate's exact representation, algorithm or model, fitting data,
cutoffs, source binder, failure policy, and execution budget before its first
run. It must compare the full combined delivered stream with the unchanged
current engine, review every diagnosis and supplied edit, count unresolved
advice separately, and hold transfer pages by repository. No candidate can
advance on these pages merely by classifying a supplied, manually localized
pair. It must locate candidate spans and aligned evidence automatically.

Even a passing transfer packet would not close #349. The original 123 exposed
events need complete-cohort reporting at at least 80% FULL-event recall and
85% accepted delivered findings, followed by separate reserved prospective
confirmation. Missing capabilities and operational errors remain failures.
