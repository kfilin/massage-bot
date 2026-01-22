# Multi-stage Dockerfile for Massage Bot
# Stage 1: Builder
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata
ENV TZ=Europe/Istanbul

WORKDIR /build

# Copy module files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags='-w -s' -o massage-bot ./cmd/bot/

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security (Runtime only)
RUN addgroup -S app && adduser -S app -G app

# Create app directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /build/massage-bot .

# Pre-create data and logs directories with correct permissions
RUN mkdir -p /app/data /app/logs && chown -R app:app /app/data /app/logs

# Switch to non-root user
USER app

# Expose health check port
EXPOSE 8081

# Health check (matches .env port)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# Run the application
CMD ["./massage-bot"]
