# --- Stage 1: Builder ---
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git (sometimes required for fetching dependencies)
RUN apk add --no-cache git

# 1. Dependency Caching
# We copy ONLY the mod files first. Docker caches this layer.
# If you change code but not dependencies, this step is skipped (fast build).
COPY go.mod go.sum ./
RUN go mod download

# 2. Source Code
# Copy everything else (respecting .dockerignore)
COPY . .

# 3. Build
# CGO_ENABLED=0 creates a static binary (no external C library dependencies)
# -o cert-server: names the binary
# ./cmd/server: points to your main.go location
RUN CGO_ENABLED=0 GOOS=linux go build -o cert-server ./cmd/server

# --- Stage 2: Production Runner ---
# Use a tiny, secure Alpine image
FROM alpine:latest

WORKDIR /app

# Install CA Certs (Essential for sending Emails via SMTP/TLS)
RUN apk --no-cache add ca-certificates tzdata

# 1. Copy the Binary
COPY --from=builder /app/cert-server .

# 2. Copy the "Public" folder
# This contains your Agent binaries and config.yaml
# The path must match your local structure relative to the Dockerfile
COPY public/ public/

# Expose the port defined in your config (8080)
EXPOSE 8080

# Run
CMD ["./cert-server"]