# goscaffold upgrade Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `goscaffold upgrade` command that brings existing projects in line with current templates without clobbering custom content.

**Architecture:** Three upgrade strategies — full overwrite for infrastructure files, target-level Makefile replacement via pattern matching, and markdown patching (badges + sections). The upgrade command detects layers from file presence and delegates to each strategy.

**Tech Stack:** Go, text/template, string manipulation (no external parsing libraries)

---

## File Structure

```
internal/
  makefile/
    merge.go           # existing: HasSection, AppendSection
    merge_test.go      # existing
    target.go          # NEW: ParseTarget, ReplaceTarget, ReplaceVariable, ReplaceDefine
    target_test.go     # NEW
  markdown/
    badges.go          # NEW: PatchBadges
    badges_test.go     # NEW
    section.go         # NEW: ReplaceSection
    section_test.go    # NEW
  scaffold/
    upgrade.go         # NEW: Upgrade(), OverwriteFiles(), DetectLayers()
    upgrade_test.go    # NEW
cmd/goscaffold/
    main.go            # MODIFY: add upgrade subcommand
```

---

### Task 1: Makefile Target Parsing and Replacement

The core parsing logic for target-level Makefile upgrades. TDD — tests first.

**Files:**
- Create: `internal/makefile/target.go`
- Create: `internal/makefile/target_test.go`

- [ ] **Step 1: Write failing tests for ReplaceVariable**

Create `internal/makefile/target_test.go`:

```go
package makefile

import (
	"testing"
)

func TestReplaceVariable(t *testing.T) {
	content := "PROJECT_NAME ?= oldname\nLOCALBIN ?= $(shell pwd)/bin\nFOO ?= bar\n"

	result := ReplaceVariable(content, "PROJECT_NAME", "PROJECT_NAME ?= newname")
	if result != "PROJECT_NAME ?= newname\nLOCALBIN ?= $(shell pwd)/bin\nFOO ?= bar\n" {
		t.Errorf("ReplaceVariable wrong:\n%s", result)
	}
}

func TestReplaceVariable_NotFound(t *testing.T) {
	content := "FOO ?= bar\n"
	result := ReplaceVariable(content, "MISSING", "MISSING ?= val")
	if result != content {
		t.Error("ReplaceVariable should return content unchanged when variable not found")
	}
}

func TestReplaceVariable_EqualsSign(t *testing.T) {
	content := "GOLANGCI_LINT = $(LOCALBIN)/golangci-lint\nGOVULNCHECK = $(LOCALBIN)/govulncheck\n"
	result := ReplaceVariable(content, "GOLANGCI_LINT", "GOLANGCI_LINT = $(LOCALBIN)/new-lint")
	if result != "GOLANGCI_LINT = $(LOCALBIN)/new-lint\nGOVULNCHECK = $(LOCALBIN)/govulncheck\n" {
		t.Errorf("ReplaceVariable with = wrong:\n%s", result)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/makefile/ -v -run TestReplaceVariable`
Expected: FAIL — `ReplaceVariable` not defined

- [ ] **Step 3: Implement ReplaceVariable**

Create `internal/makefile/target.go`:

```go
package makefile

import "strings"

func ReplaceVariable(content, name, replacement string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, name+" ?=") || strings.HasPrefix(line, name+" =") {
			lines[i] = replacement
			return strings.Join(lines, "\n")
		}
	}
	return content
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/makefile/ -v -run TestReplaceVariable`
Expected: PASS

- [ ] **Step 5: Write failing tests for ReplaceTarget**

Add to `internal/makefile/target_test.go`:

```go
func TestReplaceTarget(t *testing.T) {
	content := `.PHONY: build
build: ## Build binary.
	go build -o bin/old ./cmd/old

.PHONY: test
test: ## Run tests.
	go test ./...

custom-target:
	echo custom
`
	replacement := `.PHONY: build
build: ## Build binary.
	go build -o bin/new ./cmd/new
`
	result := ReplaceTarget(content, "build", replacement)

	if !strings.Contains(result, "go build -o bin/new ./cmd/new") {
		t.Error("build target should be replaced")
	}
	if !strings.Contains(result, "go test ./...") {
		t.Error("test target should be preserved")
	}
	if !strings.Contains(result, "echo custom") {
		t.Error("custom target should be preserved")
	}
}

func TestReplaceTarget_NotFound(t *testing.T) {
	content := ".PHONY: build\nbuild:\n\techo build\n"
	result := ReplaceTarget(content, "missing", ".PHONY: missing\nmissing:\n\techo hi\n")
	if result != content {
		t.Error("ReplaceTarget should return content unchanged when target not found")
	}
}

func TestReplaceTarget_CustomBetween(t *testing.T) {
	content := `.PHONY: build
build: ## Build.
	go build ./...

my-custom:
	echo custom

.PHONY: test
test: ## Test.
	go test ./...
