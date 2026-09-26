# Generic introductions before conditional outcomes

`filler.document-justification` version 6 adds a contextual construction: a
bare announcement of several possible outcomes followed immediately by two
complete conditional sentences in the same prose block. The announcement is
the primary range; both branches are related evidence. The editing advice is
to remove the announcement while preserving every condition and outcome.

The rule's message and suggestion now cover both document and outcome
announcements. Its ID, weight, severity, gate and thresholds do not change.
Runtime still uses local Go code and the existing extraction, NLP, source
mapping, resource limits and scoring.

## Scope and controls

An isolated list introduction does not qualify. Exact counts, domain limits,
concrete outcome categories, attribution, questions, quotations, protected code
and block boundaries remain controls. Surface predicates distinguish complete
branches from fragments. A comma alone is not enough. Two bounded lexical
predicates address the backend's noun tags for *cache matches/misses*; arbitrary
plural nouns do not count as verbs. This is a heuristic, not a dependency parse
or proof that every announcement is unnecessary.

The implementation was isolated from the recovered September 21 draft. Its 32
constructed examples were previously exposed; their recovery is not a new
blind freeze. Ten additional examples were frozen before this port. A further
plural-noun regression guards the lexical fallback. Tests exercise negation,
source ranges, two related locations, allowances, MDX, Go comments, Python
strings, Unicode, BOM and CRLF. The CLI fixture also covers protected code and
JSON/SARIF output.

## Complete-page results

The baseline is `6591c318f998bc9be9f90bb294f3080d1d89c3b9`. Both profiles replay
all 36 complete source files. The cohorts are separate; their totals must not
be pooled into a population estimate. Review is by the same Codex assistant,
under ADR 0041, not by an independent human annotator.

| Cohort | Pages | Frozen defects | Before | After |
| --- | ---: | ---: | ---: | ---: |
| September 17 full-page audit | 26 | 57 | 15 | 15 |
| Exposed contextual reference | 6 | 46 | 4 | 5 |
| Current Ptah follow-up (mixed exposure) | 4 | 20 | 1 | 1 |

The context denominator uses **all 46 defects** in the retained
[reference annotations](../2026-09-17-proposition-repetition/confirmation/annotations.json).
It does not substitute the smaller 28-event model-reference subset mentioned
in older experiments. Two multi-layer context defects remain partial because
an action-setup warning does not establish that every annotated redundant
layer was diagnosed. All three earlier partial events in the 26-page audit
also remain partial. Overlapping length or punctuation warnings receive no
rhetorical-event credit; an existing complexity event can retain its actual
length diagnosis.

The one added finding is BuildKit's *There are a couple of possibilities how
this check may end up*, event `context/c04-d11`. The preceding sentence names
the check and the following branches supply its outcomes. Removing just the
announcement preserves the cancellation and completion conditions. The same
assistant accepts this addition. One accepted development finding is not an
estimate of general precision.

Technical findings rise from 164 to 165; strict findings rise from 185 to 186.
All earlier findings retain their locations, evidence and scores. The only
intentional metadata changes are the document-justification version, message
and editing suggestion. The verifier excludes those documented metadata fields
from the identity comparison, not diagnostic content or source ranges.

## Separate confirmation and limitations

Four complete Ptah migration guides were selected at commit
`6318f98bf5937266431766d15f75093da5a3b39e` by a fixed SHA-256 path rank before
reading their prose or diagnostic outputs. The initial short-Markdown frame
contained only two eligible files; it was expanded to MDX and 14,000 bytes
before selection. The manifest records that change and the excluded paths.
The initial exclusion scan used retained JSON path mentions and known earlier
pages. A post-output provenance audit found the Atlas and overview pages named
in a recovered README from September 21. Their bytes are identical to that
previously reviewed snapshot. `exposure-amendment.json` records this discovery;
these two pages are **previously exposed**, not fresh confirmation. No source
or label was dropped to improve the measurements.

The other two pages, Liquibase and golang-migrate, have no known earlier review
in the retained records. Their result is **0/8**, while the two known reviewed
pages account for **1/12**. These strata are generated in `summary.json`.
Neither a universal lack of conversational exposure nor lack of model-training
exposure is claimed.

The same assistant read and labeled all four sources before implementing this
port or viewing their outputs. All four selected pages belong to the migration
guide genre and share tutorial conventions; they are not four independent
documentation styles. Their 20 defects, acceptable examples and uncertain
alternatives are frozen. No behavior was tuned after inspecting confirmation
outputs.

The complete follow-up remains **1/20 full-event recall** in both profiles. The new
construction has no positive occurrence here. Technical emits eight findings:
three accepted, four uncertain and one rejected. Strict adds one rejected
punctuation finding. Uncertain and rejected findings remain in the denominator.
A repeated-output-caption warning is a pre-existing false positive: the captions
introduce different command examples. Post-freeze length advice does not add
new events to the recall denominator.

This change recovers a specific real defect and preserves the tested controls.
It does not demonstrate transfer to Ptah, satisfy the broader recall goal, or
establish the 80% recall and 85% precision targets. No authorship, model
probability, public playground deployment or release claim follows from it.

## Reproduction

Existing reference archives retain their source notices. Confirmation sources
come from maintainer-authorized Ptah documentation under `LICENSE.ptah`.
`input-freeze.json`, `label-freeze.json`, `additional-freeze.json` and
`execution-freeze.json` record the actual sequence and hashes. Earlier failed
constructed-test runs led to the documented predicate corrections before
confirmation output review; frozen labels were not changed to match output.

```sh
python3 -B research/reviews/2026-09-26-outcome-announcements/verify.py
python3 -B -m unittest discover -s research/reviews/2026-09-26-outcome-announcements -p test_verify.py -v
CGO_ENABLED=0 go build -o /tmp/unswell-outcomes ./cmd/unswell
python3 -B research/reviews/2026-09-26-outcome-announcements/replay.py --binary /tmp/unswell-outcomes --output /tmp/unswell-outcome-replay
```

The output directory must be new. The offline verifier checks all source
identities, complete reports, byte ranges, event denominators, retained
findings and every new or confirmation disposition. Its negative probes reject
source drift, omitted reviews and unsupported credit. It validates recorded
judgments rather than independently judging English prose.

`summary.json` is generated from the verifier. Full product acceptance belongs
to the PR's terminal CI evidence; this record does not substitute for CI.
