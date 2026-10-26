# Diagnosis-first contextual pilot

This folder records a completed, negative exposed-development pilot for
[#349](https://github.com/stokaro/unswell/issues/349). See [RESULTS.md](RESULTS.md)
for the measured result, [DESIGN.md](DESIGN.md) for the original hypothesis and
limitations, and [protocol.json](protocol.json) for the execution contract
frozen before model calls. The prior authorization covered only that terminal
run; no further model execution is authorized by this archive.

The preparer copies exact source bytes and engine unit spans from the earlier
frozen packet. It selects the six exposed context pages without sending cohort
names or labels to the model. Preparation, verification, and tests make zero
model calls. The default runner command is an offline plan; a new explicit
authorization is required before the guarded live flag may be used.

The prepared packet contains seven requests covering all 310 original units on
those six pages. Its request JSON totals 312,442 bytes before prompt overhead;
this is a byte count, not a token or monetary estimate.
It also retains the hashed original manifest so offline scoring can count all
36 source pages, including the 30 pages unavailable to this pilot.

## Offline reproduction

From the repository root:

```sh
study=research/reviews/2026-10-26-diagnosis-first
prior=research/reviews/2026-09-26-contextual-evidence
scratch=$(mktemp -d)
tar -xzf "$prior/evidence.tar.gz" -C "$scratch" packet
python3 -B "$study/diagnosis.py" "$scratch/packet" "$scratch/diagnosis-first"
python3 -B "$study/execute.py" "$scratch/diagnosis-first"
python3 -B -m unittest discover -s "$study" -p 'test_*.py' -v
```

The runner command above verifies the frozen packet and prints its offline
plan. It never launches a model without both a new explicit user authorization
and `--authorize-model-calls --output NEW_DIRECTORY`. Output directories cannot
be reused. The packet and evaluator freeze bind the unchanged source packet,
old source binder, original 123 reference events, new code, prompt, and protocol.

The previous accounting run and all earlier judgments remain unchanged.
A passing pilot would have required a separately authorized full 36-page screen
and then separate prospective confirmation. This pilot failed, so it does not
justify either step. No runtime, gate, or API changes follow from this archive.
