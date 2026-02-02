# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags '-static'" \
    -o /app/bin/server ./cmd/server

# Runtime stage - Distroless
FROM gcr.io/distroless/static-debian12:nonroot

# Copy binary from builder
COPY --from=builder /app/bin/server /server

# Copy config files
COPY configs/ /configs/

# Use non-root user (UID 65532)
USER nonroot:nonroot

# Expose port
EXPOSE 8001

# Run the server
ENTRYPOINT ["/server"]
