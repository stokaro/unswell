# Contextual unit accounting: prepared, not executed

This candidate addresses [#349](https://github.com/stokaro/unswell/issues/349).
It does not provide a new detection result. The previous failed Astra run and
its judgments remain unchanged. [DESIGN.md](DESIGN.md) records the hypothesis,
limitations, and ownership contract; [protocol.json](protocol.json) records the
proposed run and unchanged acceptance thresholds.

The 36 exposed pages become 88 requests covering 4,549 engine units exactly once.
Every request retains full-page context and exact raw unit slices. Each focuses
on at most 64 units and requires short editorial contribution assessments linked
to warranted diagnostics. Structural coverage is not treated as semantic recall.

The proposed execution is GPT-6-Astra medium through the current ChatGPT
subscription, sequentially, without tools, retries, or fallback models, for at
most four hours. It consumes subscription quota. Cost and token counts are not
known in advance. **New explicit authorization is required.** The previous
36-request experiment is terminal and does not authorize this run.

## Offline reproduction

From the repository root, use a new directory outside the checkout:

```sh
study=research/reviews/2026-09-27-contextual-accounting
prior=research/reviews/2026-09-26-contextual-evidence
scratch=$(mktemp -d)
tar -xzf "$prior/evidence.tar.gz" -C "$scratch" packet
python3 -B "$study/accounting.py" "$scratch/packet" "$scratch/accounting"
python3 -B "$study/execute.py" "$scratch/accounting"
python3 -B -m unittest discover -s "$study" -p 'test_*.py' -v
```

The default runner command only validates the frozen packet and prints the plan.
Tests use an in-process fake CLI. They do not send content to a model. Actual
execution additionally requires `--authorize-model-calls --output NEW_DIRECTORY`
after user approval. Output directories cannot be reused; no resume exists.

`evaluation-freeze.json` binds the packet, code, prompt, protocol, prior source
binder, prior scoring code, and all 123 exposed events. `freeze.py` creates a new
freeze only during preparation and refuses to overwrite one. It must not be used
to revise an experiment after execution. Prospective confirmation inputs retain
their separate role and are not present in model requests.

After a terminal authorized run, `evaluate_accounting.py PACKET RUN --new-review
REVIEW.json` creates an empty review ledger. After source-based review, use
`--review REVIEW.json` to compute results. The old scoring implementation remains
responsible for complete-page event credit and cohort thresholds. Partial-page
emissions have a separate, mandatory disposition and are also included in the
reported accepted fraction of all valid-request emissions. No contribution
assessment itself earns event credit.

The output contract distinguishes retain/revise/uncertain. The source binder
still rejects quotations outside their named unit and protected-only targets.
A cross-window finding belongs to its earliest targeted unit. Review must check
diagnosis, relevance, duplicate findings, and suggested changes independently;
schema validation cannot establish these properties.
