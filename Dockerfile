# Use the official Go image for building the application
FROM golang:1.22-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./

# Download dependencies only if go.mod or go.sum has changed
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go app
RUN go build -o main .

# Final stage with a minimal image
FROM alpine:3.18
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/docs ./docs

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./main"]
