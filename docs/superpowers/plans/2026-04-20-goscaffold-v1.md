# goscaffold Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI that scaffolds new Go projects with composable layers (cli, controller, helm), following Jeff's engineering standards.

**Architecture:** Cobra CLI with embedded `text/template` files via `embed.FS`. Three layers (base, cli, controller, helm) render into a target directory. Makefile is the only file that merges content from multiple layers. Config loaded from `~/.config/goscaffold/config.yaml`.

**Tech Stack:** Go 1.24+, cobra, embed.FS, text/template, gopkg.in/yaml.v3

---

## File Structure

```
goscaffold/
├── cmd/goscaffold/
│   └── main.go                    # cobra root + init/add subcommands
├── internal/
│   ├── config/
│   │   ├── config.go              # Load ~/.config/goscaffold/config.yaml
│   │   └── config_test.go
│   ├── scaffold/
│   │   ├── engine.go              # RenderTemplate, WriteFile, RenderLayer
│   │   ├── engine_test.go
│   │   ├── init.go                # Init() orchestration
│   │   ├── init_test.go
│   │   ├── add.go                 # Add() orchestration
│   │   ├── add_test.go
│   │   ├── embed.go               # embed.FS declaration
│   │   └── templates/
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
│       ├── merge.go                # HasSection, AppendSection
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

---

### Task 1: Project Bootstrap

Bootstrap goscaffold's own project infrastructure — it eats its own dog food.

**Files:**
- Create: `Makefile`
- Create: `.golangci.yml`
- Create: `.gitignore`
- Create: `hack/ci-checks.sh`
- Create: `.githooks/pre-push`
- Create: `.github/workflows/ci.yml`
- Create: `.github/workflows/release.yml`
- Create: `.goreleaser.yaml`
- Create: `AGENTS.md`
- Create: `README.md`
- Create: `go.mod`
- Create: `cmd/goscaffold/main.go`

- [ ] **Step 1: Create go.mod**

```
module github.com/jholm117/goscaffold

go 1.24.2

require (
	github.com/spf13/cobra v1.9.1
	gopkg.in/yaml.v3 v3.0.1
)
```

Run: `go mod tidy`

- [ ] **Step 2: Create minimal main.go**

Create `cmd/goscaffold/main.go`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	root := &cobra.Command{
		Use:     "goscaffold",
		Short:   "Scaffold Go projects with composable layers",
		Version: version,
	}

	root.AddCommand(newInitCmd())
	root.AddCommand(newAddCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <project-name>",
		Short: "Create a new Go project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("init not implemented")
			return nil
		},
	}
	cmd.Flags().String("module", "", "Go module path (default: <module_prefix>/<project-name>)")
	cmd.Flags().Bool("cli", false, "Include CLI distribution layer")
	cmd.Flags().Bool("controller", false, "Include K8s controller layer")
	cmd.Flags().Bool("helm", false, "Include Helm chart layer")
	return cmd
}

func newAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "add <layer>",
		Short:     "Add a layer to an existing project",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"cli", "controller", "helm"},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("add not implemented")
			return nil
		},
	}
}
```

- [ ] **Step 3: Create Makefile**

```makefile
PROJECT_NAME ?= goscaffold
LOCALBIN ?= $(shell pwd)/bin

## Tool Versions
GOLANGCI_LINT_VERSION ?= v2.8.0
GOVULNCHECK_VERSION ?= v1.1.4
GORELEASER_VERSION ?= v2.8.2

.PHONY: build
build: ## Build binary.
	go build -o bin/$(PROJECT_NAME) ./cmd/$(PROJECT_NAME)

.PHONY: install
install: ## Install to $GOPATH/bin.
	go install ./cmd/$(PROJECT_NAME)

.PHONY: test
test: ## Run tests.
	go test ./... -coverprofile cover.out

.PHONY: lint
lint: golangci-lint ## Run linter.
	"$(GOLANGCI_LINT)" run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run linter with auto-fixes.
	"$(GOLANGCI_LINT)" run --fix

.PHONY: lint-config
lint-config: golangci-lint ## Verify linter config.
	"$(GOLANGCI_LINT)" config verify

.PHONY: fmt
fmt: ## Run go fmt.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet.
	go vet ./...

.PHONY: govulncheck
govulncheck: $(GOVULNCHECK) ## Run vulnerability check.
	"$(GOVULNCHECK)" ./...

.PHONY: release-snapshot
release-snapshot: goreleaser ## Build release snapshot (no publish).
	"$(GORELEASER)" release --snapshot --clean

.PHONY: setup-hooks
setup-hooks: ## Install git hooks.
	@git config core.hooksPath .githooks
	@echo "Installed hooks via core.hooksPath -> .githooks/"

.PHONY: clean
clean: ## Remove build artifacts.
	rm -rf bin/ dist/ cover.out

.PHONY: help
help: ## Show this help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

## Tools

GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
GOVULNCHECK = $(LOCALBIN)/govulncheck
GORELEASER = $(LOCALBIN)/goreleaser

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

.PHONY: govulncheck
govulncheck: $(GOVULNCHECK)
$(GOVULNCHECK): $(LOCALBIN)
	$(call go-install-tool,$(GOVULNCHECK),golang.org/x/vuln/cmd/govulncheck,$(GOVULNCHECK_VERSION))

.PHONY: goreleaser
goreleaser: $(GORELEASER)
$(GORELEASER): $(LOCALBIN)
	$(call go-install-tool,$(GORELEASER),github.com/goreleaser/goreleaser/v2,$(GORELEASER_VERSION))

$(LOCALBIN):
	mkdir -p "$(LOCALBIN)"

define go-install-tool
@[ -f "$(1)-$(3)" ] && [ "$$(readlink -- "$(1)" 2>/dev/null)" = "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f "$(1)" ;\
GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) GOBIN="$(LOCALBIN)" go install $${package} ;\
mv "$(LOCALBIN)/$$(basename "$(1)")" "$(1)-$(3)" ;\
} ;\
ln -sf "$$(realpath "$(1)-$(3)")" "$(1)"
endef
```

- [ ] **Step 4: Create remaining infrastructure files**

Create `.golangci.yml` (same as go-standards baseline config):
```yaml
version: "2"
run:
  allow-parallel-runners: true
linters:
  default: none
  enable:
    - copyloopvar
    - dupl
    - errcheck
    - goconst
    - gocyclo
    - govet
    - ineffassign
    - lll
    - modernize
    - misspell
    - nakedret
    - prealloc
    - revive
    - staticcheck
    - unconvert
    - unparam
    - unused
  settings:
    revive:
      rules:
        - name: comment-spacings
        - name: import-shadowing
  exclusions:
    generated: lax
    paths:
      - third_party$
      - builtin$
      - examples$
formatters:
  enable:
    - gofmt
    - goimports
  exclusions:
    generated: lax
    paths:
      - third_party$
      - builtin$
      - examples$
```

Create `.gitignore`:
```
.worktrees/
bin/
dist/
testbin/
.idea/
*.swp
*.swo
*~
cover.out
```

Create `hack/ci-checks.sh` (executable):
```bash
#!/usr/bin/env bash
set -euo pipefail

parallel=false
for arg in "$@"; do
    case "$arg" in
        --parallel) parallel=true ;;
    esac
done

failed=0

run_lint() {
    echo "--- make lint-config"
    make lint-config
    echo "--- make lint"
    make lint
}

run_tidy_check() {
    echo "--- go mod tidy (checking for drift)"
    cp go.mod go.mod.bak
    cp go.sum go.sum.bak
    go mod tidy
    if ! diff -q go.mod go.mod.bak >/dev/null 2>&1 || ! diff -q go.sum go.sum.bak >/dev/null 2>&1; then
        echo "ERROR: go.mod or go.sum changed after 'go mod tidy'. Commit the changes."
        mv go.mod.bak go.mod
        mv go.sum.bak go.sum
        return 1
    fi
    rm -f go.mod.bak go.sum.bak
}

run_test() {
    echo "--- make test"
    make test
}

run_govulncheck() {
    echo "--- make govulncheck"
    make govulncheck
}

run_tidy_check || { echo "FAIL: go mod tidy"; exit 1; }

if [ "$parallel" = true ]; then
    echo "==> running checks in parallel"
    run_lint &
    lint_pid=$!
    run_test &
    test_pid=$!
    run_govulncheck &
    vuln_pid=$!

    if ! wait "$lint_pid"; then echo "FAIL: lint"; failed=1; fi
    if ! wait "$test_pid"; then echo "FAIL: test"; failed=1; fi
    if ! wait "$vuln_pid"; then echo "FAIL: govulncheck"; failed=1; fi
else
    echo "==> running checks sequentially"
    run_lint || { echo "FAIL: lint"; failed=1; }
    run_test || { echo "FAIL: test"; failed=1; }
    run_govulncheck || { echo "FAIL: govulncheck"; failed=1; }
fi

if [ "$failed" -ne 0 ]; then exit 1; fi
echo "==> all checks passed"
```

Create `.githooks/pre-push` (executable):
```bash
#!/usr/bin/env bash
set -euo pipefail
exec hack/ci-checks.sh
```

Create `.github/workflows/ci.yml`:
```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:

concurrency:
  group: ci-${{ github.ref }}
  cancel-in-progress: true

jobs:
  ci:
    name: Lint & Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - uses: actions/cache@v4
        with:
          path: bin/
          key: tools-${{ runner.os }}-${{ hashFiles('Makefile') }}

      - name: Run checks
        run: hack/ci-checks.sh --parallel
```

