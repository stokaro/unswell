# Installation and distribution verification

Unswell is distributed as a Go library, CLI archives, two container images, a
Homebrew formula and a GitHub Action. The MCP image is also described in the
official MCP Registry. These routes all use the same engine and policy model.

The first alpha archives, containers, MCP Registry entry, and Action tag are
published. The tap's versioned formula is merged on main.
See the [alpha distribution evidence](alpha-distribution-evidence.md) for the
verified versions and remaining automation work.

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
verifies all four Unix archive hashes, prepares a formula branch, and starts its
installation checks. A maintainer creates and merges the PR after those checks
and the required review pass. Automatic PR creation and merging are still tracked
in [#66](https://github.com/stokaro/unswell/issues/66).

## GitHub Action

The action lives in [stokaro/unswell-action](https://github.com/stokaro/unswell-action).
Pin an action commit and an independent CLI version:

```yaml
- uses: stokaro/unswell-action@5c4e109d71c1ec16429a62ac1860b97baee3974e
  id: unswell
  with:
    version: 0.1.0-alpha.1
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

The proposed Action updater verifies all six archives and opens a PR for the
default CLI version. Its implementation and automatic tag publication are still
under review in [Action PR #2](https://github.com/stokaro/unswell-action/pull/2).
Do not rely on automatic updates until #66 records a successful release cycle.

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

The following setup belongs to the automation proposed in #66. The receiving
workflows are still under review in
[tap PR #2](https://github.com/stokaro/homebrew-unswell/pull/2) and Action PR #2.
The checked-in `Update distributions` workflow alone does not prove delivery.
Its successful end-to-end execution remains an acceptance requirement.

After those workflows and settings are available, `Update distributions` accepts
an existing release tag to retry delivery. Repeated requests must preserve an
already selected release. An existing update branch must match the regenerated
files before reuse; unrelated changes must not be overwritten.

The organization variable `PUBLISH_APP_ID` and secret `PUBLISH_APP_KEY` must be
available to Unswell, the tap, and the Action repository. The app needs Contents
and Pull requests write access. Each job requests a short-lived installation
token limited to its target repositories and required permissions.

Enable auto-merge in the tap and Action repositories. Their main-branch review
rules allow only the publish app to merge without manual approval; required CI
checks and resolved conversations still apply. Keep one approving review for
other authors and keep force pushes and main deletion disabled. Organization-wide
permission for GitHub Actions to approve PRs is not needed.
