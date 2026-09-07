# MCP Registry publication

The official server name is `io.github.stokaro/unswell`. Its
[registry metadata](../mcp/server.json) points to the separate
`ghcr.io/stokaro/unswell-mcp` image with stdio transport. The image declares the same
server name in its `io.modelcontextprotocol.server.name` label.

This entry lets clients discover the offline checker for AI-style wording. The
server provides `unswell_check` for draft bytes and `unswell_describe` for the fixed
policy. It does not host a public HTTP endpoint. See [MCP configuration](mcp.md)
and [container commands](containers.md), including explicit Docker contexts.

## Release sequence

The tag, Go version constant, registry version and OCI image tag must agree. Update
`result.go` and `mcp/server.json` together when preparing a new version.

1. CI validates the metadata and proves that an invalid server name is rejected.
2. The release builds both container architectures, publishes the CLI and MCP
   images, and verifies anonymous pulls and real protocol behavior by image digest.
3. The archive publication job succeeds before registry publication starts.
4. The registry job authenticates through GitHub OIDC and publishes `mcp/server.json`.
5. A public API read checks the exact name, version, active status, repository and
   OCI package. The response is retained as workflow evidence.

The publisher is pinned in [.mcp-publisher-version](../.mcp-publisher-version).
Its setup script checks the release archive's recorded SHA-256 before extraction.
Updating the publisher requires updating the version and matching platform hashes.
The workflow uses `id-token: write` for short-lived OIDC authentication and removes
the publisher's saved credentials afterward. No personal access token is required.

## Validate before publishing

```sh
bash scripts/setup-mcp-publisher.sh
make check-registry
```

These commands download the pinned publisher and use its registry validator. They
require network access and do not publish an entry. The usual `make check` remains
independent of this registry service; CI runs both checks.

After publication, verify the public record again with:

```sh
bash scripts/verify-registry.sh
```

A missing version, mismatched package or inactive entry fails verification. Files
and workflows alone do not establish publication: a release is discoverable only
after this public check succeeds. The registry is currently a preview service, so
preserve release metadata and verify the entry after any upstream data reset.

The workflow follows the registry's [GitHub Actions publication guide](https://github.com/modelcontextprotocol/registry/blob/main/docs/modelcontextprotocol-io/github-actions.mdx)
and [server metadata specification](https://github.com/modelcontextprotocol/registry/blob/main/docs/reference/server-json/generic-server-json.md).
