# Key-Value Store

An in-memory key-value store service built with Go, featuring a RESTful API interface and Docker support.

## Features

- In-memory key-value storage
- RESTful API with JSON responses
- Configurable key length and value size limits
- Docker and Docker Compose support
- Comprehensive test suite including benchmarks
- OpenAPI specification
- CORS support

## Prerequisites

- Go 1.22 or higher
- Docker and Docker Compose (for containerized deployment)
- Task (optional, for running task commands)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/jeffy-mathew/key-value-store.git
   cd key-value-store
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. (Optional) Install Task runner, visit https://taskfile.dev/installation/

## Configuration

The service can be configured using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| SERVER_ADDRESS | Server listen address | 0.0.0.0:8081 |
| READ_TIMEOUT | HTTP read timeout | 5s |
| WRITE_TIMEOUT | HTTP write timeout | 5s |
| SHUTDOWN_TIMEOUT | Graceful shutdown timeout | 5s |
| MAX_KEY_LENGTH | Maximum key length | 256 |
| MAX_VALUE_SIZE | Maximum value size in bytes | 1048576 |

## Usage

### Using Task Runner

We use [Task](https://taskfile.dev) for managing development tasks. Here are the available commands:

```bash
# Build the binary
task build

# Run the service
task run

# Run tests
task test                # Run all tests
task test:integration    # Run integration tests
task test:benchmark      # Run benchmark tests

# Docker operations
task docker:build        # Build Docker image
task docker:run         # Run with Docker Compose
task docker:stop        # Stop Docker containers

# Development tools
task lint              # Run linters
task clean             # Clean build artifacts
```

### Manual Usage

If you prefer not to use Task:

1. Build the binary:
   ```bash
   go build -o store ./cmd/store
   ```

2. Run the service:
   ```bash
   ./store
   ```

### Docker Usage

1. Build and run using Docker Compose:
   ```bash
   docker-compose up
   ```

## API Endpoints

### Set Key
```http
curl --location 'http://localhost8081/key/' \
--header 'Content-Type: application/json' \
--data '{
    "key": "hello",
    "value": "world"
}'
```

### Get Key
```http
curl --location 'http://localhost8081/key/hello' 
```

### Delete Key
```http
curl --location --request DELETE 'http://localhost8081/key/hello'
```

For detailed API documentation, refer to the OpenAPI specification in [openapi.yaml](openapi.yaml).

## Testing

Run different types of tests:

```bash
# Unit tests
go test ./...

# Integration tests
go test ./tests -tags=integration

# Benchmark tests
go test -bench=. -benchmem ./tests/
```
