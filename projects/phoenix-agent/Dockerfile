# Build stage
FROM golang:1.24-alpine AS builder

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
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o phoenix-agent cmd/phoenix-agent/main.go

# Final stage
FROM alpine:3.18

# Install required packages
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    bash \
    curl \
    wget \
    stress-ng \
    bc

# Download OTel Collector Contrib
RUN wget -q https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.95.0/otelcol-contrib_0.95.0_linux_amd64.tar.gz && \
    tar -xzf otelcol-contrib_0.95.0_linux_amd64.tar.gz && \
    mv otelcol-contrib /usr/local/bin/ && \
    chmod +x /usr/local/bin/otelcol-contrib && \
    rm otelcol-contrib_0.95.0_linux_amd64.tar.gz

# Create directories
RUN mkdir -p /etc/phoenix-agent /var/log/phoenix-agent

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/phoenix-agent /app/

# Create non-root user (but agent needs some privileges)
RUN addgroup -g 1000 -S phoenix && \
    adduser -u 1000 -S phoenix -G phoenix

# Change ownership
RUN chown -R phoenix:phoenix /app /etc/phoenix-agent /var/log/phoenix-agent

# Note: Agent typically runs as root for process management
# but can be configured to drop privileges after startup

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD pgrep phoenix-agent || exit 1

# Run the binary
CMD ["/app/phoenix-agent"]