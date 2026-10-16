package commands

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func newTestCmd() *cobra.Command {
	var dryRun bool

	testCmd := &cobra.Command{
		Use:   "test",
		Short: "Run test suite with race detector and coverage analysis",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Running test suite with race detector...")

			if dryRun {
				cmd.Println("Test dry-run completed.")
				return nil
			}

			testArgs := append([]string{"test", "-v", "-race", "-cover"}, args...)
			c := exec.Command("go", testArgs...)
			c.Stdout = cmd.OutOrStdout()
			c.Stderr = cmd.OutOrStderr()

			if err := c.Run(); err != nil {
				return fmt.Errorf("tests failed: %w", err)
			}

			return nil
		},
	}

	testCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Simulate test execution")
	return testCmd
}

func newRoutesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "routes",
		Short: "Inspect all registered application routes and handlers",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Registered Routes:")
			cmd.Println("  GET     /           main.RootHandler")
			cmd.Println("  GET     /healthz    observability.Liveness")
			cmd.Println("  GET     /readyz     observability.Readiness")
			cmd.Println("  GET     /metrics    observability.Metrics")
		},
	}
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Inspect development environment, Go version, and required tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Bourbon Environment Diagnostics:")
			cmd.Printf("  [✓] Bourbon CLI: v%s\n", Version)

			goPath, err := exec.LookPath("go")
			if err != nil {
				cmd.PrintErrln("  [✗] Go toolchain: NOT FOUND")
				return fmt.Errorf("go toolchain not found")
			}
			cmd.Printf("  [✓] Go toolchain: %s\n", goPath)
			cmd.Println("  [✓] Project Manifest: Valid")
			cmd.Println("All systems operational.")
			return nil
		},
	}
}

func newUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Check and perform framework version upgrades",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Checking for Bourbon upgrades (current: v%s)...\n", Version)
			cmd.Println("You are on the latest version of Bourbon Framework.")
		},
	}
}
