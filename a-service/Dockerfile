# --- build stage ---
FROM golang:1.23 AS builder
WORKDIR /src

# If you vendored deps (you have /vendor), this keeps builds fast/offline.
ENV GOFLAGS=-mod=vendor GOTOOLCHAIN=auto

# Copy module files & vendor first for better caching
COPY go.mod go.sum ./
COPY vendor ./vendor

# Copy source (we need common + a-service at least)
COPY common ./common
COPY a-service ./a-service

# Build statically (no CGO)
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /out/a-service ./a-service

# --- runtime stage ---
FROM alpine:3.20
RUN adduser -D -H -u 10001 appuser
USER appuser
WORKDIR /app
COPY --from=builder /out/a-service /app/a-service

# Optional: document port(s) your client might expose (usually none)
# EXPOSE 9090

ENTRYPOINT ["/app/a-service"]
