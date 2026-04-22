package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	mf "github.com/jholm117/goscaffold/internal/makefile"
	"github.com/jholm117/goscaffold/internal/markdown"
)

const GoVersion = "1.26.2"

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

// Upgrade brings an existing project in line with current templates.
// It detects layers from file presence and applies three strategies:
// overwrite for infrastructure files, target-level Makefile replacement,
// and markdown patching (badges + sections).
func Upgrade(targetDir string, dryRun bool) error {
	params, err := DetectProject(targetDir)
	if err != nil {
		return err
	}

	layers := DetectLayers(targetDir)
	params.GoVersion = GoVersion
	params.CLI = layers.CLI
	params.Controller = layers.Controller
	params.Helm = layers.Helm
	if _, err := os.Stat(filepath.Join(targetDir, "pkg")); err == nil {
		params.Pkg = true
	}

	var layerNames []string
	if layers.CLI {
		layerNames = append(layerNames, "cli")
	}
	if layers.Controller {
		layerNames = append(layerNames, "controller")
	}
	if layers.Helm {
		layerNames = append(layerNames, "helm")
	}

	mode := ""
	if dryRun {
		mode = " (dry-run)"
	}
	fmt.Printf("Upgrading %s (%s)%s\n", params.ProjectName, strings.Join(layerNames, ", "), mode)

	if err := upgradeOverwriteFiles(targetDir, params, layers, dryRun); err != nil {
		return err
	}

	if err := upgradeGoMod(targetDir, dryRun); err != nil {
		return err
	}

	if err := upgradeMakefile(targetDir, params, layers, dryRun); err != nil {
		return err
	}

	if err := upgradeReadmeBadges(targetDir, params, dryRun); err != nil {
		return err
	}

	if err := upgradeAgentsSections(targetDir, params, dryRun); err != nil {
		return err
	}

	fmt.Println("\nDone. Review changes with: git diff")
	return nil
}

func upgradeGoMod(targetDir string, dryRun bool) error {
	goModPath := filepath.Join(targetDir, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(content), "\n")
	changed := false
	for i, line := range lines {
		if strings.HasPrefix(line, "go ") {
			want := "go " + GoVersion
			if line != want {
				lines[i] = want
				changed = true
			}
			break
		}
	}

	if !changed {
		return nil
	}

	if dryRun {
		fmt.Println("  gomod     go.mod (go " + GoVersion + ")")
		return nil
	}

	if err := os.WriteFile(goModPath, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return fmt.Errorf("write go.mod: %w", err)
	}
	fmt.Println("  gomod     go.mod (go " + GoVersion + ")")

	fmt.Println("  tidy      go mod tidy")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = targetDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}
	return nil
}

func upgradeOverwriteFiles(targetDir string, params Params, layers Layers, dryRun bool) error {
	type overwriteSpec struct {
		templatePath string
		outputPath   string
		perm         fs.FileMode
	}

	specs := []overwriteSpec{
		{"templates/base/golangci.yml.tmpl", ".golangci.yml", 0o644},
		{"templates/base/ci-checks.sh.tmpl", "hack/ci-checks.sh", 0o755},
		{"templates/base/pre-push.tmpl", ".githooks/pre-push", 0o755},
		{"templates/base/ci.yml.tmpl", ".github/workflows/ci.yml", 0o644},
		{"templates/base/gitignore.tmpl", ".gitignore", 0o644},
	}

	if layers.CLI {
		specs = append(specs,
			overwriteSpec{"templates/cli/goreleaser.yaml.tmpl", ".goreleaser.yaml", 0o644},
			overwriteSpec{"templates/cli/release.yml.tmpl", ".github/workflows/release.yml", 0o644},
		)
	}

	if layers.Controller {
		specs = append(specs,
			overwriteSpec{"templates/controller/Dockerfile.tmpl", "Dockerfile", 0o644},
			overwriteSpec{"templates/controller/dockerignore.tmpl", ".dockerignore", 0o644},
			overwriteSpec{"templates/controller/e2e-up.sh.tmpl", "hack/e2e-up.sh", 0o755},
			overwriteSpec{"templates/controller/e2e-down.sh.tmpl", "hack/e2e-down.sh", 0o755},
			overwriteSpec{"templates/controller/kind-config.yaml.tmpl", "hack/kind-config.yaml", 0o644},
		)
	}

	if layers.Helm {
		fmt.Println("  skip      charts/ (project-specific)")
	}

	for _, spec := range specs {
		tmplData, err := templates.ReadFile(spec.templatePath)
		if err != nil {
			return fmt.Errorf("read template %s: %w", spec.templatePath, err)
		}
		rendered, err := RenderTemplate(spec.templatePath, string(tmplData), params)
		if err != nil {
			return err
		}

		outPath := filepath.Join(targetDir, spec.outputPath)
		if dryRun {
			existing, _ := os.ReadFile(outPath)
			if string(existing) != rendered {
				fmt.Printf("  overwrite %s\n", spec.outputPath)
			}
			continue
		}

		changed, err := OverwriteFile(outPath, rendered, spec.perm)
		if err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
		if changed {
			fmt.Printf("  overwrite %s\n", spec.outputPath)
		}
	}

	return nil
}

