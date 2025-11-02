# ---------------------------------------------
# Stage 1 — Build the Go binary
# ---------------------------------------------
FROM golang:1.25-alpine AS builder

# Enable Go modules and install git (if private deps)
RUN apk add --no-cache git

# Set working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first (for dependency caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .
     
# Build the binary with optimizations
RUN go build -o web-analyzer main.go

# ---------------------------------------------
# Stage 2 — Create lightweight runtime image
# ---------------------------------------------
FROM alpine:latest

WORKDIR /app

# Copy only the built binary and web assets from builder
COPY --from=builder /app/web-analyzer .
COPY --from=builder /app/web ./web

# Expose server port
EXPOSE 8080

# Set the default command
CMD ["./web-analyzer"]
