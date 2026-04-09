# Technical Decisions

## Architecture overview

ShieldCI is structured as a Docker-based GitHub Action. The binary is a single statically-linked Go executable compiled at build time and embedded in an Alpine container.

```
action.yml                  ← GitHub Action entrypoint (inputs/outputs declaration)
Dockerfile                  ← Multi-stage build: golang:alpine → alpine runtime
cmd/shieldci/main.go        ← Entry point: reads env, calls detect → generate → pr
internal/
  config/config.go          ← Maps INPUT_* env vars to a typed Config struct
  detect/detect.go          ← Inspects the workspace to determine language/docker/k8s
  generate/generate.go      ← Renders embedded text/template files into workflow YAML
  generate/templates/       ← Embedded workflow templates (base, go, node, python, java, rust, docker)
  pr/pr.go                  ← GitHub API client: creates branch, commits files, opens PR
```

**Data flow:**

1. `config` reads `INPUT_*` environment variables injected by the GitHub Actions runner
2. `detect` walks the workspace filesystem to identify the language stack and optional components (Docker, Kubernetes, monorepo)
3. `generate` renders the appropriate `text/template` files using the detected stack, producing workflow YAML strings
4. `pr` uses the GitHub Contents API to commit each file to a new branch and open a pull request

**Trust boundaries:**

- The action runs inside a Docker container on the GitHub Actions runner (ephemeral, isolated VM)
- It receives a PAT from the caller via an environment variable — the token is never logged or included in outputs
- It writes only to `.github/workflows/` on the caller's repository via the GitHub API — no other filesystem or network access occurs
- Templates are embedded at compile time; no external template fetching occurs at runtime

## Assurance case

### Threat model

| Threat | Mitigation |
|--------|-----------|
| Supply chain attack via compromised action | All third-party actions pinned to full commit SHAs; Dependabot monitors for updates |
| Supply chain attack via compromised base image | Dockerfile base images pinned by SHA digest |
| Secret leakage via committed credentials | Gitleaks scans every commit; CONTRIBUTING.md mandates no secrets |
| Container vulnerability exploitation | Trivy scans with `--exit-code 1` on CRITICAL/HIGH; Alpine minimal base image |
| Dependency vulnerability exploitation | Dependabot weekly updates; Trivy Go module scanning |
| Malicious input via action inputs | All inputs validated and typed via `config.go`; language/docker/kubernetes are enum values |
| Build tampering | SLSA Level 3 provenance generated on every release; stored in Rekor transparency log |
| Token over-permissioning | PAT scopes limited to `repo` + `workflow`; documented in README |
| Workflow injection via generated YAML | Templates use `text/template` with no user-controlled interpolation into shell commands |

### Secure design principles applied

- **Least privilege**: job-level permissions in all CI workflows; Docker container runs with no escalated host privileges
- **Defense in depth**: Gitleaks + CodeQL + Trivy + Dependabot + SLSA — multiple independent layers
- **Minimal attack surface**: 2 direct Go dependencies; Alpine base; static binary with no runtime file dependencies
- **Fail secure**: Trivy and CodeQL configured to fail the build on policy violations
- **Separation of concerns**: token never touches the filesystem; written only to environment variable

### Common implementation weaknesses countered

- **CWE-20 (Improper Input Validation)**: All action inputs are validated in `config.go`; enum inputs reject unknown values
- **CWE-312 (Cleartext Storage of Sensitive Information)**: PAT is stored in GitHub Secrets and passed as an env var; never written to disk or logs
- **CWE-78 (OS Command Injection)**: The action makes no shell exec calls; all GitHub API operations use the typed Go client
- **CWE-494 (Download of Code Without Integrity Check)**: All external actions and base images are pinned to SHAs

## Docker action vs composite action

ShieldCI uses a Docker action instead of a composite action because:
- The Go binary needs to be compiled — composite actions only support shell scripts and other actions
- Docker gives a fully reproducible environment with pinned dependencies
- The binary can be tested locally with `act`

## `text/template` vs external templating engine

Go's standard `text/template` was chosen over alternatives (Jinja2, Handlebars, etc.) because:
- Zero external dependencies for the templating itself
- Ships with Go — no extra `go get`
- `embed.FS` + `text/template` gives a single self-contained binary with templates baked in

## `embed.FS` for templates

Templates are embedded into the binary at compile time using `//go:embed`. This means:
- The Docker image only needs the compiled binary — no need to copy template files separately
- Templates can't be accidentally missing at runtime

## GitHub API via `google/go-github`

The official Go client for the GitHub API was chosen because:
- Typed structs for all API responses — no manual JSON parsing
- Actively maintained by Google
- Full coverage of the Git Data API needed for branch/commit/PR creation

