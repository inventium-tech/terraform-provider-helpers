# Contributing

Developer and automation change process for this repository.

See [AGENTS.md](./AGENTS.md) for canonical execution commands and automation constraints.
See [ARCHITECTURE.md](./ARCHITECTURE.md) for design context.

## Local Setup

Prerequisites:

- Go `1.25.x`
- Terraform CLI `>= 1.11`

Typical local bootstrap:

```sh
go mod download
```

## Branching Conventions

Create focused branches using one of these prefixes:

| Prefix            | Purpose                          |
|-------------------|----------------------------------|
| `feature/<title>` | Introduce a new feature          |
| `fix/<title>`     | Fix an existing bug              |
| `docs/<title>`    | Documentation-only changes       |
| `perf/<title>`    | Performance improvements         |
| `ci/<title>`      | CI/CD workflow changes           |
| `chore/<title>`   | Refactoring or maintenance tasks |

Keep each branch scoped to a single logical change.

## Commit Message Format

This repository uses Conventional Commits, interpreted by Semantic Release (`.releaserc.json`).

Format:

```text
type(scope): short description
```

Supported types in this repo include:
`release`, `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `ci`, `chore`.

Commit messages should explain **why** the change exists, not only what changed.

## Pull Request Process

1. Sync with the latest target branch (`main` or `dev`) before starting work.
2. Create a branch that follows [Branching Conventions](#branching-conventions).
3. Implement the change and keep commits focused.
4. Run relevant checks from [AGENTS.md](./AGENTS.md) (build/tests/docs generation/linting as applicable). For local
   linting, run MegaLinter (Docker), for example:

   ```sh
   docker run --rm -v $PWD:/tmp/lint:rw ghcr.io/oxsecurity/megalinter-go:latest
   ```

5. Open a PR with a title that follows [Commit Message Format](#commit-message-format).
6. Include change context and any testing evidence in the PR description.
7. Request review from maintainers (see [CODEOWNERS](./.github/CODEOWNERS)).
8. Ensure CI checks pass before marking the PR ready to merge.

## Proposing Design Changes

For structural or design-impacting changes:

1. Open an issue describing the problem and proposal.
2. Align on approach before implementation.
3. Update [ARCHITECTURE.md](./ARCHITECTURE.md) in the same PR when architectural behavior changes.

## Code of Conduct

<!-- TODO: add a link when a CODE_OF_CONDUCT.md (or external policy) is adopted -->
