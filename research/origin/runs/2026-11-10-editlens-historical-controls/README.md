# EditLens historical technical-document control transfer

**Decision: negative.** The fixed academic five-window threshold flagged
20 of 462 historical technical documents (4.33%). The registered control
transfer gate required zero flags. No threshold was adjusted and no document
was removed after scoring. This run does not qualify an AI-origin detector or
change issue #349 editorial recall.

## Frozen execution

The [protocol](PROTOCOL.md), exact source frame, candidate audits, runner,
independent verifier, and 13 focused tests were committed as `5d98e58`
before the single local CPU inference. The [freeze](freeze.json) pins the
model and tokenizer hashes, preprocessing and window code, 462 source
identities, 3,735 windows, and threshold `0.27447814408224075`. The model
file SHA-256 was verified before inference. Source pages were fetched from
pinned Git blobs and validated by Git blob SHA-1 and raw SHA-256. No source
text is committed in this run.

`verify_selection.py` confirmed the accounting chain from 196 inspected repositories
through 506 candidate files to 462 selected files and 44 exclusions. The
recorded manual exclusion reasons were preserved, but the verifier did not
adjudicate their editorial rationale. The
model scored all 3,735 windows once, with no error. `verify.py` independently
recomputed the document means, threshold decisions, group counts, and result
summary from the sealed [window scores](window_scores.json). The
[decision record](decision.json) pins the three result-file hashes.

## Observed result

| Measure | Result |
| --- | ---: |
| Historical controls scored | 462/462 documents, 3,735/3,735 windows |
| Repositories represented | 153 |
| Flagged at the fixed threshold | **20/462 (4.33%)** |
| Repositories with at least one flag | 16/153 |
| Median document score index | 0.0819 |
| 90th percentile score index | 0.2107 |
| Registered zero-flag gate | **Failed** |

Sixteen of the 20 flags occurred on documents with one to four windows;
four occurred with five to nine windows, and none occurred on the 97
documents with ten or more windows. Eight selected documents had fewer than
100 cleaned words after the exact preprocessing, and two of those were
flagged. These length patterns are exploratory, not a revised gate or an
independent validation of a length-conditioned policy. The complete
[result](result.json) retains language, repository, length, quantile, and
flagged-source accounting; [documents.json](documents.json) retains every
document score and source identity.

The source commits predate August 15, 2021. This makes current LLM writing
unlikely in the selected Git snapshots, but it does not verify each author's
identity or exclude older automated text. The reported 4.33% is a
**historical-control flag rate**, not a verified human false-positive rate.
Possible training-source overlap with EditLens remains unknown. The
academic threshold was learned on a different genre and a five-window
paper score, so transfer failure is informative but does not isolate whether
the model, document length, domain, or aggregation caused the flags.

The earlier paired technical-document experiment in
[`2026-11-10-editlens-technical-fullwindows`](../2026-11-10-editlens-technical-fullwindows/README.md)
still shows an exposed directional signal. This control failure prevents
using that signal as a standalone classifier. Any length-conditioned
abstention or new threshold requires a new predeclared, source-disjoint
control set with checked human provenance and an independent generator
holdout. No production scoring or editorial rule changed here.

## Replay

Fetch the selected blobs with `fetch.py --sources-dir <directory>` after
authenticating `gh`; the fetcher checks each pinned hash. To score again,
install the frozen Python runtime, provide the pinned EditLens modelkit, and
run:

```sh
python3 run.py --model-dir <modelkit> --sources-dir <directory> \
  --output-dir <result-directory>
python3 verify.py --output-dir <result-directory>
python3 verify_selection.py
python3 -m unittest -q test_run.py test_verify.py test_verify_selection.py
```

The committed result files can be verified without the model or raw pages:

```sh
python3 verify.py
python3 verify_selection.py
```
