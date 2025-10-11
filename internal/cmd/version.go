package cmd

import (
	"github.com/spf13/cobra"
)

const version = "v0.1.0-dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display the current version of Slacky.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("Slacky %s\n", version)
	},
}
