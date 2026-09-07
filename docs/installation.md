# Installation and distribution verification

Unswell is distributed as a Go library, CLI archives, two container images, a
Homebrew formula and a GitHub Action. The MCP image is also described in the
official MCP Registry. These routes all use the same engine and policy model.

The first alpha is being prepared. Source builds and the tap's `--HEAD` formula
are available before the versioned archives. A workflow or repository existing
does not mean its corresponding release artifact has been published.

## Homebrew

The dedicated tap is [stokaro/homebrew-unswell](https://github.com/stokaro/homebrew-unswell).
Install the current public source with:

```sh
brew install --HEAD stokaro/unswell/unswell
brew test stokaro/unswell/unswell
```

The development formula builds `main` with Go. Once the versioned formula is
published, `brew install stokaro/unswell/unswell` selects a checked release archive
for macOS or Linux on ARM64 or AMD64. Its tests require clean prose to pass, bad
prose to return 1, and malformed C# to return 2. The installed executable also
checks the tap's own Markdown, YAML and Python scripts in CI.

After publishing an Unswell release, run the tap's `Update formula` workflow with
that tag. The updater checks all four archive hashes, pushes a formula branch and
starts installation CI. Its summary links to the comparison for creating a PR.
Review the generated formula and complete all installation checks before merging.

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
