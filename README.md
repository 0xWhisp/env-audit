# env-audit

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/0xWhisp/env-audit)](https://goreportcard.com/report/github.com/0xWhisp/env-audit)
[![Coverage Status](https://coveralls.io/repos/github/0xWhisp/env-audit/badge.svg?branch=master)](https://coveralls.io/github/0xWhisp/env-audit?branch=master)

Scan environment variables and `.env` files for misconfigurations. Catches empty values, missing required vars, and flags sensitive keys—without ever exposing secrets.

## Why?

Misconfigured env vars break deployments. This tool catches issues before they hit production:

- Empty `DATABASE_URL` that slipped through
- Missing `API_KEY` your app expects
- Sensitive keys that shouldn't be logged

## Install

```bash
go install github.com/0xWhisp/env-audit@latest
```

Or build from source:

```bash
git clone https://github.com/0xWhisp/env-audit.git
cd env-audit
go build -o env-audit .
```

## Usage

```bash
# Scan current environment
env-audit

# Check required vars exist
env-audit --required DATABASE_URL,API_KEY,SECRET_TOKEN

# Scan a .env file
env-audit --file .env

# Dump config (secrets redacted)
env-audit --dump

# Combine flags
env-audit --file .env.production --required DATABASE_URL -d
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--file` | `-f` | Path to `.env` file |
| `--required` | `-r` | Comma-separated required vars |
| `--dump` | `-d` | Print config with redacted secrets |
| `--help` | `-h` | Show help |

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Clean |
| 1 | Issues found |
| 2 | Fatal error |

## Example

```
$ env-audit --file .env --required API_SECRET

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

## CI/CD

GitHub Actions:
```yaml
- name: Audit env
  run: |
    go install github.com/0xWhisp/env-audit@latest
    env-audit --required DATABASE_URL,API_KEY
```

GitLab CI:
```yaml
audit:
  script:
    - go install github.com/0xWhisp/env-audit@latest
    - env-audit --file .env --required DATABASE_URL
```

## Sensitive Key Patterns

Keys matching these patterns (case-insensitive) are flagged and redacted:

`SECRET` `PASSWORD` `TOKEN` `API_KEY` `APIKEY` `*KEY` `CREDENTIAL` `PRIVATE` `AUTH`

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](SECURITY.md).

## License

MIT
