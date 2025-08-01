# Multi-stage build for efficient image size
FROM golang:1.21-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies (no need for git since modules are cached)
RUN go mod download

# Copy protobuf definition
COPY testcommunicator.proto ./

# Copy generated protobuf files
COPY proto/ ./proto/

# Copy source code
COPY main.go .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o test-communicator .

# Final stage - use distroless for security and small size
FROM gcr.io/distroless/static-debian11:nonroot

# Copy the binary from builder stage
COPY --from=builder /app/test-communicator /test-communicator

# Expose ports for different protocols
EXPOSE 8080 8081 8082 8083
EXPOSE 9080 9081 9082 9083
EXPOSE 7080 7081 7082 7083

# Run the application
ENTRYPOINT ["/test-communicator"]