Create `.github/workflows/release.yml`:
```yaml
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    name: GoReleaser
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

Create `.goreleaser.yaml`:
```yaml
version: 2

builds:
  - main: ./cmd/goscaffold
    binary: goscaffold
    env:
      - CGO_ENABLED=0
    ldflags:
      - -s -w -X main.version={{.Version}}
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64

archives:
  - formats: ["tar.gz"]
    name_template: >-
      {{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^ci:"

brews:
  - repository:
      owner: jholm117
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
    homepage: "https://github.com/jholm117/goscaffold"
    description: "Scaffold Go projects with composable layers"
```

Create `AGENTS.md`:
```markdown
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
```

Create `README.md`:
```markdown
# goscaffold

Scaffold Go projects with composable layers (CLI, K8s controller, Helm chart) following consistent engineering standards.

## Install

```bash
brew install jholm117/tap/goscaffold
```

## Usage

```bash
# Create a CLI project
goscaffold init myapp --cli

# Create a K8s controller with Helm chart
goscaffold init mycontroller --controller --helm

# Create a project with everything
goscaffold init myproject --cli --controller --helm

# Add a layer to an existing project
cd myapp
goscaffold add controller
```

## Configuration

Create `~/.config/goscaffold/config.yaml`:

```yaml
module_prefix: github.com/youruser
homebrew_tap_token: "<GitHub PAT with contents:write on homebrew-tap repo>"
```

## Development

| Command | Description |
|---|---|
| `make build` | Build binary |
| `make test` | Run tests |
| `make lint` | Run golangci-lint |
| `make setup-hooks` | Install pre-push hook |
| `make help` | Show all targets |
```

- [ ] **Step 5: Verify build and lint**

Run: `go mod tidy && make build && make lint-config`
Expected: build succeeds, lint config valid

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "Bootstrap project infrastructure"
```

---

### Task 2: Config Package

Load user configuration from `~/.config/goscaffold/config.yaml`.

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/config/config_test.go`:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FromFile(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(cfgFile, []byte("module_prefix: github.com/testuser\nhomebrew_tap_token: tok123\n"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(cfgFile)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}
	if cfg.ModulePrefix != "github.com/testuser" {
		t.Errorf("ModulePrefix = %q, want %q", cfg.ModulePrefix, "github.com/testuser")
	}
	if cfg.HomebrewTapToken != "tok123" {
		t.Errorf("HomebrewTapToken = %q, want %q", cfg.HomebrewTapToken, "tok123")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	cfg, err := LoadFrom("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("LoadFrom should not error on missing file: %v", err)
	}
	if cfg.ModulePrefix != "" {
		t.Errorf("ModulePrefix = %q, want empty", cfg.ModulePrefix)
	}
}

func TestLoad_DefaultPath(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	_ = cfg
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -v`
Expected: FAIL — `LoadFrom` not defined

- [ ] **Step 3: Write implementation**

Create `internal/config/config.go`:

```go
package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ModulePrefix     string `yaml:"module_prefix"`
	HomebrewTapToken string `yaml:"homebrew_tap_token"`
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return &Config{}, nil
	}
	return LoadFrom(filepath.Join(home, ".config", "goscaffold", "config.yaml"))
}

func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go mod tidy && go test ./internal/config/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "Add config package for ~/.config/goscaffold/config.yaml"
```

---

### Task 3: Core Scaffold Engine

The engine renders Go text/templates with `Params` and writes files to disk.

**Files:**
- Create: `internal/scaffold/engine.go`
- Create: `internal/scaffold/engine_test.go`
- Create: `internal/scaffold/embed.go`
- Create: `internal/scaffold/templates/base/.gitkeep` (placeholder so embed.FS works)

- [ ] **Step 1: Write the failing test**

Create `internal/scaffold/engine_test.go`:

```go
package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	tmpl := "Hello {{.ProjectName}} at {{.Module}}"
	params := Params{ProjectName: "myapp", Module: "github.com/test/myapp"}

	result, err := RenderTemplate("test", tmpl, params)
	if err != nil {
		t.Fatalf("RenderTemplate: %v", err)
	}
	want := "Hello myapp at github.com/test/myapp"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestWriteFile_CreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "file.txt")

	written, err := WriteFile(path, "content", 0o644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if !written {
		t.Error("WriteFile returned false, want true")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "content" {
		t.Errorf("file content = %q, want %q", string(data), "content")
	}
}

func TestWriteFile_SkipsExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	os.WriteFile(path, []byte("original"), 0o644)

	written, err := WriteFile(path, "new content", 0o644)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if written {
		t.Error("WriteFile returned true for existing file, want false")
	}

	data, _ := os.ReadFile(path)
	if string(data) != "original" {
		t.Errorf("file was overwritten: got %q", string(data))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/scaffold/ -v`
Expected: FAIL — types not defined

- [ ] **Step 3: Write implementation**

Create `internal/scaffold/embed.go`:

```go
package scaffold

import "embed"

//go:embed templates/*
var templates embed.FS
```

Create a placeholder `internal/scaffold/templates/base/.gitkeep` (empty file, so `embed.FS` has something to embed).

Create `internal/scaffold/engine.go`:

```go
package scaffold

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"
)

type Params struct {
	ProjectName string
	Module      string
	CLI         bool
	Controller  bool
	Helm        bool
}

func RenderTemplate(name, text string, params Params) (string, error) {
	tmpl, err := template.New(name).Parse(text)
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("execute template %s: %w", name, err)
	}
	return buf.String(), nil
}

func WriteFile(path, content string, perm fs.FileMode) (written bool, err error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		return false, err
	}
	return true, nil
}

type FileSpec struct {
	TemplatePath string
	OutputPath   string
	Perm         fs.FileMode
}

func RenderLayer(layer string, targetDir string, params Params, specs []FileSpec) error {
	for _, spec := range specs {
		tmplData, err := templates.ReadFile(spec.TemplatePath)
		if err != nil {
			return fmt.Errorf("read template %s: %w", spec.TemplatePath, err)
		}
		rendered, err := RenderTemplate(spec.TemplatePath, string(tmplData), params)
		if err != nil {
			return err
		}
		outPath := filepath.Join(targetDir, spec.OutputPath)
		written, err := WriteFile(outPath, rendered, spec.Perm)
		if err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
		if written {
			fmt.Printf("  create %s\n", spec.OutputPath)
		} else {
			fmt.Printf("  skip   %s (exists)\n", spec.OutputPath)
		}
	}
	return nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/scaffold/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/scaffold/
git commit -m "Add core scaffold engine: render, write, skip-existing"
```

---

### Task 4: Makefile Merge Package

The Makefile is the one file where multiple layers merge into a single output. The `add` command needs to detect and append layer sections.

**Files:**
- Create: `internal/makefile/merge.go`
- Create: `internal/makefile/merge_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/makefile/merge_test.go`:

```go
package makefile

import (
	"strings"
	"testing"
)

func TestHasSection(t *testing.T) {
	content := "## Base\nfoo:\n\techo foo\n\n## CLI Targets\nbar:\n\techo bar\n"

	if !HasSection(content, "## CLI Targets") {
		t.Error("HasSection should find '## CLI Targets'")
	}
	if HasSection(content, "## Controller Targets") {
		t.Error("HasSection should not find '## Controller Targets'")
	}
}

func TestAppendSection(t *testing.T) {
	base := "## Base\nfoo:\n\techo foo\n"
	section := "\n## CLI Targets\nbar:\n\techo bar\n"

	result := AppendSection(base, section)

	if !strings.Contains(result, "## CLI Targets") {
		t.Error("result should contain CLI Targets section")
	}
	if !strings.Contains(result, "## Base") {
		t.Error("result should still contain Base section")
	}
}

func TestAppendSection_BeforeToolsBlock(t *testing.T) {
	base := "## Base\nfoo:\n\techo foo\n\n## Tools\n\nGOLANGCI_LINT = bin/golangci-lint\n"
	section := "\n## CLI Targets\nbar:\n\techo bar\n"

	result := AppendSection(base, section)

	cliIdx := strings.Index(result, "## CLI Targets")
	toolsIdx := strings.Index(result, "## Tools")
	if cliIdx > toolsIdx {
		t.Error("CLI Targets should appear before ## Tools block")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/makefile/ -v`
Expected: FAIL — functions not defined

- [ ] **Step 3: Write implementation**

Create `internal/makefile/merge.go`:

```go
package makefile

import "strings"

func HasSection(content, sentinel string) bool {
	return strings.Contains(content, sentinel)
}

func AppendSection(content, section string) string {
	toolsMarker := "## Tools"
	if idx := strings.Index(content, toolsMarker); idx != -1 {
		return content[:idx] + section + "\n" + content[idx:]
	}
	return content + section
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/makefile/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/makefile/
git commit -m "Add Makefile merge package: HasSection, AppendSection"
```

---

### Task 5: Base Layer Templates

Create all base layer template files. These are rendered for every project regardless of flags.

**Files:**
- Create: `internal/scaffold/templates/base/Makefile.tmpl`
- Create: `internal/scaffold/templates/base/golangci.yml.tmpl`
- Create: `internal/scaffold/templates/base/gitignore.tmpl`
- Create: `internal/scaffold/templates/base/ci.yml.tmpl`
- Create: `internal/scaffold/templates/base/ci-checks.sh.tmpl`
- Create: `internal/scaffold/templates/base/pre-push.tmpl`
- Create: `internal/scaffold/templates/base/main.go.tmpl`
- Create: `internal/scaffold/templates/base/go.mod.tmpl`
- Create: `internal/scaffold/templates/base/agents.md.tmpl`
- Create: `internal/scaffold/templates/base/readme.md.tmpl`
- Create: `internal/scaffold/base.go`
- Create: `internal/scaffold/base_test.go`

- [ ] **Step 1: Create template files**

Delete `internal/scaffold/templates/base/.gitkeep`.

Create `internal/scaffold/templates/base/Makefile.tmpl`:
```
PROJECT_NAME ?= {{.ProjectName}}
LOCALBIN ?= $(shell pwd)/bin

## Tool Versions
GOLANGCI_LINT_VERSION ?= v2.8.0
GOVULNCHECK_VERSION ?= v1.1.4

.PHONY: build
build: ## Build binary.
	go build -o bin/$(PROJECT_NAME) ./cmd/$(PROJECT_NAME)

.PHONY: install
install: ## Install to $GOPATH/bin.
	go install ./cmd/$(PROJECT_NAME)

.PHONY: test
test: ## Run tests.
	go test ./... -coverprofile cover.out

.PHONY: lint
lint: golangci-lint ## Run linter.
	"$(GOLANGCI_LINT)" run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run linter with auto-fixes.
	"$(GOLANGCI_LINT)" run --fix

.PHONY: lint-config
lint-config: golangci-lint ## Verify linter config.
	"$(GOLANGCI_LINT)" config verify

.PHONY: fmt
fmt: ## Run go fmt.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet.
	go vet ./...

.PHONY: govulncheck
govulncheck: $(GOVULNCHECK) ## Run vulnerability check.
	"$(GOVULNCHECK)" ./...

.PHONY: setup-hooks
setup-hooks: ## Install git hooks.
	@git config core.hooksPath .githooks
	@echo "Installed hooks via core.hooksPath -> .githooks/"

.PHONY: clean
clean: ## Remove build artifacts.
	rm -rf bin/ dist/ cover.out

.PHONY: help
help: ## Show this help.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

## Tools

GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
GOVULNCHECK = $(LOCALBIN)/govulncheck

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

.PHONY: govulncheck
govulncheck: $(GOVULNCHECK)
$(GOVULNCHECK): $(LOCALBIN)
	$(call go-install-tool,$(GOVULNCHECK),golang.org/x/vuln/cmd/govulncheck,$(GOVULNCHECK_VERSION))

$(LOCALBIN):
	mkdir -p "$(LOCALBIN)"

define go-install-tool
@[ -f "$(1)-$(3)" ] && [ "$$(readlink -- "$(1)" 2>/dev/null)" = "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f "$(1)" ;\
GOOS=$(shell go env GOHOSTOS) GOARCH=$(shell go env GOHOSTARCH) GOBIN="$(LOCALBIN)" go install $${package} ;\
mv "$(LOCALBIN)/$$(basename "$(1)")" "$(1)-$(3)" ;\
} ;\
ln -sf "$$(realpath "$(1)-$(3)")" "$(1)"
endef
```

Create `internal/scaffold/templates/base/main.go.tmpl`:
```
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	root := &cobra.Command{
		Use:     "{{.ProjectName}}",
		Short:   "{{.ProjectName}} does things",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("hello from {{.ProjectName}}")
			return nil
		},
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
```

Create `internal/scaffold/templates/base/go.mod.tmpl`:
```
module {{.Module}}

go 1.24.2

require github.com/spf13/cobra v1.9.1

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
)
```

Create `internal/scaffold/templates/base/golangci.yml.tmpl`:
```
version: "2"
run:
  allow-parallel-runners: true
linters:
  default: none
  enable:
    - copyloopvar
    - dupl
    - errcheck
    - goconst
    - gocyclo
    - govet
    - ineffassign
    - lll
    - modernize
    - misspell
    - nakedret
    - prealloc
    - revive
    - staticcheck
    - unconvert
    - unparam
    - unused
  settings:
    revive:
      rules:
        - name: comment-spacings
        - name: import-shadowing
  exclusions:
    generated: lax
    paths:
      - third_party$
      - builtin$
      - examples$
formatters:
  enable:
    - gofmt
    - goimports
  exclusions:
    generated: lax
    paths:
      - third_party$
      - builtin$
      - examples$
```

Create `internal/scaffold/templates/base/gitignore.tmpl`:
```
.worktrees/
bin/
dist/
testbin/
.idea/
*.swp
*.swo
*~
cover.out
```

Create `internal/scaffold/templates/base/ci.yml.tmpl`:
```
name: CI

on:
  push:
    branches: [main]
  pull_request:

concurrency:
  group: ci-${{"{{"}} github.ref {{"}}"}}
  cancel-in-progress: true

jobs:
  ci:
    name: Lint & Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - uses: actions/cache@v4
        with:
          path: bin/
          key: tools-${{"{{"}} runner.os {{"}}"}}-${{"{{"}} hashFiles('Makefile') {{"}}"}}

      - name: Run checks
        run: hack/ci-checks.sh --parallel
```

Create `internal/scaffold/templates/base/ci-checks.sh.tmpl`:
```
#!/usr/bin/env bash
set -euo pipefail

parallel=false
for arg in "$@"; do
    case "$arg" in
        --parallel) parallel=true ;;
    esac
done

failed=0

run_lint() {
    echo "--- make lint-config"
    make lint-config
    echo "--- make lint"
    make lint
}

run_tidy_check() {
    echo "--- go mod tidy (checking for drift)"
    cp go.mod go.mod.bak
    cp go.sum go.sum.bak
    go mod tidy
    if ! diff -q go.mod go.mod.bak >/dev/null 2>&1 || ! diff -q go.sum go.sum.bak >/dev/null 2>&1; then
        echo "ERROR: go.mod or go.sum changed after 'go mod tidy'. Commit the changes."
        mv go.mod.bak go.mod
        mv go.sum.bak go.sum
        return 1
    fi
    rm -f go.mod.bak go.sum.bak
}

run_test() {
    echo "--- make test"
    make test
}

run_govulncheck() {
    echo "--- make govulncheck"
    make govulncheck
}

run_tidy_check || { echo "FAIL: go mod tidy"; exit 1; }

if [ "$parallel" = true ]; then
    echo "==> running checks in parallel"
    run_lint &
    lint_pid=$!
    run_test &
    test_pid=$!
    run_govulncheck &
    vuln_pid=$!

    if ! wait "$lint_pid"; then echo "FAIL: lint"; failed=1; fi
    if ! wait "$test_pid"; then echo "FAIL: test"; failed=1; fi
    if ! wait "$vuln_pid"; then echo "FAIL: govulncheck"; failed=1; fi
else
    echo "==> running checks sequentially"
    run_lint || { echo "FAIL: lint"; failed=1; }
    run_test || { echo "FAIL: test"; failed=1; }
    run_govulncheck || { echo "FAIL: govulncheck"; failed=1; }
fi

if [ "$failed" -ne 0 ]; then exit 1; fi
echo "==> all checks passed"
```

Create `internal/scaffold/templates/base/pre-push.tmpl`:
```
#!/usr/bin/env bash
set -euo pipefail
exec hack/ci-checks.sh
```

Create `internal/scaffold/templates/base/agents.md.tmpl`:
```
# {{.ProjectName}}

## Build & Test

- `make build` — build binary to `bin/{{.ProjectName}}`
- `make test` — run all tests
- `make lint` — run golangci-lint
- `make setup-hooks` — install pre-push hook

## Project Layout

- `cmd/{{.ProjectName}}/` — CLI entry point (cobra)
- `internal/` — private packages
- `pkg/` — public packages
- `hack/ci-checks.sh` — shared CI checks (hook + CI parity)
```

Create `internal/scaffold/templates/base/readme.md.tmpl`:
```
# {{.ProjectName}}

Short description of what this project does.

## Getting Started

` + "```" + `bash
git clone {{.Module}}
cd {{.ProjectName}}
make setup-hooks
make build
./bin/{{.ProjectName}}
` + "```" + `

## Development

| Command | Description |
|---|---|
| ` + "`" + `make build` + "`" + ` | Build binary |
| ` + "`" + `make test` + "`" + ` | Run tests with coverage |
| ` + "`" + `make lint` + "`" + ` | Run golangci-lint |
| ` + "`" + `make clean` + "`" + ` | Remove build artifacts |
| ` + "`" + `make help` + "`" + ` | Show all targets |
```

- [ ] **Step 2: Create base.go with file specs**

Create `internal/scaffold/base.go`:

```go
package scaffold

import "io/fs"

func BaseSpecs(params Params) []FileSpec {
	return []FileSpec{
		{TemplatePath: "templates/base/go.mod.tmpl", OutputPath: "go.mod", Perm: 0o644},
		{TemplatePath: "templates/base/main.go.tmpl", OutputPath: "cmd/" + params.ProjectName + "/main.go", Perm: 0o644},
		{TemplatePath: "templates/base/Makefile.tmpl", OutputPath: "Makefile", Perm: 0o644},
		{TemplatePath: "templates/base/golangci.yml.tmpl", OutputPath: ".golangci.yml", Perm: 0o644},
		{TemplatePath: "templates/base/gitignore.tmpl", OutputPath: ".gitignore", Perm: 0o644},
		{TemplatePath: "templates/base/ci.yml.tmpl", OutputPath: ".github/workflows/ci.yml", Perm: 0o644},
		{TemplatePath: "templates/base/ci-checks.sh.tmpl", OutputPath: "hack/ci-checks.sh", Perm: 0o755},
		{TemplatePath: "templates/base/pre-push.tmpl", OutputPath: ".githooks/pre-push", Perm: 0o755},
		{TemplatePath: "templates/base/agents.md.tmpl", OutputPath: "AGENTS.md", Perm: 0o644},
		{TemplatePath: "templates/base/readme.md.tmpl", OutputPath: "README.md", Perm: 0o644},
	}
}

func PlaceholderDirs() []string {
	return []string{"internal", "pkg"}
}

func WritePlaceholders(targetDir string) error {
	for _, dir := range PlaceholderDirs() {
		_, err := WriteFile(targetDir+"/"+dir+"/.gitkeep", "", fs.FileMode(0o644))
		if err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 3: Write the test**

Create `internal/scaffold/base_test.go`:

```go
package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBaseSpecs_RenderAll(t *testing.T) {
	params := Params{
		ProjectName: "testapp",
		Module:      "github.com/test/testapp",
	}

	dir := t.TempDir()
	if err := RenderLayer("base", dir, params, BaseSpecs(params)); err != nil {
		t.Fatalf("RenderLayer base: %v", err)
	}

	wantFiles := []string{
		"go.mod",
		"cmd/testapp/main.go",
		"Makefile",
		".golangci.yml",
		".gitignore",
		".github/workflows/ci.yml",
		"hack/ci-checks.sh",
		".githooks/pre-push",
		"AGENTS.md",
		"README.md",
	}

	for _, f := range wantFiles {
		path := filepath.Join(dir, f)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("missing file %s: %v", f, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("file %s is empty", f)
		}
	}

	makefile, _ := os.ReadFile(filepath.Join(dir, "Makefile"))
	if !strings.Contains(string(makefile), "PROJECT_NAME ?= testapp") {
		t.Error("Makefile should contain project name")
	}

	mainGo, _ := os.ReadFile(filepath.Join(dir, "cmd/testapp/main.go"))
	if !strings.Contains(string(mainGo), `Use:     "testapp"`) {
		t.Error("main.go should contain project name in Use field")
	}

	goMod, _ := os.ReadFile(filepath.Join(dir, "go.mod"))
	if !strings.Contains(string(goMod), "module github.com/test/testapp") {
		t.Error("go.mod should contain module path")
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/scaffold/ -v -run TestBaseSpecs`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/scaffold/
git commit -m "Add base layer templates and specs"
```

---

### Task 6: CLI Layer Templates

GoReleaser config, release workflow, and Makefile CLI section.

**Files:**
- Create: `internal/scaffold/templates/cli/goreleaser.yaml.tmpl`
- Create: `internal/scaffold/templates/cli/release.yml.tmpl`
- Create: `internal/scaffold/templates/cli/makefile-cli.tmpl`
- Create: `internal/scaffold/cli.go`
- Create: `internal/scaffold/cli_test.go`

- [ ] **Step 1: Create template files**

Create `internal/scaffold/templates/cli/goreleaser.yaml.tmpl`:
```
version: 2

builds:
  - main: ./cmd/{{.ProjectName}}
    binary: {{.ProjectName}}
    env:
      - CGO_ENABLED=0
    ldflags:
      - -s -w -X main.version={{"{{"}}{{".Version"}}{{"}}"}}
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64

archives:
  - formats: ["tar.gz"]
    name_template: >-
      {{"{{"}} .ProjectName {{"}}"}}_{{"{{"}}.Version{{"}}"}}_{{"{{"}}.Os{{"}}"}}_{{"{{"}}.Arch{{"}}"}}

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
      - "^ci:"

brews:
  - repository:
      owner: jholm117
      name: homebrew-tap
      token: "{{"{{"}} .Env.HOMEBREW_TAP_GITHUB_TOKEN {{"}}"}}"
    homepage: "https://{{.Module}}"
    description: "{{.ProjectName}}"
```

Create `internal/scaffold/templates/cli/release.yml.tmpl`:
```
name: Release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    name: GoReleaser
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{"{{"}} secrets.GITHUB_TOKEN {{"}}"}}
          HOMEBREW_TAP_GITHUB_TOKEN: ${{"{{"}} secrets.HOMEBREW_TAP_GITHUB_TOKEN {{"}}"}}
```

Create `internal/scaffold/templates/cli/makefile-cli.tmpl`:
```
## CLI Targets

GORELEASER_VERSION ?= v2.8.2

.PHONY: release-snapshot
release-snapshot: goreleaser ## Build release snapshot (no publish).
	"$(GORELEASER)" release --snapshot --clean

GORELEASER = $(LOCALBIN)/goreleaser

.PHONY: goreleaser
goreleaser: $(GORELEASER)
$(GORELEASER): $(LOCALBIN)
	$(call go-install-tool,$(GORELEASER),github.com/goreleaser/goreleaser/v2,$(GORELEASER_VERSION))
```

- [ ] **Step 2: Create cli.go with file specs and test**

Create `internal/scaffold/cli.go`:

```go
package scaffold

func CLISpecs() []FileSpec {
	return []FileSpec{
		{TemplatePath: "templates/cli/goreleaser.yaml.tmpl", OutputPath: ".goreleaser.yaml", Perm: 0o644},
		{TemplatePath: "templates/cli/release.yml.tmpl", OutputPath: ".github/workflows/release.yml", Perm: 0o644},
	}
}

func CLIMakefileTemplate() string {
	return "templates/cli/makefile-cli.tmpl"
}
```

Create `internal/scaffold/cli_test.go`:

```go
package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLISpecs_RenderAll(t *testing.T) {
	params := Params{
		ProjectName: "myapp",
		Module:      "github.com/test/myapp",
		CLI:         true,
	}

	dir := t.TempDir()
	if err := RenderLayer("cli", dir, params, CLISpecs()); err != nil {
		t.Fatalf("RenderLayer cli: %v", err)
	}

	goreleaser, err := os.ReadFile(filepath.Join(dir, ".goreleaser.yaml"))
	if err != nil {
		t.Fatalf("missing .goreleaser.yaml: %v", err)
	}
	if !strings.Contains(string(goreleaser), "main: ./cmd/myapp") {
		t.Error(".goreleaser.yaml should reference cmd/myapp")
	}
	if !strings.Contains(string(goreleaser), "homebrew-tap") {
		t.Error(".goreleaser.yaml should contain homebrew-tap config")
	}

	release, err := os.ReadFile(filepath.Join(dir, ".github/workflows/release.yml"))
	if err != nil {
		t.Fatalf("missing release.yml: %v", err)
	}
	if !strings.Contains(string(release), "HOMEBREW_TAP_GITHUB_TOKEN") {
		t.Error("release.yml should reference HOMEBREW_TAP_GITHUB_TOKEN")
	}
}
```

- [ ] **Step 3: Run test to verify it passes**

Run: `go test ./internal/scaffold/ -v -run TestCLI`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/scaffold/
git commit -m "Add CLI layer templates: GoReleaser, release workflow, Makefile section"
```

---

### Task 7: Controller Layer Templates

Dockerfile, reconciler skeleton, e2e scripts, and Makefile controller section.

**Files:**
- Create: `internal/scaffold/templates/controller/Dockerfile.tmpl`
- Create: `internal/scaffold/templates/controller/dockerignore.tmpl`
- Create: `internal/scaffold/templates/controller/reconciler.go.tmpl`
- Create: `internal/scaffold/templates/controller/e2e-up.sh.tmpl`
- Create: `internal/scaffold/templates/controller/e2e-down.sh.tmpl`
- Create: `internal/scaffold/templates/controller/kind-config.yaml.tmpl`
- Create: `internal/scaffold/templates/controller/makefile-controller.tmpl`
- Create: `internal/scaffold/controller.go`
- Create: `internal/scaffold/controller_test.go`

- [ ] **Step 1: Create template files**

Create `internal/scaffold/templates/controller/Dockerfile.tmpl`:
```
ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags "-s -w" -o /out/{{.ProjectName}} ./cmd/{{.ProjectName}}

FROM gcr.io/distroless/static-debian12:nonroot AS runtime

WORKDIR /
COPY --from=build /out/{{.ProjectName}} /{{.ProjectName}}

USER 65532:65532
ENTRYPOINT ["/{{.ProjectName}}"]
```

Create `internal/scaffold/templates/controller/dockerignore.tmpl`:
```
.git
.github
.githooks
.worktrees
.vscode
.idea
.DS_Store
bin
dist
charts
docs
test
hack
Makefile
README.md
*.md
```

Create `internal/scaffold/templates/controller/reconciler.go.tmpl`:
```
package controller

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Reconciler struct {
	client.Client
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("reconciling", "name", req.Name, "namespace", req.Namespace)

	return ctrl.Result{}, fmt.Errorf("not implemented")
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("{{.ProjectName}}").
		Complete(r)
}
```

Create `internal/scaffold/templates/controller/e2e-up.sh.tmpl`:
```
#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-{{.ProjectName}}-e2e}"
IMAGE_REPO="${IMAGE_REPO:-{{.ProjectName}}}"
IMAGE_TAG="${IMAGE_TAG:-e2e}"
KIND_CONFIG="${KIND_CONFIG:-$(dirname "$0")/kind-config.yaml}"

repo_root="$(cd "$(dirname "$0")/.." && pwd)"
cd "${repo_root}"

require() {
    if ! command -v "$1" >/dev/null 2>&1; then
        echo "error: required command $1 not found on PATH" >&2
        exit 1
    fi
}

require kind
require docker
require kubectl

if ! kind get clusters | grep -qx "${CLUSTER_NAME}"; then
    echo "==> creating kind cluster ${CLUSTER_NAME}"
    kind create cluster --name "${CLUSTER_NAME}" --config "${KIND_CONFIG}" --wait 120s
else
    echo "==> kind cluster ${CLUSTER_NAME} already exists"
fi

kubectl config use-context "kind-${CLUSTER_NAME}" >/dev/null

echo "==> building image ${IMAGE_REPO}:${IMAGE_TAG}"
docker build -t "${IMAGE_REPO}:${IMAGE_TAG}" .

echo "==> loading image into kind"
kind load docker-image "${IMAGE_REPO}:${IMAGE_TAG}" --name "${CLUSTER_NAME}"

echo "==> ready"
```

Create `internal/scaffold/templates/controller/e2e-down.sh.tmpl`:
```
#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${CLUSTER_NAME:-{{.ProjectName}}-e2e}"

if kind get clusters | grep -qx "${CLUSTER_NAME}"; then
    echo "==> deleting kind cluster ${CLUSTER_NAME}"
    kind delete cluster --name "${CLUSTER_NAME}"
else
    echo "==> no kind cluster named ${CLUSTER_NAME}; nothing to do"
fi
```

Create `internal/scaffold/templates/controller/kind-config.yaml.tmpl`:
```
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: {{.ProjectName}}-e2e
nodes:
  - role: control-plane
```

Create `internal/scaffold/templates/controller/makefile-controller.tmpl`:
```
## Controller Targets

IMG ?= {{.ProjectName}}:latest

.PHONY: docker-build
docker-build: ## Build Docker image.
	docker build -t $(IMG) .

.PHONY: docker-push
docker-push: ## Push Docker image.
	docker push $(IMG)

.PHONY: e2e
e2e: e2e-up ## Run E2E tests against kind cluster.
	go test ./test/e2e/ -v -tags e2e -timeout 10m

.PHONY: e2e-up
e2e-up: ## Bootstrap kind cluster for E2E tests.
	hack/e2e-up.sh

.PHONY: e2e-down
e2e-down: ## Delete kind E2E cluster.
	hack/e2e-down.sh
```

- [ ] **Step 2: Create controller.go with file specs and test**

Create `internal/scaffold/controller.go`:

```go
package scaffold

func ControllerSpecs(params Params) []FileSpec {
	return []FileSpec{
		{TemplatePath: "templates/controller/Dockerfile.tmpl", OutputPath: "Dockerfile", Perm: 0o644},
		{TemplatePath: "templates/controller/dockerignore.tmpl", OutputPath: ".dockerignore", Perm: 0o644},
		{TemplatePath: "templates/controller/reconciler.go.tmpl", OutputPath: "internal/controller/reconciler.go", Perm: 0o644},
		{TemplatePath: "templates/controller/e2e-up.sh.tmpl", OutputPath: "hack/e2e-up.sh", Perm: 0o755},
		{TemplatePath: "templates/controller/e2e-down.sh.tmpl", OutputPath: "hack/e2e-down.sh", Perm: 0o755},
		{TemplatePath: "templates/controller/kind-config.yaml.tmpl", OutputPath: "hack/kind-config.yaml", Perm: 0o644},
	}
}

func ControllerMakefileTemplate() string {
	return "templates/controller/makefile-controller.tmpl"
}
```

Create `internal/scaffold/controller_test.go`:

```go
package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestControllerSpecs_RenderAll(t *testing.T) {
	params := Params{
		ProjectName: "myctrl",
		Module:      "github.com/test/myctrl",
		Controller:  true,
	}

	dir := t.TempDir()
	if err := RenderLayer("controller", dir, params, ControllerSpecs(params)); err != nil {
		t.Fatalf("RenderLayer controller: %v", err)
	}

	wantFiles := []string{
		"Dockerfile",
		".dockerignore",
		"internal/controller/reconciler.go",
		"hack/e2e-up.sh",
		"hack/e2e-down.sh",
		"hack/kind-config.yaml",
	}
	for _, f := range wantFiles {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("missing file %s: %v", f, err)
		}
	}

	dockerfile, _ := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if !strings.Contains(string(dockerfile), "go build -trimpath") {
		t.Error("Dockerfile should use -trimpath")
	}
	if !strings.Contains(string(dockerfile), "/myctrl") {
		t.Error("Dockerfile should reference project name")
	}

	reconciler, _ := os.ReadFile(filepath.Join(dir, "internal/controller/reconciler.go"))
	if !strings.Contains(string(reconciler), `Named("myctrl")`) {
		t.Error("reconciler should use project name in controller name")
	}
}
```

- [ ] **Step 3: Run test to verify it passes**

Run: `go test ./internal/scaffold/ -v -run TestController`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/scaffold/
git commit -m "Add controller layer templates: Dockerfile, reconciler, e2e scripts"
```

---

### Task 8: Helm Layer Templates

Helm chart following helm-standards.md: standard labels, hardened security, RBAC.

**Files:**
- Create: `internal/scaffold/templates/helm/Chart.yaml.tmpl`
- Create: `internal/scaffold/templates/helm/values.yaml.tmpl`
- Create: `internal/scaffold/templates/helm/helmignore.tmpl`
- Create: `internal/scaffold/templates/helm/helpers.tpl.tmpl`
- Create: `internal/scaffold/templates/helm/NOTES.txt.tmpl`
- Create: `internal/scaffold/templates/helm/deployment.yaml.tmpl`
- Create: `internal/scaffold/templates/helm/service.yaml.tmpl`
- Create: `internal/scaffold/templates/helm/serviceaccount.yaml.tmpl`
- Create: `internal/scaffold/templates/helm/clusterrole.yaml.tmpl`
- Create: `internal/scaffold/templates/helm/clusterrolebinding.yaml.tmpl`
- Create: `internal/scaffold/templates/helm/makefile-helm.tmpl`
- Create: `internal/scaffold/helm.go`
- Create: `internal/scaffold/helm_test.go`

- [ ] **Step 1: Create template files**

Create `internal/scaffold/templates/helm/Chart.yaml.tmpl`:
```
apiVersion: v2
name: {{.ProjectName}}
description: A Helm chart for {{.ProjectName}}
type: application
version: 0.1.0
appVersion: "0.1.0"
```

Create `internal/scaffold/templates/helm/values.yaml.tmpl`:
```
# -- Number of replicas
replicaCount: 1

image:
  # -- Container image repository
  repository: {{.ProjectName}}
  # -- Image pull policy
  pullPolicy: IfNotPresent
  # -- Image tag (defaults to chart appVersion)
  tag: ""

serviceAccount:
  # -- Create a ServiceAccount
  create: true
  # -- ServiceAccount name (generated if empty)
  name: ""

rbac:
  # -- Create ClusterRole and ClusterRoleBinding
  create: true

service:
  # -- Service port for metrics
  port: 8080

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    memory: 256Mi

serviceMonitor:
  # -- Create a ServiceMonitor (requires Prometheus Operator CRD)
  enabled: false
```

Create `internal/scaffold/templates/helm/helmignore.tmpl`:
```
.DS_Store
*.swp
*.swo
*~
.git
.gitignore
.worktrees
```

Create `internal/scaffold/templates/helm/helpers.tpl.tmpl` (note: use `{{"{{"}}`/`{{"}}"}}`  to escape Helm template delimiters from Go template delimiters):
```
{{"{{"}}/*
Expand the name of the chart.
*/{{"}}"}}
{{"{{"}}- define "{{.ProjectName}}.name" -{{"}}"}}
{{"{{"}}- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" {{"}}"}}
{{"{{"}}- end {{"}}"}}

{{"{{"}}/*
Create a default fully qualified app name.
*/{{"}}"}}
{{"{{"}}- define "{{.ProjectName}}.fullname" -{{"}}"}}
{{"{{"}}- if .Values.fullnameOverride {{"}}"}}
{{"{{"}}- .Values.fullnameOverride | trunc 63 | trimSuffix "-" {{"}}"}}
{{"{{"}}- else {{"}}"}}
{{"{{"}}- $name := default .Chart.Name .Values.nameOverride {{"}}"}}
{{"{{"}}- if contains $name .Release.Name {{"}}"}}
{{"{{"}}- .Release.Name | trunc 63 | trimSuffix "-" {{"}}"}}
{{"{{"}}- else {{"}}"}}
{{"{{"}}- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" {{"}}"}}
{{"{{"}}- end {{"}}"}}
{{"{{"}}- end {{"}}"}}
{{"{{"}}- end {{"}}"}}

{{"{{"}}/*
Create chart name and version as used by the chart label.
*/{{"}}"}}
{{"{{"}}- define "{{.ProjectName}}.chart" -{{"}}"}}
{{"{{"}}- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" {{"}}"}}
{{"{{"}}- end {{"}}"}}

{{"{{"}}/*
Common labels
*/{{"}}"}}
{{"{{"}}- define "{{.ProjectName}}.labels" -{{"}}"}}
helm.sh/chart: {{"{{"}} include "{{.ProjectName}}.chart" . {{"}}"}}
{{"{{"}} include "{{.ProjectName}}.selectorLabels" . {{"}}"}}
app.kubernetes.io/version: {{"{{"}} .Chart.AppVersion | quote {{"}}"}}
app.kubernetes.io/managed-by: {{"{{"}} .Release.Service {{"}}"}}
app.kubernetes.io/part-of: {{.ProjectName}}
{{"{{"}}- end {{"}}"}}

{{"{{"}}/*
Selector labels
*/{{"}}"}}
{{"{{"}}- define "{{.ProjectName}}.selectorLabels" -{{"}}"}}
app.kubernetes.io/name: {{"{{"}} include "{{.ProjectName}}.name" . {{"}}"}}
app.kubernetes.io/instance: {{"{{"}} .Release.Name {{"}}"}}
{{"{{"}}- end {{"}}"}}

{{"{{"}}/*
ServiceAccount name
*/{{"}}"}}
{{"{{"}}- define "{{.ProjectName}}.serviceAccountName" -{{"}}"}}
{{"{{"}}- if .Values.serviceAccount.create {{"}}"}}
{{"{{"}}- default (include "{{.ProjectName}}.fullname" .) .Values.serviceAccount.name {{"}}"}}
{{"{{"}}- else {{"}}"}}
{{"{{"}}- default "default" .Values.serviceAccount.name {{"}}"}}
{{"{{"}}- end {{"}}"}}
{{"{{"}}- end {{"}}"}}
```

Create `internal/scaffold/templates/helm/NOTES.txt.tmpl`:
```
{{.ProjectName}} has been installed.

Release: {{"{{"}} .Release.Name {{"}}"}}
Namespace: {{"{{"}} .Release.Namespace {{"}}"}}
```

Create `internal/scaffold/templates/helm/deployment.yaml.tmpl`:
```
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{"{{"}} include "{{.ProjectName}}.fullname" . {{"}}"}}
  labels:
    {{"{{"}}- include "{{.ProjectName}}.labels" . | nindent 4 {{"}}"}}
spec:
  replicas: {{"{{"}} .Values.replicaCount {{"}}"}}
  selector:
    matchLabels:
      {{"{{"}}- include "{{.ProjectName}}.selectorLabels" . | nindent 6 {{"}}"}}
  template:
    metadata:
      labels:
        {{"{{"}}- include "{{.ProjectName}}.selectorLabels" . | nindent 8 {{"}}"}}
    spec:
      serviceAccountName: {{"{{"}} include "{{.ProjectName}}.serviceAccountName" . {{"}}"}}
      securityContext:
        runAsNonRoot: true
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: {{"{{"}} .Chart.Name {{"}}"}}
          image: "{{"{{"}} .Values.image.repository {{"}}"}}:{{"{{"}} .Values.image.tag | default .Chart.AppVersion {{"}}"}}"
          imagePullPolicy: {{"{{"}} .Values.image.pullPolicy {{"}}"}}
          ports:
            - name: http-metrics
              containerPort: 8080
              protocol: TCP
          livenessProbe:
            httpGet:
              path: /healthz
              port: http-metrics
          readinessProbe:
            httpGet:
              path: /readyz
              port: http-metrics
          resources:
            {{"{{"}}- toYaml .Values.resources | nindent 12 {{"}}"}}
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop:
                - ALL
            runAsNonRoot: true
            runAsUser: 65532
```

Create `internal/scaffold/templates/helm/service.yaml.tmpl`:
```
apiVersion: v1
kind: Service
metadata:
  name: {{"{{"}} include "{{.ProjectName}}.fullname" . {{"}}"}}
  labels:
    {{"{{"}}- include "{{.ProjectName}}.labels" . | nindent 4 {{"}}"}}
spec:
  ports:
    - name: http-metrics
      port: {{"{{"}} .Values.service.port {{"}}"}}
      targetPort: http-metrics
      protocol: TCP
  selector:
    {{"{{"}}- include "{{.ProjectName}}.selectorLabels" . | nindent 4 {{"}}"}}
```

Create `internal/scaffold/templates/helm/serviceaccount.yaml.tmpl`:
```
{{"{{"}}- if .Values.serviceAccount.create -{{"}}"}}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{"{{"}} include "{{.ProjectName}}.serviceAccountName" . {{"}}"}}
  labels:
    {{"{{"}}- include "{{.ProjectName}}.labels" . | nindent 4 {{"}}"}}
{{"{{"}}- end {{"}}"}}
```

Create `internal/scaffold/templates/helm/clusterrole.yaml.tmpl`:
```
{{"{{"}}- if .Values.rbac.create -{{"}}"}}
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: {{"{{"}} include "{{.ProjectName}}.fullname" . {{"}}"}}
  labels:
    {{"{{"}}- include "{{.ProjectName}}.labels" . | nindent 4 {{"}}"}}
rules:
  - apiGroups: [""]
    resources: ["pods", "services", "configmaps"]
    verbs: ["get", "list", "watch"]
{{"{{"}}- end {{"}}"}}
```

Create `internal/scaffold/templates/helm/clusterrolebinding.yaml.tmpl`:
```
{{"{{"}}- if .Values.rbac.create -{{"}}"}}
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{"{{"}} include "{{.ProjectName}}.fullname" . {{"}}"}}
  labels:
    {{"{{"}}- include "{{.ProjectName}}.labels" . | nindent 4 {{"}}"}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: {{"{{"}} include "{{.ProjectName}}.fullname" . {{"}}"}}
subjects:
  - kind: ServiceAccount
    name: {{"{{"}} include "{{.ProjectName}}.serviceAccountName" . {{"}}"}}
    namespace: {{"{{"}} .Release.Namespace {{"}}"}}
{{"{{"}}- end {{"}}"}}
```

Create `internal/scaffold/templates/helm/makefile-helm.tmpl`:
```
## Helm Targets

.PHONY: helm-lint
helm-lint: ## Lint Helm chart.
	helm lint charts/{{.ProjectName}}

.PHONY: helm-template
helm-template: ## Render Helm chart templates.
	helm template release charts/{{.ProjectName}}
```

- [ ] **Step 2: Create helm.go with file specs and test**

Create `internal/scaffold/helm.go`:

```go
package scaffold

func HelmSpecs(params Params) []FileSpec {
	return []FileSpec{
		{TemplatePath: "templates/helm/Chart.yaml.tmpl", OutputPath: "charts/" + params.ProjectName + "/Chart.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/values.yaml.tmpl", OutputPath: "charts/" + params.ProjectName + "/values.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/helmignore.tmpl", OutputPath: "charts/" + params.ProjectName + "/.helmignore", Perm: 0o644},
		{TemplatePath: "templates/helm/helpers.tpl.tmpl", OutputPath: "charts/" + params.ProjectName + "/templates/_helpers.tpl", Perm: 0o644},
		{TemplatePath: "templates/helm/NOTES.txt.tmpl", OutputPath: "charts/" + params.ProjectName + "/templates/NOTES.txt", Perm: 0o644},
		{TemplatePath: "templates/helm/deployment.yaml.tmpl", OutputPath: "charts/" + params.ProjectName + "/templates/deployment.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/service.yaml.tmpl", OutputPath: "charts/" + params.ProjectName + "/templates/service.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/serviceaccount.yaml.tmpl", OutputPath: "charts/" + params.ProjectName + "/templates/serviceaccount.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/clusterrole.yaml.tmpl", OutputPath: "charts/" + params.ProjectName + "/templates/clusterrole.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/clusterrolebinding.yaml.tmpl", OutputPath: "charts/" + params.ProjectName + "/templates/clusterrolebinding.yaml", Perm: 0o644},
	}
}

func HelmMakefileTemplate() string {
	return "templates/helm/makefile-helm.tmpl"
}
```

Create `internal/scaffold/helm_test.go`:

```go
package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelmSpecs_RenderAll(t *testing.T) {
	params := Params{
		ProjectName: "myapp",
		Module:      "github.com/test/myapp",
		Helm:        true,
	}

	dir := t.TempDir()
	if err := RenderLayer("helm", dir, params, HelmSpecs(params)); err != nil {
		t.Fatalf("RenderLayer helm: %v", err)
	}

	wantFiles := []string{
		"charts/myapp/Chart.yaml",
		"charts/myapp/values.yaml",
		"charts/myapp/.helmignore",
		"charts/myapp/templates/_helpers.tpl",
		"charts/myapp/templates/NOTES.txt",
		"charts/myapp/templates/deployment.yaml",
		"charts/myapp/templates/service.yaml",
		"charts/myapp/templates/serviceaccount.yaml",
		"charts/myapp/templates/clusterrole.yaml",
		"charts/myapp/templates/clusterrolebinding.yaml",
	}
	for _, f := range wantFiles {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("missing file %s: %v", f, err)
		}
	}

	chart, _ := os.ReadFile(filepath.Join(dir, "charts/myapp/Chart.yaml"))
	if !strings.Contains(string(chart), "name: myapp") {
		t.Error("Chart.yaml should contain project name")
	}

	deployment, _ := os.ReadFile(filepath.Join(dir, "charts/myapp/templates/deployment.yaml"))
	if !strings.Contains(string(deployment), "runAsNonRoot: true") {
		t.Error("deployment should have hardened security context")
	}
	if !strings.Contains(string(deployment), "http-metrics") {
		t.Error("deployment should expose http-metrics port")
	}

	helpers, _ := os.ReadFile(filepath.Join(dir, "charts/myapp/templates/_helpers.tpl"))
	if !strings.Contains(string(helpers), "myapp.fullname") {
		t.Error("helpers should define fullname template")
	}
}
```

- [ ] **Step 3: Run test to verify it passes**

Run: `go test ./internal/scaffold/ -v -run TestHelm`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/scaffold/
git commit -m "Add Helm layer templates: chart, values, RBAC, deployment, service"
```

---

### Task 9: Init Command

Wire everything together: create dir, git init, render layers, merge Makefile, go mod tidy, setup hooks, set secrets.

**Files:**
- Create: `internal/scaffold/init.go`
- Create: `internal/scaffold/init_test.go`
- Modify: `cmd/goscaffold/main.go`

- [ ] **Step 1: Write the failing test**

Create `internal/scaffold/init_test.go`:

```go
package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit_CLIOnly(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "testproj")

	params := Params{
		ProjectName: "testproj",
		Module:      "github.com/test/testproj",
		CLI:         true,
	}

	if err := Init(dir, params); err != nil {
		t.Fatalf("Init: %v", err)
	}

	wantFiles := []string{
		"go.mod",
		"cmd/testproj/main.go",
		"Makefile",
		".golangci.yml",
		".gitignore",
		".github/workflows/ci.yml",
		"hack/ci-checks.sh",
		".githooks/pre-push",
		"AGENTS.md",
		"README.md",
		"internal/.gitkeep",
		"pkg/.gitkeep",
		".goreleaser.yaml",
		".github/workflows/release.yml",
	}
	for _, f := range wantFiles {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("missing %s: %v", f, err)
		}
	}

	shouldNotExist := []string{"Dockerfile", ".dockerignore", "charts/testproj/Chart.yaml"}
	for _, f := range shouldNotExist {
		if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
			t.Errorf("unexpected file %s should not exist for CLI-only", f)
		}
	}

	makefile, _ := os.ReadFile(filepath.Join(dir, "Makefile"))
	mf := string(makefile)
	if !strings.Contains(mf, "## CLI Targets") {
		t.Error("Makefile should contain CLI Targets section")
	}
	if strings.Contains(mf, "## Controller Targets") {
		t.Error("Makefile should NOT contain Controller Targets for CLI-only")
	}
}