`
	replacement := `.PHONY: build
build: ## Build.
	go build -o bin/app ./cmd/app
`
	result := ReplaceTarget(content, "build", replacement)

	if !strings.Contains(result, "go build -o bin/app ./cmd/app") {
		t.Error("build target should be replaced")
	}
	if !strings.Contains(result, "echo custom") {
		t.Error("custom target between managed targets should survive")
	}
}
```

- [ ] **Step 6: Run tests to verify they fail**

Run: `go test ./internal/makefile/ -v -run TestReplaceTarget`
Expected: FAIL — `ReplaceTarget` not defined

- [ ] **Step 7: Implement ReplaceTarget**

Add to `internal/makefile/target.go`:

```go
func ReplaceTarget(content, name, replacement string) string {
	marker := ".PHONY: " + name
	start := strings.Index(content, marker)
	if start == -1 {
		return content
	}

	end := findTargetEnd(content, start+len(marker))
	return content[:start] + strings.TrimRight(replacement, "\n") + "\n" + content[end:]
}

func findTargetEnd(content string, after int) int {
	lines := strings.Split(content[after:], "\n")
	pos := after
	inRecipe := false

	for i, line := range lines {
		if i == 0 {
			pos += len(line) + 1
			continue
		}

		if strings.HasPrefix(line, "\t") {
			inRecipe = true
			pos += len(line) + 1
			continue
		}

		if inRecipe && line == "" {
			pos += 1
			continue
		}

		if inRecipe {
			break
		}

		if line == "" {
			pos += 1
			continue
		}

		if strings.HasPrefix(line, "\t") {
			inRecipe = true
			pos += len(line) + 1
			continue
		}

		pos += len(line) + 1
		inRecipe = true
	}

	return pos
}
```

Note: the `findTargetEnd` implementation above is a starting point. The tests will drive the exact behavior. The key rule: a target block is `.PHONY: name`, followed by the target line (`name: deps`), followed by tab-indented recipe lines, ending when we hit a non-tab, non-blank line.

- [ ] **Step 8: Run tests, iterate until they pass**

Run: `go test ./internal/makefile/ -v -run TestReplaceTarget`
Expected: PASS (may need to adjust `findTargetEnd` logic)

- [ ] **Step 9: Write failing tests for ReplaceDefine**

Add to `internal/makefile/target_test.go`:

```go
func TestReplaceDefine(t *testing.T) {
	content := "stuff before\n\ndefine go-install-tool\nold content\nold line 2\nendef\n\nstuff after\n"
	replacement := "define go-install-tool\nnew content\nendef"

	result := ReplaceDefine(content, "go-install-tool", replacement)

	if !strings.Contains(result, "new content") {
		t.Error("define block should be replaced")
	}
	if !strings.Contains(result, "stuff before") {
		t.Error("content before should be preserved")
	}
	if !strings.Contains(result, "stuff after") {
		t.Error("content after should be preserved")
	}
	if strings.Contains(result, "old content") {
		t.Error("old define content should be gone")
	}
}
```

- [ ] **Step 10: Run test to verify it fails**

Run: `go test ./internal/makefile/ -v -run TestReplaceDefine`
Expected: FAIL — `ReplaceDefine` not defined

- [ ] **Step 11: Implement ReplaceDefine**

Add to `internal/makefile/target.go`:

```go
func ReplaceDefine(content, name, replacement string) string {
	marker := "define " + name
	start := strings.Index(content, marker)
	if start == -1 {
		return content
	}

	endMarker := "endef"
	endIdx := strings.Index(content[start:], endMarker)
	if endIdx == -1 {
		return content
	}
	end := start + endIdx + len(endMarker)

	if end < len(content) && content[end] == '\n' {
		end++
	}

	return content[:start] + replacement + "\n" + content[end:]
}
```

- [ ] **Step 12: Run tests to verify they pass**

Run: `go test ./internal/makefile/ -v -run TestReplaceDefine`
Expected: PASS

- [ ] **Step 13: Write failing test for ReplaceSpecialTarget ($(LOCALBIN): etc.)**

Add to `internal/makefile/target_test.go`:

```go
func TestReplaceSpecialTarget(t *testing.T) {
	content := "$(GOLANGCI_LINT): $(LOCALBIN)\n\t$(call go-install-tool,old)\n\n$(GOVULNCHECK): $(LOCALBIN)\n\t$(call go-install-tool,other)\n"

	replacement := "$(GOLANGCI_LINT): $(LOCALBIN)\n\t$(call go-install-tool,new)"

	result := ReplaceSpecialTarget(content, "$(GOLANGCI_LINT):", replacement)

	if !strings.Contains(result, "go-install-tool,new") {
		t.Error("special target should be replaced")
	}
	if !strings.Contains(result, "go-install-tool,other") {
		t.Error("other special target should be preserved")
	}
}
```

- [ ] **Step 14: Implement ReplaceSpecialTarget**

