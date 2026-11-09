# Exposed structured-candidate screen: negative

This archive records the November 8, 2026 local experiment recovered and
independently replayed on November 9. It evaluates an archived six-stream
candidate bank; its 48/57 unfiltered FULL count is **not** the shipped
current-main result. It neither changes the detector nor resolves #349.

V1 stopped at the whole-page inventory guard before fitting or scoring:
there are 12 whole-page repository groups, although the complete inventory
contains 13. V2 changed only that guard and its protocol wording, then ran
once with the frozen files in `freeze-v2.json`.

The v2 repository-held-out AUC for 2,101 whole-page candidate aliases was
0.77747. The fixed 0.5 decision retained 1,183/1,358 accepted criticisms
(87.1%) but only 37/57 FULL events, with four unresolved supplied edits.
Across all distinct held-out scores, the optimistic post-hoc threshold
envelope reached at most 45/57 FULL events while keeping at least 85% useful
actual criticisms: 1,321/1,554 (85.006%), with five unresolved edits.
The required whole-page development count is 46/57 FULL; no pooled threshold
met both numeric gates. Unfiltered context and historical confirmation
cohorts remained 39/46 and 17/20 FULL, respectively. This does not test
standalone generation, advice safety, or new-page confirmation.

The result is exploratory, based on already exposed assistant grades. The
post-hoc threshold is not a policy. This exact feature family is retired
without further fitting or threshold tuning on these records. A new method
must test source-specific phrase function and retention on new page-disjoint
sources before any full 123-event development run and separate prospective
confirmation.

`make check-structured-admission` verifies the public aggregate, frozen
protocol and script hashes, denominators, fold accounting, and explicit
failure claims without private development inputs. Negative tests reject
changed input identities, denominators, folds, fractions, and a false pass.

To recompute the aggregate from the archived inputs, use Python 3.12 with
`numpy==2.5.4`, `scipy==1.18.1`, and `scikit-learn==1.7.2`, then run:

```sh
python scan-v2.py --development-root /path/to/archived/development \
  --output /tmp/structured-admission-replay.json
cmp result-v2.json /tmp/structured-admission-replay.json
```

The five input SHA-256 identities are embedded in `scan-v2.py` and the
published result. The local replay matched `result-v2.json` byte for byte
(SHA-256 `df663463350602d3592fc930caa89fa0589a073d3e3e00f4b9983f61cff38765`).
The source-level inputs are archived outside this repository; the public
verifier does not pretend to reconstruct them.
