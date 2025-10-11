package main

import (
	"fmt"
	"os"

	"github.com/jcserv/slacky/internal/cmd"
	"github.com/jcserv/slacky/internal/i18n"
)

func main() {
	// Initialize i18n bundle
	if err := i18n.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing i18n: %v\n", err)
		os.Exit(1)
	}

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
