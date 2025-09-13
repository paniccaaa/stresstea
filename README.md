# Stresstea

A modern CLI tool for load testing HTTP and gRPC services with a beautiful TUI interface.

## Features

- 🚀 **HTTP/HTTPS Load Testing** - support for all HTTP methods
- 🔌 **gRPC Load Testing** - testing gRPC services (planned)
- 📝 **Declarative YAML Configurations** - simple and clear scenarios
- 🎨 **Compact TUI Interface** - minimal and efficient
- 📊 **Console Output** - perfect for scripts and CI/CD
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
- `-o, --output` - output mode (compact, console, default compact)

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

## Output Modes

Stresstea supports two output modes:

### Compact Mode (Default)
Interactive TUI interface with:
- Real-time metrics display
- Progress tracking
- Interactive controls (pause, help)
- Color-coded status indicators

Controls:
- `h` - show/hide help
- `p` - pause/resume test
- `q` - exit application
- `Ctrl+C` - force quit

### Console Mode
Simple console output perfect for:
- Scripts and automation
- CI/CD pipelines
- Background monitoring

Controls:
- `Ctrl+C` - stop test and exit

## Usage Examples

### Testing REST API

```bash
# Compact mode (default)
stresstea run -t http://api.example.com/health -r 200 -d 60s

# Console mode for automation
stresstea run -t http://api.example.com/health -r 200 -d 60s -o console

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