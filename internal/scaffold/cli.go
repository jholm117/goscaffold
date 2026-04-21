package scaffold

func CLISpecs() []FileSpec {
	return []FileSpec{
		{TemplatePath: "templates/cli/goreleaser.yaml.tmpl", OutputPath: ".goreleaser.yaml", Perm: 0o644},
		{TemplatePath: "templates/cli/release.yml.tmpl", OutputPath: ".github/workflows/release.yml", Perm: 0o644},
	}
}

func CLIMakefileTemplate() string {
	return "templates/cli/makefile-cli.tmpl"
}
