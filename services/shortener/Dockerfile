# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy source code
COPY . .

# Initialize a new module with the correct Go version
RUN rm -f go.mod go.sum && \
    go mod init github.com/ashtonholgate/url-minifier/services/shortener && \
    go mod tidy && \
    go get github.com/fsnotify/fsnotify@v1.7.0 && \
    go get golang.org/x/sys/unix@latest && \
    go get github.com/spf13/viper@v1.19.0 && \
    go get go.mongodb.org/mongo-driver@v1.17.3 && \
    go get github.com/go-redis/redis/v8@v8.11.5 && \
    go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/server .

# Create a non-root user
RUN adduser -D appuser
USER appuser

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./server"]