Add to `internal/makefile/target.go`:

```go
func ReplaceSpecialTarget(content, prefix, replacement string) string {
	start := strings.Index(content, prefix)
	if start == -1 {
		return content
	}

	end := findTargetEnd(content, start+len(prefix))
	return content[:start] + strings.TrimRight(replacement, "\n") + "\n" + content[end:]
}
```

- [ ] **Step 15: Run all makefile tests**

Run: `go test ./internal/makefile/ -v`
Expected: PASS

- [ ] **Step 16: Commit**

```bash
git add internal/makefile/target.go internal/makefile/target_test.go
git commit -m "Add Makefile target-level parsing: ReplaceVariable, ReplaceTarget, ReplaceDefine, ReplaceSpecialTarget"
```

---

### Task 2: Markdown Badge Patching

**Files:**
- Create: `internal/markdown/badges.go`
- Create: `internal/markdown/badges_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/markdown/badges_test.go`:

```go
package markdown

import (
	"strings"
	"testing"
)

func TestPatchBadges_InsertWhenNone(t *testing.T) {
	content := "# myproject\n\nSome description.\n\n## Getting Started\n"
	badges := "[![CI](https://example.com/ci.svg)](https://example.com/ci)\n[![Release](https://example.com/release.svg)](https://example.com/release)"

	result := PatchBadges(content, badges)

	if !strings.Contains(result, "[![CI]") {
		t.Error("badges should be inserted")
	}
	if !strings.Contains(result, "Some description.") {
		t.Error("description should be preserved")
	}
	ciIdx := strings.Index(result, "[![CI]")
	descIdx := strings.Index(result, "Some description.")
	if ciIdx > descIdx {
		t.Error("badges should appear before description")
	}
}

func TestPatchBadges_ReplaceExisting(t *testing.T) {
	content := "# myproject\n\n[![OldBadge](https://old.svg)](https://old)\n\nSome description.\n"
	badges := "[![CI](https://new.svg)](https://new)"

	result := PatchBadges(content, badges)

	if strings.Contains(result, "OldBadge") {
		t.Error("old badges should be removed")
	}
	if !strings.Contains(result, "[![CI]") {
		t.Error("new badges should be present")
	}
	if !strings.Contains(result, "Some description.") {
		t.Error("description should be preserved")
	}
}

func TestPatchBadges_PreserveContentBelow(t *testing.T) {
	content := "# myproject\n\n[![Old](https://old.svg)](https://old)\n[![Old2](https://old2.svg)](https://old2)\n\nDescription here.\n\n## Section\n\nMore content.\n"
	badges := "[![New](https://new.svg)](https://new)"

	result := PatchBadges(content, badges)

	if strings.Contains(result, "Old") {
		t.Error("old badges should be removed")
	}
	if !strings.Contains(result, "Description here.") {
		t.Error("description should be preserved")
	}
	if !strings.Contains(result, "## Section") {
		t.Error("sections should be preserved")
	}
}

func TestPatchBadges_NoTitle(t *testing.T) {
	content := "No title here.\n"
	badges := "[![CI](https://ci.svg)](https://ci)"

	result := PatchBadges(content, badges)

	if result != content {
		t.Error("should return content unchanged when no # title found")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/markdown/ -v -run TestPatchBadges`
Expected: FAIL — package/function not defined

- [ ] **Step 3: Implement PatchBadges**

Create `internal/markdown/badges.go`:

```go
package markdown

import "strings"

func PatchBadges(content, badges string) string {
	lines := strings.Split(content, "\n")

	titleIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			titleIdx = i
			break
		}
	}
	if titleIdx == -1 {
		return content
	}

	badgeStart := -1
	badgeEnd := -1
	for i := titleIdx + 1; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}
		if strings.HasPrefix(lines[i], "[![") {
			if badgeStart == -1 {
				badgeStart = i
			}
			badgeEnd = i + 1
		} else {
			break
		}
	}

	badgeLines := strings.Split(badges, "\n")

	if badgeStart != -1 {
		result := make([]string, 0, len(lines))
		result = append(result, lines[:badgeStart]...)
		result = append(result, badgeLines...)
		result = append(result, lines[badgeEnd:]...)
		return strings.Join(result, "\n")
	}

	result := make([]string, 0, len(lines)+len(badgeLines)+1)
	result = append(result, lines[titleIdx])
	result = append(result, "")
	result = append(result, badgeLines...)
	result = append(result, lines[titleIdx+1:]...)
	return strings.Join(result, "\n")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/markdown/ -v -run TestPatchBadges`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/markdown/
git commit -m "Add markdown badge patching: PatchBadges"
```

---

### Task 3: Markdown Section Replacement

For AGENTS.md `## Build & Test` and `## Project Layout` sections.

**Files:**
- Create: `internal/markdown/section.go`
- Create: `internal/markdown/section_test.go`

