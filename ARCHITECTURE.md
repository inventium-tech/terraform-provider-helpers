# Architecture

System design reference for `terraform-provider-helpers`.

See [README.md](./README.md) for end-user quickstart and [AGENTS.md](./AGENTS.md) for machine execution commands.

## 1. System Overview

This repository implements a Terraform provider that exposes **custom functions** (Terraform 1.8+) to extend the
Terraform language. It does not manage infrastructure resources or data sources.

The provider is focused on reusable helper behavior in three areas: collection transforms, object manipulation, and
environment-aware OS helpers.

## 2. Architectural Style

Layered package-oriented Go project:

- `internal/provider` holds Terraform framework function registrations and runtime behavior.
- `internal/validators` and `internal/custom_types` provide domain-specific validation/type support.
- `internal/utils` contains reusable helper logic consumed by provider functions.

## 3. Component Breakdown

| Component               | Responsibility                                          |
|-------------------------|---------------------------------------------------------|
| `main.go`               | Provider server entrypoint for Terraform plugin runtime |
| `internal/provider`     | Function definitions, schema, and tests                 |
| `internal/validators`   | Input and schema validation helpers                     |
| `internal/custom_types` | Shared custom type abstractions                         |
| `internal/utils`        | Reusable low-level utility functions                    |
| `templates/functions`   | Human-authored docs templates                           |
| `examples/functions`    | Runnable Terraform examples used by docs/tests          |
| `docs/functions`        | Generated end-user function reference                   |
| `tools/tools.go`        | `go generate` bridge for doc generation                 |

## 4. Data & Control Flow

1. Terraform loads the provider plugin.
2. Provider registers available helper functions.
3. A Terraform configuration calls `provider::helpers::<function_name>`.
4. The framework validates arguments and invokes the function implementation in `internal/provider`.
5. Provider logic delegates to validators/utils/custom types as needed and returns a Terraform value.

Documentation flow:

1. Contributors update `templates/functions/*` and `examples/functions/*`.
2. `go generate ./tools` formats examples and regenerates `docs/functions/*`.

## 5. External Dependencies

| Dependency                             | Type            | Purpose                                             |
|----------------------------------------|-----------------|-----------------------------------------------------|
| Terraform CLI                          | Tooling/runtime | Runs plans/applies and acceptance/integration tests |
| Terraform Plugin Framework             | Go library      | Provider function implementation model              |
| terraform-plugin-docs (`tfplugindocs`) | Tooling         | Generates function docs from templates/examples     |
| GitHub Actions                         | CI/CD platform  | Build/test/lint/release workflows                   |
| semantic-release + goreleaser          | Release tooling | Versioning, release notes, and artifact publishing  |

## 6. Repository Structure

Top-level structure is optimized for a provider codebase with generated docs and testable examples.

- Runtime code: `internal/**`
- Generated docs: `docs/**`
- Docs sources: `templates/**`, `examples/**`
- Automation: `.github/workflows/**`, `tools/tools.go`

## 7. Design Constraints & Decisions

- **Function-only scope:** keep provider focused on language helpers instead of resource management.
- **Generated documentation:** `docs/` is an output artifact; source-of-truth lives in `templates/` and `examples/`.
- **Test-first change safety:** provider behavior changes should be validated with `go test` and, when relevant,
  `TF_ACC=1` provider tests.
- **Release automation compatibility:** commit semantics are designed to work with semantic-release rules in
  `.releaserc.json`.

## 8. Evolution Strategy

- Add new helpers by implementing in `internal/provider`, plus aligned tests, examples, and templates.
- Keep function naming and argument patterns consistent with existing provider conventions.
- When introducing behavioral changes, update tests and generated docs in the same change.
