# Multi-stage Dockerfile for Massage Bot
# Stage 1: Builder
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Create and use non-root user for build
RUN adduser -D -u 10001 builder
USER builder
WORKDIR /build

# Copy module files first for better caching
COPY --chown=builder:builder go.mod go.sum ./
RUN go mod download

# Copy source code
COPY --chown=builder:builder . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags='-w -s' -o massage-bot ./cmd/bot/

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN addgroup -S app && adduser -S app -G app

# Create app directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder --chown=app:app /build/massage-bot .
COPY --from=builder --chown=app:app /build/.env .env

# Switch to non-root user
USER app

# Expose health check port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./massage-bot"]
