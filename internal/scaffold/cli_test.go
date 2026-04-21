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
