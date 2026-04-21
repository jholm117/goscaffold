# goscaffold upgrade Design Spec

## Overview

Add an `upgrade` command to goscaffold that brings existing Go projects in line with the latest scaffolding patterns. Propagates improvements — updated CI checks, new linters, tool version bumps, badges — across all projects without clobbering project-specific content.

## Problem

goscaffold encodes best practices into templates. When those practices evolve (e.g., adding a `make tools` target, gating releases on CI, bumping golangci-lint), existing projects don't benefit. The only option today is manual file-by-file updates across every project.

## Command

```
goscaffold upgrade [--dry-run]
```

Run inside an existing project directory. Detects which layers are present, then upgrades all managed files to match goscaffold's current templates.

- **Default**: autopilot — writes all changes immediately
- **`--dry-run`**: prints what would change without writing

Output:

```
Upgrading myproject (cli, controller, helm)
  overwrite .golangci.yml
  overwrite hack/ci-checks.sh
  overwrite .github/workflows/ci.yml
  overwrite .github/workflows/release.yml
  overwrite .goreleaser.yaml
  overwrite Dockerfile
  overwrite .dockerignore
  overwrite hack/e2e-up.sh
  overwrite hack/e2e-down.sh
  overwrite hack/kind-config.yaml
  targets   Makefile (14 targets, 6 variables)
  badges    README.md
  section   AGENTS.md (Build & Test, Project Layout)
  skip      cmd/myproject/main.go (project-specific)
  skip      go.mod (project-specific)

Done. Review changes with: git diff
```

## Layer Detection

Detect which layers the project uses by file presence:

| Layer | Detected by |
|---|---|
| cli | `.goreleaser.yaml` exists |
| controller | `Dockerfile` exists |
| helm | `charts/` directory exists |

Project name and module come from `DetectProject()` (reads go.mod), which already exists.

## File Upgrade Strategies

### Strategy 1: Full Overwrite

Write rendered template directly. No diffing, no merging. For files with no project-specific content.

**Files:**

| Layer | Files |
|---|---|
| base | `.golangci.yml`, `hack/ci-checks.sh`, `.githooks/pre-push`, `.github/workflows/ci.yml` |
| cli | `.goreleaser.yaml`, `.github/workflows/release.yml` |
| controller | `Dockerfile`, `.dockerignore`, `hack/e2e-up.sh`, `hack/e2e-down.sh`, `hack/kind-config.yaml` |
| helm | All files under `charts/<name>/` |

New files (e.g., a template added in a future goscaffold version) are created if they don't exist.

### Strategy 2: Makefile Target-Level Replacement

goscaffold knows every target, variable, and define block it manages. Upgrade finds each one in the existing Makefile and replaces just that block. Custom targets anywhere in the file are untouched.

**Managed elements:**

Variables (find `NAME ?=` or `NAME =` line, replace):
- `PROJECT_NAME`
- `LOCALBIN`
- `GOLANGCI_LINT_VERSION`
- `GOVULNCHECK_VERSION`
- `GORELEASER_VERSION` (cli)
- `IMG` (controller)
- `GOLANGCI_LINT`
- `GOVULNCHECK`
- `GORELEASER` (cli)

Targets (find `.PHONY: name` + target line + recipe, replace):
- Base: `build`, `install`, `test`, `lint`, `lint-fix`, `lint-config`, `fmt`, `vet`, `govulncheck`, `setup-hooks`, `clean`, `help`, `tools`
- CLI: `release-snapshot`, `goreleaser` (tool install)
- Controller: `docker-build`, `docker-push`, `e2e`, `e2e-up`, `e2e-down`
- Helm: `helm-lint`, `helm-template`

Special blocks:
- `define go-install-tool` through `endef`
- `$(LOCALBIN):` target (mkdir)
- `$(GOLANGCI_LINT): $(LOCALBIN)` tool install targets
- `$(GOVULNCHECK): $(LOCALBIN)` tool install targets
- `$(GORELEASER): $(LOCALBIN)` tool install targets (cli)

**Parsing approach:** Pattern matching on a predictable structure, not a full Makefile parser. Each target block follows:
```
.PHONY: name
name: [deps] [## help text]
\trecipe line(s)
```

A block ends at the next `.PHONY:`, next variable assignment, next `##` comment, `define`, or a blank line not followed by a tab-indented line.

**New targets:** If goscaffold adds a new managed target in a future version and it doesn't exist in the project's Makefile, append it in the appropriate location.

**Removed targets:** If a layer was removed (e.g., user deleted Dockerfile but Makefile still has `docker-build`), don't remove targets. The Makefile might have custom dependencies on them.

### Strategy 3: README Badge Patch

Find consecutive lines starting with `[![` after the `# title` line. Replace them with the current badge set. If no badge lines exist, insert them after the title line with a blank line separator.

### Strategy 4: AGENTS.md Section Replace

Find `## Build & Test` header, replace content through to the next `##` header. Same for `## Project Layout`. Leave all other sections untouched (e.g., project-specific architecture docs the user added).

## What `upgrade` Does NOT Touch

- `cmd/*/main.go` — project-specific logic
- `go.mod` / `go.sum` — dependency management
- `internal/`, `pkg/` — project code
- `AGENTS.md` sections other than Build & Test and Project Layout
- `README.md` body below badges
- Any file not in goscaffold's template set
- Custom Makefile targets

## Architecture

New/modified files in goscaffold:

```
internal/
  scaffold/
    upgrade.go          # Upgrade() orchestration
    upgrade_test.go
  makefile/
    parse.go            # ParseTarget, ParseVariable, ParseDefine
    parse_test.go
    replace.go          # ReplaceTarget, ReplaceVariable, ReplaceDefine
    replace_test.go
  markdown/
    badges.go           # PatchBadges — find/replace badge lines
    badges_test.go
    section.go          # ReplaceSection — find header, replace to next header
    section_test.go
cmd/goscaffold/
    main.go             # Add upgrade subcommand
```

## Testing

### Makefile Parsing & Replacement
- Parse a target block from a Makefile string, verify name and content extracted
- Replace a known target, verify surrounding content preserved
- Replace a variable, verify other variables untouched
- Add a new target that doesn't exist, verify it's appended
- Replace with custom targets interspersed, verify custom targets survive

### README Badges
- Project with no badges → badges inserted after title
- Project with old badges → badges replaced
- Project with badges + content below → content below preserved

### AGENTS.md Sections
- Replace Build & Test section, verify other sections preserved
- Replace Project Layout section, verify custom sections preserved

### Integration
- Init a project, modify managed files, run upgrade, verify files restored to template output
- Init a project, add custom Makefile targets, run upgrade, verify custom targets survive
- Init a project, run upgrade with no changes, verify output says nothing to do (or silently succeeds)
- Dry run prints changes without writing
