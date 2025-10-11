package cli

import "fmt"

const version = "v0.1.0-dev"

// GetVersion returns the current application version
func GetVersion() string {
	return version
}

// PrintHelp displays the help message
func PrintHelp() {
	fmt.Println("Slacky - A terminal client for Slack")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  slacky           Start the Slack client")
	fmt.Println("  slacky init      Run setup wizard to create config file")
	fmt.Println("  slacky help      Show this help message")
	fmt.Println("  slacky version   Show version information")
	fmt.Println()
}

// PrintVersion displays version information
func PrintVersion() {
	fmt.Printf("Slacky %s\n", GetVersion())
}
