# Governance

## Project roles

| Role | Who | Responsibilities |
|------|-----|-----------------|
| **Maintainer** | [@Richonn](https://github.com/Richonn) | All decisions: code review, releases, roadmap, security response, repository administration |

ShieldCI is a solo-maintained project. The maintainer holds all roles and is responsible for:

- Reviewing and merging changes to the codebase
- Publishing releases and updating the floating `v1` tag
- Triaging bug reports and feature requests via GitHub Issues
- Responding to security vulnerability reports within 7 days (see [SECURITY.md](SECURITY.md))
- Managing repository settings, branch protection rules, and secrets

## Decision-making

Decisions are made unilaterally by the maintainer. For any significant design change, the rationale is documented in [DECISIONS.md](DECISIONS.md). Community members are welcome to open issues to propose changes or raise concerns.

## Contributions

This project does not currently accept pull requests from external contributors. Bug reports and feature suggestions via GitHub Issues are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

## Continuity

To ensure the project can continue operating in the event the maintainer becomes unavailable:

- The full source code and history are publicly available under the MIT license, allowing any party to fork and continue the project independently.
- The MIT license explicitly grants all rights needed to use, modify, and redistribute the software without the original maintainer's involvement.
- Release artifacts (Docker images, SLSA provenance) are published to public registries (`ghcr.io`) and the Rekor transparency log, ensuring existing releases remain verifiable and accessible indefinitely.
- All CI/CD configuration, templates, and documentation are version-controlled and self-contained in this repository — there are no external dependencies required to build or run the project beyond what is declared in `go.mod` and `Dockerfile`.

In the event of extended maintainer unavailability, the community is encouraged to fork the repository. The floating `v1` tag points to the last published release and will continue to function for existing users.

## Bus factor

The current bus factor is 1 (single maintainer). This is a known limitation of solo-maintained projects. The continuity measures above are designed to mitigate this risk.
