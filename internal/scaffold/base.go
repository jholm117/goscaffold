package scaffold

import "io/fs"

func BaseSpecs(params Params) []FileSpec {
	return []FileSpec{
		{TemplatePath: "templates/base/go.mod.tmpl", OutputPath: "go.mod", Perm: 0o644},
		{TemplatePath: "templates/base/main.go.tmpl", OutputPath: "cmd/" + params.ProjectName + "/main.go", Perm: 0o644},
		{TemplatePath: "templates/base/Makefile.tmpl", OutputPath: "Makefile", Perm: 0o644},
		{TemplatePath: "templates/base/golangci.yml.tmpl", OutputPath: ".golangci.yml", Perm: 0o644},
		{TemplatePath: "templates/base/gitignore.tmpl", OutputPath: ".gitignore", Perm: 0o644},
		{TemplatePath: "templates/base/ci.yml.tmpl", OutputPath: ".github/workflows/ci.yml", Perm: 0o644},
		{TemplatePath: "templates/base/ci-checks.sh.tmpl", OutputPath: "hack/ci-checks.sh", Perm: 0o755},
		{TemplatePath: "templates/base/pre-push.tmpl", OutputPath: ".githooks/pre-push", Perm: 0o755},
		{TemplatePath: "templates/base/agents.md.tmpl", OutputPath: "AGENTS.md", Perm: 0o644},
		{TemplatePath: "templates/base/readme.md.tmpl", OutputPath: "README.md", Perm: 0o644},
	}
}

func PlaceholderDirs() []string {
	return []string{"internal"}
}

func WritePlaceholders(targetDir string) error {
	for _, dir := range PlaceholderDirs() {
		_, err := WriteFile(targetDir+"/"+dir+"/.gitkeep", "", fs.FileMode(0o644))
		if err != nil {
			return err
		}
	}
	return nil
}
