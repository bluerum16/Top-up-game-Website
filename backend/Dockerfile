# ════════════════════════════════════════════
# Stage 1: Builder
# ════════════════════════════════════════════
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata make

WORKDIR /app

# Copy go.mod & go.sum dulu → layer cache
# Jika tidak ada perubahan dependency, go mod download di-skip
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy seluruh source code
COPY . .

# Build binary
# -ldflags="-w -s"  → strip debug info, perkecil binary
# CGO_ENABLED=0     → static binary, tidak butuh glibc
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o /app/bin/main \
    ./cmd/api

# ════════════════════════════════════════════
# Stage 2: Runner — image seminimal mungkin
# ════════════════════════════════════════════
FROM alpine:3.19

# ca-certificates → untuk HTTPS calls ke Midtrans, publisher API
# tzdata          → untuk timezone Jakarta di logging & cron
RUN apk --no-cache add ca-certificates tzdata wget && \
    cp /usr/share/zoneinfo/Asia/Jakarta /etc/localtime && \
    echo "Asia/Jakarta" > /etc/timezone

# Jalankan sebagai non-root user (best practice keamanan)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy hanya binary dari builder stage
COPY --from=builder /app/bin/main .
RUN chown appuser:appgroup /app/main

USER appuser

EXPOSE 8080

# Healthcheck → Docker & docker-compose tahu service sudah siap
HEALTHCHECK --interval=30s --timeout=10s --start-period=20s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./main"]