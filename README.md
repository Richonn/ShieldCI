# ShieldCI

> GitHub Action that auto-generates hardened CI/CD DevSecOps pipelines and opens a PR with the generated workflows.

[![CI](https://github.com/Richonn/ShieldCI/actions/workflows/ci.yml/badge.svg)](https://github.com/Richonn/ShieldCI/actions/workflows/ci.yml)
[![OpenSSF Best Practices](https://www.bestpractices.dev/projects/12352/badge)](https://www.bestpractices.dev/projects/12352)

## Quick start

**1. Create a Personal Access Token** with scopes `repo` + `workflow` and store it as a secret (e.g. `GH_TOKEN`) in your repository.

**2. Add the action to your workflow:**

```yaml
- uses: Richonn/ShieldCI@v1
  with:
    github-token: ${{ secrets.GH_TOKEN }}
```

ShieldCI will detect your stack, generate the appropriate workflows, and open a PR.

> **Why a PAT?** GitHub blocks writes to `.github/workflows/` for `GITHUB_TOKEN` by design. A PAT with `workflow` scope is required to create workflow files.

## Inputs

| Input | Required | Default | Description |
|---|---|---|---|
| `github-token` | ✅ | — | Token for creating branches and PRs |
| `language` | ❌ | `auto` | Language override: `node`, `python`, `java`, `go`, `auto` |
| `docker` | ❌ | `auto` | Docker detection: `true`, `false`, `auto` |
| `kubernetes` | ❌ | `false` | Include K8s deploy workflow |
| `enable-trivy` | ❌ | `true` | Add Trivy image scan job |
| `enable-gitleaks` | ❌ | `true` | Add Gitleaks secret scan job |
| `enable-sast` | ❌ | `true` | Add SAST job (CodeQL or Semgrep) |
| `sast-tool` | ❌ | `codeql` | SAST tool: `codeql` or `semgrep` |
| `branch-name` | ❌ | `shieldci/generated-workflows` | Branch to push generated workflows to |
| `pr-title` | ❌ | `[ShieldCI] Add CI/CD DevSecOps pipeline` | PR title |
| `dry-run` | ❌ | `false` | If `true`, print generated workflows to the Job Summary without creating a branch or PR |
| `max-depth` | ❌ | `3` | Max directory depth for monorepo component detection |

## Outputs

| Output | Description |
|---|---|
| `pr-url` | URL of the created pull request |
| `detected-stack` | Detected stack as JSON |
| `generated-files` | Comma-separated list of generated file paths |

## Using outputs in downstream steps

```yaml
- name: Generate pipelines
  id: shieldci
  uses: Richonn/ShieldCI@v1
  with:
    github-token: ${{ secrets.GH_TOKEN }}

- name: Print detected stack
  run: |
    echo "Stack: ${{ steps.shieldci.outputs.detected-stack }}"
    echo "PR: ${{ steps.shieldci.outputs.pr-url }}"

- name: Conditional step based on detected stack
  if: ${{ fromJson(steps.shieldci.outputs.detected-stack).language == 'go' }}
  run: echo "Go project detected — run extra Go-specific steps here"
```

> `detected-stack` is a JSON string — use `fromJson()` to access individual fields (`language`, `docker`, `k8s`).

## Supported stacks

| Language | CI | Lint | Test | Build |
|---|---|---|---|---|
| Go | ✅ | golangci-lint | go test -race | go build |
| Node.js | ✅ | eslint | jest | npm/yarn build |
| Python | ✅ | ruff | pytest | build/poetry |
| Java | ✅ | — | mvn/gradle | mvn/gradle |
| Rust | ✅ | cargo clippy | cargo test | cargo build |

Docker and Kubernetes workflows are generated automatically when detected.

## Security tools

- **Gitleaks** — secret detection in git history
- **Trivy** — container vulnerability scanning with SARIF upload to GitHub Security tab
- **CodeQL / Semgrep** — static analysis (SAST)
- **Syft** — SBOM generation (Software Bill of Materials)
- **OpenSSF Scorecard** — automated security posture scoring (weekly + on push), results uploaded to GitHub Security tab
- **SLSA provenance** — cryptographic attestation of the build process (level 3), stored in the Rekor transparency log

### Semgrep custom rules

When `sast-tool: semgrep` is set and no `.semgrep/` directory exists in the target repo, ShieldCI generates a `.semgrep/rules/example.yml` file with a commented example rule to get you started.

If `.semgrep/` already exists, ShieldCI uses your existing rules (`--config=.semgrep/`) instead of the default community ruleset (`--config=auto`).

## Dry-run mode

Set `dry-run: "true"` to preview the generated workflows in the GitHub Actions Job Summary without touching your repository:

```yaml
- uses: Richonn/ShieldCI@v1
  with:
    github-token: ${{ secrets.GH_TOKEN }}
    dry-run: "true"
```

The Job Summary will display each generated workflow file as a fenced YAML block. No branch or PR is created.

## Versioning

ShieldCI uses a floating major tag (`v1`) that always points to the latest release in the `v1.x.x` series. This means `Richonn/ShieldCI@v1` automatically picks up new features and fixes without any change on your side.

The floating tag is updated automatically via a GitHub Actions workflow on every new release.

If you need reproducibility, pin to a specific version:

```yaml
- uses: Richonn/ShieldCI@v1.1.1
```

## Image signing with Cosign

When a `Dockerfile` is detected, ShieldCI generates a Docker workflow that automatically signs the built image using [Cosign](https://github.com/sigstore/cosign) in keyless mode via GitHub Actions OIDC.

No keys or secrets to manage — the signature is tied to the GitHub Actions identity and stored in the public [Rekor](https://rekor.sigstore.dev) transparency log.

The image is pushed to `ghcr.io/<owner>/<repo>:<sha>` and signed immediately after the build. Anyone can verify the signature with:

```sh
cosign verify ghcr.io/<owner>/<repo>:<sha> \
  --certificate-identity-regexp="https://github.com/<owner>/<repo>" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com"
```

## SBOM generation

ShieldCI generates two SBOM workflows via [Syft](https://github.com/anchore/syft):

- **`sbom.yml`** — always generated, analyses the repository source and dependencies
- **`sbom-docker.yml`** — generated when a `Dockerfile` is detected, builds the image and generates a SBOM from it

SBOM files are uploaded as artifacts and available from the Actions run summary.

## Monorepo support

ShieldCI automatically detects monorepos by scanning subdirectories up to a configurable depth. A separate workflow is generated per detected component, named `<component>-ci.yml`, `<component>-lint.yml`, etc.

**Supported monorepo layouts:**

```
my-monorepo/
├── backend-services/
│   ├── user-service/       # Go component → user-service-ci.yml
│   └── media-service/      # Rust component → media-service-ci.yml
└── tools/
    └── inspector/          # Python component → inspector-ci.yml
```

The following directories are automatically excluded from scanning: `node_modules`, `vendor`, `dist`, `build`, `target`, `docs`, `scripts`, and others.

Adjust scan depth with `max-depth` (default: `3`):

```yaml
- uses: Richonn/ShieldCI@v1
  with:
    github-token: ${{ secrets.GH_TOKEN }}
    max-depth: '4'
```

## Roadmap

- [x] Rust support
- [x] `dry-run` mode
- [x] Pinned action SHAs in generated workflows
- [x] Semgrep custom rules support
- [x] SBOM via Syft
- [x] Monorepo support
- [x] Image signing with Cosign (keyless via OIDC)
- [x] SLSA provenance via `slsa-github-generator` (level 3)
- [x] Build caching in generated workflows (Go modules, pip/poetry, npm/yarn, maven/gradle)
- [x] Multi-version matrix testing in generated workflows (Go, Rust, Java, Node, Python)
- [x] OpenSSF Scorecard integration
- [x] Concurrency groups in generated workflows (`cancel-in-progress`)
- [x] Go fuzz tests (`detect`, `generate`)
- [x] Workflow permission hardening (least privilege, job-level write scopes)
- [x] Dockerfile base image SHA pinning
- [x] Security policy (`SECURITY.md`)

## License

MIT
