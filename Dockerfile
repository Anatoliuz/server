# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder

# Install necessary tools and dependencies
RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker layer caching
COPY go.mod go.sum ./

# Download dependencies only if go.mod or go.sum has changed
RUN go mod download

# Copy the application source code
COPY . .

# Build the Go app
RUN go build -o main ./cmd/app

# Stage 2: Create a minimal image for running the application
FROM alpine:3.18

# Set the working directory in the final stage
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Add a non-root user for security (optional but recommended)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./main"]