func upgradeMakefile(targetDir string, params Params, layers Layers, dryRun bool) error {
	makefilePath := filepath.Join(targetDir, "Makefile")
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		return fmt.Errorf("read Makefile: %w", err)
	}

	result := string(content)

	// Render the base Makefile template.
	baseTmpl, err := templates.ReadFile("templates/base/Makefile.tmpl")
	if err != nil {
		return err
	}
	renderedBase, err := RenderTemplate("Makefile.tmpl", string(baseTmpl), params)
	if err != nil {
		return err
	}

	// Build composite rendered output from all enabled layer Makefile templates.
	allRendered := renderedBase
	if layers.CLI {
		tmplData, err := templates.ReadFile("templates/cli/makefile-cli.tmpl")
		if err != nil {
			return err
		}
		rendered, err := RenderTemplate("makefile-cli.tmpl", string(tmplData), params)
		if err != nil {
			return err
		}
		allRendered += "\n" + rendered
	}
	if layers.Controller {
		tmplData, err := templates.ReadFile("templates/controller/makefile-controller.tmpl")
		if err != nil {
			return err
		}
		rendered, err := RenderTemplate("makefile-controller.tmpl", string(tmplData), params)
		if err != nil {
			return err
		}
		allRendered += "\n" + rendered
	}
	if layers.Helm {
		tmplData, err := templates.ReadFile("templates/helm/makefile-helm.tmpl")
		if err != nil {
			return err
		}
		rendered, err := RenderTemplate("makefile-helm.tmpl", string(tmplData), params)
		if err != nil {
			return err
		}
		allRendered += "\n" + rendered
	}

	// Replace managed variables.
	managedVars := []string{
		"PROJECT_NAME", "LOCALBIN",
		"GOLANGCI_LINT_VERSION", "GOVULNCHECK_VERSION",
		"GOLANGCI_LINT", "GOVULNCHECK",
	}
	if layers.CLI {
		managedVars = append(managedVars, "GORELEASER_VERSION", "GORELEASER")
	}
	if layers.Controller {
		managedVars = append(managedVars, "IMG")
	}

	for _, v := range managedVars {
		line := extractVariable(allRendered, v)
		if line != "" {
			result = mf.ReplaceVariable(result, v, line)
		}
	}

	// Replace managed targets.
	managedTargets := []string{
		"build", "install", "test", "lint", "lint-fix", "lint-config",
		"fmt", "vet", "govulncheck", "setup-hooks", "scaffold-upgrade", "clean", "help", "tools",
		"golangci-lint", "goreleaser",
	}
	if layers.CLI {
		managedTargets = append(managedTargets, "release-snapshot")
	}
	if layers.Controller {
		managedTargets = append(managedTargets, "docker-build", "docker-push", "e2e", "e2e-up", "e2e-down")
	}
	if layers.Helm {
		managedTargets = append(managedTargets, "helm-lint", "helm-template")
	}

	result = replaceOrInsertTargets(result, allRendered, managedTargets)

	// Replace define blocks.
	defineBlock := extractDefine(allRendered, "go-install-tool")
	if defineBlock != "" {
		result = mf.ReplaceDefine(result, "go-install-tool", defineBlock)
	}

	// Replace special targets (tool dependency targets).
	specialTargets := []string{"$(LOCALBIN):", "$(GOLANGCI_LINT):", "$(GOVULNCHECK):"}
	if layers.CLI {
		specialTargets = append(specialTargets, "$(GORELEASER):")
	}
	for _, prefix := range specialTargets {
		block := extractSpecialTarget(allRendered, prefix)
		if block != "" {
			result = mf.ReplaceSpecialTarget(result, prefix, block)
		}
	}

	if dryRun {
		if result != string(content) {
			fmt.Println("  targets   Makefile")
		}
		return nil
	}

	if result != string(content) {
		if err := os.WriteFile(makefilePath, []byte(result), 0o644); err != nil {
			return fmt.Errorf("write Makefile: %w", err)
		}
		fmt.Println("  targets   Makefile")
	}

	return nil
}

