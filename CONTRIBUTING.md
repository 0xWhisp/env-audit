# Contributing to env-audit

Thank you for your interest in contributing to env-audit! This document provides guidelines and instructions for contributing.

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/yourusername/env-audit.git
   cd env-audit
   ```
3. Create a feature branch:
   ```bash
   git checkout -b feature/amazing-feature
   ```

## Development Requirements

- Go 1.21 or higher
- No external runtime dependencies required

## Project Structure

```
env-audit/
├── main.go          # CLI entry point, argument parsing
├── main_test.go     # CLI tests
├── scanner.go       # Coordinates checks, aggregates results
├── checks.go        # Audit rules (empty, missing, sensitive)
├── checks_test.go   # Audit rule tests
├── parser.go        # .env file parsing and formatting
├── parser_test.go   # Parser tests
├── output.go        # Summary formatting, redaction
├── output_test.go   # Output tests
└── env.go           # Environment variable reader
```

## Development Workflow

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -v -run TestCheckEmpty ./...
```

### Building

```bash
# Build for current platform
go build -o env-audit .

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o env-audit-linux .
GOOS=darwin GOARCH=amd64 go build -o env-audit-darwin .
GOOS=windows GOARCH=amd64 go build -o env-audit.exe .
```

### Code Style

- Follow Go idioms and conventions
- Use `gofmt` for formatting
- Keep functions small and focused
- Write clear, descriptive names
- Add comments for exported functions

## Submitting Changes

1. Ensure all tests pass:
   ```bash
   go test -v ./...
   ```

2. Commit your changes with a clear message:
   ```bash
   git commit -m 'Add amazing feature'
   ```
   
   Commit message guidelines:
   - Use present tense ("Add feature" not "Added feature")
   - Keep the first line under 50 characters
   - One logical change per commit

3. Push to your branch:
   ```bash
   git push origin feature/amazing-feature
   ```

4. Open a Pull Request

## Pull Request Guidelines

- Provide a clear description of the changes
- Reference any related issues
- Ensure CI checks pass
- Keep PRs focused on a single change
- Update documentation if needed

## Testing Guidelines

### Unit Tests
- Test specific examples and edge cases
- Cover error conditions
- Use descriptive test names

### Property-Based Tests
- Use `github.com/leanovate/gopter` for property tests
- Run minimum 100 iterations
- Tag tests with the correctness property they validate:
  ```go
  // Feature: env-audit, Property 1: Empty value detection completeness
  ```

## Security Considerations

Before submitting:

- [ ] No sensitive values in test fixtures
- [ ] No hardcoded credentials
- [ ] Redaction logic preserved for all output paths
- [ ] Tests pass without exposing real secrets

See [SECURITY.md](SECURITY.md) for the full security policy.

## Reporting Issues

- Use GitHub Issues for bug reports and feature requests
- Include steps to reproduce for bugs
- Check existing issues before creating new ones

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow

## Questions?

Feel free to open an issue for any questions about contributing.

Thank you for contributing to env-audit!
