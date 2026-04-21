package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

func Init(targetDir string, params Params) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	fmt.Printf("Creating %s in %s\n", params.ProjectName, targetDir)

	if err := RenderLayer("base", targetDir, params, BaseSpecs(params)); err != nil {
		return fmt.Errorf("render base layer: %w", err)
	}

	if err := WritePlaceholders(targetDir); err != nil {
		return fmt.Errorf("write placeholders: %w", err)
	}

	if err := appendMakefileSections(targetDir, params); err != nil {
		return err
	}

	if params.CLI {
		if err := RenderLayer("cli", targetDir, params, CLISpecs()); err != nil {
			return fmt.Errorf("render cli layer: %w", err)
		}
	}

	if params.Controller {
		if err := RenderLayer("controller", targetDir, params, ControllerSpecs(params)); err != nil {
			return fmt.Errorf("render controller layer: %w", err)
		}
	}

	if params.Helm {
		if err := RenderLayer("helm", targetDir, params, HelmSpecs(params)); err != nil {
			return fmt.Errorf("render helm layer: %w", err)
		}
	}

	return nil
}

func appendMakefileSections(targetDir string, params Params) error {
	makefilePath := filepath.Join(targetDir, "Makefile")
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return fmt.Errorf("read Makefile: %w", err)
	}

	mf := string(content)
	layers := []struct {
		enabled  bool
		template string
	}{
		{params.CLI, CLIMakefileTemplate()},
		{params.Controller, ControllerMakefileTemplate()},
		{params.Helm, HelmMakefileTemplate()},
	}

	for _, layer := range layers {
		if !layer.enabled {
			continue
		}
		tmplData, err := templates.ReadFile(layer.template)
		if err != nil {
			return fmt.Errorf("read makefile template: %w", err)
		}
		rendered, err := RenderTemplate(layer.template, string(tmplData), params)
		if err != nil {
			return err
		}
		mf = appendBeforeTools(mf, rendered)
	}

	if err := os.WriteFile(makefilePath, []byte(mf), 0o644); err != nil {
		return fmt.Errorf("write Makefile: %w", err)
	}
	return nil
}

func appendBeforeTools(content, section string) string {
	const toolsMarker = "## Tools"
	idx := len(content)
	if i := indexOf(content, toolsMarker); i >= 0 {
		idx = i
	}
	return content[:idx] + section + "\n" + content[idx:]
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
