package scaffold

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Params struct {
	ProjectName string
	Module      string
	GoVersion   string
	CLI         bool
	Controller  bool
	Helm        bool
	Pkg         bool
}

func (p Params) OwnerRepo() string {
	parts := strings.SplitN(p.Module, "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return p.Module
}

func RenderTemplate(name, text string, params Params) (string, error) {
	tmpl, err := template.New(name).Parse(text)
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("execute template %s: %w", name, err)
	}
	return buf.String(), nil
}

func WriteFile(path, content string, perm fs.FileMode) (written bool, err error) {
	if _, err := os.Stat(path); err == nil {
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

type FileSpec struct {
	TemplatePath string
	OutputPath   string
	Perm         fs.FileMode
}

func RenderLayer(layer string, targetDir string, params Params, specs []FileSpec) error {
	for _, spec := range specs {
		tmplData, err := templates.ReadFile(spec.TemplatePath)
		if err != nil {
			return fmt.Errorf("read template %s: %w", spec.TemplatePath, err)
		}
		rendered, err := RenderTemplate(spec.TemplatePath, string(tmplData), params)
		if err != nil {
			return err
		}
		outPath := filepath.Join(targetDir, spec.OutputPath)
		written, err := WriteFile(outPath, rendered, spec.Perm)
		if err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
		if written {
			fmt.Printf("  create %s\n", spec.OutputPath)
		} else {
			fmt.Printf("  skip   %s (exists)\n", spec.OutputPath)
		}
	}
	return nil
}
