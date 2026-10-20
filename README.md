# yamlchecker

yamlchecker is a Go-based CLI utility for validating YAML configuration files and monitoring HTTP endpoints described in YAML. It is built to help teams catch configuration issues early, enforce required keys, and keep track of service health without introducing a heavy operational stack.

The project is intentionally lightweight and focused on two core workflows:

- YAML validation for required fields and basic structure
- HTTP endpoint monitoring for services declared in YAML definitions

## Why yamlchecker?

Configuration drift and malformed YAML are common causes of service failures. yamlchecker helps you:

- verify that essential fields are present in YAML files
- validate HTTP targets before or during deployment
- monitor endpoint health in a repeatable, scriptable way
- integrate into CI/CD and operational tooling with minimal setup

## Features

- Validate YAML files in a directory against required fields
- Enforce required keys with simple CLI arguments
- Parse YAML into Go maps for flexible config inspection
- Monitor HTTP endpoints declared in YAML files
- Run checks concurrently with configurable worker counts
- Expose Prometheus-compatible metrics over an HTTP endpoint
- Support structured JSON and text logging
- Built with Go and Cobra for a clean command-line interface

## Project status

This repository is designed to evolve with release automation and tag-based publishing workflows. Planned improvements include:

- GoReleaser-based binary builds and packaging
- GitHub Actions release automation triggered by version tags such as `v1.2.3`
- Automated artifact publishing for Linux, macOS, and Windows

## Installation

### Prerequisites

- Go 1.20 or later
- Git

### Clone and build

```bash
git clone https://github.com/phanivarun05/yamlchecker.git
cd yamlchecker
go build -o yamlchecker .
```

You can also build with a module-aware command:

```bash
go mod tidy
go build ./...
```

## Usage

Run the CLI with the help command:

```bash
./yamlchecker --help
```

### Global flags

The root command supports the following flags:

```bash
./yamlchecker --log-level info --log-format text --version
```

Available global flags:

- `--log-level` — log verbosity (`debug`, `info`, `warn`, `error`)
- `--log-format` — logging output format (`text` or `json`)
- `--version` — print the application version

## Commands

### 1. check

Validate YAML files in a directory and ensure required keys are present.

Usage:

```bash
yamlchecker check --dir <directory> --required <field1,field2,...>
```

Example:

```bash
./yamlchecker check --dir ./testdata --required type,name,url
```

Flags:

- `-d, --dir` — directory to scan for YAML files
- `-r, --required` — required fields, comma-separated

This command loads all YAML files under the directory, parses them, and prints warnings or errors when required fields are missing.

### 2. monitor

Read HTTP endpoints defined in YAML and monitor them over time.

Usage:

```bash
yamlchecker monitor --dir <directory> --workers <count> --interval <duration> --metrics-addr <host:port>
```

Example:

```bash
./yamlchecker monitor --dir ./testdata-monitor --workers 10 --interval 30s
```

Flags:

- `-d, --dir` — directory containing YAML entry definitions
- `-w, --workers` — number of concurrent HTTP checks
- `-i, --interval` — interval between health-check rounds
- `--metrics-addr` — address for the optional Prometheus metrics server

This command will:

- load endpoint definitions from YAML
- validate each endpoint definition
- perform HTTP health checks
- classify each result as healthy, unhealthy, or unreachable
- optionally expose metrics for observation systems

## YAML format

The project supports YAML content that maps to Go `map[string]interface{}` values. The validator checks for required keys only; it is intentionally simple and flexible.

### Example validation config

```yaml
type: http
name: Google
url: https://www.google.com
```

Validation command:

```bash
./yamlchecker check --dir ./testdata --required type,name,url
```

### Example monitoring config

```yaml
endpoints:
  - type: http
    name: httpbin-get
    url: "https://httpbin.org/get"

  - type: http
    name: example-root
    url: "https://example.com/"
```

Monitoring command:

```bash
./yamlchecker monitor --dir ./testdata-monitor --workers 5 --interval 30s
```

## Health semantics

The monitor evaluates each HTTP target and classifies it as one of the following:

- `healthy` — HTTP request succeeded with status code 200
- `unhealthy` — HTTP request reached the endpoint but returned a non-200 status
- `unreachable` — network failure, timeout, invalid URL, or connection issue

## Exit behavior

The CLI returns a non-zero exit status when:

- required fields are missing
- YAML files cannot be parsed
- a target endpoint fails health checks
- invalid monitor parameters are supplied

## Example output

```text
Loading from Directory: ./testdata
YAML Script parsed without errors:./testdata/app.yaml -> map[type:http name:Google url:http://www.google.com]
```

During monitoring, output will include logs such as:

```text
endpoint healthy
endpoint unhealthy
endpoint unreachable
```

## Project structure

```text
.
├── checker/
│   └── http_checker.go
├── cmd/
│   ├── check.go
│   ├── root.go
│   └── version.go
├── internal/
│   ├── config/
│   ├── metrics/
│   ├── scanner/
│   └── validator/
├── main.go
├── go.mod
├── go.sum
├── testdata/
├── testdata-monitor/
├── README.md
├── LICENSE
└── .gitignore
```

## Development

### Format code

```bash
go fmt ./...
```

### Run tests

```bash
go test ./...
```

## Future release automation

This project is set up with a path toward professional releases. Planned automation includes:

- GoReleaser for cross-platform binary builds
- GitHub Actions workflow triggered by version tags like `v1.0.0`
- Release notes generation and artifact publishing to GitHub Releases

Example tag pattern:

```bash
git tag v1.0.0
git push origin v1.0.0
```

A GitHub Actions workflow can then build and publish release artifacts automatically when the tag is pushed.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.

## Contributing

Contributions are welcome. If you plan to submit changes:

1. Fork the repository
2. Create a topic branch
3. Make your changes
4. Run formatting and tests
5. Open a pull request with a clear description

## Maintainer

- phanivarun05

## Repository

https://github.com/phanivarun05/yamlchecker
