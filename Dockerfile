# syntax=docker/dockerfile:1
#
# trident-mcp container image (stdio MCP server):
#   docker run -i --rm -e TRIPO_API_KEY -v "$HOME/generated_models:/output" \
#     --user "$(id -u):$(id -g)" ghcr.io/mordor-forge/trident-mcp
#
# Artifacts are written to /output/<task_id>/. --user makes files in the
# bind-mounted directory belong to you. Local-path inputs must be mounted too.
#
# Build: docker build --build-arg VERSION=0.2.0 -t trident-mcp .

# Cross-compile on the build machine's architecture: no emulation for multi-arch builds.
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY cmd ./cmd
COPY internal ./internal

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath \
      -ldflags "-s -w -X main.version=${VERSION}" \
      -o /out/trident-mcp ./cmd/trident-mcp

# distroless has no shell to create the mount point, so make it here: owned by
# the default user and world-writable, so --user "$(id -u):$(id -g)" works too.
RUN mkdir -p /out/rootfs/output && chmod 1777 /out/rootfs/output

# distroless/static: CA certificates, tzdata and a non-root user (65532), nothing else.
FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=build --chown=65532:65532 /out/rootfs/ /
COPY --from=build /out/trident-mcp /usr/local/bin/trident-mcp
COPY LICENSE /usr/share/doc/trident-mcp/LICENSE

ENV MODEL_OUTPUT_DIR=/output
VOLUME ["/output"]
USER nonroot:nonroot

ENTRYPOINT ["trident-mcp"]
