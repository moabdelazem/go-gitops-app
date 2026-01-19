FROM golang:1.25-alpine AS builder

WORKDIR /app

# Download dependencies first (caching layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd

FROM alpine:latest

WORKDIR /app

# Copy only the binary from the builder stage
COPY --from=builder /app/server .

# Expose the application port
EXPOSE 8080

# Run the binary
CMD ["./server"]