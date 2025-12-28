package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"google-docs-manager/internal/auth"
	"google-docs-manager/internal/conversion"
	"google-docs-manager/internal/document"

	"github.com/spf13/cobra"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
)

var (
	copyCmd = &cobra.Command{
		Args:  cobra.ExactArgs(2),
		RunE:  runCopy,
		Short: "Copy an existing Google Doc to create a new document",
		Use:   "copy <source-document-id> <new-title>",
	}

	createCmd = &cobra.Command{
		Args:  cobra.ExactArgs(1),
		RunE:  runCreate,
		Short: "Create a new Google Doc",
		Use:   "create <title>",
	}

	getStructureCmd = &cobra.Command{
		Args:  cobra.ExactArgs(1),
		RunE:  runGetStructure,
		Short: "Get document structure (headings)",
		Use:   "get-structure <document-id>",
	}

	infoCmd = &cobra.Command{
		Args:  cobra.ExactArgs(1),
		RunE:  runInfo,
		Short: "Get document information",
		Use:   "info <document-id>",
	}

	readCmd = &cobra.Command{
		Args:  cobra.ExactArgs(1),
		RunE:  runRead,
		Short: "Read a document and output as markdown",
		Use:   "read <document-id>",
	}
)

func initDocumentCommands() {
	copyCmd.Flags().String("folder", "", "Folder ID to place the copied document in")
	createCmd.Flags().String("folder", "", "Folder ID to create document in")
}

func runCopy(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	sourceDocID := args[0]
	newTitle := args[1]

	driveService, err := auth.GetDriveService(ctx)
	if err != nil {
		return err
	}

	copyMetadata := &drive.File{
		Name: newTitle,
	}

	folderID, _ := cmd.Flags().GetString("folder")
	if folderID != "" {
		copyMetadata.Parents = []string{folderID}
	}

	copiedFile, err := driveService.Files.Copy(sourceDocID, copyMetadata).Do()
	if err != nil {
		return fmt.Errorf("error copying document: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", green("✅ Document copied: "+copiedFile.Name))
	fmt.Fprintf(os.Stderr, "%s\n", green("   ID: "+copiedFile.Id))
	fmt.Println(copiedFile.Id)

	return nil
}

func runCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	title := args[0]

	service, err := auth.GetDocsService(ctx)
	if err != nil {
		return err
	}

	doc := &docs.Document{
		Title: title,
	}

	result, err := service.Documents.Create(doc).Do()
	if err != nil {
		return fmt.Errorf("error creating document: %w", err)
	}

	folderID, _ := cmd.Flags().GetString("folder")
	if folderID != "" {
		driveService, err := auth.GetDriveService(ctx)
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

func runGetStructure(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]

	service, err := auth.GetDocsService(ctx)
	if err != nil {
		return err
	}

	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}

	sections := document.GetStructure(doc)
	return printJSON(sections)
}

func runInfo(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]

	service, err := auth.GetDocsService(ctx)
	if err != nil {
		return err
	}

	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}

	info := map[string]interface{}{
		"documentId":          doc.DocumentId,
		"revisionId":          doc.RevisionId,
		"suggestionsViewMode": doc.SuggestionsViewMode,
		"title":               doc.Title,
	}

	return printJSON(info)
}

func runRead(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]

	service, err := auth.GetDocsService(ctx)
	if err != nil {
		return err
	}

	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error reading document: %w", err)
	}

	markdown := conversion.DocsToMarkdown(doc)
	fmt.Println(markdown)

	return nil
}

func printJSON(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
