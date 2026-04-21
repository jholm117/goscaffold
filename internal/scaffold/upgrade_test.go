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
	os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM golang"), 0o644)

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

	data, _ := os.ReadFile(path)
	if string(data) != "content" {
		t.Errorf("content = %q, want %q", string(data), "content")
	}
}