- [ ] **Step 1: Write failing tests**

Create `internal/markdown/section_test.go`:

```go
package markdown

import (
	"strings"
	"testing"
)

func TestReplaceSection(t *testing.T) {
	content := `# myproject

## Build & Test

- old build command
- old test command

## Architecture

Custom architecture docs here.

## Other

More custom stuff.
`

	newSection := `## Build & Test

- ` + "`make build`" + ` — build binary
- ` + "`make test`" + ` — run tests
- ` + "`make lint`" + ` — run golangci-lint
`

	result := ReplaceSection(content, "## Build & Test", newSection)

	if !strings.Contains(result, "make lint") {
		t.Error("new section content should be present")
	}
	if strings.Contains(result, "old build command") {
		t.Error("old section content should be gone")
	}
	if !strings.Contains(result, "Custom architecture docs here.") {
		t.Error("other sections should be preserved")
	}
	if !strings.Contains(result, "More custom stuff.") {
		t.Error("subsequent sections should be preserved")
	}
}

func TestReplaceSection_NotFound(t *testing.T) {
	content := "# myproject\n\n## Other\n\nstuff\n"
	result := ReplaceSection(content, "## Build & Test", "## Build & Test\n\nnew\n")
	if result != content {
		t.Error("should return content unchanged when section not found")
	}
}

func TestReplaceSection_LastSection(t *testing.T) {
	content := "# myproject\n\n## Build & Test\n\nold content\n"
	newSection := "## Build & Test\n\nnew content\n"

	result := ReplaceSection(content, "## Build & Test", newSection)

	if !strings.Contains(result, "new content") {
		t.Error("section should be replaced even when it's the last section")
	}
	if strings.Contains(result, "old content") {
		t.Error("old content should be gone")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/markdown/ -v -run TestReplaceSection`
Expected: FAIL — `ReplaceSection` not defined

- [ ] **Step 3: Implement ReplaceSection**

Create `internal/markdown/section.go`:

```go
package markdown

import "strings"

func ReplaceSection(content, header, replacement string) string {
	start := strings.Index(content, header)
	if start == -1 {
		return content
	}

	afterHeader := start + len(header)
	rest := content[afterHeader:]

	end := -1
	lines := strings.Split(rest, "\n")
	pos := 0
	for i, line := range lines {
		if i == 0 {
			pos += len(line) + 1
			continue
		}
		if strings.HasPrefix(line, "## ") {
			end = afterHeader + pos
			break
		}
		pos += len(line) + 1
	}

	if end == -1 {
		return content[:start] + strings.TrimRight(replacement, "\n") + "\n"
	}

	return content[:start] + strings.TrimRight(replacement, "\n") + "\n\n" + content[end:]
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/markdown/ -v -run TestReplaceSection`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/markdown/section.go internal/markdown/section_test.go
git commit -m "Add markdown section replacement: ReplaceSection"
```

---

### Task 4: Upgrade Orchestration — Overwrite Strategy

The `Upgrade()` function that ties overwrite files together with layer detection.

**Files:**
- Create: `internal/scaffold/upgrade.go`
- Create: `internal/scaffold/upgrade_test.go`

- [ ] **Step 1: Write failing test for DetectLayers**

Create `internal/scaffold/upgrade_test.go`:

```go
package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectLayers(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".goreleaser.yaml"), []byte(""), 0o644)
	os.MkdirAll(filepath.Join(dir, "charts", "test"), 0o755)

	layers := DetectLayers(dir)

	if !layers.CLI {
		t.Error("CLI should be detected from .goreleaser.yaml")
	}
	if layers.Controller {
		t.Error("Controller should NOT be detected (no Dockerfile)")
	}
	if !layers.Helm {
		t.Error("Helm should be detected from charts/ dir")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/scaffold/ -v -run TestDetectLayers`
Expected: FAIL — `DetectLayers` not defined

- [ ] **Step 3: Implement DetectLayers**

Create `internal/scaffold/upgrade.go`:

```go
package scaffold

import "os"

type Layers struct {
	CLI        bool
	Controller bool
	Helm       bool
}

func DetectLayers(dir string) Layers {
	var l Layers
	if _, err := os.Stat(dir + "/.goreleaser.yaml"); err == nil {
		l.CLI = true
	}
	if _, err := os.Stat(dir + "/Dockerfile"); err == nil {
		l.Controller = true
	}
	if entries, err := os.ReadDir(dir + "/charts"); err == nil && len(entries) > 0 {
		l.Helm = true
	}
	return l
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/scaffold/ -v -run TestDetectLayers`
Expected: PASS

- [ ] **Step 5: Write failing test for OverwriteFile (force-write, unlike WriteFile which skips)**

Add to `internal/scaffold/upgrade_test.go`:

```go
func TestOverwriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("old"), 0o644)

	changed, err := OverwriteFile(path, "new", 0o644)
	if err != nil {
		t.Fatalf("OverwriteFile: %v", err)
	}
	if !changed {
		t.Error("should report changed")
	}

	data, _ := os.ReadFile(path)
	if string(data) != "new" {
		t.Errorf("content = %q, want %q", string(data), "new")
	}
}

