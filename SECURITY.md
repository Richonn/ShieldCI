# Security Policy

## Supported Versions

| Version | Supported | Support scope |
|---------|-----------|---------------|
| v1.x    | ✅        | Security fixes, critical bug fixes, feature updates |

**Support scope**: The active v1.x release line receives security patches, critical bug fixes, and new features. Non-critical bugs are addressed on a best-effort basis.

**Duration**: v1.x support continues until a v2.x major release is published. There is no fixed end date for v1.x at this time.

**End-of-life notice**: When a new major version (v2.x) is released, v1.x will enter a **90-day critical-security-only maintenance window**, after which it will be marked end-of-life. An announcement will be made in the GitHub repository and in the release notes with at least 90 days' notice before support ends.

## SCA (Software Composition Analysis) Policy

**Remediation threshold**: All **CRITICAL** and **HIGH** severity vulnerabilities in dependencies and container base images must be resolved before merging. MEDIUM and LOW severity findings are tracked and addressed on a best-effort basis in the next release cycle.

**Pre-release gate**: The CI pipeline runs Trivy with `--severity CRITICAL,HIGH --exit-code 1` on every PR and push. Any unresolved CRITICAL or HIGH finding blocks the build and prevents a release from being published. Releases are not created unless all CI jobs pass.

**License policy**: Dependencies must be compatible with the MIT license. Go module licenses are validated by the `anchore/sbom-action` SBOM output.

**Dependency updates**: Dependabot is configured for weekly automated dependency updates (Go modules + GitHub Actions), ensuring vulnerabilities are caught and patched promptly.

## Secrets and Credentials Policy

- **PAT tokens**: Users store their PAT in GitHub Secrets (`Settings → Secrets`). The minimum required scopes are `repo` + `workflow` — no broader access should be granted.
- **No long-lived signing keys**: ShieldCI uses keyless signing (Cosign via GitHub OIDC) for Docker images. No private keys are stored in the repository or secrets.
- **Secret scanning**: Gitleaks runs on every commit to block accidental credential commits (enforced in CI — see `.github/workflows/security.yml`).
- **Runtime secrets**: The action receives the token as an environment variable (`SHIELDCI_TOKEN`). It is never logged or exposed in outputs.
- **Rotation**: If a PAT is compromised, revoke it immediately in GitHub Developer Settings and generate a new one with the minimum required scopes.
- **GITHUB_TOKEN**: Used for ephemeral operations (package registry login, provenance). Scoped to minimum permissions per job.

## Reporting a Vulnerability

Please **do not** open a public GitHub issue for security vulnerabilities.

Report vulnerabilities privately via [GitHub Security Advisories](https://github.com/Richonn/ShieldCI/security/advisories/new).

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

You will receive a response within 7 days.
