FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine@sha256:cf6fca6641884b8433441b2b0652976f975e1d0fdd26d177eaaf8596087f3125 AS build
WORKDIR /src
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
COPY go.mod go.sum ./
COPY mcp/go.mod mcp/go.sum ./mcp/
RUN cd mcp && go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG BUILD_COMMIT=development
RUN cd mcp && GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -buildvcs=false \
    -ldflags "-s -w -X github.com/stokaro/unswell.BuildCommit=$BUILD_COMMIT" \
    -o /out/unswell-mcp ./cmd/unswell-mcp

FROM scratch
ARG VERSION=development
ARG BUILD_COMMIT=development
LABEL org.opencontainers.image.title="Unswell MCP" \
      org.opencontainers.image.description="Offline AI draft self-checking through the Unswell engine and MCP." \
      org.opencontainers.image.source="https://github.com/stokaro/unswell" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.version=$VERSION \
      org.opencontainers.image.revision=$BUILD_COMMIT \
      io.modelcontextprotocol.server.name="io.github.stokaro/unswell"
COPY --from=build /out/unswell-mcp /usr/local/bin/unswell-mcp
COPY licenses /licenses
COPY LICENSE THIRD_PARTY_NOTICES.md /licenses/
WORKDIR /work
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/unswell-mcp"]
