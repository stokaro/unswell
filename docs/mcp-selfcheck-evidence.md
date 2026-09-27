# MCP self-check failure evidence

This record addresses [#326](https://github.com/stokaro/unswell/issues/326).
It separates the original batch error from a later verification timeout.
Neither result measures editorial recall.

## Historical batch 25

The AMD64 artifact from [run 35278009093](https://github.com/stokaro/unswell/actions/runs/35278009093)
contains a complete CLI report for 1,076 documents and 878 findings. Its
`cli.json` SHA-256 is
`7b5f88403a126124d5df82fee48e212bee492649cf18859663ad68c5e7c9dd50`.
The reported build is `b3f77e87cddb5132e037c780f1d512a36aeeb90b`, and its policy
hash is `2ec4c15eac740349664d74a48cc27e0c91d8b49863481c64b7232b8eb0dc9ad4`.

A diagnostic replay used the batch planner and MCP server built from that exact
commit, its committed `.unswell.yaml`, and the retained source bytes. Discovery
verified the original policy and build identities. The planner produced 25
batches. The last batch contains only `workflow_sources_test.go`: 1,434 bytes,
zero prose words, source SHA-256
`bf8d57c6da4072f374234f9279665e247cac448f9ee371b2d8569d6a27fe7f41`.

The original server returned an operational error:

```text
scan contains no applicable English prose
```

The original checker required `pass` for every batch. The server correctly
applied `fail_on_empty` to the isolated code-only request. PR #342 already fixed
that expectation: an empty batch's error is checked against the discovered
policy, and its documents and measurements must still match. The public engine
continues to reject unexpected empty scans.

This was a source-build replay on macOS, not a claim that the historical Linux
container run became successful. A whole-report comparison on this host first
encountered last-bit floating-point differences in batch 4. To diagnose the
reported batch without weakening parity checks, a separate one-off probe used
the original planner to submit batch 25 directly. The [retained response](evidence/mcp-326-batch25.json),
including source identities and the actual error, has SHA-256
`83082b844ae613ca413ac9939d9fb551b8e0490c250236e77fefab4c321ae2cf`.

## Total verification deadline

A later [PR #352 run](https://github.com/stokaro/unswell/actions/runs/36328891930)
failed with `context deadline exceeded` in repository batch 63. A local parser
migration check reached batch 56 before the same deadline. The
[main run after #353](https://github.com/stokaro/unswell/actions/runs/36330808581)
finished the repository batches but exhausted the deadline at `mcp-negative.md`.

The developer checker imposed one fixed two-minute context on subprocess
startup, discovery, every batch, and the probes. CI now selects five minutes
explicitly through `--timeout`; the command retains its two-minute default and
rejects nonpositive values or values above 30 minutes. This changes the total
verification budget only. Server request limits, document coverage, normalized
result equality, and error exit codes remain enforced.

## Retained failure contract

Evidence identifies its schema, build, policy, and completeness. A failed run
retains matched batches, completed probes, the failed stage and request source
identities, and any actual structured response. A disconnected or timed-out
request has no invented response. Source bodies and snippets are removed.
The failed batch is separate from the matched batches. The command still exits
unsuccessfully, including when it cannot write the evidence file.

Blackbox tests start real checker and MCP subprocesses. They exercise a later
structured analysis error, a later transport disconnect, invalid deadlines,
and expiration before connection. They check nonzero exit codes, retained
partial results and hashes, and absence of source text. Native and container
acceptance still require fresh complete CLI/MCP evidence from the final commit.
