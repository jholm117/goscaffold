package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	root := &cobra.Command{
		Use:     "/tmp/goscaffold-smoke-test",
		Short:   "/tmp/goscaffold-smoke-test does things",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("hello from /tmp/goscaffold-smoke-test")
			return nil
		},
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
