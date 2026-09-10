# Execute current CLI code in e2e checks

The module runner uses `go test -count=1` for runtime and consumer modules.
`make test`, `make race`, and native CI use this runner. `scripts/coverage.sh`
keeps its own loop, because each module needs its own profile path, and applies
the same `-count=1` rule to the instrumented run `make check` uses.
Additional test flags still pass through. The Go build cache remains enabled;
successful test results are not reused by these required checks.

Root e2e tests build research commands from nested modules using `os/exec`.
Those source files are outside the parent test package's import graph. Editing
one can leave the parent's cached success valid, even when the current command
would fail its golden test. Applying the fresh-run default across the module
ledger also covers other subprocess tests without duplicating Go's dependency
discovery. Repeated checks take longer because ordinary tests run again.

For a focused check, use:

```sh
go test ./e2e -run '^TestAnnotationProtocol$' -count=1
```

## Reproduce the stale result

Use an isolated checkout of `6fe8adf354376bea0d29e2c92c818253118463c9`, the parent
of this runner change, and preserve its Go cache between invocations. Do not
modify a working checkout that contains unrelated edits.

1. Run `go test ./e2e -run '^TestAnnotationProtocol$'` twice. The second pass uses
   the test-result cache.
2. In `research/annotation/cmd/annotate/main.go`, replace only
   `result, err = round.Packet(ctx)` with
   `return fmt.Errorf("cache regression probe")`. The command still compiles.
3. Repeat the unchanged test command. It reports a cached pass despite the broken
   packet operation.
4. Repeat with `-count=1`. The blinded-packet test fails and reports
   `stderr: cache regression probe`.
5. Apply this runner change and run
   `bash scripts/modules.sh test -run '^TestAnnotationProtocol$'`.
   The required runner propagates that failure.
6. Restore the command's exact original bytes and repeat the runner invocation.
   The focused protocol test passes.

The September 9, 2026 local reproduction produced:

| Input and command | Exit | Result |
| --- | --- | --- |
| Original command, warmup | 0 | Executed pass |
| Original command, repeat | 0 | Cached pass |
| Broken packet command, ordinary test | 0 | Stale cached pass |
| Broken packet command, `-count=1` | 1 | Blinded-packet failure |
| Broken packet command, updated required runner | 1 | Failure propagated |
| Restored packet command, updated required runner | 0 | Executed pass |

This probe establishes freshness and failure propagation for the real protocol
e2e. It does not replace the full suite or main-branch CI acceptance for #106.
The deliberate command failure is not part of the committed source.