func TestInit_AllLayers(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "fullproj")

	params := Params{
		ProjectName: "fullproj",
		Module:      "github.com/test/fullproj",
		CLI:         true,
		Controller:  true,
		Helm:        true,
	}

	if err := Init(dir, params); err != nil {
		t.Fatalf("Init: %v", err)
	}

	wantFiles := []string{
		"Makefile",
		".goreleaser.yaml",
		"Dockerfile",
		".dockerignore",
		"internal/controller/reconciler.go",
		"hack/e2e-up.sh",
		"charts/fullproj/Chart.yaml",
		"charts/fullproj/templates/deployment.yaml",
	}
	for _, f := range wantFiles {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("missing %s: %v", f, err)
		}
	}

	makefile, _ := os.ReadFile(filepath.Join(dir, "Makefile"))
	mf := string(makefile)
	if !strings.Contains(mf, "## CLI Targets") {
		t.Error("Makefile should contain CLI Targets")
	}
	if !strings.Contains(mf, "## Controller Targets") {
		t.Error("Makefile should contain Controller Targets")
	}
	if !strings.Contains(mf, "## Helm Targets") {
		t.Error("Makefile should contain Helm Targets")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/scaffold/ -v -run TestInit`
Expected: FAIL — `Init` not defined

- [ ] **Step 3: Write implementation**

Create `internal/scaffold/init.go`:

```go
package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

func Init(targetDir string, params Params) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	fmt.Printf("Creating %s in %s\n", params.ProjectName, targetDir)

	if err := RenderLayer("base", targetDir, params, BaseSpecs(params)); err != nil {
		return fmt.Errorf("render base layer: %w", err)
	}

	if err := WritePlaceholders(targetDir); err != nil {
		return fmt.Errorf("write placeholders: %w", err)
	}

	if err := appendMakefileSections(targetDir, params); err != nil {
		return err
	}

	if params.CLI {
		if err := RenderLayer("cli", targetDir, params, CLISpecs()); err != nil {
			return fmt.Errorf("render cli layer: %w", err)
		}
	}

	if params.Controller {
		if err := RenderLayer("controller", targetDir, params, ControllerSpecs(params)); err != nil {
			return fmt.Errorf("render controller layer: %w", err)
		}
	}

	if params.Helm {
		if err := RenderLayer("helm", targetDir, params, HelmSpecs(params)); err != nil {
			return fmt.Errorf("render helm layer: %w", err)
		}
	}

	return nil
}

