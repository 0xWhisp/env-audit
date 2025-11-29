# env-audit

A minimal CLI tool that scans environment variables and .env files for common misconfigurations. Identifies security risks and configuration issues without exposing sensitive values.

## Features

- Detect empty environment variables
- Find missing required variables
- Flag sensitive keys (SECRET, PASSWORD, TOKEN, API_KEY, etc.)
- Parse and validate .env files
- Detect duplicate key definitions
- CI/CD friendly exit codes

## Installation

### From Source

```bash
go install github.com/[username]/env-audit@latest
```

### Build Locally

```bash
git clone https://github.com/[username]/env-audit.git
cd env-audit
go build -o env-audit .
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

# Show help
env-audit --help
```

### Options

| Flag | Short | Description |
|------|-------|-------------|
| `--file` | `-f` | Path to .env file to scan |
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

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
