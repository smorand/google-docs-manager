package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
)

// createCmd creates a new document
var createCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new Google Doc",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

// readCmd reads a document and outputs as markdown
var readCmd = &cobra.Command{
	Use:   "read <document-id>",
	Short: "Read a document and output as markdown",
	Args:  cobra.ExactArgs(1),
	RunE:  runRead,
}

// infoCmd gets document information
var infoCmd = &cobra.Command{
	Use:   "info <document-id>",
	Short: "Get document information",
	Args:  cobra.ExactArgs(1),
	RunE:  runInfo,
}

// getStructureCmd gets document structure
var getStructureCmd = &cobra.Command{
	Use:   "get-structure <document-id>",
	Short: "Get document structure (headings)",
	Args:  cobra.ExactArgs(1),
	RunE:  runGetStructure,
}

func initDocumentCommands() {
	createCmd.Flags().String("folder", "", "Folder ID to create document in")
}

func runCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	title := args[0]

	service, err := getDocsService(ctx)
	if err != nil {
		return err
	}

	// Create document
	doc := &docs.Document{
		Title: title,
	}

	result, err := service.Documents.Create(doc).Do()
	if err != nil {
		return fmt.Errorf("error creating document: %w", err)
	}

	// Move to folder if specified
	folderID, _ := cmd.Flags().GetString("folder")
	if folderID != "" {
		driveService, err := getDriveService(ctx)
		if err != nil {
			return err
		}

		_, err = driveService.Files.Update(result.DocumentId, &drive.File{}).AddParents(folderID).Do()
		if err != nil {
			return fmt.Errorf("error moving to folder: %w", err)
		}
	}

	fmt.Fprintf(os.Stderr, "%s\n", green("✅ Document created: "+result.Title))
	fmt.Fprintf(os.Stderr, "%s\n", green("   ID: "+result.DocumentId))
	fmt.Println(result.DocumentId)

	return nil
}

func runRead(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]

	service, err := getDocsService(ctx)
	if err != nil {
		return err
	}

	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error reading document: %w", err)
	}

	markdown := docsToMarkdown(doc)
	fmt.Println(markdown)

	return nil
}

func runInfo(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]

	service, err := getDocsService(ctx)
	if err != nil {
		return err
	}

	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}

	info := map[string]interface{}{
		"title":               doc.Title,
		"documentId":          doc.DocumentId,
		"revisionId":          doc.RevisionId,
		"suggestionsViewMode": doc.SuggestionsViewMode,
	}

	return printJSON(info)
}

func runGetStructure(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]

	service, err := getDocsService(ctx)
	if err != nil {
		return err
	}

	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}

	sections := getDocumentStructure(doc)
	return printJSON(sections)
}

func printJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
