# RACE external transfer screen

The [protocol](PROTOCOL.md) tests a published four-class origin checkpoint on
60 ARB human sources and 240 paired LLM rewrites from three source domains.
The score and advance rule are fixed before inference. This is a feasibility
screen on an exposed public corpus, not a product origin model or an editorial
[#349](https://github.com/stokaro/unswell/issues/349) result.

`prepare.py` selects only test-role sources from the previously frozen ARB
split. `selection.json` pins all 60 source IDs, 300 raw-text hashes, and the
untracked `input.jsonl` SHA-256. `parse.py` invokes the pinned English RST
parser and `score.py` applies the predeclared checkpoint score. `freeze.json`
pins code, source, model, and environment identities. None of the raw texts
or per-document predictions belongs in Git.

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

The remote CPU container and downloaded model assets are temporary research
infrastructure. They will be removed after the bounded run and artifact
verification. No Unswell runtime code, policy, gate, or editorial label changes.