func TestOverwriteFile_NoChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("same"), 0o644)

	changed, err := OverwriteFile(path, "same", 0o644)
	if err != nil {
		t.Fatalf("OverwriteFile: %v", err)
	}
	if changed {
		t.Error("should report not changed when content identical")
	}
}

func TestOverwriteFile_Creates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "new.txt")

	changed, err := OverwriteFile(path, "content", 0o644)
	if err != nil {
		t.Fatalf("OverwriteFile: %v", err)
	}
	if !changed {
		t.Error("should report changed for new file")
	}
}
```

- [ ] **Step 6: Implement OverwriteFile**

Add to `internal/scaffold/upgrade.go`:

```go
import (
	"io/fs"
	"os"
	"path/filepath"
)

func OverwriteFile(path, content string, perm fs.FileMode) (changed bool, err error) {
	existing, err := os.ReadFile(path)
	if err == nil && string(existing) == content {
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
```

- [ ] **Step 7: Run tests to verify they pass**

Run: `go test ./internal/scaffold/ -v -run TestOverwrite`
Expected: PASS

- [ ] **Step 8: Commit**

```bash
git add internal/scaffold/upgrade.go internal/scaffold/upgrade_test.go
git commit -m "Add DetectLayers and OverwriteFile for upgrade command"
```

---

### Task 5: Upgrade Command — Full Integration

Wire all strategies into the `Upgrade()` function and add the cobra command.

**Files:**
- Modify: `internal/scaffold/upgrade.go`
- Modify: `internal/scaffold/upgrade_test.go`
- Modify: `cmd/goscaffold/main.go`

- [ ] **Step 1: Write failing integration test**

Add to `internal/scaffold/upgrade_test.go`:

```go
func TestUpgrade_RestoresModifiedFiles(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "proj")
	params := Params{
		ProjectName: "proj",
		Module:      "github.com/test/proj",
		CLI:         true,
	}
	if err := Init(dir, params); err != nil {
		t.Fatalf("Init: %v", err)
	}

	golangciPath := filepath.Join(dir, ".golangci.yml")
	os.WriteFile(golangciPath, []byte("corrupted"), 0o644)

	ciPath := filepath.Join(dir, "hack/ci-checks.sh")
	os.WriteFile(ciPath, []byte("corrupted"), 0o755)

	if err := Upgrade(dir, false); err != nil {
		t.Fatalf("Upgrade: %v", err)
	}

	golangci, _ := os.ReadFile(golangciPath)
	if string(golangci) == "corrupted" {
		t.Error(".golangci.yml should be restored")
	}

	ci, _ := os.ReadFile(ciPath)
	if string(ci) == "corrupted" {
		t.Error("ci-checks.sh should be restored")
	}
}

func TestUpgrade_PreservesCustomMakefileTargets(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "proj")
	params := Params{
		ProjectName: "proj",
		Module:      "github.com/test/proj",
		CLI:         true,
	}
	if err := Init(dir, params); err != nil {
		t.Fatalf("Init: %v", err)
	}

	makefilePath := filepath.Join(dir, "Makefile")
	mf, _ := os.ReadFile(makefilePath)
	custom := string(mf) + "\n.PHONY: my-custom\nmy-custom: ## My custom target.\n\techo custom\n"
	os.WriteFile(makefilePath, []byte(custom), 0o644)

	if err := Upgrade(dir, false); err != nil {
		t.Fatalf("Upgrade: %v", err)
	}

	result, _ := os.ReadFile(makefilePath)
	if !strings.Contains(string(result), "my-custom") {
		t.Error("custom Makefile target should survive upgrade")
	}
}

func TestUpgrade_DryRun(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "proj")
	params := Params{
		ProjectName: "proj",
		Module:      "github.com/test/proj",
		CLI:         true,
	}
	if err := Init(dir, params); err != nil {
		t.Fatalf("Init: %v", err)
	}

	golangciPath := filepath.Join(dir, ".golangci.yml")
	os.WriteFile(golangciPath, []byte("corrupted"), 0o644)

	if err := Upgrade(dir, true); err != nil {
		t.Fatalf("Upgrade dry-run: %v", err)
	}

	data, _ := os.ReadFile(golangciPath)
	if string(data) != "corrupted" {
		t.Error("dry-run should not modify files")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/scaffold/ -v -run TestUpgrade`
Expected: FAIL — `Upgrade` not defined

- [ ] **Step 3: Implement Upgrade()**

Add to `internal/scaffold/upgrade.go`:

```go
import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	mf "github.com/jholm117/goscaffold/internal/makefile"
	"github.com/jholm117/goscaffold/internal/markdown"
)