func upgradeReadmeBadges(targetDir string, params Params, dryRun bool) error {
	readmePath := filepath.Join(targetDir, "README.md")
	content, err := os.ReadFile(readmePath)
	if err != nil {
		return nil // README not found is not an error
	}

	badges := fmt.Sprintf(
		"[![CI](https://%s/actions/workflows/ci.yml/badge.svg)](https://%s/actions/workflows/ci.yml)\n"+
			"[![Release](https://img.shields.io/github/v/release/%s)](https://%s/releases/latest)",
		params.Module, params.Module, params.OwnerRepo(), params.Module,
	)

	result := markdown.PatchBadges(string(content), badges)

	if dryRun {
		if result != string(content) {
			fmt.Println("  badges    README.md")
		}
		return nil
	}

	if result != string(content) {
		if err := os.WriteFile(readmePath, []byte(result), 0o644); err != nil {
			return fmt.Errorf("write README.md: %w", err)
		}
		fmt.Println("  badges    README.md")
	}

	return nil
}

func upgradeAgentsSections(targetDir string, params Params, dryRun bool) error {
	agentsPath := filepath.Join(targetDir, "AGENTS.md")
	content, err := os.ReadFile(agentsPath)
	if err != nil {
		return nil // AGENTS.md not found is not an error
	}

	tmplData, err := templates.ReadFile("templates/base/agents.md.tmpl")
	if err != nil {
		return err
	}
	rendered, err := RenderTemplate("agents.md.tmpl", string(tmplData), params)
	if err != nil {
		return err
	}

	result := string(content)
	for _, header := range []string{"## Build & Test", "## Project Layout"} {
		section := extractMarkdownSection(rendered, header)
		if section != "" {
			result = markdown.ReplaceSection(result, header, section)
		}
	}

	if dryRun {
		if result != string(content) {
			fmt.Println("  section   AGENTS.md (Build & Test, Project Layout)")
		}
		return nil
	}

	if result != string(content) {
		if err := os.WriteFile(agentsPath, []byte(result), 0o644); err != nil {
			return fmt.Errorf("write AGENTS.md: %w", err)
		}
		fmt.Println("  section   AGENTS.md (Build & Test, Project Layout)")
	}

	return nil
}

func replaceOrInsertTargets(content, allRendered string, targets []string) string {
	for _, name := range targets {
		block := extractTarget(allRendered, name)
		if block == "" {
			continue
		}
		updated := mf.ReplaceTarget(content, name, block)
		if updated == content {
			content = mf.InsertTarget(content, block, ".PHONY: golangci-lint")
		} else {
			content = updated
		}
	}
	return content
}

// extractVariable extracts a variable assignment line from rendered Makefile content.
func extractVariable(content, name string) string {
	for line := range strings.SplitSeq(content, "\n") {
		if strings.HasPrefix(line, name+" ?=") || strings.HasPrefix(line, name+" =") {
			return line
		}
	}
	return ""
}

// extractTarget extracts a .PHONY target block from rendered Makefile content.
func extractTarget(content, name string) string {
	marker := ".PHONY: " + name + "\n"
	start := strings.Index(content, marker)
	if start == -1 {
		return ""
	}
	end := min(mf.FindTargetEnd(content, start+len(marker)-1), len(content))
	return strings.TrimRight(content[start:end], "\n") + "\n"
}

// extractDefine extracts a define...endef block from rendered Makefile content.
func extractDefine(content, name string) string {
	marker := "define " + name
	start := strings.Index(content, marker)
	if start == -1 {
		return ""
	}
	endMarker := "endef"
	endIdx := strings.Index(content[start:], endMarker)
	if endIdx == -1 {
		return ""
	}
	end := start + endIdx + len(endMarker)
	return content[start:end]
}

// extractSpecialTarget extracts a special target block (e.g. $(LOCALBIN):) from content.
func extractSpecialTarget(content, prefix string) string {
	start := strings.Index(content, prefix)
	if start == -1 {
		return ""
	}
	end := min(mf.FindTargetEnd(content, start+len(prefix)), len(content))
	return strings.TrimRight(content[start:end], "\n") + "\n"
}

// extractMarkdownSection extracts a ## section from rendered markdown content.
func extractMarkdownSection(content, header string) string {
	start := strings.Index(content, header)
	if start == -1 {
		return ""
	}
	afterHeader := start + len(header)
	rest := content[afterHeader:]
	lines := strings.Split(rest, "\n")
	pos := 0
	for i, line := range lines {
		if i == 0 {
			pos += len(line) + 1
			continue
		}
		if strings.HasPrefix(line, "## ") {
			return strings.TrimRight(content[start:afterHeader+pos], "\n") + "\n"
		}
		pos += len(line) + 1
	}
	return strings.TrimRight(content[start:], "\n") + "\n"
}
