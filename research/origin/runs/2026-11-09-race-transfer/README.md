# RACE external transfer screen

The [protocol](PROTOCOL.md) tests a published four-class origin checkpoint on
60 ARB human sources and 240 paired LLM rewrites from three source domains.
The score and advance rule were fixed before inference. This is a feasibility
screen on an exposed public corpus, not a product origin model or an editorial
[#349](https://github.com/stokaro/unswell/issues/349) result.

## Result: retire this checkpoint

The single V2 inference pass scored all 300 selected documents. Seventeen
exceeded the checkpoint's 512-token limit and were truncated by RACE. The
predeclared machine-involvement index gave pooled source-matched AUROC **0.692**
(source-bootstrap 95% interval 0.581–0.795) and matched-pair win fraction
**0.683** (interval 0.567–0.792). The advance rule required at least 0.75
AUROC, at least 0.70 paired wins, and at least 0.60 paired wins in every domain.
It failed the pooled thresholds and the domain floor.

| ARB source domain | Paired rewrites | AUROC | Paired win fraction |
| --- | ---: | ---: | ---: |
| `owt` | 80 | 0.176 | 0.150 |
| `wp` | 80 | 0.876 | 0.900 |
| `xsum` | 80 | 1.000 | 1.000 |

The `owt` ordering reverses the intended direction. The large domain spread
precludes treating the strong `wp` or `xsum` subsets as a transfer pass. The
unchanged full-document richness comparator had AUROC 0.643 and paired-win
fraction 0.706 on the same 240 pairs. This comparison is descriptive: the
small AUROC difference does not qualify either detector, and the comparator's
paired-win fraction alone does not satisfy the RACE advance rule.

The [aggregate result](result.json), [inference journal](infer-journal.json),
and [hash manifest](run-summary.json) preserve the outcome. The
post-result [independent verifier](verify_result.py) recomputes the frozen
scorer, brute-force AUROC from the saved logits, domain AUROCs, the original
selection, and the integrity links. It confirmed the negative result. The
published paper's HART results and this ARB transfer use different corpora and
split rules; neither the HART test accuracy nor this screen establishes a
1% human false-positive operating point, software-document transfer, or
mixed-origin span detection.

`prepare.py` selects only test-role sources from the previously frozen ARB
split. `selection.json` pins all 60 source IDs, 300 raw-text hashes, and the
untracked `input.jsonl` SHA-256. `parse.py` invokes the pinned English RST
parser and `score.py` applies the predeclared checkpoint score. `freeze.json`
pins code, source, model, and environment identities. None of the raw texts
or per-document predictions belongs in Git.

The first freeze stopped before inference on a mistaken tokenizer hash. The
[V1 preflight record](preflight-v1.json) preserves the exact mismatch. V2
corrects the hash without changing the selected documents or decision rule.

Reproduce the local input and integrity checks with the pinned public ARB
file and an environment containing `duckdb==1.4.2`:

```sh
python3 -B research/origin/runs/2026-11-09-race-transfer/prepare.py \
  --arb /path/to/arb.parquet --output /private/tmp/unswell-race-input
python3 -B research/origin/runs/2026-11-09-race-transfer/verify_freeze.py \
  --arb /path/to/arb.parquet --input /private/tmp/unswell-race-input/input.jsonl
python3 -B -m unittest discover \
  -s research/origin/runs/2026-11-09-race-transfer -p 'test_*.py'
```

To verify the recorded decision from a completed execution, supply the local
`input.jsonl`, `parse-status.json`, and `test_predictions.json` retained with
their hashes. From the repository root:

```sh
python3 -B research/origin/runs/2026-11-09-race-transfer/verify_result.py \
  --arb /path/to/arb.parquet \
  --input /path/to/input.jsonl \
  --status /path/to/parse-status.json \
  --predictions /path/to/test_predictions.json \
  --journal research/origin/runs/2026-11-09-race-transfer/infer-journal.json \
  --result research/origin/runs/2026-11-09-race-transfer/result.json \
  --summary research/origin/runs/2026-11-09-race-transfer/run-summary.json
```

Raw source texts and per-document logits remain local. The bounded remote CPU
container and downloaded model assets were temporary research infrastructure.
No Unswell runtime code, policy, gate, or editorial label changed.
