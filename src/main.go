package main

import (
	"fmt"
	"os"
)

func main() {
	// Initialize all commands and flags
	initCommands()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
