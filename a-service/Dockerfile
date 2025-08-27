# --- build stage ---
FROM golang:1.23 AS builder
WORKDIR /src

# Speed up builds & keep toolchain pinned
ENV GOFLAGS=-mod=vendor GOTOOLCHAIN=auto

# Cache go build layers
COPY go.mod go.sum ./
COPY vendor ./vendor

# Copy sources
COPY common ./common
COPY a-service ./a-service

# Build static binary (tiny, reproducible)
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
      -trimpath -ldflags='-s -w' \
      -o /out/a-service ./a-service

# --- runtime stage ---
FROM alpine:3.20
# Install CA roots if your service calls HTTPS endpoints
RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -H -u 10001 appuser
ENV TZ=Asia/Jakarta

USER appuser
WORKDIR /app
COPY .env /app
COPY --from=builder /out/a-service /app/a-service

# uncomment if you want docs/self-documenting
EXPOSE 9090   
ENTRYPOINT ["/app/a-service"]
