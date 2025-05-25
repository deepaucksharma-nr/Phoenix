# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o control-service ./cmd/control-service

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl

# Create non-root user
RUN addgroup -g 1000 phoenix && \
    adduser -D -s /bin/sh -u 1000 -G phoenix phoenix

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/control-service .

# Change ownership
RUN chown -R phoenix:phoenix /app

# Switch to non-root user
USER phoenix

# Expose ports
EXPOSE 50053 8083

# Run the service
CMD ["./control-service"]