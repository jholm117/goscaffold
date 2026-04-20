# goscaffold Design Spec

## Overview

A Go CLI tool that scaffolds new Go projects following Jeff's engineering standards (`~/.claude/go-standards.md`, `~/.claude/general-standards.md`). Replaces the `go-template` GitHub template repo with a parameterized, composable generator that supports project evolution.

## Problem

Go projects fall on a spectrum — pure CLIs (hackerrank-cli, infra-audit), K8s controllers (kubecracker, helm-sync), and projects that start as CLIs and evolve into controllers (backflow). A single template repo either includes irrelevant files that need deletion or misses files that need manual addition. At ~monthly project creation frequency, this friction compounds.

## Solution

A standalone CLI with two commands:

- `goscaffold init` — creates a new project from scratch
- `goscaffold add` — layers capabilities onto an existing project

## Commands

### `goscaffold init <project-name>`

Creates a new directory, initializes git, writes all files, runs `go mod tidy`, sets up hooks.

```
goscaffold init backflow --cli
goscaffold init helm-sync --controller
goscaffold init kubecracker --cli --controller
goscaffold init backflow --cli --module github.com/jholm117/backflow
```

Flags:
- `--module <path>` — Go module path. Defaults to `github.com/jholm117/<project-name>`.
- `--cli` — include CLI distribution layer (GoReleaser, Homebrew tap, release workflow)
- `--controller` — include K8s controller layer (Dockerfile, docker targets, kind e2e scaffolding)

At least one of `--cli` or `--controller` is required.

### `goscaffold add <layer>`

Runs inside an existing project directory. Adds files for the specified layer. Skips files that already exist (prints a message, does not overwrite). Updates the Makefile to include the new layer's targets.

```
cd ~/wip-repos/backflow
goscaffold add controller
```

Layers: `cli`, `controller`.

## File Generation Matrix

| File | Base | `cli` | `controller` |
|---|---|---|---|
| `cmd/<name>/main.go` | x | | |
| `go.mod` | x | | |
| `internal/.gitkeep` | x | | |
| `pkg/.gitkeep` | x | | |
| `Makefile` | x | merged | merged |
| `.golangci.yml` | x | | |
| `.gitignore` | x | | |
| `hack/ci-checks.sh` | x | | |
| `.githooks/pre-push` | x | | |
| `.github/workflows/ci.yml` | x | | |
| `AGENTS.md` | x | | |
| `README.md` | x | | |
| `.goreleaser.yaml` | | x | |
| `.github/workflows/release.yml` | | x | |
| `Dockerfile` | | | x |
| `.dockerignore` | | | x |
| `hack/e2e-up.sh` | | | x |
| `hack/e2e-down.sh` | | | x |
| `hack/kind-config.yaml` | | | x |

## Makefile Merging

The Makefile is the one file where content from multiple layers combines into a single output. Base provides: `build`, `install`, `test`, `lint`, `lint-fix`, `lint-config`, `fmt`, `vet`, `govulncheck`, `setup-hooks`, `clean`, `help`, and the tool installation infrastructure.

- `cli` adds: `release-snapshot` target, `GORELEASER_VERSION` variable, `goreleaser` tool install
- `controller` adds: `docker-build`, `docker-push`, `e2e`, `e2e-up`, `e2e-down` targets, `IMG` variable

The `add` command reads the existing Makefile and appends the new layer's section. It detects whether a layer's targets are already present by checking for a sentinel comment (e.g., `## CLI Targets` or `## Controller Targets`).

## Template Parameters

```go
type Params struct {
    ProjectName string // e.g. "backflow"
    Module      string // e.g. "github.com/jholm117/backflow"
    CLI         bool
    Controller  bool
}
```

All templates receive the full `Params` struct. Templates use `text/template` syntax.

## Architecture