func Upgrade(targetDir string, dryRun bool) error {
	params, err := DetectProject(targetDir)
	if err != nil {
		return err
	}

	layers := DetectLayers(targetDir)
	params.CLI = layers.CLI
	params.Controller = layers.Controller
	params.Helm = layers.Helm

	layerNames := []string{}
	if layers.CLI {
		layerNames = append(layerNames, "cli")
	}
	if layers.Controller {
		layerNames = append(layerNames, "controller")
	}
	if layers.Helm {
		layerNames = append(layerNames, "helm")
	}

	mode := ""
	if dryRun {
		mode = " (dry-run)"
	}
	fmt.Printf("Upgrading %s (%s)%s\n", params.ProjectName, strings.Join(layerNames, ", "), mode)

	if err := upgradeOverwriteFiles(targetDir, params, layers, dryRun); err != nil {
		return err
	}

	if err := upgradeMakefile(targetDir, params, layers, dryRun); err != nil {
		return err
	}

	if err := upgradeReadmeBadges(targetDir, params, dryRun); err != nil {
		return err
	}

	if err := upgradeAgentsSections(targetDir, params, dryRun); err != nil {
		return err
	}

	fmt.Println("\nDone. Review changes with: git diff")
	return nil
}

func upgradeOverwriteFiles(targetDir string, params Params, layers Layers, dryRun bool) error {
	type overwriteSpec struct {
		templatePath string
		outputPath   string
		perm         fs.FileMode
	}

	specs := []overwriteSpec{
		{"templates/base/golangci.yml.tmpl", ".golangci.yml", 0o644},
		{"templates/base/ci-checks.sh.tmpl", "hack/ci-checks.sh", 0o755},
		{"templates/base/pre-push.tmpl", ".githooks/pre-push", 0o755},
		{"templates/base/ci.yml.tmpl", ".github/workflows/ci.yml", 0o644},
		{"templates/base/gitignore.tmpl", ".gitignore", 0o644},
	}

	if layers.CLI {
		specs = append(specs,
			overwriteSpec{"templates/cli/goreleaser.yaml.tmpl", ".goreleaser.yaml", 0o644},
			overwriteSpec{"templates/cli/release.yml.tmpl", ".github/workflows/release.yml", 0o644},
		)
	}

	if layers.Controller {
		specs = append(specs,
			overwriteSpec{"templates/controller/Dockerfile.tmpl", "Dockerfile", 0o644},
			overwriteSpec{"templates/controller/dockerignore.tmpl", ".dockerignore", 0o644},
			overwriteSpec{"templates/controller/e2e-up.sh.tmpl", "hack/e2e-up.sh", 0o755},
			overwriteSpec{"templates/controller/e2e-down.sh.tmpl", "hack/e2e-down.sh", 0o755},
			overwriteSpec{"templates/controller/kind-config.yaml.tmpl", "hack/kind-config.yaml", 0o644},
		)
	}

	if layers.Helm {
		for _, spec := range HelmSpecs(params) {
			specs = append(specs, overwriteSpec{spec.TemplatePath, spec.OutputPath, spec.Perm})
		}
	}

	for _, spec := range specs {
		tmplData, err := templates.ReadFile(spec.templatePath)
		if err != nil {
			return fmt.Errorf("read template %s: %w", spec.templatePath, err)
		}
		rendered, err := RenderTemplate(spec.templatePath, string(tmplData), params)
		if err != nil {
			return err
		}

		outPath := filepath.Join(targetDir, spec.outputPath)
		if dryRun {
			existing, _ := os.ReadFile(outPath)
			if string(existing) != rendered {
				fmt.Printf("  overwrite %s\n", spec.outputPath)
			}
			continue
		}

		changed, err := OverwriteFile(outPath, rendered, spec.perm)
		if err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
		if changed {
			fmt.Printf("  overwrite %s\n", spec.outputPath)
		}
	}

	return nil
}

