# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o config-generator cmd/generator/main.go

# Final stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates git

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/config-generator .

CMD ["./config-generator"]