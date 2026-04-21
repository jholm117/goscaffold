package scaffold

func HelmSpecs(params Params) []FileSpec {
	chart := "charts/" + params.ProjectName
	tmpl := chart + "/templates"
	return []FileSpec{
		{TemplatePath: "templates/helm/Chart.yaml.tmpl", OutputPath: chart + "/Chart.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/values.yaml.tmpl", OutputPath: chart + "/values.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/helmignore.tmpl", OutputPath: chart + "/.helmignore", Perm: 0o644},
		{TemplatePath: "templates/helm/helpers.tpl.tmpl", OutputPath: tmpl + "/_helpers.tpl", Perm: 0o644},
		{TemplatePath: "templates/helm/NOTES.txt.tmpl", OutputPath: tmpl + "/NOTES.txt", Perm: 0o644},
		{TemplatePath: "templates/helm/deployment.yaml.tmpl", OutputPath: tmpl + "/deployment.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/service.yaml.tmpl", OutputPath: tmpl + "/service.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/serviceaccount.yaml.tmpl", OutputPath: tmpl + "/serviceaccount.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/clusterrole.yaml.tmpl", OutputPath: tmpl + "/clusterrole.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/clusterrolebinding.yaml.tmpl",
			OutputPath: tmpl + "/clusterrolebinding.yaml", Perm: 0o644},
	}
}

func HelmMakefileTemplate() string {
	return "templates/helm/makefile-helm.tmpl"
}
