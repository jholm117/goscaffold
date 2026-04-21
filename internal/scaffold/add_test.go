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

	addParams := Params{ProjectName: "proj", Module: "github.com/test/proj", Controller: true}
	if err := Add(dir, "controller", addParams); err != nil {
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
