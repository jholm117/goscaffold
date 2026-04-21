package scaffold

func ControllerSpecs(params Params) []FileSpec {
	return []FileSpec{
		{TemplatePath: "templates/controller/Dockerfile.tmpl", OutputPath: "Dockerfile", Perm: 0o644},
		{TemplatePath: "templates/controller/dockerignore.tmpl", OutputPath: ".dockerignore", Perm: 0o644},
		{TemplatePath: "templates/controller/reconciler.go.tmpl",
			OutputPath: "internal/controller/reconciler.go", Perm: 0o644},
		{TemplatePath: "templates/controller/e2e-up.sh.tmpl", OutputPath: "hack/e2e-up.sh", Perm: 0o755},
		{TemplatePath: "templates/controller/e2e-down.sh.tmpl", OutputPath: "hack/e2e-down.sh", Perm: 0o755},
		{TemplatePath: "templates/controller/kind-config.yaml.tmpl", OutputPath: "hack/kind-config.yaml", Perm: 0o644},
	}
}

func ControllerMakefileTemplate() string {
	return "templates/controller/makefile-controller.tmpl"
}
