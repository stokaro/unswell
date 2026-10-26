# Diagnosis-first pilot result

This is the terminal result of the seven-request exposed-context pilot for
[#349](https://github.com/stokaro/unswell/issues/349). The candidate failed its
frozen pilot screen. It does not establish a full development result or product
qualification.

The explicitly authorized GPT-6-Astra medium run completed all seven requests
with valid bound responses. All six context pages were complete, and all 310
focus units received an assessment. The assistant reviewer identified in
`review.json` examined all 38 emitted diagnoses against the original page text
and separately judged all 46 context events. The other 30 pages and 77 frozen
events were unavailable in this pilot, not semantic misses.

| Measure | Result | Frozen pilot minimum |
| --- | ---: | ---: |
| Complete context pages | 6/6 | 6/6 |
| FULL context events | 19/46 (41.3%) | 37/46 (80%) |
| Accepted diagnoses | 36/38 (94.7%) | at least 85% |
| Valid requests | 7/7 | 7/7 |
| Assessed focus units | 310/310 | 310/310 |

The model chose `no_burden` for 275 units and `diagnose` for 35; it chose
`uncertain` for none. Three events received PARTIAL credit and 24 were missed;
neither counts toward FULL recall. The high diagnostic acceptance fraction does
not compensate for these 27 incomplete events. In particular, source units on the first two
pages were often assessed as having no burden despite the per-unit instruction
to separate technical payload from editorial burden. Removing repair advice and
forcing two assessments per unit did not recover the required contextual recall.
The frozen threshold and reference events were not changed after the run.

`evidence.tar.gz` retains the exact frozen packet, raw model traces, stderr,
responses, bound results, and terminal run ledger. `artifact.json` binds the
archive and assistant review by SHA-256. `summary.json` is the evaluator output;
the evaluator checks original UTF-8 source spans, prompt identity, call order,
CLI flags, response bounds, raw artifact hashes, and all frozen event decisions.

For a local replay, extract the archive's `packet` and `run` directories into
the original absolute locations recorded in `run/run.json`: the packet at
`/private/tmp/unswell-diagnosis-first-packet-20261026` and the run at
`/private/tmp/unswell-diagnosis-first-run-20261026-authorized-01`. The current
evaluator records these execution paths as part of the immutable call identity.
Then run from the repository root:

```sh
python3 -B research/reviews/2026-10-26-diagnosis-first/evaluate.py \
  /private/tmp/unswell-diagnosis-first-packet-20261026 \
  /private/tmp/unswell-diagnosis-first-run-20261026-authorized-01 \
  --review research/reviews/2026-10-26-diagnosis-first/review.json
```

The result rejects this candidate. The full 36-page development screen and
separate prospective confirmation remain undone. No runtime detector, gate, or
public API was changed by this pilot.
