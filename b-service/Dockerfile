# --- build stage ---
FROM golang:1.23 AS builder
WORKDIR /src
ENV GOFLAGS=-mod=vendor GOTOOLCHAIN=auto

COPY go.mod go.sum ./
COPY vendor ./vendor

COPY common ./common
COPY b-service ./b-service

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o /out/b-service ./b-service

# --- runtime stage ---
FROM alpine:3.20
RUN adduser -D -H -u 10001 appuser
USER appuser
WORKDIR /app
COPY --from=builder /out/b-service /app/b-service

# If b-service exposes REST & gRPC:
EXPOSE 8080 50051
ENTRYPOINT ["/app/b-service"]
