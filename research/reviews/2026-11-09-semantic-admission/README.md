# Frozen semantic admission screen: negative

The archived six-stream candidate bank contains enough accepted candidates
for an optimistic upper bound of 48/57 whole-page, 39/46 contextual, and
17/20 historical FULL events. Removing every nonaccepted alias from the
already graded bank preserves those FULL counts and makes its accepted fraction
100%. This **oracle uses exposed labels**; it is evidence of candidate coverage,
not a deployable selector. Three accepted aliases still have unresolved supplied
edits, so even the oracle does not establish advice safety.

This screen tested a distinct, frozen local semantic representation of candidate
diagnoses, exact original source context, and retention explanations. It made
no hosted model calls, changed no detector code, and did not alter labels.
Leave-one-repository-out AUC on 2,101 whole-page candidates was 0.70928. The
fixed 0.5 threshold retained 1,164/1,386 accepted findings (84.0%) and 39/57
FULL events, with three unresolved supplied edits. The optimistic post-hoc
threshold envelope at 85% acceptance reached only 36/57 FULL events:
1,078/1,268 accepted (85.02%), again with three unresolved edits. Contextual
and historical candidate banks were left unfiltered at 39/46 and 17/20 FULL.

The fixed policy misses both whole-page gates: at least 46/57 FULL and 85%
accepted findings. This exact encoder/feature/classifier configuration is
retired on these exposed pages. No threshold adjustment or additional fitting
on the same labels is justified. The result does not test new-page transfer,
standalone candidate generation, or advice safety. Issue #349 remains open.

`PROTOCOL.md`, `scan.py`, and the local checkpoint identities were frozen in
`freeze.json` before inference. `result.json` and the separate label-only
`oracle.json` contain only aggregates. The original 36 pages and grades remain
in the private archive. Run
`python3 verify_result.py` and `python3 -m unittest test_verify_result.py` to
check the public artifacts and negative tests. The exact replay additionally
requires the SHA-pinned private development archive and local APE checkpoint;
the commands are:

```sh
HF_HUB_OFFLINE=1 TRANSFORMERS_OFFLINE=1 python scan.py \
  --development-root /path/to/archived/development \
  --model /path/to/pinned/ape-model \
  --backbone /path/to/pinned/bge-backbone \
  --output /tmp/semantic-admission-replay.json
cmp result.json /tmp/semantic-admission-replay.json
python oracle.py --development-root /path/to/archived/development \
  --output /tmp/accepted-only-oracle-replay.json
cmp oracle.json /tmp/accepted-only-oracle-replay.json
```

The next experiment must start from new page-disjoint sources and a new
predeclared mechanism for source-specific phrase function and retention. Its
complete stream must be checked separately for recall, accepted fraction,
abstention, and edit safety before any product change.