## Input mapping in `action.yml`

Docker actions do NOT automatically expose inputs as `INPUT_*` environment variables (unlike JavaScript actions). The `env:` block under `runs:` is mandatory to bridge inputs to the container.

The token input is mapped to `SHIELDCI_TOKEN` (not `GITHUB_TOKEN`) to avoid collision with the runner's auto-injected `GITHUB_TOKEN`, which would otherwise override our mapping.

## Output via `$GITHUB_OUTPUT`

The deprecated `::set-output::` workflow command is ignored on current runners. All outputs are written by appending `key=value` to the file at `$GITHUB_OUTPUT`.

Multi-line values are not supported in this format — `generated-files` uses comma-separated paths instead of newlines.

## PAT required — `GITHUB_TOKEN` cannot write to `.github/workflows/`

GitHub blocks any write to `.github/workflows/` from `GITHUB_TOKEN`, regardless of the `permissions:` block in the workflow YAML. This is a deliberate security measure to prevent workflow injection attacks. The `workflows` scope is not exposed as a valid key in the `permissions:` block — it is only available via PAT (classic, `repo` + `workflow` scopes) or a GitHub App.

ShieldCI therefore requires a PAT. This is the same constraint faced by any action that creates or modifies workflow files.

## Pinned action SHAs in generated workflows

All third-party GitHub Actions referenced in generated workflows are pinned to their commit SHA rather than a version tag. This protects against supply chain attacks where a tag could be silently moved to a different (potentially malicious) commit.

Version comments (`# vX.Y.Z`) are included alongside each SHA so Dependabot can parse and update them automatically when new versions are released.

ShieldCI's own CI workflows follow the same convention.

**Exception — `slsa-framework/slsa-github-generator`**: this reusable workflow explicitly rejects SHA-based refs at runtime and requires a `refs/tags/vX.Y.Z` reference. It is therefore pinned by tag only. This is an intentional design choice by the SLSA generator maintainers to guarantee the workflow version matches a signed release. The Scorecard `Pinned-Dependencies` check will report a score of 9/10 as a result; this is expected and acceptable.

## Binary release artifacts and Signed-Releases

ShieldCI is distributed as a Docker image (GitHub Action consumers pull from GHCR), so binary artifacts are not required for the action itself. They are built and published on every release specifically to satisfy the OpenSSF Scorecard `Signed-Releases` check, which requires release artifacts to carry cryptographic provenance.

On each tag push, `release.yml` builds five static binaries (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64) using `-trimpath` for reproducibility, computes a `checksums.txt` SHA256 manifest, creates the GitHub Release, and hands the hashes to `generator_generic_slsa3` (SLSA Level 3) which generates and attaches a `.intoto.jsonl` provenance file.

This approach was chosen over GPG or cosign blob signing because provenance via SLSA satisfies both the `not signed` and `no provenance` Scorecard warnings in a single step, is keyless (no key management), and is consistent with the container provenance already generated by `generator_container_slsa3`.

## Semgrep image rename

Semgrep rebranded their Docker image from `returntocorp/semgrep` to `semgrep/semgrep`. Generated workflows use the new image name.

## Semgrep custom rules bootstrap

When `sast-tool: semgrep` is selected and no `.semgrep/` directory is detected in the target repo, ShieldCI generates a `.semgrep/rules/example.yml` file with a commented starter rule. This lowers the barrier to writing custom rules while keeping the workflow functional out of the box.

If `.semgrep/` already exists, ShieldCI passes `--config=.semgrep/` to use the existing rules instead of `--config=auto`.

## Contents API vs Git Data API

The initial implementation used the Git Data API (blob → tree → commit → UpdateRef) to create a single atomic commit. This approach returned `403 Resource not accessible by integration` consistently despite correct permissions.

The Contents API (`PUT /repos/{owner}/{repo}/contents/{path}`) was adopted instead. It creates one commit per file but works reliably with both `GITHUB_TOKEN` (for non-workflow paths) and PAT. The tradeoff — N commits instead of 1 — is acceptable given the use case (one-shot PR generation).

## SBOM split: repo vs Docker image

SBOM generation is split into two separate workflows (`sbom.yml` and `sbom-docker.yml`) rather than one combined workflow because:
- `sbom.yml` is always relevant — every project has source dependencies
- `sbom-docker.yml` is only relevant when a `Dockerfile` is detected
- Merging them would require building the Docker image unconditionally, adding unnecessary overhead to repos without Docker
- Each workflow has a single clear responsibility

The Docker SBOM job builds and analyses the image in a single job (rather than two separate jobs) to avoid the overhead of uploading/downloading a potentially large image tarball between runner instances.
