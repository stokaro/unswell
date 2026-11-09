# #349 source-context candidate generator: negative location screen

**Status: terminal negative exploratory screen; no product change.** This
frozen offline run tested a non-generative classifier trained on the current
source inventory and the 123 exposed event targets. It used local scikit-learn
only. No hosted model call, network lookup, new label, or reserved prospective
page was involved.

The candidate used immediately neighboring eligible units in the same source
block and trained three category classifiers with leave-one-repository-group-out
fitting. Each held-out page received up to one distinct candidate location for
`empty_framing`, `wordiness`, and `vague_claims`. Selection was fixed at three
locations per page before scoring. The selected locations remained bound to
original UTF-8 eligible segments; protected bytes were never made editable.

| Cohort | Current reviewed FULL | Candidate locations covering every target byte | Previously non-FULL events newly location-covered |
| --- | ---: | ---: | ---: |
| Whole page | 28/57 | 3/57 | 2 |
| Contextual | 16/46 | 3/46 | 2 |
| Historical confirmation | 4/20 | 1/20 | 0 |

Across 36 pages and 7,625 eligible units, the method selected 108 locations.
Only 11 selected units overlapped a frozen target in one of the three modeled
categories. Four previously non-FULL events were fully location-covered in
total, below the frozen minimum of eight, and none came from historical
confirmation. One of the four is an incidental `unjustified_intensifiers`
overlap, not evidence that the model learned that category. No finding text or
edit was generated, so these counts are **not** FULL-event recall, accepted
delivery, or suggestion safety. Existing baseline findings remain in place;
this result does not replace their stream.

The exact configuration is retired on these exposed pages. A new threshold,
another choice of top locations, or a retrospective category adjustment would
not be a continuation of this frozen run. The failure is consistent with the
need to assess a statement's role and necessity in the complete source, plus
separate accepted-delivery triage. It does not rule out every small classifier
or source-context method. Issue #349 remains open under its unchanged 80% FULL,
85% accepted, complete-cohort, source-integrity, and advice-safety gates.

`PROTOCOL.md`, `run.py`, source archive, review ledgers, exact main tree,
exporter binary, and local dependency versions were frozen in `freeze.json`
before execution. `result.json` is the public aggregate; per-unit scores and
event IDs remain in a local untracked artifact. The public result SHA-256 is
`81e772f8611012b09119b1c25b1da78e0e01da7e0c4ca5ba8d0673327fb05142`.
Run `python3 -B verify_result.py` and `python3 -B -m unittest discover -s .
-p 'test_*.py'` from this directory to verify artifact identity, accounting,
source-bound selection behavior, and false-success rejection paths.
