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
goscaffold init helm-sync --controller --helm
goscaffold init kubecracker --cli --controller --helm
goscaffold init backflow --cli --module github.com/jholm117/backflow
```

Flags:
- `--module <path>` — Go module path. Defaults to `github.com/jholm117/<project-name>`.
- `--cli` — include CLI distribution layer (GoReleaser, Homebrew tap, release workflow)
- `--controller` — include K8s controller layer (Dockerfile, docker targets, kind e2e, controller-runtime reconciler skeleton)
- `--helm` — include Helm chart scaffolding (chart layout, templates, values, RBAC, security context per helm-standards.md)

At least one of `--cli`, `--controller`, or `--helm` is required. Flags are composable — `--cli --controller --helm` includes everything.

### `goscaffold add <layer>`

Runs inside an existing project directory. Adds files for the specified layer. Skips files that already exist (prints a message, does not overwrite). Updates the Makefile to include the new layer's targets.

```
cd ~/wip-repos/backflow
goscaffold add controller
goscaffold add helm
```

Layers: `cli`, `controller`, `helm`.

## File Generation Matrix

| File | Base | `cli` | `controller` | `helm` |
|---|---|---|---|---|
| `cmd/<name>/main.go` | x | | | |
| `go.mod` | x | | | |
| `internal/.gitkeep` | x | | | |
| `pkg/.gitkeep` | x | | | |
| `Makefile` | x | merged | merged | merged |
| `.golangci.yml` | x | | | |
| `.gitignore` | x | | | |
| `hack/ci-checks.sh` | x | | | |
| `.githooks/pre-push` | x | | | |
| `.github/workflows/ci.yml` | x | | | |
| `AGENTS.md` | x | | | |
| `README.md` | x | | | |
| `.goreleaser.yaml` | | x | | |
| `.github/workflows/release.yml` | | x | | |
| `Dockerfile` | | | x | |
| `.dockerignore` | | | x | |
| `internal/controller/reconciler.go` | | | x | |
| `hack/e2e-up.sh` | | | x | |
| `hack/e2e-down.sh` | | | x | |
| `hack/kind-config.yaml` | | | x | |
| `charts/<name>/Chart.yaml` | | | | x |
| `charts/<name>/values.yaml` | | | | x |
| `charts/<name>/.helmignore` | | | | x |
| `charts/<name>/templates/_helpers.tpl` | | | | x |
| `charts/<name>/templates/NOTES.txt` | | | | x |
| `charts/<name>/templates/deployment.yaml` | | | | x |
| `charts/<name>/templates/service.yaml` | | | | x |
| `charts/<name>/templates/serviceaccount.yaml` | | | | x |
| `charts/<name>/templates/clusterrole.yaml` | | | | x |
| `charts/<name>/templates/clusterrolebinding.yaml` | | | | x |

## Makefile Merging

The Makefile is the one file where content from multiple layers combines into a single output. Base provides: `build`, `install`, `test`, `lint`, `lint-fix`, `lint-config`, `fmt`, `vet`, `govulncheck`, `setup-hooks`, `clean`, `help`, and the tool installation infrastructure.

- `cli` adds: `release-snapshot` target, `GORELEASER_VERSION` variable, `goreleaser` tool install
- `controller` adds: `docker-build`, `docker-push`, `e2e`, `e2e-up`, `e2e-down` targets, `IMG` variable
- `helm` adds: `helm-lint`, `helm-template` targets

The `add` command reads the existing Makefile and appends the new layer's section. It detects whether a layer's targets are already present by checking for a sentinel comment (e.g., `## CLI Targets`, `## Controller Targets`, `## Helm Targets`).

## Configuration

goscaffold reads `~/.config/goscaffold/config.yaml` for user-level settings:

```yaml
module_prefix: github.com/jholm117
homebrew_tap_token: "<fine-grained PAT with contents:write on homebrew-tap repo>"
```

- `module_prefix` — default Go module prefix. `--module` defaults to `<module_prefix>/<project-name>` if omitted.
- `homebrew_tap_token` — GitHub PAT for pushing Homebrew formulas. When `init --cli` creates a repo, goscaffold runs `gh secret set HOMEBREW_TAP_GITHUB_TOKEN` using this token automatically.

## Template Parameters