func appendMakefileSections(targetDir string, params Params) error {
	makefilePath := filepath.Join(targetDir, "Makefile")
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return fmt.Errorf("read Makefile: %w", err)
	}

	mf := string(content)
	layers := []struct {
		enabled  bool
		template string
	}{
		{params.CLI, CLIMakefileTemplate()},
		{params.Controller, ControllerMakefileTemplate()},
		{params.Helm, HelmMakefileTemplate()},
	}

	for _, layer := range layers {
		if !layer.enabled {
			continue
		}
		tmplData, err := templates.ReadFile(layer.template)
		if err != nil {
			return fmt.Errorf("read makefile template: %w", err)
		}
		rendered, err := RenderTemplate(layer.template, string(tmplData), params)
		if err != nil {
			return err
		}
		mf = appendBeforeTools(mf, rendered)
	}

	if err := os.WriteFile(makefilePath, []byte(mf), 0o644); err != nil {
		return fmt.Errorf("write Makefile: %w", err)
	}
	return nil
}

func appendBeforeTools(content, section string) string {
	const toolsMarker = "## Tools"
	idx := len(content)
	if i := indexOf(content, toolsMarker); i >= 0 {
		idx = i
	}
	return content[:idx] + section + "\n" + content[idx:]
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/scaffold/ -v -run TestInit`
Expected: PASS

- [ ] **Step 5: Wire into main.go**

Update `cmd/goscaffold/main.go` — replace the `newInitCmd` function:

```go
func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <project-name>",
		Short: "Create a new Go project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			cli, _ := cmd.Flags().GetBool("cli")
			controller, _ := cmd.Flags().GetBool("controller")
			helm, _ := cmd.Flags().GetBool("helm")

			if !cli && !controller && !helm {
				return fmt.Errorf("at least one of --cli, --controller, or --helm is required")
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			module, _ := cmd.Flags().GetString("module")
			if module == "" {
				prefix := cfg.ModulePrefix
				if prefix == "" {
					return fmt.Errorf("--module is required (or set module_prefix in ~/.config/goscaffold/config.yaml)")
				}
				module = prefix + "/" + name
			}

			params := scaffold.Params{
				ProjectName: name,
				Module:      module,
				CLI:         cli,
				Controller:  controller,
				Helm:        helm,
			}

			targetDir := filepath.Join(".", name)
			if err := scaffold.Init(targetDir, params); err != nil {
				return err
			}

			fmt.Printf("\nProject %s created successfully.\n", name)
			fmt.Println("Next steps:")
			fmt.Printf("  cd %s\n", name)
			fmt.Println("  git init && git add -A && git commit -m 'Initial commit'")
			fmt.Println("  make setup-hooks")
			return nil
		},
	}
	cmd.Flags().String("module", "", "Go module path (default: <module_prefix>/<project-name>)")
	cmd.Flags().Bool("cli", false, "Include CLI distribution layer")
	cmd.Flags().Bool("controller", false, "Include K8s controller layer")
	cmd.Flags().Bool("helm", false, "Include Helm chart layer")
	return cmd
}
```

Add imports for `config` and `scaffold` packages and `path/filepath` at the top of `main.go`.

- [ ] **Step 6: Verify build**

Run: `go mod tidy && make build`
Expected: build succeeds

- [ ] **Step 7: Commit**

```bash
git add cmd/ internal/scaffold/init.go internal/scaffold/init_test.go
git commit -m "Add init command: orchestrates layer rendering and Makefile merging"
```

---

### Task 10: Add Command

Detect existing project, render missing files for a layer, merge Makefile section.

**Files:**
- Create: `internal/scaffold/add.go`
- Create: `internal/scaffold/add_test.go`
- Modify: `cmd/goscaffold/main.go`

- [ ] **Step 1: Write the failing test**

Create `internal/scaffold/add_test.go`:

```go
package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdd_Controller(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "proj")

	params := Params{
		ProjectName: "proj",
		Module:      "github.com/test/proj",
		CLI:         true,
	}
	if err := Init(dir, params); err != nil {
		t.Fatalf("Init: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "Dockerfile")); err == nil {
		t.Fatal("Dockerfile should not exist before add")
	}

	if err := Add(dir, "controller", Params{ProjectName: "proj", Module: "github.com/test/proj", Controller: true}); err != nil {
		t.Fatalf("Add controller: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "Dockerfile")); err != nil {
		t.Error("Dockerfile should exist after add controller")
	}
	if _, err := os.Stat(filepath.Join(dir, "internal/controller/reconciler.go")); err != nil {
		t.Error("reconciler.go should exist after add controller")
	}

	makefile, _ := os.ReadFile(filepath.Join(dir, "Makefile"))
	if !strings.Contains(string(makefile), "## Controller Targets") {
		t.Error("Makefile should contain Controller Targets after add")
	}
}

func TestAdd_SkipsExisting(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "proj")

	params := Params{
		ProjectName: "proj",
		Module:      "github.com/test/proj",
		CLI:         true,
	}
	if err := Init(dir, params); err != nil {
		t.Fatalf("Init: %v", err)
	}

	err := Add(dir, "cli", Params{ProjectName: "proj", Module: "github.com/test/proj", CLI: true})
	if err != nil {
		t.Fatalf("Add cli (should succeed with skips): %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/scaffold/ -v -run TestAdd`
Expected: FAIL — `Add` not defined

- [ ] **Step 3: Write implementation**

Create `internal/scaffold/add.go`:

```go
package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	mf "github.com/jholm117/goscaffold/internal/makefile"
)

func Add(targetDir string, layer string, params Params) error {
	var specs []FileSpec
	var makefileTmpl string
	var sentinel string

	switch layer {
	case "cli":
		specs = CLISpecs()
		makefileTmpl = CLIMakefileTemplate()
		sentinel = "## CLI Targets"
	case "controller":
		specs = ControllerSpecs(params)
		makefileTmpl = ControllerMakefileTemplate()
		sentinel = "## Controller Targets"
	case "helm":
		specs = HelmSpecs(params)
		makefileTmpl = HelmMakefileTemplate()
		sentinel = "## Helm Targets"
	default:
		return fmt.Errorf("unknown layer: %s (valid: cli, controller, helm)", layer)
	}

	fmt.Printf("Adding %s layer to %s\n", layer, targetDir)

	if err := RenderLayer(layer, targetDir, params, specs); err != nil {
		return fmt.Errorf("render %s layer: %w", layer, err)
	}

	makefilePath := filepath.Join(targetDir, "Makefile")
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return fmt.Errorf("read Makefile: %w", err)
	}

	if mf.HasSection(string(content), sentinel) {
		fmt.Printf("  skip   Makefile %s (already present)\n", sentinel)
		return nil
	}

	tmplData, err := templates.ReadFile(makefileTmpl)
	if err != nil {
		return fmt.Errorf("read makefile template: %w", err)
	}
	rendered, err := RenderTemplate(makefileTmpl, string(tmplData), params)
	if err != nil {
		return err
	}

	updated := mf.AppendSection(string(content), rendered)
	if err := os.WriteFile(makefilePath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write Makefile: %w", err)
	}
	fmt.Printf("  update Makefile (added %s)\n", sentinel)

	return nil
}

func DetectProject(dir string) (Params, error) {
	goMod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return Params{}, fmt.Errorf("not a Go project (no go.mod): %w", err)
	}

	var module string
	for _, line := range strings.Split(string(goMod), "\n") {
		if strings.HasPrefix(line, "module ") {
			module = strings.TrimPrefix(line, "module ")
			break
		}
	}

	parts := strings.Split(module, "/")
	name := parts[len(parts)-1]

	return Params{
		ProjectName: name,
		Module:      module,
	}, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/scaffold/ -v -run TestAdd`
Expected: PASS

- [ ] **Step 5: Wire into main.go**

Update `cmd/goscaffold/main.go` — replace the `newAddCmd` function:

```go
func newAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "add <layer>",
		Short:     "Add a layer to an existing project",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"cli", "controller", "helm"},
		RunE: func(cmd *cobra.Command, args []string) error {
			layer := args[0]

			params, err := scaffold.DetectProject(".")
			if err != nil {
				return err
			}

			switch layer {
			case "cli":
				params.CLI = true
			case "controller":
				params.Controller = true
			case "helm":
				params.Helm = true
			}

			return scaffold.Add(".", layer, params)
		},
	}
}
```

- [ ] **Step 6: Verify build**

Run: `make build`
Expected: build succeeds

- [ ] **Step 7: Commit**

```bash
git add cmd/ internal/scaffold/add.go internal/scaffold/add_test.go
git commit -m "Add add command: layer files into existing projects"
```

---

### Task 11: Integration Tests

Full workflow tests: init a project, verify it builds, verify `add` works.

**Files:**
- Create: `internal/scaffold/integration_test.go`

- [ ] **Step 1: Write integration tests**

Create `internal/scaffold/integration_test.go`:

```go
//go:build integration

package scaffold_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jholm117/goscaffold/internal/scaffold"
)

func TestIntegration_InitCLI_Builds(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "myapp")
	params := scaffold.Params{
		ProjectName: "myapp",
		Module:      "github.com/test/myapp",
		CLI:         true,
	}

	if err := scaffold.Init(dir, params); err != nil {
		t.Fatalf("Init: %v", err)
	}

	runInDir(t, dir, "go", "mod", "tidy")
	runInDir(t, dir, "go", "build", "./...")
	runInDir(t, dir, "go", "vet", "./...")
}

func TestIntegration_InitController_Builds(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "myctrl")
	params := scaffold.Params{
		ProjectName: "myctrl",
		Module:      "github.com/test/myctrl",
		Controller:  true,
	}

	if err := scaffold.Init(dir, params); err != nil {
		t.Fatalf("Init: %v", err)
	}

	runInDir(t, dir, "go", "mod", "tidy")
	runInDir(t, dir, "go", "build", "./cmd/myctrl/")
	runInDir(t, dir, "go", "vet", "./...")
}

func TestIntegration_InitAll_Builds(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "fullproj")
	params := scaffold.Params{
		ProjectName: "fullproj",
		Module:      "github.com/test/fullproj",
		CLI:         true,
		Controller:  true,
		Helm:        true,
	}

	if err := scaffold.Init(dir, params); err != nil {
		t.Fatalf("Init: %v", err)
	}

	runInDir(t, dir, "go", "mod", "tidy")
	runInDir(t, dir, "go", "build", "./cmd/fullproj/")
	runInDir(t, dir, "go", "vet", "./...")
}

func TestIntegration_AddController_AfterCLI(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "proj")
	params := scaffold.Params{
		ProjectName: "proj",
		Module:      "github.com/test/proj",
		CLI:         true,
	}

	if err := scaffold.Init(dir, params); err != nil {
		t.Fatalf("Init: %v", err)
	}

	addParams := scaffold.Params{
		ProjectName: "proj",
		Module:      "github.com/test/proj",
		Controller:  true,
	}
	if err := scaffold.Add(dir, "controller", addParams); err != nil {
		t.Fatalf("Add: %v", err)
	}

	runInDir(t, dir, "go", "mod", "tidy")
	runInDir(t, dir, "go", "build", "./cmd/proj/")
	runInDir(t, dir, "go", "vet", "./...")

	if _, err := os.Stat(filepath.Join(dir, "Dockerfile")); err != nil {
		t.Error("Dockerfile should exist after add controller")
	}
}

func runInDir(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s %v failed: %v", name, args, err)
	}
}
```

- [ ] **Step 2: Run integration tests**

Run: `go test ./internal/scaffold/ -v -tags integration -timeout 5m`
Expected: PASS (may need to fix template issues discovered during real builds)

- [ ] **Step 3: Commit**

```bash
git add internal/scaffold/integration_test.go
git commit -m "Add integration tests: verify generated projects build and vet"
```

---

### Task 12: Run Full Test Suite and Merge

Run everything, merge to main, push.

- [ ] **Step 1: Run all unit tests**

Run: `go test ./... -v`
Expected: PASS

- [ ] **Step 2: Run lint**

Run: `make lint`
Expected: PASS (fix any issues found)

- [ ] **Step 3: Run integration tests**

Run: `go test ./internal/scaffold/ -v -tags integration -timeout 5m`
Expected: PASS

- [ ] **Step 4: Manual smoke test**

Run:
```bash
make build
./bin/goscaffold init /tmp/test-cli --cli --module github.com/test/test-cli
cd /tmp/test-cli && go mod tidy && go build ./... && echo "CLI OK"

./bin/goscaffold init /tmp/test-ctrl --controller --module github.com/test/test-ctrl
cd /tmp/test-ctrl && go mod tidy && go build ./cmd/test-ctrl/ && echo "Controller OK"

./bin/goscaffold init /tmp/test-all --cli --controller --helm --module github.com/test/test-all
cd /tmp/test-all && go mod tidy && go build ./cmd/test-all/ && helm lint charts/test-all && echo "All OK"
```

- [ ] **Step 5: Merge to main and push**

```bash
cd ~/wip-repos/goscaffold
git merge design
git worktree remove .worktrees/design
git branch -d design
gh repo create jholm117/goscaffold --public --source=. --description "Scaffold Go projects with composable layers" --push
```

- [ ] **Step 6: Tag and release**

```bash
git tag v0.1.0
git push origin v0.1.0
```

GoReleaser will build, create the GitHub release, and publish the Homebrew formula.
