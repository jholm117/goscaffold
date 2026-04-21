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
