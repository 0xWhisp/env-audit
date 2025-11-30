# Security Policy

## Supported Versions

| Version | Supported |
| ------- | --------- |
| 1.x.x   | ✓         |

## How It Works

- Sensitive values never appear in output—only `[REDACTED]`
- No network calls, no telemetry
- Zero runtime dependencies

### Sensitive Patterns

Keys matching these (case-insensitive) get flagged:

`SECRET` `PASSWORD` `TOKEN` `API_KEY` `APIKEY` `*KEY` `CREDENTIAL` `PRIVATE` `AUTH`

## Reporting Vulnerabilities

Don't open a public issue. Email: 0xWhisp@proton.me

Include:
- What you found
- How to reproduce
- Impact assessment

| Response | Timeline |
|----------|----------|
| Acknowledgment | 48h |
| Assessment | 7 days |
| Fix | 14-30 days |

## Limitations

- Pattern-based detection only—unconventional secret names may slip through
- Checks key names, not values—`DATA=hunter2` won't be flagged
- Not a secrets manager—just an auditor
