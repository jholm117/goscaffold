package main

import (
	"fmt"
	"os"

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
			fmt.Println("init not implemented")
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
			fmt.Println("add not implemented")
			return nil
		},
	}
}
