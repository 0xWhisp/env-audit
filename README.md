# env-audit

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-27%20passed-brightgreen)](https://github.com/yourusername/env-audit/actions)
[![Build](https://img.shields.io/badge/Build-passing-brightgreen)]()

A minimal, production-grade CLI tool that scans environment variables and `.env` files for common misconfigurations. Identifies security risks and configuration issues without exposing sensitive values.

## Features

- **Empty Value Detection** - Find environment variables with empty string values
- **Missing Required Variables** - Verify required variables are present
- **Sensitive Key Detection** - Flag keys matching patterns like `SECRET`, `PASSWORD`, `TOKEN`, `API_KEY`, `CREDENTIAL`, `PRIVATE`, `AUTH`
- **`.env` File Parsing** - Parse and validate configuration files
- **Duplicate Key Detection** - Identify duplicate definitions in config files
- **Value Redaction** - Never exposes sensitive values in output
- **CI/CD Integration** - Exit codes designed for pipeline integration

## Installation

### From Source

```bash
go install github.com/yourusername/env-audit@latest
```

### Build Locally

```bash
git clone https://github.com/yourusername/env-audit.git
cd env-audit
go build -o env-audit .
```

### Cross-Platform Binaries

Build for multiple platforms:

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o env-audit-linux .

# macOS
GOOS=darwin GOARCH=amd64 go build -o env-audit-darwin .

# Windows
GOOS=windows GOARCH=amd64 go build -o env-audit.exe .
```

## Usage

```bash
# Scan current environment
env-audit

# Scan with required variables
env-audit --required DATABASE_URL,API_KEY,SECRET_TOKEN

# Scan a .env file
env-audit --file .env

# Dump parsed config (with redaction)
env-audit --dump

# Combine options
env-audit --file .env --required DATABASE_URL,REDIS_HOST --dump

# Show help
env-audit --help
```

### Options

| Flag | Short | Description |
|------|-------|-------------|
| `--file` | `-f` | Path to `.env` file to scan |
| `--required` | `-r` | Comma-separated list of required variables |
| `--dump` | `-d` | Output parsed configuration (with redaction) |
| `--help` | `-h` | Show help message |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No risks found |
| 1 | Risks detected |
| 2 | Fatal error (invalid arguments, file not found) |

## Example Output

```
env-audit scan results
======================

Empty Values (2):
  - DATABASE_URL
  - REDIS_HOST

Missing Required (1):
  - API_SECRET

Sensitive Keys Detected (3):
  - AWS_SECRET_KEY: [REDACTED]
  - DATABASE_PASSWORD: [REDACTED]
  - JWT_TOKEN: [REDACTED]

Summary: 6 issues found
```

### Dump Mode Output

```
DATABASE_URL=
REDIS_HOST=
AWS_SECRET_KEY=[REDACTED]
DATABASE_PASSWORD=[REDACTED]
JWT_TOKEN=[REDACTED]
APP_NAME=myapp
DEBUG=true
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Audit Environment
  run: |
    go install github.com/yourusername/env-audit@latest
    env-audit --required DATABASE_URL,API_KEY
```

### GitLab CI

```yaml
audit:
  script:
    - go install github.com/yourusername/env-audit@latest
    - env-audit --file .env --required DATABASE_URL
```

## Testing

The project includes comprehensive test coverage with both unit tests and property-based tests.

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -cover ./...
```

### Test Summary

| Category | Tests | Status |
|----------|-------|--------|
| Property-Based Tests | 9 | ✅ Passed |
| Unit Tests | 18 | ✅ Passed |
| **Total** | **27** | ✅ **All Passed** |

Property-based tests run 100 iterations each, validating correctness properties including:
- Empty value detection completeness
- Missing required detection completeness
- Sensitive key pattern matching
- Sensitive value redaction
- `.env` parsing round-trip consistency
- Duplicate key detection
- Exit code correctness

## Architecture

```
env-audit/
├── main.go      # CLI entry point, argument parsing
├── scanner.go   # Coordinates checks, aggregates results
├── checks.go    # Audit rules (empty, missing, sensitive)
├── parser.go    # .env file parsing and formatting
├── output.go    # Summary formatting, redaction
└── env.go       # Environment variable reader
```

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Security

See [SECURITY.md](SECURITY.md) for security policy and reporting vulnerabilities.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Related Projects

- [godotenv](https://github.com/joho/godotenv) - Go port of Ruby dotenv
- [envconfig](https://github.com/kelseyhightower/envconfig) - Go library for managing configuration data from environment variables
