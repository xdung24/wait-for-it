# wait-for-it

A Go utility that polls a TCP host/port until it becomes available, then optionally executes a command. Compiles to a single self-contained binary — no shell, no runtime dependencies.

## Features

- **TCP availability check** — repeatedly attempts to connect to a host/port until it succeeds or times out
- **Flexible host/port input** — accepts the combined `host:port` shorthand or separate `-h`/`-p` flags
- **Configurable timeout** — set a maximum wait time with `-t`/`--timeout` (default: 15 s); use `0` for no timeout
- **Quiet mode** (`-q`/`--quiet`) — suppresses all status messages for use in scripts
- **Strict mode** (`-s`/`--strict`) — only executes the trailing command if the port check **succeeds**; skips it on timeout
- **Command execution** — run any command after the check with `-- COMMAND ARGS`; exit code is propagated to the caller
- **Command runs on timeout by default** — without `--strict`, the trailing command still runs even if the timeout was reached
- **Custom DNS server** (`-d`/`--dns`) — resolve hostnames via a specific DNS server instead of the system default
- **SIGINT/SIGTERM support** — handles Ctrl+C and termination signals cleanly
- **Cross-platform** — runs on Linux, macOS, and Windows without any dependencies

## Build

```bash
# Build for the current platform
go build -o wait-for-it wait-for-it.go

# Build a fully static binary (recommended for containers)
CGO_ENABLED=0 go build -o wait-for-it wait-for-it.go

# Cross-compile
GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -o wait-for-it       wait-for-it.go
GOOS=windows GOARCH=amd64               go build -o wait-for-it.exe    wait-for-it.go
GOOS=darwin  GOARCH=arm64               go build -o wait-for-it-darwin  wait-for-it.go
```

## Usage

```bash
# Wait for a service using combined host:port notation (15 s default timeout)
./wait-for-it localhost:5432

# Set an explicit timeout
./wait-for-it localhost:5432 -t 30

# Wait with no timeout
./wait-for-it localhost:5432 -t 0

# Suppress all output
./wait-for-it localhost:5432 -q

# Run a command after the port is available
./wait-for-it localhost:5432 -t 30 -- echo "Database is up"

# Only run the command if the port check succeeded (strict mode)
./wait-for-it localhost:5432 -t 30 -s -- python app.py

# Use a custom DNS server to resolve the host
./wait-for-it myservice.internal:5432 -d 8.8.8.8

# Use separate host and port flags
./wait-for-it -h localhost -p 5432 -t 30
```

## Arguments

| Argument | Description |
|---|---|
| `host:port` | Combined host and port (positional) |
| `-h HOST` \| `--host=HOST` | Target hostname or IP address |
| `-p PORT` \| `--port=PORT` | Target TCP port number |
| `-t SECONDS` \| `--timeout=SECONDS` | Seconds to wait before giving up; `0` = no timeout (default: `15`) |
| `-q` \| `--quiet` | Suppress all status messages |
| `-s` \| `--strict` | Only execute the subcommand if the port check succeeds |
| `-d DNS` \| `--dns=DNS` | Custom DNS server, e.g. `8.8.8.8` or `8.8.8.8:53` (default port: `53`) |
| `-- COMMAND [ARGS]` | Command to execute after the check completes |
| `--help` | Print usage information and exit |

## Exit Codes

| Code | Meaning |
|---|---|
| `0` | Port became available (and trailing command succeeded, if given) |
| `1` | Missing required arguments or unknown argument |
| non-zero | Timeout elapsed (no trailing command), or exit code of the trailing command |
