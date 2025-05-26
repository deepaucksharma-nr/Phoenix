# Build stage
FROM golang:1.21-alpine AS builder

# Install dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o phoenix-api cmd/api/main.go

# Final stage
FROM alpine:3.18

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 -S phoenix && \
    adduser -u 1000 -S phoenix -G phoenix

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/phoenix-api /app/
COPY --from=builder /build/migrations /app/migrations

# Install migrate tool
RUN wget -q https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz && \
    tar -xf migrate.linux-amd64.tar.gz && \
    mv migrate /usr/local/bin/migrate && \
    rm migrate.linux-amd64.tar.gz

# Change ownership
RUN chown -R phoenix:phoenix /app

# Switch to non-root user
USER phoenix

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
CMD ["/app/phoenix-api"]