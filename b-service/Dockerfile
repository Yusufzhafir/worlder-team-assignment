# --- build stage ---
FROM golang:1.23 AS builder
WORKDIR /src
ENV GOFLAGS=-mod=vendor GOTOOLCHAIN=auto

COPY go.mod go.sum ./
COPY vendor ./vendor
COPY common ./common
COPY b-service ./b-service

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
      -trimpath -ldflags='-s -w' \
      -o /out/b-service ./b-service

# --- runtime stage ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata curl && \
    adduser -D -H -u 10001 appuser
ENV TZ=Asia/Jakarta

USER appuser
WORKDIR /app
COPY .env /app
COPY --from=builder /out/b-service /app/b-service

# REST + gRPC (documented)
EXPOSE 8080 50051
HEALTHCHECK --interval=10s --timeout=3s --retries=10 \
  CMD curl -fsS http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/app/b-service"]
