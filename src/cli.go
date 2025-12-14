package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "google-docs-manager",
	Short: "Google Docs Manager",
	Long:  "Comprehensive Google Docs operations: create, read, format, tables, images, and more",
}

// initCommands initializes all CLI commands and their flags
func initCommands() {
	// Initialize command flags
	initDocumentCommands()
	initFormattingCommands()
	initTableCommands()
	initImageCommands()

	// Document operations
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(getStructureCmd)

	// Content operations
	rootCmd.AddCommand(setMarkdownCmd)
	rootCmd.AddCommand(updateSectionCmd)
	rootCmd.AddCommand(insertAfterCmd)
	rootCmd.AddCommand(deleteTextCmd)

	// Formatting operations
	rootCmd.AddCommand(formatTextCmd)
	rootCmd.AddCommand(alignParagraphCmd)
	rootCmd.AddCommand(createBulletsCmd)
	rootCmd.AddCommand(createNumberedCmd)
	rootCmd.AddCommand(removeBulletsCmd)

	// Table operations
	rootCmd.AddCommand(insertTableCmd)
	rootCmd.AddCommand(updateTableCellCmd)
	rootCmd.AddCommand(styleTableCellCmd)

	// Image operations
	rootCmd.AddCommand(insertImageCmd)

	// Structure operations
	rootCmd.AddCommand(addHeaderCmd)
	rootCmd.AddCommand(addFooterCmd)
}
