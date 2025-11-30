# Contributing

## Setup

```bash
git clone https://github.com/0xWhisp/env-audit.git
cd env-audit
go test ./...
```

Requires Go 1.21+.

## Structure

```
main.go      # CLI entry, args
scanner.go   # Orchestrates checks
checks.go    # Audit logic
parser.go    # .env parsing
output.go    # Formatting, redaction
env.go       # Reads os.Environ()
```

## Workflow

```bash
# Test
go test -v ./...

# Coverage
go test -cover ./...

# Build
go build -o env-audit .
```

## Commits

- Present tense: "Add feature" not "Added feature"
- One change per commit
- Keep it short

## PRs

- Tests must pass
- One logical change
- Update docs if needed

## Tests

Property-based tests use `github.com/leanovate/gopter` with 100+ iterations. Unit tests cover edge cases.

## Security

Before submitting:
- No secrets in test fixtures
- Redaction logic intact
- Tests don't leak real values

See [SECURITY.md](SECURITY.md).
