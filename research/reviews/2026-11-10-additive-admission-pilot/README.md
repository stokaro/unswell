# #349 additive source-bound admission: frozen source preparation

**Status: prepared, not executed.** No model call, product change, FULL-event
gain, or accepted-delivery improvement is claimed here. The [protocol](PROTOCOL.md)
registers a two-stage pilot: generate only new diagnoses from exact eligible
source units, then judge each proposal against its complete source without
withholding existing engine findings.

Four public README snapshots from Nushell, hyperfine, eza, and cargo-binstall
are pinned by commit, Git blob, and SHA-256 in [sources.json](sources.json).
The current technical-v1 exporter produced 565 eligible units across all four
complete pages. A source-only Codex assistant review under ADR 0041 froze four
editorial events and four retain controls in [reference.json](reference.json).
The sample is purposive and small; it is only a pilot of generation, source
binding, and admission behavior.

The source freeze includes the original UTF-8 files, complete exporter input
and output, source-only reference, protocol, verifier, and negative tests.
Verify it offline:

```sh
python3 -B research/reviews/2026-11-10-additive-admission-pilot/verify_source.py
(cd research/reviews/2026-11-10-additive-admission-pilot && \
  python3 -B -m unittest -q test_verify_source.py)
```

Before any model call, run `python3 -B run.py prepare`, commit the generated
`packet/` directory, and verify it with `python3 -B run.py verify`. The prompt,
schemas, runner, scorer, failure policy, and bounded budget are then frozen.
The bounded contract tests run with `python3 -B -m unittest -q test_run.py
test_score.py`. A later `RESULT.md` will record execution and review without
changing this source and packet freeze. The unchanged
36-page/123-event #349 development gates and separate prospective confirmation
remain required after this pilot. The original event labels are untouched.
