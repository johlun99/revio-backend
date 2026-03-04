# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Download dependencies first (layer cache)
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/bin/api \
    ./cmd/api

# ---- Runtime stage ----
FROM scratch

# CA certs for outbound TLS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Binary
COPY --from=builder /app/bin/api /api

# Migrations (embedded via go:embed, but keep for reference)
COPY --from=builder /build/internal/db/migrations /migrations

USER 65534:65534

EXPOSE 8080

ENTRYPOINT ["/api"]
