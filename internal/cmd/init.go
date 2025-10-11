package cmd

import (
	"fmt"

	_init "github.com/jcserv/slacky/internal/app/init"
	"github.com/jcserv/slacky/internal/version"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Run setup wizard to create config file",
	Long: `Run the interactive setup wizard to configure Slacky.
This will guide you through setting up your Slack bot token and app token.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := runInitWizard()
		if err != nil {
			return fmt.Errorf("init failed: %w", err)
		}

		// If init completed successfully, optionally continue to main app
		if result.ShouldContinue {
			fmt.Println("Setup complete! Starting Slacky...")
			fmt.Println()
			// Note: We could call the main app here, but for now just exit
			// The user can run 'slacky' to start the app
		}

		return nil
	},
}

// runInitWizard runs the initialization wizard
func runInitWizard() (*InitResult, error) {
	// Call the internal/init package
	result, err := _init.InitWithVersion(version.Version)
	if err != nil {
		return nil, err
	}
	return &InitResult{
		ShouldContinue: result.ShouldContinue,
	}, nil
}

// InitResult represents the result of the initialization process
type InitResult struct {
	ShouldContinue bool
}
