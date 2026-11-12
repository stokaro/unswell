FROM --platform=$BUILDPLATFORM golang:1.27.4-alpine@sha256:0a5f7fe796599ff487605987432e7619ae45beeb2fba4c2ce0b907879932aebf AS build
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
FROM alpine/git@sha256:6a1fdb962814ca11508ca8d1bab94375b5b7ac4e4a9a40981c29fc32c02f095d
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
