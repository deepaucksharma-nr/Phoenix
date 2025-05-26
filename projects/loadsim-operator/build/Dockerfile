# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./
COPY ../../../go.work ../../../go.work
COPY ../../../go.work.sum ../../../go.work.sum

# Copy source code
COPY . .

# Download dependencies
RUN go mod download

# Build the simulator
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o load-simulator cmd/simulator/main.go

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates bash coreutils

# Create non-root user
RUN addgroup -g 1000 phoenix && \
    adduser -D -u 1000 -G phoenix phoenix

# Copy binary from builder
COPY --from=builder /workspace/load-simulator /usr/local/bin/load-simulator

# Set ownership
RUN chown phoenix:phoenix /usr/local/bin/load-simulator && \
    chmod +x /usr/local/bin/load-simulator

USER phoenix

ENTRYPOINT ["/usr/local/bin/load-simulator"]