# Contributing

Discuss substantial behavior or API changes in an issue before implementing them.
Use American English and keep changes focused. Run `make check`; concurrency
changes also require `make race`. CI tests the minimum compiler from `go.mod` on
Linux, macOS, and Windows with automatic toolchain upgrades disabled. Tools have
their own pinned module and compiler requirements.

Add every Go module to `.gomodules` with its role. Add every public package to
`docs/public_api.md`. Keep library code independent of the CLI, environment,
filesystem discovery, network, and process execution. Repository policy tests
include deliberately invalid fixtures; preserve their ability to reject changes.

Rules need source-coordinate tests, realistic counterexamples, and documented
limitations. A few examples do not establish precision. Experimental rules become
stable only after the evaluation described in `docs/roadmap.md`.

Update the API snapshot with the pinned `apidiff -m -w docs/api/alpha.api
github.com/stokaro/unswell` executable from the root module. CI compares against
the preceding release and rejects incompatible changes. The first alpha bootstraps
that history with its reviewed snapshot. Do not update snapshots to conceal a break.

Release tags are immutable. Release preparation requires green checks for the
exact commit, an updated changelog, license notices, and reviewed artifacts. The
release workflow runs the native and quality checks again before publishing.

Pull requests can modify their own workflow. Branch protection and a maintainer's
review are the trust boundary; repository scripts cannot prevent an authorized
maintainer from replacing them. No pull request receives release credentials.
