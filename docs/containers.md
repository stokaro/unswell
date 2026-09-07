# CLI and MCP containers

Unswell provides two separate Linux images. Both contain the same engine, English
model and grammars and run without runtime downloads. Release images support
`linux/amd64` and `linux/arm64`.

| Interface | GHCR | Docker Hub mirror |
| --- | --- | --- |
| CLI commands, filesystem discovery and reports | `ghcr.io/stokaro/unswell` | `docker.io/cabyrc/unswell` |
| MCP over stdin/stdout, with source bytes supplied by the client | `ghcr.io/stokaro/unswell-mcp` | `docker.io/cabyrc/unswell-mcp` |

The release tag is the version without its leading `v`, such as `0.1.0-alpha.1`.
Version `0.1.0-alpha.1` is public in both registries. Each mirror preserves the
GHCR image index, both architectures, and its provenance/SBOM attestations. Release
evidence records immutable image digests; pin a digest in reproducible installations.
Either registry name works in the examples below.

## Run the CLI

With an explicit Docker context, mount a repository at `/work` and keep it read-only:

```sh
docker --context remote-dev-container run --rm --read-only --network none \
  --mount type=bind,source=/absolute/path/on/daemon/repository,target=/work,readonly \
  ghcr.io/stokaro/unswell:0.1.0-alpha.1 \
  check . --config .unswell.yaml --report json:-
```

Bind paths belong to the Docker daemon's machine. A remote context cannot mount a
directory from the client's machine. For stdin checks, pipe the draft into `run -i`
and use `check --stdin --filename draft.md`; no source mount is required unless a
custom policy is also needed. Exit codes remain 0 for success, 1 for a policy
violation and 2 for an operational error.

The CLI image includes Git for tracked-file discovery. Its system Git configuration
trusts the explicit `/work` mount when host ownership differs from the container's
UID. Git discovery disables filesystem-monitor commands. Reports can go to stdout;
file reports need a separately mounted writable destination.

## Connect an MCP client

Use `docker` as the client's command with these arguments:

```json
[
  "--context", "remote-dev-container", "run", "--rm", "-i",
  "--read-only", "--network", "none",
  "ghcr.io/stokaro/unswell-mcp:0.1.0-alpha.1"
]
```

Keep stdin open with `-i` and omit `-t`: stdout carries the MCP protocol. Diagnostics
go to stderr. The image uses a `scratch` runtime and contains no shell or Git.
To select a policy, mount that file read-only and append `--config /path/in/container`.
Clients submit draft bytes to `unswell_check`; document names are labels, not paths
that the server opens. See the [MCP tools and outcomes](mcp.md).

Both images run as UID/GID 65532. Mounted source and policy files must be readable
by that user. Neither interface requires a listening port.

## Build and verify

Build each Dockerfile from the repository root with an explicit Docker context and
separate tags. Base images are pinned by digest; Dependabot proposes updates.
The Go build stage targets the requested architecture without cgo. The CLI keeps
its Git runtime, while the MCP runtime contains only the executable and notices.

```sh
docker --context remote-dev-container build -f Dockerfile -t unswell-local:cli .
docker --context remote-dev-container build -f mcp/Dockerfile -t unswell-local:mcp .
bash scripts/check-containers.sh remote-dev-container unswell-local:cli unswell-local:mcp
```

The check script transfers tracked files into a private volume, checks this repository
through the CLI, and submits identical bytes to a real MCP process. It verifies
matching policy and engine results, deliberate violations, a clean rewrite and
malformed source. Both containers run read-only with networking disabled. The script
removes its containers and volume; remove the named build images when finished.

CI runs these checks on native AMD64 and ARM64 Linux runners. A tag release waits
for CI, publishes both images with SBOM and provenance attestations, then pulls them
by digest on fresh runners without registry credentials and repeats the checks.
Public download or protocol failures stop the archive publication job. Container
results and immutable image references remain attached to the workflow run.

## Docker Hub mirroring

The release workflow copies the verified GHCR indexes to Docker Hub with pinned
Skopeo, without rebuilding them. It reads `DOCKER_HUB_USER` and `DOCKER_HUB_TOKEN`
from repository Actions secrets. The token reaches `skopeo login` through stdin;
the authfile stays in a temporary container tmpfs. The script removes the container
and any tool image it had to pull.

The mirror gate fetches both registries anonymously, requires identical index
bytes, and checks that both Linux architectures and attestations are present.
Missing, differing, malformed or incomplete indexes fail. The fetched indexes are
retained as workflow artifacts. Archive publication waits for this gate.

Retry the `Mirror containers to Docker Hub` workflow on `main` with an existing
image version. Its input accepts the version with or without a leading `v`.
To check the current public mirrors without credentials or a Docker daemon:

```sh
bash scripts/check-image-mirrors.sh 0.1.0-alpha.1 cabyrc
```

`make check` includes the mirror gate's offline positive and negative cases.