func upgradeMakefile(targetDir string, params Params, layers Layers, dryRun bool) error {
	makefilePath := filepath.Join(targetDir, "Makefile")
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return fmt.Errorf("read Makefile: %w", err)
	}

	result := string(content)

	baseTmpl, err := templates.ReadFile("templates/base/Makefile.tmpl")
	if err != nil {
		return err
	}
	renderedBase, err := RenderTemplate("Makefile.tmpl", string(baseTmpl), params)
	if err != nil {
		return err
	}

	managedVars := []string{
		"PROJECT_NAME", "LOCALBIN",
		"GOLANGCI_LINT_VERSION", "GOVULNCHECK_VERSION",
		"GOLANGCI_LINT", "GOVULNCHECK",
	}
	if layers.CLI {
		managedVars = append(managedVars, "GORELEASER_VERSION", "GORELEASER")
	}
	if layers.Controller {
		managedVars = append(managedVars, "IMG")
	}

	for _, v := range managedVars {
		line := extractVariable(renderedBase, v)
		if line == "" {
			for _, tmplName := range []string{"templates/cli/makefile-cli.tmpl", "templates/controller/makefile-controller.tmpl"} {
				tmplData, _ := templates.ReadFile(tmplName)
				rendered, _ := RenderTemplate(tmplName, string(tmplData), params)
				line = extractVariable(rendered, v)
				if line != "" {
					break
				}
			}
		}
		if line != "" {
			result = mf.ReplaceVariable(result, v, line)
		}
	}

	managedTargets := []string{
		"build", "install", "test", "lint", "lint-fix", "lint-config",
		"fmt", "vet", "govulncheck", "setup-hooks", "clean", "help", "tools",
	}
	if layers.CLI {
		managedTargets = append(managedTargets, "release-snapshot")
	}
	if layers.Controller {
		managedTargets = append(managedTargets, "docker-build", "docker-push", "e2e", "e2e-up", "e2e-down")
	}
	if layers.Helm {
		managedTargets = append(managedTargets, "helm-lint", "helm-template")
	}

	allRendered := renderedBase
	if layers.CLI {
		tmplData, _ := templates.ReadFile("templates/cli/makefile-cli.tmpl")
		rendered, _ := RenderTemplate("makefile-cli.tmpl", string(tmplData), params)
		allRendered += "\n" + rendered
	}
	if layers.Controller {
		tmplData, _ := templates.ReadFile("templates/controller/makefile-controller.tmpl")
		rendered, _ := RenderTemplate("makefile-controller.tmpl", string(tmplData), params)
		allRendered += "\n" + rendered
	}
	if layers.Helm {
		tmplData, _ := templates.ReadFile("templates/helm/makefile-helm.tmpl")
		rendered, _ := RenderTemplate("makefile-helm.tmpl", string(tmplData), params)
		allRendered += "\n" + rendered
	}

	for _, name := range managedTargets {
		block := extractTarget(allRendered, name)
		if block != "" {
			result = mf.ReplaceTarget(result, name, block)
		}
	}

	result = mf.ReplaceDefine(result, "go-install-tool", extractDefine(renderedBase, "go-install-tool"))

	specialTargets := []string{"$(LOCALBIN):", "$(GOLANGCI_LINT):", "$(GOVULNCHECK):"}
	if layers.CLI {
		specialTargets = append(specialTargets, "$(GORELEASER):")
	}
	for _, prefix := range specialTargets {
		block := extractSpecialTarget(allRendered, prefix)
		if block != "" {
			result = mf.ReplaceSpecialTarget(result, prefix, block)
		}
	}

	if dryRun {
		if result != string(content) {
			fmt.Println("  targets   Makefile")
		}
		return nil
	}

	if result != string(content) {
		if err := os.WriteFile(makefilePath, []byte(result), 0o644); err != nil {
			return fmt.Errorf("write Makefile: %w", err)
		}
		fmt.Println("  targets   Makefile")
	}

	return nil
}

func upgradeReadmeBadges(targetDir string, params Params, dryRun bool) error {
	readmePath := filepath.Join(targetDir, "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return nil
	}

	badges := fmt.Sprintf(
		"[![CI](https://%s/actions/workflows/ci.yml/badge.svg)](https://%s/actions/workflows/ci.yml)\n"+
			"[![Release](https://img.shields.io/github/v/release/%s)](https://%s/releases/latest)",
		params.Module, params.Module, params.OwnerRepo(), params.Module,
	)

	result := markdown.PatchBadges(string(content), badges)

	if dryRun {
		if result != string(content) {
			fmt.Println("  badges    README.md")
		}
		return nil
	}

	if result != string(content) {
		if err := os.WriteFile(readmePath, []byte(result), 0o644); err != nil {
			return fmt.Errorf("write README.md: %w", err)
		}
		fmt.Println("  badges    README.md")
	}

	return nil
}

func upgradeAgentsSections(targetDir string, params Params, dryRun bool) error {
	agentsPath := filepath.Join(targetDir, "AGENTS.md")
	content, err := os.ReadFile(agentsPath)
	if err != nil {
		return nil
	}

	tmplData, err := templates.ReadFile("templates/base/agents.md.tmpl")
	if err != nil {
		return err
	}
	rendered, err := RenderTemplate("agents.md.tmpl", string(tmplData), params)
	if err != nil {
		return err
	}

	result := string(content)
	for _, header := range []string{"## Build & Test", "## Project Layout"} {
		section := extractMarkdownSection(rendered, header)
		if section != "" {
			result = markdown.ReplaceSection(result, header, section)
		}
	}

	if dryRun {
		if result != string(content) {
			fmt.Println("  section   AGENTS.md (Build & Test, Project Layout)")
		}
		return nil
	}

	if result != string(content) {
		if err := os.WriteFile(agentsPath, []byte(result), 0o644); err != nil {
			return fmt.Errorf("write AGENTS.md: %w", err)
		}
		fmt.Println("  section   AGENTS.md (Build & Test, Project Layout)")
	}

	return nil
}

