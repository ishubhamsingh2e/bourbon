package commands

import (
	"fmt"

	"github.com/ishubhamsingh2e/bourbon/cmd/bourbon/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Bourbon developer configuration and remembered defaults",
	}

	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Clear all remembered 'bourbon new' choices on this machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.ResetPreferences(); err != nil {
				return fmt.Errorf("failed resetting preferences: %w", err)
			}
			cmd.Println("Successfully reset all remembered developer choices.")
			return nil
		},
	}

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Display current remembered developer choices",
		RunE: func(cmd *cobra.Command, args []string) error {
			prefs, err := config.LoadPreferences()
			if err != nil {
				return fmt.Errorf("failed loading preferences: %w", err)
			}
			cmd.Printf("Remembered Preferences (%s):\n", config.GetPreferencesPath())
			cmd.Printf("  Profile:       %s\n", prefs.Profile)
			cmd.Printf("  Structure:     %s\n", prefs.Structure)
			cmd.Printf("  Database:      %s\n", prefs.Database)
			cmd.Printf("  Cache:         %s\n", prefs.Cache)
			cmd.Printf("  Auth:          %s\n", prefs.Auth)
			cmd.Printf("  Capabilities:  %v\n", prefs.Capabilities)
			return nil
		},
	}

	configCmd.AddCommand(resetCmd, showCmd)
	return configCmd
}
