package cli

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	green = color.New(color.FgGreen).SprintFunc()
	cyan  = color.New(color.FgCyan).SprintFunc()
	red   = color.New(color.FgRed).SprintFunc()
)

var rootCmd = &cobra.Command{
	Long:  "Comprehensive Google Docs operations: create, read, format, tables, images, and more",
	Short: "Google Docs Manager",
	Use:   "google-docs-manager",
}

// Execute runs the root command
func Execute() error {
	initCommands()
	return rootCmd.Execute()
}

func initCommands() {
	initDocumentCommands()
	initFormattingCommands()
	initImageCommands()
	initTableCommands()

	// Document operations
	rootCmd.AddCommand(alignParagraphCmd)
	rootCmd.AddCommand(copyCmd)
	rootCmd.AddCommand(createBulletsCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(createNumberedCmd)
	rootCmd.AddCommand(deleteTextCmd)
	rootCmd.AddCommand(formatTextCmd)
	rootCmd.AddCommand(getStructureCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(insertAfterCmd)
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(removeBulletsCmd)
	rootCmd.AddCommand(setMarkdownCmd)
	rootCmd.AddCommand(updateSectionCmd)

	// Structure operations
	rootCmd.AddCommand(addFooterCmd)
	rootCmd.AddCommand(addHeaderCmd)

	// Image operations
	rootCmd.AddCommand(insertImageCmd)

	// Table operations
	rootCmd.AddCommand(insertTableCmd)
	rootCmd.AddCommand(styleTableCellCmd)
	rootCmd.AddCommand(updateTableCellCmd)
}
