# CLI startup notice compatibility amendment

The maintainer authorized GPT-6-Astra medium for up to 36 requests on September
27, 2026. The original frozen runner made one request and stopped. Codex CLI
0.157.1 emitted its `skip_host_skill_discovery` unstable-feature warning as an
`item.completed` event with item type `error`, before `turn.started`. The actual
model turn completed with exit code 0, no tool events, complete unit coverage,
and valid source-bound findings. The strict original trace validator rejected
the startup notice.

`cli_compat.py` accepts exactly that notice, at most once, between the single
initial `thread.started` event and `turn.started`. Any other error, tool event,
repeated notice, or misplaced notice remains invalid. The original trace and
failed run ledger are retained. The existing response is revalidated without a
new request. The adapter continues only the 35 unattempted pages and uses the
original output-directory creation time to preserve the two-hour deadline.

There is no change to the model, effort, page set, prompt, response schema,
source-binding rules, labels, or acceptance thresholds. No model request is
retried. This is a disclosed execution-compatibility amendment after one output,
not an unchanged preregistered execution. The original frozen 42-file packet
and evaluation files remain unchanged. The adapter, exact notice, original
ledger hash, and deadline are bound in the execution amendment before continuation.

Collection must use the compatibility trace checker explicitly and still run
all original evaluator checks. The original frozen evaluator on its own will
continue rejecting the notice; do not rewrite its history or silently replace
the raw output with filtered events. Three offline tests cover absent and exact
notices, genuine errors, tool events, duplicates, and misplaced notices.
