# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |

## Security Design Principles

env-audit is designed with security as a core principle:

### Value Redaction
- Sensitive values are **never** logged or displayed in output
- The `[REDACTED]` placeholder is used for all sensitive key values
- Redaction applies to both scan results and dump mode output

### Sensitive Key Detection
Keys are flagged as sensitive if they match these patterns (case-insensitive):
- `SECRET`
- `PASSWORD`
- `TOKEN`
- `API_KEY` / `APIKEY`
- `KEY` (as suffix, e.g., `STRIPE_KEY`)
- `CREDENTIAL`
- `PRIVATE`
- `AUTH`

### No Network Access
- env-audit operates entirely offline
- No data is transmitted externally
- No telemetry or analytics

### Minimal Dependencies
- Zero external runtime dependencies
- Single static binary
- Reduced attack surface

## Reporting a Vulnerability

If you discover a security vulnerability, please report it responsibly:

1. **Do NOT** open a public GitHub issue
2. Email security concerns to: [your-email@example.com]
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### Response Timeline

| Action | Timeframe |
|--------|-----------|
| Initial acknowledgment | 48 hours |
| Preliminary assessment | 7 days |
| Fix development | 14-30 days |
| Public disclosure | After fix release |

## Security Best Practices

When using env-audit in your projects:

### CI/CD Pipelines
```yaml
# Good: Check for misconfigurations
- run: env-audit --required DATABASE_URL,API_KEY

# Good: Validate .env files before deployment
- run: env-audit --file .env.production
```

### Local Development
```bash
# Audit before committing
env-audit --file .env.local

# Verify required vars are set
env-audit --required $(cat .env.required | tr '\n' ',')
```

### Production Environments
- Run env-audit as part of container startup validation
- Use exit codes to prevent deployment with missing configuration
- Never commit `.env` files to version control

## Known Limitations

1. **Pattern-based detection**: Sensitive key detection relies on naming patterns. Unconventionally named secrets may not be flagged.

2. **Value content not analyzed**: env-audit checks key names, not value content. A key named `DATA` containing a password would not be flagged.

3. **No encryption**: env-audit does not encrypt or decrypt values. It's an auditing tool, not a secrets manager.

## Security Checklist for Contributors

Before submitting a PR:

- [ ] No sensitive values in test fixtures
- [ ] No hardcoded credentials
- [ ] Redaction logic preserved for all output paths
- [ ] Tests pass without exposing real secrets
- [ ] No new external dependencies without security review

## Acknowledgments

We appreciate responsible disclosure from security researchers. Contributors who report valid vulnerabilities will be acknowledged (with permission) in release notes.
