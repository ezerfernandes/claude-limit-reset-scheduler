# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o calgo ./cmd/calgo

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for Google API HTTPS calls
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /app/calgo /usr/local/bin/calgo

# Create non-root user
RUN adduser -D -g '' calgo
USER calgo

ENTRYPOINT ["calgo"]
