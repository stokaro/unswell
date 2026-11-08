# Real-source admission transfer v1: environment stop

The offline packet was frozen and committed in `71a3f83` before execution.
It contained 24 source-bound real-text requests from 16 complete exposed pages,
with 12 archived accepted and 12 archived uncertain diagnoses. The exact
request prompt and JSON response schema were inherited unchanged from the
synthetic evidence-flip pilot. The packet freeze SHA-256 is
`bd9f5787fa912e12b1a0fd1598702fef5594a220f67286e5e2a039c33a931967`.

The first `codex exec` process exited 1 in 0.233 seconds. Its stderr reports
`failed to initialize in-process app-server client: Operation not permitted`
after a PATH-alias warning. The event stream is empty, there is no response
file, and no usage is recorded. Per the frozen stop policy, this run ended
without a retry. The remaining 23 requests were not attempted. Its result
records zero valid calls and no diagnostic-quality measurement. The original
123 events, old assistant grades, and #349 thresholds are unchanged.

A separate v2 protocol may use the same frozen source and diagnosis inputs
with an execution environment that permits Codex CLI initialization. Such a
run needs a new pre-result freeze and its own no-retry journal; v1 remains
terminal. This operational stop neither supports nor refutes transfer of the
admission method to real pages.
