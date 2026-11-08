# Mandatory unit verdict pilot for #349

This selected development pilot tests a specific failure found in the earlier
[exhaustive inventory](../2026-11-08-exhaustive-inventory-pilot/README.md): all
four frozen problems were represented in the first-pass proposition and support
records, but the page-level diagnosis pass produced only partial findings. A
post hoc support-gap phrase screen found the four target units among 37 of 169
units. That is a location observation, not a diagnosis or transferable score.

The new hypothesis separates two judgments. First, the model must classify
every eligible unit's proposition and its relation to support on the *complete*
page. Then a separate model call must decide the editorial necessity of a
correction for **every** unit marked narrower, contradicted, or unestablished.
The verifier rejects a missing verdict. No global top-N output may silently
discard a candidate. Ordinary factual descriptions, qualified technical
claims, warnings, and authorial voice remain counterexamples; absent proof in a
README is not automatically a defect. Source roles are a useful distinction in
[functional discourse research](https://aclanthology.org/2022.acl-short.28/),
but that paper's news deletion results do not establish technical-editorial
quality here.

## Frozen material and authority

Four complete, public README files from distinct repositories are pinned by
commit, Git blob, SHA-256, and exact bytes in `sources.json` and `sources/`.
The current Go source inventory exported 388 eligible units with original
UTF-8 segments and protected exclusions into `inventory.json`. All five
source-only assistant-reviewed target spans lie wholly in eligible segments.
`labels.json` freezes five independently useful semantic defects, five
necessary counterexamples, two unresolved judgments, and one complete page
with no established defect before any model output. The reviewer is the Codex
assistant, not a human annotator. The sample was selected after inspecting the
pages and is not a random prevalence estimate.

The first-pass requests, prompts, schemas, model, effort, candidate policy,
failure policy, and evaluation script are SHA-256-pinned by `packet/freeze.json`.
The exact packet must be committed before model execution. A second-pass
request is generated only from validated first-pass records by the frozen
`packet.py`; its bytes and SHA-256 enter the journal. No labels, review grades,
or earlier candidate answers enter either request.

`packet/protocol.json` allows at most nine sequential GPT-6-Astra medium calls
through the existing Codex subscription, one attempt per request, no tools or
external search, a 600-second limit per call and 10,800 seconds overall.
Execution stops at the first CLI, schema, completeness, candidate-budget,
source-binding, or trace error without repair or retry. Fastfetch's 176 units
are split into two exhaustive chunks; each other page uses one chunk. At most
64 selected candidates per page are allowed. The final decision call returns
one verdict for every selected unit, including explicit abstentions. Diagnoses
must quote one eligible original UTF-8 span with explicit byte offsets and no
supplied replacement edit.

## Registered assessment

The assistant reviews every *actual delivered* diagnosis against the complete
page and frozen labels. The verifier reports candidate-stage event exposure,
FULL versus partial versus missed events, accepted versus uncertain versus
nonactionable diagnoses, abstentions, invalid pages, protected controls, and
edit safety separately. Uncertain diagnoses remain in the delivery denominator.
The pilot advances only with at least 4/5 FULL events, at least 85% accepted
actual diagnoses, all four valid complete pages, zero protected counterexamples
flagged, and no supplied edits. Zero delivered diagnoses cannot pass precision.

This is an exposed, selected development pilot. Even an advance would require
a separate repository-disjoint pilot, the unchanged complete 123-event
development accounting in #349, and untouched prospective confirmation. A
small-sample pass is never product qualification or evidence of AI authorship.
The original frozen labels and denominators remain unchanged.