```
goscaffold/
├── cmd/goscaffold/
│   └── main.go                  # cobra root + init/add subcommands
├── internal/
│   ├── scaffold/
│   │   ├── scaffold.go          # Core: render templates, write files, skip existing
│   │   ├── scaffold_test.go
│   │   ├── init.go              # init command logic
│   │   ├── init_test.go
│   │   ├── add.go               # add command logic
│   │   ├── add_test.go
│   │   └── templates/           # embed.FS
│   │       ├── base/
│   │       │   ├── Makefile.tmpl
│   │       │   ├── golangci.yml.tmpl
│   │       │   ├── gitignore.tmpl
│   │       │   ├── ci.yml.tmpl
│   │       │   ├── ci-checks.sh.tmpl
│   │       │   ├── pre-push.tmpl
│   │       │   ├── main.go.tmpl
│   │       │   ├── go.mod.tmpl
│   │       │   ├── agents.md.tmpl
│   │       │   └── readme.md.tmpl
│   │       ├── cli/
│   │       │   ├── goreleaser.yaml.tmpl
│   │       │   ├── release.yml.tmpl
│   │       │   └── makefile-cli.tmpl
│   │       └── controller/
│   │           ├── Dockerfile.tmpl
│   │           ├── dockerignore.tmpl
│   │           ├── e2e-up.sh.tmpl
│   │           ├── e2e-down.sh.tmpl
│   │           ├── kind-config.yaml.tmpl
│   │           └── makefile-controller.tmpl
│   └── makefile/
│       ├── merge.go             # Append layer sections to existing Makefile
│       └── merge_test.go
├── .golangci.yml
├── .gitignore
├── .githooks/pre-push
├── .github/workflows/
│   ├── ci.yml
│   └── release.yml
├── .goreleaser.yaml
├── hack/ci-checks.sh
├── Makefile
├── AGENTS.md
├── README.md
├── go.mod
└── go.sum
```

Key decisions:
- **`embed.FS`** for templates — single binary distribution, no external files
- **`text/template`** — templates are readable, close to final output
- **File existence checks** for `add` — no state file, just check what's on disk
- **Sentinel comments** in Makefile for layer detection

## Generated File Content

All generated files follow the standards defined in `~/.claude/go-standards.md` and `~/.claude/general-standards.md`. Specifically:

- **Makefile**: versioned tool symlinks in `bin/`, `go-install-tool` function, `LOCALBIN` pattern
- **ci-checks.sh**: `--parallel` flag support (sequential locally by default, parallel in CI). Single source of truth for hooks and CI.
- **CI workflow**: reads Go version from `go.mod`, caches `bin/`, concurrency guard, triggers on push to main + PRs
- **golangci-lint**: v2 config with the standard linter set
- **pre-push hook**: calls `hack/ci-checks.sh`
- **main.go**: cobra root command with `version` var for ldflags injection
- **AGENTS.md**: stub documenting project layout, build commands, and conventions
- **README.md**: overview, getting started, and development sections. Content adapts based on `--cli`/`--controller` flags.
- **GoReleaser** (cli): linux/darwin x amd64/arm64, ldflags version injection, Homebrew tap at `jholm117/homebrew-tap`
- **Release workflow** (cli): triggers on `v*` tags, uses `goreleaser/goreleaser-action@v6`
- **Dockerfile** (controller): multi-stage, `golang:1.25-alpine` builder, `gcr.io/distroless/static:nonroot` runtime, `CGO_ENABLED=0`, `-trimpath`, `-s -w` ldflags
- **.dockerignore** (controller): excludes `.git`, `.github`, `.githooks`, `.worktrees`, `bin/`, `dist/`
- **kind e2e scripts** (controller): idempotent `e2e-up.sh` (create cluster, build image, load, deploy), `e2e-down.sh` (delete cluster), `kind-config.yaml` (single control-plane)

## Testing

### Unit Tests

Test template rendering in isolation. Given `Params`, assert rendered output matches golden files in `testdata/`.

- Template rendering: each template produces expected content for each flag combination
- Makefile merging: base + cli, base + controller, base + cli + controller all produce valid Makefiles

### Integration Tests

One test per mode (`--cli`, `--controller`, `--cli --controller`):

1. Run `init` into `t.TempDir()`
2. Assert all expected files exist
3. Assert `go build ./...` succeeds
4. Assert `go vet ./...` passes
5. Assert `make lint-config` passes (validates generated `.golangci.yml`)

### Add Tests

1. Run `init --cli` into temp dir
2. Run `add controller` in same dir
3. Assert controller files appear
4. Assert Makefile now has docker targets
5. Run `add cli` again — assert it skips without overwriting, prints message

### No E2E for Heavy Dependencies

No tests that run `kind`, `goreleaser`, `docker build`, or `brew`. Those depend on external tools. The integration tests prove the generated project is valid and buildable.

## Distribution

GoReleaser + Homebrew tap (`jholm117/homebrew-tap`). Install via:

```
brew install jholm117/tap/goscaffold
```

goscaffold itself is built with the same standards it generates — it eats its own dog food.

## What Happens to go-template

Archive `jholm117/go-template` once goscaffold is working. It's superseded.
