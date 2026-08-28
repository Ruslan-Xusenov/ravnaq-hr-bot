# Build stage
FROM golang:alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
COPY . .
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o hrbot ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/hrbot .
# Copy locales
COPY --from=builder /app/internal/locales ./internal/locales

# Expose HTTP port
EXPOSE 8080

# Run the binary
CMD ["./hrbot"]
