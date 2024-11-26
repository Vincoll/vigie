# syntax=docker/dockerfile:1.4

# Advanced Dockerfiles are located in ./build/release
# Read ./build/release/ReadMe.md

# Multi-Stage Build
# (Go Builder + Source) => Final Image
# ARGS are passed via docker build --build-arg

# --- Build Image ---
ARG GO_VERSION=1.23
FROM golang:${GO_VERSION}-alpine AS builder

# Build arguments
ARG VIGIE_VERSION
ARG COMMIT
ARG DATE

# Install essential build dependencies
RUN apk add --no-cache --update \
    libcap \
    ca-certificates \
    && rm -rf /var/cache/apk/*

WORKDIR /src

# Copy Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -ldflags="-w -s \
        -X github.com/vincoll/vigie/cmd/vigie/version.LdVersion=${VIGIE_VERSION} \
        -X github.com/vincoll/vigie/cmd/vigie/version.LdBuildDate=${DATE} \
        -X github.com/vincoll/vigie/cmd/vigie/version.LdGitCommit=${COMMIT}" \
    -o /bin/vigie .

# Create non-root user
RUN addgroup -S -g 1001 vigie && \
    adduser -S -u 1001 -G vigie -h /home/vigie -s /sbin/nologin vigie && \
    chown vigie:vigie /bin/vigie

# --- Final Image ---
FROM alpine:3.19 AS final

# Security: Run as non-root user
COPY --from=builder /etc/passwd /etc/group /etc/
USER vigie

# Copy necessary files
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder --chmod=755 /bin/vigie /vigie

# Create and set proper permissions for app directory
WORKDIR /app
RUN mkdir -p /app/config

# Metadata
LABEL org.opencontainers.image.title="Vigie"
LABEL org.opencontainers.image.source="https://github.com/vincoll/vigie"
LABEL org.opencontainers.image.version="${VIGIE_VERSION}"
LABEL org.opencontainers.image.created="${DATE}"
LABEL org.opencontainers.image.revision="${COMMIT}"

# Configuration
EXPOSE 8080
ENV CONFIG_PATH=/app/config/vigie.toml

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/vigie"]
CMD ["api", "--config", "${CONFIG_PATH}"]
