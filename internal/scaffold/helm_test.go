package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelmSpecs_RenderAll(t *testing.T) {
	params := Params{
		ProjectName: "myapp",
		Module:      "github.com/test/myapp",
		Helm:        true,
	}

	dir := t.TempDir()
	if err := RenderLayer("helm", dir, params, HelmSpecs(params)); err != nil {
		t.Fatalf("RenderLayer helm: %v", err)
	}

	wantFiles := []string{
		"charts/myapp/Chart.yaml",
		"charts/myapp/values.yaml",
		"charts/myapp/.helmignore",
		"charts/myapp/templates/_helpers.tpl",
		"charts/myapp/templates/NOTES.txt",
		"charts/myapp/templates/deployment.yaml",
		"charts/myapp/templates/service.yaml",
		"charts/myapp/templates/serviceaccount.yaml",
		"charts/myapp/templates/clusterrole.yaml",
		"charts/myapp/templates/clusterrolebinding.yaml",
	}
	for _, f := range wantFiles {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("missing file %s: %v", f, err)
		}
	}

	chart, _ := os.ReadFile(filepath.Join(dir, "charts/myapp/Chart.yaml"))
	if !strings.Contains(string(chart), "name: myapp") {
		t.Error("Chart.yaml should contain project name")
	}

	deployment, _ := os.ReadFile(filepath.Join(dir, "charts/myapp/templates/deployment.yaml"))
	if !strings.Contains(string(deployment), "runAsNonRoot: true") {
		t.Error("deployment should have hardened security context")
	}
	if !strings.Contains(string(deployment), "http-metrics") {
		t.Error("deployment should expose http-metrics port")
	}

	helpers, _ := os.ReadFile(filepath.Join(dir, "charts/myapp/templates/_helpers.tpl"))
	if !strings.Contains(string(helpers), "myapp.fullname") {
		t.Error("helpers should define fullname template")
	}
}
