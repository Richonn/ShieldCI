# Contributing

Thanks for your interest in ShieldCI.

This project is currently maintained as a personal project and is not open to external contributions. Bug reports and feature suggestions are welcome via [GitHub Issues](https://github.com/Richonn/ShieldCI/issues), but pull requests from external contributors will not be accepted at this time.

If you find a bug, feel free to open an issue — it may be addressed in a future release. Issue templates are provided for bug reports and feature requests. Please use the appropriate template when opening an issue.

## Developer Certificate of Origin (DCO)

All commits must include a `Signed-off-by` line asserting that you are legally authorized
to contribute the code under the project's MIT license, per the
[Developer Certificate of Origin v1.1](https://developercertificate.org/).

Add it automatically with:

```sh
git commit -s -m "your message"
```

This adds `Signed-off-by: Your Name <your@email.com>` to the commit. The DCO check runs
automatically in CI and blocks merges on unsigned commits.

## Coding standards

ShieldCI follows the official Go coding style:

- **Formatting**: [`gofmt`](https://pkg.go.dev/cmd/gofmt) — code must be `gofmt`-clean before merging
- **Linting**: [`golangci-lint`](https://golangci-lint.run/) — enforced automatically in CI (`.github/workflows/lint.yml`); any lint failure blocks the merge
- **Idioms**: standard Go idioms as described in [Effective Go](https://go.dev/doc/effective_go)

## Requirements for contributions

All changes (including maintainer commits) must meet the following criteria:

- Pass `golangci-lint` and `go test -race ./...`
- New features include tests (target >80% statement coverage on `detect` and `generate`)
- Generated workflow templates produce valid YAML
- No secrets or credentials committed — Gitleaks runs on every PR
- Action SHAs in templates must be pinned to a full commit SHA

## Developer quick setup

```sh
git clone https://github.com/Richonn/ShieldCI.git
cd ShieldCI

# Run tests
go test -race ./...

# Run tests with coverage (core packages)
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -E "(detect|generate|total)"

# Run linter
golangci-lint run ./...

# Build the binary
go build ./cmd/shieldci/...

# Run locally with act (requires Docker)
act -j test
```

All dependencies are declared in `go.mod` and fetched automatically by `go test` / `go build`. No additional installation steps are required.

## Reporting vulnerabilities

Do not open public issues for security vulnerabilities. Use [GitHub Security Advisories](https://github.com/Richonn/ShieldCI/security/advisories/new) instead.
