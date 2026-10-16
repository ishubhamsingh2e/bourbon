package commands

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func newBuildCmd() *cobra.Command {
	var dryRun bool
	var output string

	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Compile production binary with embedded assets and optimizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Building Bourbon production binary...")

			if dryRun {
				cmd.Println("Build dry-run completed.")
				return nil
			}

			buildArgs := []string{"build", "-ldflags=-s -w", "-o", output, "."}
			c := exec.Command("go", buildArgs...)
			c.Stdout = cmd.OutOrStdout()
			c.Stderr = cmd.OutOrStderr()

			if err := c.Run(); err != nil {
				return fmt.Errorf("error building binary: %w", err)
			}

			cmd.Printf("Production binary compiled successfully: %s\n", output)
			return nil
		},
	}

	buildCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate build without compiling")
	buildCmd.Flags().StringVarP(&output, "output", "o", "bin/app", "Target executable path")
	return buildCmd
}