func extractVariable(content, name string) string {
	for line := range strings.SplitSeq(content, "\n") {
		if strings.HasPrefix(line, name+" ?=") || strings.HasPrefix(line, name+" =") {
			return line
		}
	}
	return ""
}

func extractTarget(content, name string) string {
	marker := ".PHONY: " + name
	start := strings.Index(content, marker)
	if start == -1 {
		return ""
	}
	end := mf.FindTargetEnd(content, start+len(marker))
	return strings.TrimRight(content[start:end], "\n") + "\n"
}

func extractDefine(content, name string) string {
	marker := "define " + name
	start := strings.Index(content, marker)
	if start == -1 {
		return ""
	}
	endMarker := "endef"
	endIdx := strings.Index(content[start:], endMarker)
	if endIdx == -1 {
		return ""
	}
	end := start + endIdx + len(endMarker)
	return content[start:end]
}

func extractSpecialTarget(content, prefix string) string {
	start := strings.Index(content, prefix)
	if start == -1 {
		return ""
	}
	end := mf.FindTargetEnd(content, start+len(prefix))
	return strings.TrimRight(content[start:end], "\n") + "\n"
}

func extractMarkdownSection(content, header string) string {
	start := strings.Index(content, header)
	if start == -1 {
		return ""
	}
	afterHeader := start + len(header)
	rest := content[afterHeader:]
	lines := strings.Split(rest, "\n")
	pos := 0
	for i, line := range lines {
		if i == 0 {
			pos += len(line) + 1
			continue
		}
		if strings.HasPrefix(line, "## ") {
			return strings.TrimRight(content[start:afterHeader+pos], "\n") + "\n"
		}
		pos += len(line) + 1
	}
	return strings.TrimRight(content[start:], "\n") + "\n"
}
```

Note: This references `mf.FindTargetEnd` which needs to be exported from the makefile package. Update `findTargetEnd` to `FindTargetEnd` in `internal/makefile/target.go`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/scaffold/ -v -run TestUpgrade`
Expected: PASS (may need iterations on the Makefile parsing)

- [ ] **Step 5: Wire into main.go**

Add to `cmd/goscaffold/main.go` in `main()`:

```go
root.AddCommand(newUpgradeCmd())
```

Add function:

```go
func newUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade managed files to latest templates",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			return scaffold.Upgrade(".", dryRun)
		},
	}
	cmd.Flags().Bool("dry-run", false, "Print what would change without writing")
	return cmd
}
```

- [ ] **Step 6: Verify build**

Run: `go mod tidy && make build && make lint`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add cmd/ internal/scaffold/upgrade.go internal/scaffold/upgrade_test.go internal/makefile/target.go
git commit -m "Add upgrade command: overwrite, Makefile targets, badges, AGENTS.md sections"
```

---

### Task 6: Smoke Test, Full Suite, and Merge

**Files:** None new — verification only.

- [ ] **Step 1: Run all tests**

Run: `go test ./... -v`
Expected: PASS

- [ ] **Step 2: Run lint**

Run: `make lint`
Expected: PASS

- [ ] **Step 3: Manual smoke test**

```bash
make build

# Create a project
cd /tmp && rm -rf upgrade-test
goscaffold init upgrade-test --cli --controller --helm --module github.com/test/upgrade-test

# Corrupt some files
echo "corrupted" > /tmp/upgrade-test/.golangci.yml
echo "corrupted" > /tmp/upgrade-test/hack/ci-checks.sh

# Add a custom Makefile target
echo -e '\n.PHONY: my-custom\nmy-custom:\n\techo hello' >> /tmp/upgrade-test/Makefile

# Run upgrade
cd /tmp/upgrade-test
~/wip-repos/goscaffold/.worktrees/upgrade-spec/bin/goscaffold upgrade

# Verify
grep "allow-parallel-runners" .golangci.yml  # should find it (restored)
grep "my-custom" Makefile                     # should find it (preserved)
grep "ci.yml/badge.svg" README.md             # should find it (badges)
```

- [ ] **Step 4: Test dry-run**

```bash
echo "corrupted" > /tmp/upgrade-test/.golangci.yml
cd /tmp/upgrade-test
~/wip-repos/goscaffold/.worktrees/upgrade-spec/bin/goscaffold upgrade --dry-run
cat .golangci.yml  # should still say "corrupted"
```

- [ ] **Step 5: Merge to main and push**

```bash
cd ~/wip-repos/goscaffold
git merge upgrade-spec
git worktree remove .worktrees/upgrade-spec
git branch -d upgrade-spec
git push
```

- [ ] **Step 6: Tag v0.2.0**

```bash
git tag v0.2.0
git push origin v0.2.0
```
