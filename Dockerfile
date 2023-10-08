# Advanced Dockerfiles are located in ./build/release
# Read ./build/release/ReadMe.md

# Multi-Stage Build
# (Go Builder + Source) => Final Image
# ARGS are passed via docker build --build-arg

# --- Build Image ---

ARG GO_VERSION
FROM golang:${GO_VERSION}-alpine AS builder

# Other ARGS below because the FROM drops ARGS previously defined

ARG VIGIE_VERSION
ARG COMMIT
ARG DATE

# Install dep and tools
RUN apk add --no-cache libcap ca-certificates && \
    update-ca-certificates 2>/dev/null || true

WORKDIR /app

# Copy Go modules and dependencies to cache them
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . ./
RUN CGO_ENABLED=0 \
    go build -ldflags \
        "-X github.com/vincoll/vigie/cmd/vigie/version.LdVersion=${VIGIE_VERSION} \
        -X github.com/vincoll/vigie/cmd/vigie/version.LdBuildDate=${DATE} \
        -X github.com/vincoll/vigie/cmd/vigie/version.LdGitCommit=${COMMIT}" \
        -o /bin/vigie .

# Tweaks for final image
# Create low privilege user, add cap to the binary (Open Port <1000 and Raw Ntw for ICMP)
# To preserve binary cap DOCKER_BUILDKIT=1 must be exported

RUN addgroup --system --gid 1001 vigie && \
    adduser --system --uid 1001 --disabled-password --shell /sbin/nologin --no-create-home --gecos "" vigie && \
    chown -R vigie:vigie /bin/vigie && \
    setcap cap_net_raw,cap_net_bind_service=+ep /bin/vigie


# --- Final Image ---

FROM alpine:latest as final

# Copy CA Certificate
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# Copy low privilege user
COPY --from=builder /etc/group /etc/passwd /etc/
# Copy Vigie binary
COPY --from=builder /bin/vigie /

# Specifing the Working Dir for relative configs paths
WORKDIR /app

# Create Vigie folder structure
RUN mkdir --parents /app/config && \
    chown vigie:vigie -R /app

USER vigie

EXPOSE 8080

# Run the Vigie binary
ENTRYPOINT ["/vigie"]
CMD ["version"]
#CMD ["api","--config","/app/config/vigie.toml"]