# AGENTS.md

Machine execution rules for AI agents and automation in this repository.

## Project Structure

Important paths for automated work:

- `.github/workflows/` — CI/release automation definitions
- `internal/provider/` — Terraform provider function implementations and tests
- `internal/utils/`, `internal/validators/`, `internal/custom_types/` — shared logic and validations
- `docs/functions/` — generated function reference docs
- `templates/functions/` — source templates used for generated docs
- `examples/functions/` — Terraform examples consumed by docs/tests
- `tools/tools.go` — go:generate entrypoint for docs generation

## Commands

All commands are non-interactive and suitable for automation.

### Setup

```sh
go mod download
```

Expected exit code: `0`.

### Test

```sh
go test ./... -v
```

Expected exit code: `0`.

Acceptance/integration path (long-running):

```sh
TF_ACC=1 go test ./internal/provider -v
```

Expected exit code: `0`.

### Build / Run

```sh
go build -v ./...
```

Expected exit code: `0`.

### Lint (MegaLinter)

Preferred local run (Docker):

```sh
docker run --rm -v $PWD:/tmp/lint:rw ghcr.io/oxsecurity/megalinter-go:latest
```

For CI parity, use the pinned image tag:

```sh
docker run --rm -v $PWD:/tmp/lint:rw ghcr.io/oxsecurity/megalinter-go:v9.4.0
```

This repository uses `.mega-linter.yml` (with `LINTER_RULES_PATH: .linters`) and writes reports to `.ml-reports/`.

Expected exit code: `0`.

### Documentation Generation

```sh
go generate ./tools
```

Expected exit code: `0`.

## Constraints

- Do not commit or rely on local IDE state under `.idea/`.
- Generated docs in `docs/` should be updated via `go generate ./tools`, with source edits in `templates/` and `examples/`.
- Avoid editing release automation semantics unless explicitly requested (`.releaserc.json`, `.github/workflows/release.yml`).
- Prefer narrow, file-scoped changes; keep existing repository conventions.

## CI/CD Automation Reference

- CI workflow: `.github/workflows/ci.yml` (`go build`, linters, unit tests, integration tests, security scans)
- MegaLinter in CI: `oxsecurity/megalinter/flavors/go@v9.4.0` (reports archived from `.ml-reports/`)
- Release workflow: `.github/workflows/release.yml` (semantic-release + goreleaser)
