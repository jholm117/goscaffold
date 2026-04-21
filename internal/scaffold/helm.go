package scaffold

func HelmSpecs(params Params) []FileSpec {
	return []FileSpec{
		{TemplatePath: "templates/helm/Chart.yaml.tmpl", OutputPath: "charts/" + params.ProjectName + "/Chart.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/values.yaml.tmpl", OutputPath: "charts/" + params.ProjectName + "/values.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/helmignore.tmpl", OutputPath: "charts/" + params.ProjectName + "/.helmignore", Perm: 0o644},
		{TemplatePath: "templates/helm/helpers.tpl.tmpl", OutputPath: "charts/" + params.ProjectName + "/templates/_helpers.tpl", Perm: 0o644},
		{TemplatePath: "templates/helm/NOTES.txt.tmpl", OutputPath: "charts/" + params.ProjectName + "/templates/NOTES.txt", Perm: 0o644},
		{TemplatePath: "templates/helm/deployment.yaml.tmpl", OutputPath: "charts/" + params.ProjectName + "/templates/deployment.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/service.yaml.tmpl", OutputPath: "charts/" + params.ProjectName + "/templates/service.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/serviceaccount.yaml.tmpl", OutputPath: "charts/" + params.ProjectName + "/templates/serviceaccount.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/clusterrole.yaml.tmpl", OutputPath: "charts/" + params.ProjectName + "/templates/clusterrole.yaml", Perm: 0o644},
		{TemplatePath: "templates/helm/clusterrolebinding.yaml.tmpl", OutputPath: "charts/" + params.ProjectName + "/templates/clusterrolebinding.yaml", Perm: 0o644},
	}
}

func HelmMakefileTemplate() string {
	return "templates/helm/makefile-helm.tmpl"
}
