# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache ca-certificates git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application (static binary)
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -tags netgo,osusergo -ldflags="-s -w" -o server ./cmd/server

# Final stage (minimal)
FROM scratch

WORKDIR /app

# Copy binary and migrations
COPY --from=builder /app/server /app/server
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose port
EXPOSE 8080

# Environment variables
ENV SERVER_PORT=8080

# Run application
CMD ["/app/server"]
