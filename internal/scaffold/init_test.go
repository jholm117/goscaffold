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
