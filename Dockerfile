# ─────── STAGE 1: Build ─────────────────────────────────────────────
FROM golang:1.23.4 AS builder

WORKDIR /app

# Install git (needed for go mod when using private/public repos)
RUN apt-get update && apt-get install -y git

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go binary
RUN go build -o reconciliation-service .

# ─────── STAGE 2: Minimal runtime image ─────────────────────────────
FROM alpine:latest

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/reconciliation-service .

# Copy .env
COPY .env .env

# Run binary
ENTRYPOINT ["./reconciliation-service"]