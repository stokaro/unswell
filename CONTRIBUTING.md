# Contributing

Discuss substantial behavior or API changes in an issue before implementing them.
Use American English and keep changes focused. Run `make check` before publishing.
CI tests the minimum compiler from `go.mod` on
Linux, macOS, and Windows with automatic toolchain upgrades disabled. Tools have
their own pinned module and compiler requirements.

Race detection and active fuzzing are temporarily deferred during alpha development,
including concurrency changes. Preserve their tests, seed corpora, and the
`make race` / `make fuzz` targets; do not run them during routine implementation.
The final roadmap task, [#123](https://github.com/stokaro/unswell/issues/123), restores
both CI checks and fixes any findings before final product acceptance. Ordinary
tests still execute the existing fuzz seeds.

Run `bash scripts/setup-shellcheck.sh` once to install the pinned ShellCheck build
in `bin/`. The installer verifies the archive checksum. `make lint-shell` checks
Bash syntax, ShellCheck diagnostics through the style level, and shfmt formatting.
It discovers tracked and new Bash scripts, including scripts without extensions.
ShellCheck also checks for hidden command failures and suppressed `set -e` behavior.
The gate rejects missing tools and proves it catches three deliberately bad scripts.
To format scripts, use the pinned executable from `cd tools && go tool -n shfmt`
with `-w -ln bash -i 2 -ci`. Shell comments and quoted strings also pass the
grammar-based prose check with the committed extraction policy.

`make check` includes `make dogfood-mcp`: the just-built CLI checks this repository
using `.unswell.yaml`. Edit the prose when a finding is valid. Investigate a false
positive with a realistic regression case; do not disable a rule just to make CI
green. A real MCP client checks the same bytes and compares complete results, then
verifies failure and rewrite probes. Reports are written to `artifacts/dogfood/`.

The root [e2e package](e2e/README.md) runs the built CLI against readable source
fixtures with inline expected diagnostics and golden reports. Run
`go test ./e2e -count=1` to inspect these scenarios independently. Review fixture
expectations and golden diffs when behavior changes; ordinary tests never update them.

Add every Go module to `.gomodules` with its role. Add every public package to
`docs/public_api.md`. Keep library code independent of the CLI, environment,
filesystem discovery, network, and process execution. Repository policy tests
include deliberately invalid fixtures; preserve their ability to reject changes.

Rules need source-coordinate tests, realistic counterexamples, and documented
limitations. A few examples do not establish precision. Experimental rules become
stable only after the evaluation described in `docs/roadmap.md`.

Alpha APIs and artifact formats may change without backward compatibility. Update
the current documentation, tests, and consumers when changing a contract. Reject
unsupported artifact formats explicitly. API snapshots and comparisons with previous
releases are not required; current schemas and library boundaries remain checked.

Release tags are immutable. Release preparation requires green checks for the
exact commit, an updated changelog, license notices, and reviewed artifacts. The
release workflow runs native, quality and container checks before publishing.
Release builds use the pinned `toolchain` from `tools/go.mod` and require an empty
`dist/` directory. The runtime's minimum compiler remains independently tested.
The separate tap and action repositories have their own installation CI. Follow
[distribution verification](docs/installation.md) before declaring a release ready.

Pull requests can modify their own workflow. Branch protection and a maintainer's
review are the trust boundary; repository scripts cannot prevent an authorized
maintainer from replacing them. No pull request receives release credentials.
