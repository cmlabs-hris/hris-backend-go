# syntax=docker/dockerfile:1

# Stage 1: Build the Go binary
FROM golang:1.27-alpine AS builder

# Install ca-certificates for HTTPS (Xendit API calls) and tzdata for timezone handling
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/api/

# Stage 2: Minimal runtime image
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

# Non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY --from=builder /app/server /usr/local/bin/server

WORKDIR /app
USER appuser

EXPOSE 8080

# Health check uses existing chi heartbeat at "/"
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

CMD ["server"]
