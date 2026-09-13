# Installation and distribution verification

Unswell is distributed as a Go library, CLI archives, two container images, a
Homebrew formula and a GitHub Action. The MCP image is also described in the
official MCP Registry. These routes all use the same engine and policy model.

The verified release is `v0.1.0-alpha.3`: its archives, containers, MCP Registry
entry, tap formula, and Action tag are published. The
[alpha distribution evidence](alpha-distribution-evidence.md) records the source
commit and checks for that release. Later changes on `main` need their own release
verification.

## Homebrew

The dedicated tap is [stokaro/homebrew-unswell](https://github.com/stokaro/homebrew-unswell).
Install the published release with:

```sh
brew install stokaro/unswell/unswell
brew test stokaro/unswell/unswell
```

The formula selects a checked release archive for macOS or Linux on ARM64 or
AMD64. Use `brew install --HEAD stokaro/unswell/unswell` to build `main` with Go.
Its tests require clean prose to pass, bad
prose to return 1, and malformed C# to return 2. The installed executable also
checks the tap's own Markdown, YAML and Python scripts in CI.

The tap's current `Update formula` workflow accepts an existing release tag,
verifies all four Unix archive hashes, and automatically opens a formula PR
through the publishing app. Installation checks run on the PR. A maintainer
reviews and merges it after the required checks pass. This cycle was verified for
alpha.3 in [#66](https://github.com/stokaro/unswell/issues/66).

## GitHub Action

The action lives in [stokaro/unswell-action](https://github.com/stokaro/unswell-action).
Pin an action commit and an independent CLI version:

```yaml
- uses: stokaro/unswell-action@b7e5d5ce69d69ca458879c938726ec0870457b35
  id: unswell
  with:
    version: 0.1.0-alpha.3
    config: .unswell.yaml
    paths: |
      README.md
      docs
      src
```

The CLI release must already exist. The action verifies the downloaded archive's
SHA-256, passes each input as a process argument, and preserves pass/failure/error
outcomes. It exposes JSON and SARIF file paths as outputs. Analysis remains offline
after installation. The action README documents inputs, report retention and the
policy trust boundary.

The release workflow checks Unswell itself through this public action on Linux,
macOS and Windows after archive publication. Also run the action repository's
`Verify published release` workflow with the version. That consumer workflow proves
that bad prose and malformed source fail their action steps, with the expected
codes and report contents. Unit tests against local archives do not replace this
public download check.

The Action updater verifies all six archives and automatically opens a PR for the
default CLI version. A maintainer reviews and merges it; tag publication follows
successful checks on the merged commit. The alpha.3 tag points to the commit in
the example above. Existing release tags are not moved.

## Archives, containers and MCP Registry

CLI release archives cover Linux, macOS and Windows on AMD64 and ARM64. Verify
`SHA256SUMS` before extracting an archive. Archives include dependency notices;
per-platform CycloneDX SBOMs accompany them. The runtime requires neither Go nor cgo.

The [container documentation](containers.md) describes the distinct CLI and MCP
images, explicit Docker contexts, source mounts and stdio operation. Container CI
checks native Linux AMD64 and ARM64 builds with read-only files and no runtime
network. Release verification pulls the published digests without credentials.

The [MCP Registry workflow](mcp-registry.md) publishes `io.github.stokaro/unswell`
after those image checks and archive publication, then verifies the exact active
version through the public registry API. It uses the separate MCP image.

## Release acceptance

Keep the [roadmap issues](roadmap.md) open until their public acceptance evidence
exists. A release needs the native and quality gates, CLI/MCP repository checks,
both container architecture checks, anonymous pulls, registry verification,
Homebrew installation results and public Action consumer results. Preserve those
workflow artifacts alongside the release's immutable version and source commit.

## Publishing app setup

`Update distributions` dispatches release updates to the tap and Action
repositories. The alpha.3 cycle completed in
[tap PR #4](https://github.com/stokaro/homebrew-unswell/pull/4) and
[Action PR #6](https://github.com/stokaro/unswell-action/pull/6), followed by their
main-branch checks and Action tag publication. Repeated update requests made no
new PR or change; the [evidence record](alpha-distribution-evidence.md) links those
runs. This verifies alpha.3, not a future release.

The workflow also accepts an existing release tag to retry delivery. Repeated
requests preserve an already selected release. An existing update branch must
match the regenerated files before reuse; unrelated changes must not be
overwritten.

The organization variable `PUBLISH_APP_ID` and secret `PUBLISH_APP_KEY` are
available to Unswell, the tap, and the Action repository, and the app has
Contents and Pull requests write access. Each job requests a short-lived
installation token limited to its target repositories and required permissions.

The app opens the update pull request; it never merges one. Required checks run
on that pull request without anyone touching it, and a code-owner entry routes
the review request to a maintainer, who approves and squash-merges. Auto-merge
stays disabled and main's review bypass list stays empty in both receiving
repositories, so the app is subject to the same approving review as any other
author.

`scripts/verify-release-artifacts.sh` audits a published release. It checks every
asset against the release manifest. It checks that each archive bundles the
license and the notices, that each build record names a trimmed build without
cgo, and that each archive has its own SBOM. With `--rebuild` and the released
commit it rebuilds each binary and compares digests. The recorded result is in
[reproducible builds](reproducible-builds.md).
