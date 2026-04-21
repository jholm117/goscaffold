package scaffold

import (
	"io/fs"
	"os"
	"path/filepath"
)

// Layers represents which optional layers are detected in a project.
type Layers struct {
	CLI        bool
	Controller bool
	Helm       bool
}

// DetectLayers inspects file presence to determine which layers are active.
// CLI: .goreleaser.yaml exists. Controller: Dockerfile exists. Helm: charts/ dir has entries.
func DetectLayers(dir string) Layers {
	var l Layers
	if _, err := os.Stat(filepath.Join(dir, ".goreleaser.yaml")); err == nil {
		l.CLI = true
	}
	if _, err := os.Stat(filepath.Join(dir, "Dockerfile")); err == nil {
		l.Controller = true
	}
	if entries, err := os.ReadDir(filepath.Join(dir, "charts")); err == nil && len(entries) > 0 {
		l.Helm = true
	}
	return l
}

// OverwriteFile writes content to path, creating parent directories as needed.
// Returns false if the file already exists with identical content.
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
