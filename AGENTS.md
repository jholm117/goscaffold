# goscaffold

Go CLI scaffolding tool. Generates project structure with composable layers.

## Build & Test

- `make build` — build binary to `bin/goscaffold`
- `make test` — run all tests
- `make lint` — run golangci-lint
- `make setup-hooks` — install pre-push hook

## Architecture

- `cmd/goscaffold/main.go` — cobra CLI entry point with `init` and `add` subcommands
- `internal/config/` — loads `~/.config/goscaffold/config.yaml`
- `internal/scaffold/` — core engine: template rendering, file writing, init/add orchestration
- `internal/scaffold/templates/` — embedded Go text/template files, organized by layer (base, cli, controller, helm)
- `internal/makefile/` — Makefile section merging for the `add` command

## Conventions

- Templates use `text/template` syntax with `Params` struct (ProjectName, Module, CLI, Controller, Helm)
- Makefile uses sentinel comments (`## CLI Targets`, `## Controller Targets`, `## Helm Targets`) for layer detection
- `add` never overwrites existing files — prints a skip message instead
