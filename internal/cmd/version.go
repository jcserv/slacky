package cmd

import (
	"github.com/jcserv/slacky/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display the current version of Slacky.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("Slacky %s\n", version.Version)
	},
}
