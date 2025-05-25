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
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api ./cmd/api

# Final stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/api .

# Expose ports
EXPOSE 8080 8081 9090

# Run the binary
ENTRYPOINT ["./api"]