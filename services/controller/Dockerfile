# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./
COPY ../../packages/go-common/go.mod ../../packages/go-common/
COPY ../../packages/contracts/go.mod ../../packages/contracts/

# Download dependencies
RUN go mod download

# Copy source code
COPY . .
COPY ../../packages/go-common ../../packages/go-common
COPY ../../packages/contracts ../../packages/contracts

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o controller ./cmd/controller

# Final stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/controller .

# Expose ports
EXPOSE 8082 9090

# Run the binary
ENTRYPOINT ["./controller"]