```go
type Params struct {
    ProjectName string // e.g. "backflow"
    Module      string // e.g. "github.com/jholm117/backflow"
    CLI         bool
    Controller  bool
    Helm        bool
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
│   │       ├── controller/
│   │       │   ├── Dockerfile.tmpl
│   │       │   ├── dockerignore.tmpl
│   │       │   ├── reconciler.go.tmpl
│   │       │   ├── e2e-up.sh.tmpl
│   │       │   ├── e2e-down.sh.tmpl
│   │       │   ├── kind-config.yaml.tmpl
│   │       │   └── makefile-controller.tmpl
│   │       └── helm/
│   │           ├── Chart.yaml.tmpl
│   │           ├── values.yaml.tmpl
│   │           ├── helmignore.tmpl
│   │           ├── helpers.tpl.tmpl
│   │           ├── NOTES.txt.tmpl
│   │           ├── deployment.yaml.tmpl
│   │           ├── service.yaml.tmpl
│   │           ├── serviceaccount.yaml.tmpl
│   │           ├── clusterrole.yaml.tmpl
│   │           ├── clusterrolebinding.yaml.tmpl
│   │           └── makefile-helm.tmpl
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
- **controller-runtime reconciler** (controller): `internal/controller/reconciler.go` with a skeleton `Reconcile()` method, manager setup in `main.go`, controller-runtime dependency in `go.mod`. Compatible with layering kubebuilder on top later for CRD-based controllers.
- **kind e2e scripts** (controller): idempotent `e2e-up.sh` (create cluster, build image, load, deploy), `e2e-down.sh` (delete cluster), `kind-config.yaml` (single control-plane)
- **Helm chart** (helm): `charts/<name>/` with `Chart.yaml`, `values.yaml`, `.helmignore`, and templates for deployment, service, serviceaccount, clusterrole, clusterrolebinding, `_helpers.tpl`, `NOTES.txt`. Follows `~/.claude/helm-standards.md`: standard labels, hardened pod security context (`runAsNonRoot`, `readOnlyRootFilesystem`, `drop ALL`), resource defaults, liveness/readiness probe stubs, RBAC with explicit verbs, `http-metrics` port on service.
- **Helm Makefile targets** (helm): `helm-lint` runs `helm lint`, `helm-template` renders and validates with `kubeconform`

## Testing

### Unit Tests

Test template rendering in isolation. Given `Params`, assert rendered output matches golden files in `testdata/`.

- Template rendering: each template produces expected content for each flag combination
- Makefile merging: base + cli, base + controller, base + cli + controller all produce valid Makefiles

### Integration Tests

One test per flag combination (`--cli`, `--controller`, `--helm`, `--cli --controller --helm`):

1. Run `init` into `t.TempDir()`
2. Assert all expected files exist
3. Assert `go build ./...` succeeds
4. Assert `go vet ./...` passes
5. Assert `make lint-config` passes (validates generated `.golangci.yml`)
6. For `--helm`: assert `helm lint charts/<name>` passes

### Add Tests

1. Run `init --cli` into temp dir
2. Run `add controller` in same dir
3. Assert controller files appear
4. Assert Makefile now has docker targets
5. Run `add cli` again — assert it skips without overwriting, prints message

### No E2E for Heavy Dependencies

No tests that run `kind`, `goreleaser`, `docker build`, or `brew`. Those depend on external tools. The integration tests prove the generated project is valid and buildable.

## CRD Controllers: Out of Scope

CRD-based controllers (custom types, deepcopy, CRD YAML generation) are out of scope for v1. This section documents the research and decision.

### K8s Controller Tooling Landscape

| Tool | What it is | When to use |
|---|---|---|
| **controller-runtime** | Go library (manager, reconciler, client) | Always — the runtime for all Go controllers |
| **controller-gen** | Standalone code generation CLI (CRD YAML, deepcopy, RBAC from markers) | When you have CRDs — no kubebuilder dependency required |
| **kubebuilder** | Project scaffolding tool (kubernetes-sigs) | Full CRD-based operator project from scratch |
| **Operator SDK** | Wraps kubebuilder + adds Ansible/Helm operators + OLM | When you need OLM integration or non-Go operators |

### Key Findings

- **Kubebuilder supports non-CRD controllers** via `kubebuilder create api --resource=false`, but it's buggy — [issue #5097](https://github.com/kubernetes-sigs/kubebuilder/issues/5097) documents broken references to non-existent CRD dirs requiring manual cleanup. Not a first-class path.
- **For non-CRD controllers, the standard approach is controller-runtime directly** — no scaffolding tool. The controller-runtime repo has a canonical `examples/builtins/` example. This is what goscaffold's `--controller` flag provides.
- **controller-gen is standalone** — any project can use it directly in a Makefile without kubebuilder. It generates CRD YAML, RBAC ClusterRoles, DeepCopy methods, webhook config, schemapatch, and apply-configuration from Go marker annotations.
- **The ecosystem has converged**: controller-runtime as the library, kubebuilder as the CRD scaffolding tool, controller-gen as the code generation tool. No serious alternatives in the Go space (KUDO is dead, Metacontroller is polyglot/webhook-based).

### Why Out of Scope

goscaffold fills the gaps where no canonical tooling exists: CLIs, non-CRD controllers, and the surrounding infrastructure (CI, Makefile, linting, Helm charts). For CRD controllers, kubebuilder is the canonical tool — replicating it would be reinventing the wheel. A future `goscaffold augment` command could layer standards onto an existing kubebuilder project, but that's a v2 concern.

## Distribution

GoReleaser + Homebrew tap (`jholm117/homebrew-tap`). Install via:

```
brew install jholm117/tap/goscaffold
```

goscaffold itself is built with the same standards it generates — it eats its own dog food.

## What Happens to go-template

Archive `jholm117/go-template` once goscaffold is working. It's superseded.
