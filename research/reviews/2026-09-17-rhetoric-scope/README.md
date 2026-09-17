# Clause scope: four exposed gains, one new-page gain

This candidate repairs two confirmed instruction false positives and detects
four more frozen defects on the 106 exposed pages. On six separately frozen
complete pages, full detection rises from **5/72 to 6/72** in both profiles.
Ptah rises from **1/29 to 2/29**; the historical pages stay at **4/43**. The working
80% recall and 85% soft-precision objectives are not met. This is a bounded
correctness improvement, not evidence that the broad detection problem is solved.

## What changed

`filler.evaluative-closure` version 7 scopes attribution through the candidate
clause. An operational report after a colon no longer suppresses a preceding
value announcement. An attribution before that colon still protects reported
speech. Quoted judgments remain excluded; a quoted noun phrase can be the
subject of an unquoted reading-value judgment.

Information judgments now recognize discourse subjects declaring importance,
cognitive-worth predicates with bounded manner/frequency qualifiers and abstract
whole-value announcements. Two adjacent evaluative predicates can share one
source-bound diagnostic. Concrete purposes, costs, measured qualifiers and
conditional judgments remain controls. These are bounded surface constructions;
they do not infer semantic equivalence or dependency edges.

`filler.instruction-scaffolding` version 5 requires a nominal actor and excludes
restrictive relative subjects. This removes the imperative changelog entry
preceded by an issue link and the HTML upload-form definition reported in PR319.
The tagger's adjective label for a noun such as library is accepted only with a
nominal determiner. Existing protected identifier actors remain opaque.

The rule count remains 53 and the class manifest remains r8. Weights, gates and
thresholds are unchanged. Blackbox API and CLI tests cover reported speech,
quoted information, technical controls, original byte ranges, Unicode, CRLF,
Markdown emphasis, Go comments, Python strings, term exemptions and occurrence
policy. Existing cancellation and resource tests remain in the required checks.

## Frozen review

The selector excluded all 106 exposed references and hashes and the original
historical study. Protocol, source identities, notices and bytes were frozen
before prose review. The implementing assistant read all 1,162 extracted prose
blocks and table cells, then froze 72 defects, seven uncertainties and 24 controls
at `2026-09-17T18:46:32.206750+00:00` before runtime edits or diagnostic output. The Ptah
home page uses native MDX extraction and has no definite frozen defect.

The reviewer is the implementing Codex assistant accepted under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md). These are
assistant development judgments, not human labels or independent qualification.
One page per cohort/length cell supports counts, not population estimates.
No semantic tuning followed inspection of confirmation output. All 112 source
identities are now exposed and must be excluded from future fresh confirmation.

| Set | Pages | Frozen defects | Before full | After full |
| --- | --- | --- | --- | --- |
| development | 26 | 57 | 14 | 14 |
| exposed_repetition | 6 | 5 | 3 | 3 |
| exposed_framing | 9 | 11 | 1 | 1 |
| exposed_local | 12 | 78 | 10 | 11 |
| exposed_context | 11 | 42 | 8 | 8 |
| exposed_instruction | 12 | 99 | 29 | 29 |
| exposed_construction | 6 | 32 | 7 | 7 |
| exposed_purpose | 6 | 28 | 4 | 4 |
| exposed_verb | 6 | 43 | 2 | 2 |
| exposed_relations | 6 | 45 | 8 | 8 |
| exposed_projection | 6 | 34 | 3 | 6 |
| confirmation | 6 | 72 | 5 | 6 |

The exposed gains are the whole-value announcement in the provider guide, the
combined whole-guarantee and worth-stating announcement, quoted reading advice,
and the importance declaration in the component reference. All four are Ptah
framing events. One replaces a previously partial diagnostic. No frozen full
detection is lost. Both removed instruction findings were recorded false positives.
The generic feature introductions that remained uncertain in PR319 are unchanged.

## New complete pages

| Page | Original path | Defects | Detected |
| --- | --- | --- | --- |
| c01 | docs/site/src/content/docs/index.mdx | 0 | 0 |
| c02 | docs/site/src/content/docs/schema/sql.md | 9 | 1 |
| c03 | docs/site/src/content/docs/reference/atlas-commands.md | 20 | 1 |
| c04 | docs/INTERNALS.md | 28 | 0 |
| c05 | docs/classDiagram.md | 13 | 4 |
| c06 | docs/usage/exporting_models.md | 2 | 0 |

