# env-audit

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/0xWhisp/env-audit)](https://goreportcard.com/report/github.com/0xWhisp/env-audit)
[![Coverage Status](https://coveralls.io/repos/github/0xWhisp/env-audit/badge.svg?branch=main)](https://coveralls.io/github/0xWhisp/env-audit?branch=main)

Scan environment variables and `.env` files for misconfigurations. Catches empty values, missing required vars, detects potential secret leaks, and flags sensitive keys—without ever exposing secrets.

## Why?

Misconfigured env vars break deployments. This tool catches issues before they hit production:

- Empty `DATABASE_URL` that slipped through
- Missing `API_KEY` your app expects
- Sensitive keys that shouldn't be logged
- Potential secret leaks (API keys, tokens, high-entropy strings)
- Variables missing from your `.env.example`

## Install

```bash
go install github.com/0xWhisp/env-audit@latest
```

Or build from source:

```bash
git clone https://github.com/0xWhisp/env-audit.git
cd env-audit
go build -o env-audit ./cmd/env-audit
```

## Usage

```bash
# Scan current environment
env-audit

# Check required vars exist
env-audit --required DATABASE_URL,API_KEY,SECRET_TOKEN

# Scan a .env file
env-audit --file .env

# Compare with .env.example
env-audit --file .env --example .env.example

# Check for potential secret leaks
env-audit --file .env --check-leaks

# Generate .env.example from current env
env-audit --file .env --init

# Compare two env files
env-audit --file .env.local --diff .env.production

# Output as JSON (for scripting)
env-audit --file .env --json

# GitHub Actions format
env-audit --file .env --github

# Watch mode (re-run on file changes)
env-audit --file .env --watch

# Strict mode (warnings become errors)
env-audit --file .env --strict

# Quiet mode (only exit code, no output)
env-audit --file .env --quiet
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--file` | `-f` | Path to `.env` file to scan |
| `--required` | `-r` | Comma-separated required variables |
| `--example` | `-e` | Path to `.env.example` for comparison |
| `--ignore` | `-i` | Comma-separated keys to ignore |
| `--diff` | | Compare with another env file |
| `--dump` | `-d` | Print config with redacted secrets |
| `--init` | | Generate `.env.example` from current env |
| `--force` | | Overwrite existing files |
| `--json` | | Output results as JSON |
| `--github` | | Output in GitHub Actions format |
| `--quiet` | `-q` | Suppress stdout output |
| `--strict` | | Treat warnings as errors |
| `--check-leaks` | | Analyze values for secret patterns |
| `--no-color` | | Disable colored output |
| `--watch` | `-w` | Watch file for changes |
| `--version` | `-V` | Show version |
| `--help` | `-h` | Show help |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No issues found |
| 1 | Issues detected |
| 2 | Fatal error (invalid arguments, file not found) |

## Config File

Create `.env-audit.yaml` or `.env-audit.yml` in your project root:

```yaml
file: .env
required:
  - DATABASE_URL
  - API_KEY
example: .env.example
ignore:
  - DEBUG
  - VERBOSE
strict: true
check_leaks: true
quiet: false
json: false
github: false
no_color: false
```

CLI flags take precedence over config file values.

## Example Output

```
$ env-audit --file .env --required API_SECRET --check-leaks

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

Potential Leaks (1):
  - GITHUB_TOKEN: matches pattern 'GitHub personal access token'

Summary: 7 issues found
```

### JSON Output

```json
{
  "hasRisks": true,
  "issues": [
    {"type": "empty", "key": "DATABASE_URL", "message": "variable has empty value"},
    {"type": "missing", "key": "API_SECRET", "message": "required variable is missing"}
  ],
  "summary": {"empty": 1, "missing": 1}
}
```

### GitHub Actions Output

```
::error::API_SECRET: required variable is missing
::warning::DATABASE_URL: variable has empty value
```

## CI/CD Integration

### GitHub Actions

```yaml
- name: Audit env
  run: |
    go install github.com/0xWhisp/env-audit@latest
    env-audit --file .env --required DATABASE_URL,API_KEY --github
```

### GitLab CI

```yaml
audit:
  script:
    - go install github.com/0xWhisp/env-audit@latest
    - env-audit --file .env --required DATABASE_URL --strict
```

### Pre-commit Hook

```bash
#!/bin/sh
env-audit --file .env --quiet
```

## Leak Detection

The `--check-leaks` flag detects:

- **GitHub tokens**: `ghp_*`
- **Stripe keys**: `sk_live_*`, `sk_test_*`
- **AWS access keys**: `AKIA*`
- **JWT tokens**: `eyJ*` format
- **High entropy strings**: Potential secrets (>4.5 bits/char, >20 chars)

## Sensitive Key Patterns

Keys matching these patterns (case-insensitive) are flagged and redacted:

`SECRET` `PASSWORD` `TOKEN` `API_KEY` `APIKEY` `*KEY` `CREDENTIAL` `PRIVATE` `AUTH`

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](SECURITY.md).

## License

MIT
