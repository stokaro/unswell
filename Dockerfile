FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine@sha256:cf6fca6641884b8433441b2b0652976f975e1d0fdd26d177eaaf8596087f3125 AS build
WORKDIR /src
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG BUILD_COMMIT=development
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -buildvcs=false \
    -ldflags "-s -w -X github.com/stokaro/unswell.BuildCommit=$BUILD_COMMIT" \
    -o /out/unswell ./cmd/unswell

# Git is required to preserve tracked-file discovery inside a mounted repository.
FROM alpine/git@sha256:6f3b5029566da8e90b24945933dcd806be866b64b1e706f51828bf84faccf21b
ARG VERSION=development
ARG BUILD_COMMIT=development
LABEL org.opencontainers.image.title="Unswell CLI" \
      org.opencontainers.image.description="Reduce AI-style wording in source code and documentation." \
      org.opencontainers.image.source="https://github.com/stokaro/unswell" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.version=$VERSION \
      org.opencontainers.image.revision=$BUILD_COMMIT
COPY --from=build /out/unswell /usr/local/bin/unswell
COPY licenses /licenses
COPY LICENSE THIRD_PARTY_NOTICES.md /licenses/
COPY packaging/gitconfig /etc/gitconfig
WORKDIR /work
USER 65532:65532
ENTRYPOINT ["/usr/local/bin/unswell"]
CMD ["--help"]