The only new-page gain is the whole-criterion announcement in the Atlas command
reference. Technical emits 68 findings: six frozen actionable findings, one
additional actionable wording edit absent from frozen labels, four unresolved
judgments and 57 nonactionable findings. Strict emits 75 with the same six,
one and four, plus 64 nonactionable findings. All six pages pass both gates.

The one additional wording edit receives no frozen recall credit. The four
unresolved length judgments are retained rather than counted as successful
wording detection. Most other length and punctuation findings identify concrete
technical conditions or overlap wording defects without identifying their repair.
The unchanged test-suite sentence-opener warning repeats a necessary named subject.

The [miss ledger](CONFIRMATION-MISSES.md) records all 66 missed events. Of the
72 frozen defects, only 2/21 empty-framing and 4/27 wordiness events are detected;
none of the eight needless repetitions, eleven unjustified intensifiers or five
vague claims is detected. There are no frozen formulaic-transition or needless-
complexity defects in this sample. Fixing scope alone has little effect on broad
recall. More narrow templates must not be presented as completion of that goal.

[CHANGES.md](CHANGES.md) records every added, replaced and removed diagnostic.
`dispositions.json` also reviews all retained confirmation findings. Original
restricted scopes for the older repetition/framing studies remain explicit.

## Reproduction and limits

Before runtime: `d2989e83389751b0fbf2f9d55e0cff8546002fe2`. After runtime: `d5ef492c15c67fcdcf2a5a7883f7f21d12e0340a`.
The after binary was built from a clean detached checkout of the runtime commit.
An earlier dirty-checkout build was retained only in temporary scratch space and
was replaced by a full replay of the clean binary before acceptance of reports.
Exposed before reports retain the previous run; every after set was replayed.
JSON preserves source bytes, spans, engine and policy identities and resource
records. No source or frozen annotation was revised after output review.

The existing #312 repeated-claim budget abstention remains on exposed instruction
page c05. No new abstention or operational error appears. Technical passes all
sets; strict fails only the preexisting forbidden worth-noting phrase in the
exposed instruction set. The old CMake text-mode limitation remains inherited.

| Set | Profile | Seconds before → after | Peak MiB before → after |
| --- | --- | --- | --- |
| development | technical | 2.216 → 2.902 | 188.8 → 194.1 |
| development | strict | 1.639 → 2.119 | 189.9 → 196.6 |
| exposed_repetition | technical | 0.573 → 0.976 | 126.8 → 124.4 |
| exposed_repetition | strict | 0.592 → 0.903 | 121.0 → 120.3 |
| exposed_framing | technical | 0.979 → 1.187 | 128.5 → 144.7 |
| exposed_framing | strict | 0.827 → 1.192 | 135.3 → 134.0 |
| exposed_local | technical | 0.705 → 0.708 | 122.6 → 117.7 |
| exposed_local | strict | 0.699 → 0.694 | 121.4 → 120.1 |
| exposed_context | technical | 0.729 → 0.718 | 124.2 → 123.3 |
| exposed_context | strict | 0.717 → 0.733 | 120.1 → 120.4 |
| exposed_instruction | technical | 0.941 → 0.949 | 146.1 → 151.9 |
| exposed_instruction | strict | 0.940 → 0.947 | 166.6 → 167.3 |
| exposed_construction | technical | 0.496 → 0.498 | 103.8 → 107.7 |
| exposed_construction | strict | 0.510 → 0.502 | 103.5 → 101.7 |
| exposed_purpose | technical | 0.526 → 0.498 | 103.7 → 107.2 |
| exposed_purpose | strict | 0.491 → 0.502 | 108.9 → 106.1 |
| exposed_verb | technical | 0.544 → 0.553 | 96.9 → 97.8 |
| exposed_verb | strict | 0.552 → 0.554 | 114.2 → 107.2 |
| exposed_relations | technical | 0.577 → 0.586 | 119.8 → 125.4 |
| exposed_relations | strict | 0.579 → 0.581 | 116.0 → 108.5 |
| exposed_projection | technical | 0.487 → 0.492 | 108.8 → 108.7 |
| exposed_projection | strict | 0.558 → 0.486 | 110.8 → 103.6 |
| confirmation | technical | 1.168 → 0.735 | 178.5 → 165.3 |
| confirmation | strict | 1.079 → 0.736 | 163.7 → 153.1 |

Measurements overlapped ordinary checks; they are not a controlled performance
comparison or 2-vCPU qualification. Run `python3 tools/render.py` and
`python3 tools/test_evidence.py` here to validate and regenerate these results.
`tools/measure.py` replays an explicit binary into a new output directory.
See [VALIDATION.md](VALIDATION.md) for product checks and their actual status.
