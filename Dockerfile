# syntax=docker/dockerfile:1.7

##############################
# Builder
##############################
FROM golang:1.21-alpine AS builder

WORKDIR /src

# Install build deps and CA certs for fetching modules
RUN apk add --no-cache ca-certificates tzdata bash git

# Leverage Docker layer cache for deps
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source
COPY . .

# Build static binary
ENV CGO_ENABLED=0
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags "-s -w" -o /out/webhook-demo ./

##############################
# Runtime
##############################
FROM alpine:3.20

WORKDIR /app

# Install runtime deps (SSL certs, timezone, minimal curl for healthcheck)
RUN apk add --no-cache ca-certificates tzdata curl \
    && adduser -D -g '' appuser \
    && chown -R appuser /app

COPY --from=builder /out/webhook-demo /app/webhook-demo

# Optional example config (not used automatically)
COPY config.env.example /app/config.env.example

# Defaults; override at runtime if needed
ENV GIN_MODE=release \
    SERVER_PORT=8080

EXPOSE 8080

# Simple container health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD curl -fsS http://127.0.0.1:8080/health || exit 1

USER appuser

ENTRYPOINT ["/app/webhook-demo"]


