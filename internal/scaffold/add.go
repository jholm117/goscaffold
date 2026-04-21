package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	mf "github.com/jholm117/goscaffold/internal/makefile"
)

func Add(targetDir string, layer string, params Params) error {
	var specs []FileSpec
	var makefileTmpl string
	var sentinel string

	switch layer {
	case "cli":
		specs = CLISpecs()
		makefileTmpl = CLIMakefileTemplate()
		sentinel = "## CLI Targets"
	case "controller":
		specs = ControllerSpecs(params)
		makefileTmpl = ControllerMakefileTemplate()
		sentinel = "## Controller Targets"
	case "helm":
		specs = HelmSpecs(params)
		makefileTmpl = HelmMakefileTemplate()
		sentinel = "## Helm Targets"
	default:
		return fmt.Errorf("unknown layer: %s (valid: cli, controller, helm)", layer)
	}

	fmt.Printf("Adding %s layer to %s\n", layer, targetDir)

	if err := RenderLayer(layer, targetDir, params, specs); err != nil {
		return fmt.Errorf("render %s layer: %w", layer, err)
	}

	makefilePath := filepath.Join(targetDir, "Makefile")
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return fmt.Errorf("read Makefile: %w", err)
	}

	if mf.HasSection(string(content), sentinel) {
		fmt.Printf("  skip   Makefile %s (already present)\n", sentinel)
		return nil
	}

	tmplData, err := templates.ReadFile(makefileTmpl)
	if err != nil {
		return fmt.Errorf("read makefile template: %w", err)
	}
	rendered, err := RenderTemplate(makefileTmpl, string(tmplData), params)
	if err != nil {
		return err
	}

	updated := mf.AppendSection(string(content), rendered)
	if err := os.WriteFile(makefilePath, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("write Makefile: %w", err)
	}
	fmt.Printf("  update Makefile (added %s)\n", sentinel)

	return nil
}

func DetectProject(dir string) (Params, error) {
	goMod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return Params{}, fmt.Errorf("not a Go project (no go.mod): %w", err)
	}

	var module string
	for _, line := range strings.Split(string(goMod), "\n") {
		if strings.HasPrefix(line, "module ") {
			module = strings.TrimPrefix(line, "module ")
			break
		}
	}

	parts := strings.Split(module, "/")
	name := parts[len(parts)-1]

	return Params{
		ProjectName: name,
		Module:      module,
	}, nil
}
