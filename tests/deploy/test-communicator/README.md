# Test Communicator

A test application that supports HTTP, gRPC, and TCP communication protocols for testing the OpenTelemetry collector.

## Building

The application is built using Docker:

```bash
docker build -t test-communicator .
```

## Protocol Buffer Files

If you need to modify the gRPC service definition, edit `testcommunicator.proto` and regenerate the Go files:

### Prerequisites

Install the Protocol Buffer compiler and Go plugins:

```bash
# Install protoc
# On macOS:
brew install protobuf

# On Ubuntu/Debian:
apt-get install -y protobuf-compiler

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Regenerating Proto Files

```bash
# From the test-communicator directory
protoc --go_out=proto --go_opt=paths=source_relative \
       --go-grpc_out=proto --go-grpc_opt=paths=source_relative \
       testcommunicator.proto
```

This will regenerate the files in the `proto/` directory.

## Usage

The application supports different communication protocols configured via environment variables:

- `PROTOCOL`: "http", "grpc", "tcp", or "all" (default: "http")
- `PORT`: Main service port (default: 8080)
- `SERVICE_NAME`: Service identifier (default: "test-communicator")
- `TARGET_URL`: Target URL for HTTP calls
- `TARGET_HOST`: Target host for non-HTTP protocols
- `TARGET_PORT`: Target port for non-HTTP protocols

### Endpoints

#### HTTP
- `GET /health` - Health check
- `GET /api/data` - Sample data endpoint
- `GET /api/users/{id}` - User information endpoint
- `GET /api/call-target` - Calls configured target
- `GET /metrics` - Prometheus metrics

#### gRPC
- `Health()` - Health check
- `GetData()` - Sample data
- `CallTarget()` - Calls configured target
