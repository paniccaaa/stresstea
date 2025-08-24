# Stresstea

A modern CLI tool for load testing HTTP and gRPC services with a beautiful TUI interface.

## Features

- 🚀 **HTTP/HTTPS Load Testing** - support for all HTTP methods
- 🔌 **gRPC Load Testing** - testing gRPC services
- 📝 **Declarative YAML Configurations** - simple and clear scenarios
- 🎨 **Interactive TUI Interface**
- 📊 **Report and Chart Generation** - detailed analytics of results
- ⏱️ **Real-time Monitoring** - tracking metrics in real time

## Installation

```bash
go install github.com/paniccaaa/stresstea/cmd/stresstea@latest
```

## Quick Start

### Simple HTTP Testing

```bash
stresstea run -t http://localhost:8080 -r 100 -d 30s
```

### Using Configuration File

```bash
stresstea run -f configs/example.yaml
```

### gRPC Testing

```bash
stresstea run -t localhost:50051 -p grpc -r 50 -d 60s
```

## Configuration

Stresstea supports YAML configurations for complex scenarios:

```yaml
global:
  target: "http://localhost:8080"
  duration: 60s
  rate: 100
  concurrent: 10
  protocol: "http"

scenarios:
  - name: "API Test"
    flow:
      - http:
          method: "GET"
          url: "/api/health"
          headers:
            Content-Type: "application/json"
      
      - wait:
          duration: 1s
      
      - http:
          method: "POST"
          url: "/api/users"
          body: '{"name": "test"}'
```

## Commands

### run
Run load testing

```bash
stresstea run [target] [flags]
```

Flags:
- `-t, --target` - target URL or gRPC endpoint
- `-d, --duration` - test duration (default 30s)
- `-r, --rate` - requests per second (default 100)
- `-c, --concurrent` - number of concurrent connections (default 10)
- `-f, --config` - path to YAML configuration file
- `-p, --protocol` - protocol (http or grpc, default http)

### report
Generate report from test results

```bash
stresstea report [results-file] [flags]
```

Flags:
- `-o, --output` - output file for the report (default report.txt)
- `-f, --format` - report format (text, json, html, default text)

### version
Show Stresstea version

```bash
stresstea version
```

## TUI Interface

When running tests, an interactive TUI interface opens that shows:

- Test configuration
- Real-time statistics
- Performance metrics
- Charts and diagrams

Controls:
- `q` - exit TUI
- `Ctrl+C` - force quit

## Usage Examples

### Testing REST API

```bash
# Simple GET request
stresstea run -t http://api.example.com/health -r 200 -d 60s

# POST request with body
stresstea run -t http://api.example.com/users \
  -r 50 \
  -d 120s \
  --method POST \
  --body '{"name": "test", "email": "test@example.com"}'
```

## Architecture

The project follows standard Go project structure:

```
stresstea/
├── cmd/stresstea/          # CLI entry point
├── internal/
│   ├── cli/               # CLI commands (Cobra)
│   ├── config/            # Configuration and YAML parsing
│   ├── engine/            # Main engine
│   ├── loadtest/          # HTTP and gRPC testers
│   └── ui/                # TUI interface (bubbletea)
├── example-configs/               # Configuration examples
└── README.md
```

## Technologies

- **Go 1.24+** 
- **Cobra** 
- **Bubbletea** 
- **gRPC**
- **YAML** 
- **Prometheus** 

## Setup Development

```bash
make all
```