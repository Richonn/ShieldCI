# Contributing

Thanks for your interest in ShieldCI.

This project is currently maintained as a personal project and is not open to external contributions. Bug reports and feature suggestions are welcome via [GitHub Issues](https://github.com/Richonn/ShieldCI/issues), but pull requests from external contributors will not be accepted at this time.

If you find a bug, feel free to open an issue — it may be addressed in a future release.

## Requirements for contributions

All changes (including maintainer commits) must meet the following criteria:

- Pass `golangci-lint` and `go test -race ./...`
- New features include tests (target >80% coverage on `detect` and `generate`)
- Generated workflow templates produce valid YAML
- No secrets or credentials committed — Gitleaks runs on every PR
- Action SHAs in templates must be pinned to a full commit SHA

## Reporting vulnerabilities

Do not open public issues for security vulnerabilities. Use [GitHub Security Advisories](https://github.com/Richonn/ShieldCI/security/advisories/new) instead.
