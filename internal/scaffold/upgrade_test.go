package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const corruptedContent = "corrupted"

func TestDetectLayers(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, ".goreleaser.yaml"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "charts", "test"), 0o755); err != nil {
		t.Fatal(err)
	}

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

func TestDetectLayers_Empty(t *testing.T) {
	dir := t.TempDir()
	layers := DetectLayers(dir)

	if layers.CLI {
		t.Error("CLI should not be detected")
	}
	if layers.Controller {
		t.Error("Controller should not be detected")
	}
	if layers.Helm {
		t.Error("Helm should not be detected")
	}
}

func TestDetectLayers_Controller(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM golang"), 0o644); err != nil {
		t.Fatal(err)
	}

	layers := DetectLayers(dir)

	if !layers.Controller {
		t.Error("Controller should be detected from Dockerfile")
	}
	if layers.CLI {
		t.Error("CLI should not be detected")
	}
}

func TestOverwriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

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

	if err := os.WriteFile(path, []byte("same"), 0o644); err != nil {
		t.Fatal(err)
	}

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

	data, _ := os.ReadFile(path)
	if string(data) != "content" {
		t.Errorf("content = %q, want %q", string(data), "content")
	}
}

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
	if err := os.WriteFile(golangciPath, []byte(corruptedContent), 0o644); err != nil {
		t.Fatal(err)
	}

	ciPath := filepath.Join(dir, "hack/ci-checks.sh")
	if err := os.WriteFile(ciPath, []byte(corruptedContent), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := Upgrade(dir, false); err != nil {
		t.Fatalf("Upgrade: %v", err)
	}

	golangci, _ := os.ReadFile(golangciPath)
	if string(golangci) == corruptedContent {
		t.Error(".golangci.yml should be restored")
	}

	ci, _ := os.ReadFile(ciPath)
	if string(ci) == corruptedContent {
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

	if err := os.WriteFile(makefilePath, []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}

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
	if err := os.WriteFile(golangciPath, []byte(corruptedContent), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Upgrade(dir, true); err != nil {
		t.Fatalf("Upgrade dry-run: %v", err)
	}

	data, _ := os.ReadFile(golangciPath)
	if string(data) != corruptedContent {
		t.Error("dry-run should not modify files")
	}
}

func TestUpgrade_BumpsGoVersion(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "proj")
	params := Params{
		ProjectName: "proj",
		Module:      "github.com/test/proj",
		CLI:         true,
	}
	if err := Init(dir, params); err != nil {
		t.Fatalf("Init: %v", err)
	}

	goModPath := filepath.Join(dir, "go.mod")
	content, _ := os.ReadFile(goModPath)
	old := strings.Replace(string(content), "go 1.26.2", "go 1.24.2", 1)
	if err := os.WriteFile(goModPath, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Upgrade(dir, false); err != nil {
		t.Fatalf("Upgrade: %v", err)
	}

	updated, _ := os.ReadFile(goModPath)
	if !strings.Contains(string(updated), "go 1.26.2") {
		t.Error("go.mod should be bumped to 1.26.2")
	}
	if strings.Contains(string(updated), "go 1.24.2") {
		t.Error("old Go version should be gone")
	}
	if !strings.Contains(string(updated), "module github.com/test/proj") {
		t.Error("module path should be preserved")
	}
}
