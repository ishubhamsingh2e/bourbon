// Package commands implements all Bourbon CLI subcommands.
package commands

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

const Version = "0.1.0"

// NewRootCmd constructs the primary Bourbon CLI root command.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "bourbon",
		Short:        "Bourbon CLI - The High-Performance Go Full-Stack Framework",
		Version:      Version,
		SilenceUsage: true,
	}

	rootCmd.AddCommand(
		newNewCmd(),
		newMakeCmd(),
		newMigrateCmd(),
		newDevCmd(),
		newBuildCmd(),
		newTestCmd(),
		newRoutesCmd(),
		newDoctorCmd(),
		newUpgradeCmd(),
		newConfigCmd(),
	)

	return rootCmd
}

// Execute runs the root command with default OS arguments and std streams.
func Execute() error {
	return NewRootCmd().Execute()
}

// Runner provides a testing harness for executing CLI commands programmatically.
type Runner struct {
	stdout io.Writer
	stderr io.Writer
}

// NewRunner creates a command runner with specified stdout and stderr.
func NewRunner(stdout, stderr io.Writer) *Runner {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	return &Runner{
		stdout: stdout,
		stderr: stderr,
	}
}

// Run executes the CLI with the provided arguments and returns an exit code.
func (r *Runner) Run(args []string) int {
	cmd := NewRootCmd()
	cmd.SetArgs(args)
	cmd.SetOut(r.stdout)
	cmd.SetErr(r.stderr)

	if err := cmd.Execute(); err != nil {
		return 1
	}
	return 0
}
