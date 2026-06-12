# Gjallarhorn builds two binaries (server + migrator) into one distroless image.
# Standalone repo: the Go module lives at the repo root and go-kit is a
# published dependency (no replace directive), so the build context is this
# repo root. GOWORK=off ignores any ambient go.work.
ARG GO_VERSION=1.25
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder
ARG TARGETOS TARGETARCH
WORKDIR /src
ENV GOWORK=off

# Resolve dependencies first for better layer caching.
COPY go.mod go.sum ./
RUN go mod download

# Build the service from the module root.
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -o /out/server   ./cmd/server
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -o /out/migrator ./cmd/migrator

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/server   /app/server
COPY --from=builder /out/migrator /app/migrator
# 8080 = REST/OpenAPI, 9090 = gRPC
EXPOSE 8080 9090
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
