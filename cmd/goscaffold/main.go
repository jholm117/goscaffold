package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jholm117/goscaffold/internal/config"
	"github.com/jholm117/goscaffold/internal/scaffold"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	root := &cobra.Command{
		Use:     "goscaffold",
		Short:   "Scaffold Go projects with composable layers",
		Version: version,
	}

	root.AddCommand(newInitCmd())
	root.AddCommand(newAddCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <project-name>",
		Short: "Create a new Go project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			cli, _ := cmd.Flags().GetBool("cli")
			controller, _ := cmd.Flags().GetBool("controller")
			helm, _ := cmd.Flags().GetBool("helm")

			if !cli && !controller && !helm {
				return fmt.Errorf("at least one of --cli, --controller, or --helm is required")
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			module, _ := cmd.Flags().GetString("module")
			if module == "" {
				prefix := cfg.ModulePrefix
				if prefix == "" {
					return fmt.Errorf("--module is required (or set module_prefix in ~/.config/goscaffold/config.yaml)")
				}
				module = prefix + "/" + name
			}

			goVersion, err := detectGoVersion()
			if err != nil {
				return fmt.Errorf("detect Go version: %w", err)
			}

			params := scaffold.Params{
				ProjectName: name,
				Module:      module,
				GoVersion:   goVersion,
				CLI:         cli,
				Controller:  controller,
				Helm:        helm,
			}

			targetDir := filepath.Join(".", name)
			if err := scaffold.Init(targetDir, params); err != nil {
				return err
			}

			fmt.Printf("\nProject %s created successfully.\n", name)
			fmt.Println("Next steps:")
			fmt.Printf("  cd %s\n", name)
			fmt.Println("  git init && git add -A && git commit -m 'Initial commit'")
			fmt.Println("  make setup-hooks")
			return nil
		},
	}
	cmd.Flags().String("module", "", "Go module path (default: <module_prefix>/<project-name>)")
	cmd.Flags().Bool("cli", false, "Include CLI distribution layer")
	cmd.Flags().Bool("controller", false, "Include K8s controller layer")
	cmd.Flags().Bool("helm", false, "Include Helm chart layer")
	return cmd
}

func newAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "add <layer>",
		Short:     "Add a layer to an existing project",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"cli", "controller", "helm"},
		RunE: func(cmd *cobra.Command, args []string) error {
			layer := args[0]

			params, err := scaffold.DetectProject(".")
			if err != nil {
				return err
			}

			switch layer {
			case "cli":
				params.CLI = true
			case "controller":
				params.Controller = true
			case "helm":
				params.Helm = true
			}

			return scaffold.Add(".", layer, params)
		},
	}
}

func detectGoVersion() (string, error) {
	out, err := exec.Command("go", "env", "GOVERSION").Output()
	if err != nil {
		return "", err
	}
	v := strings.TrimSpace(string(out))
	return strings.TrimPrefix(v, "go"), nil
}
