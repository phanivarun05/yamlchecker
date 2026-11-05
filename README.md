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

- Go 1.20 or later (needed to install or build from source)
- Git (needed only to clone the repository)

### Install with Go

Install the latest tagged version as a command-line tool:

```bash
go install github.com/phanivarun05/yamlchecker@v0.1.2
```

The executable is installed to `$GOBIN` or, by default, `$GOPATH/bin`. Add that directory to your `PATH` if needed, then run:

```bash
yamlchecker --help
```

> With modern Go versions, `go get` adds or updates dependencies in a module; it does not install command-line executables. Use `go install module@version` to install yamlchecker as a CLI tool.

### Download a prebuilt release

1. Visit the [latest GitHub release](https://github.com/phanivarun05/yamlchecker/releases/latest).
2. In **Assets**, download the archive for your operating system and CPU architecture. Linux and macOS binaries are `.tar.gz` files; Windows binaries are `.zip` files.
3. Extract the archive and place the `yamlchecker` executable (or `yamlchecker.exe` on Windows) somewhere on your `PATH`.

For example, to download and install the Linux x86-64 build:

```bash
curl -LO https://github.com/phanivarun05/yamlchecker/releases/download/v0.1.2/yamlchecker_0.1.2_linux_amd64.tar.gz
tar -xzf yamlchecker_0.1.2_linux_amd64.tar.gz
sudo install yamlchecker /usr/local/bin/yamlchecker
```

Choose the matching archive from the Releases page for other platforms. A `checksums.txt` file is also available on the release page for verifying downloads.

### Download the source ZIP

Open the [latest release page](https://github.com/phanivarun05/yamlchecker/releases/latest), expand **Assets**, and download **Source code (zip)**. Extract the ZIP, open a terminal in the project directory, and build the executable:

```bash
go build -o yamlchecker .
```

Alternatively, use **Code → Download ZIP** on the repository page to download a ZIP of the current default branch. This is not tied to a specific release.

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

## Docker

The published image is available from GitHub Container Registry (GHCR). The examples use the pinned `v0.1.2` release tag to make the image version explicit:

```bash
docker pull ghcr.io/phanivarun05/yamlchecker:v0.1.2
```

The `latest` tag is also published, but it moves to the newest release when a new version is published. Pin a version tag in scripts and deployments for predictable upgrades; use `latest` only if you intend to follow the newest release automatically.

### Validate YAML files

Mount the directory containing your YAML files at `/config` as read-only, then point the `check` command at that path. Replace `./config` with your local configuration directory:

```bash
docker run --rm \
  -v "$(pwd)/config:/config:ro" \
  ghcr.io/phanivarun05/yamlchecker:v0.1.2 \
  check --dir /config --required type,name,url
```

### Monitor endpoints and expose metrics

Start the monitor in the background, publish port `2112`, and mount your configuration directory read-only:

```bash
docker run -d --name yamlchecker-monitor \
  -p 2112:2112 \
  -v "$(pwd)/config:/config:ro" \
  ghcr.io/phanivarun05/yamlchecker:v0.1.2 \
  monitor --dir /config --workers 10 --interval 30s --metrics-addr 0.0.0.0:2112
```

The metrics endpoint is available at `http://localhost:2112/metrics`. Binding to `0.0.0.0` is necessary inside the container so the metrics server listens on the container's network interfaces and is reachable through Docker's published port.

### Stop the monitor

Stop the container with Docker's standard command:

```bash
docker stop yamlchecker-monitor
```

The application handles `SIGTERM`, so `docker stop` allows it to shut down cleanly.

### Container permissions and metrics security

- **Metrics have no authentication.** The `/metrics` endpoint is unauthenticated; expose it only to networks you trust and avoid making it publicly accessible.
- **Mounted configurations must be readable by the container user.** The image runs as non-root UID `65532`. Ensure mounted files are readable and directories are traversable by this user; mounting a directory as read-only does not override host filesystem permissions.

### Image security

The image is built on a minimal distroless base and runs as a non-root user. CI fails on known, fixable HIGH or CRITICAL vulnerabilities in the Go code and dependencies (govulncheck) and in the image (Trivy). Findings with no available fix are not counted, and new vulnerabilities are published daily, so a scan result describes one moment in time.

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
