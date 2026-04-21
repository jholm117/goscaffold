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
