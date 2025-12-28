package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"google-docs-manager/internal/auth"
	"google-docs-manager/internal/conversion"
	"google-docs-manager/internal/document"

	"github.com/spf13/cobra"
	"google.golang.org/api/docs/v1"
)

var (
	deleteTextCmd = &cobra.Command{
		Args:  cobra.ExactArgs(3),
		RunE:  runDeleteText,
		Short: "Delete text in a range",
		Use:   "delete-text <document-id> <start-index> <end-index>",
	}

	insertAfterCmd = &cobra.Command{
		Args:  cobra.ExactArgs(3),
		RunE:  runInsertAfter,
		Short: "Insert text after a section",
		Use:   "insert-after <document-id> <section-name> <text>",
	}

	setMarkdownCmd = &cobra.Command{
		Args:  cobra.ExactArgs(2),
		RunE:  runSetMarkdown,
		Short: "Set document content from markdown file",
		Use:   "set-markdown <document-id> <markdown-file>",
	}

	updateSectionCmd = &cobra.Command{
		Args:  cobra.ExactArgs(3),
		RunE:  runUpdateSection,
		Short: "Update a specific section with markdown content",
		Use:   "update-section <document-id> <section-name> <markdown-file>",
	}
)

func runDeleteText(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]

	startIndex, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid start index: %w", err)
	}

	endIndex, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid end index: %w", err)
	}

	service, err := auth.GetDocsService(ctx)
	if err != nil {
		return err
	}

	requests := []*docs.Request{
		{
			DeleteContentRange: &docs.DeleteContentRangeRequest{
				Range: &docs.Range{
					EndIndex:   endIndex,
					StartIndex: startIndex,
				},
			},
		},
	}

	_, err = service.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()

	if err != nil {
		return fmt.Errorf("error deleting text: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", green(fmt.Sprintf("✅ Text deleted from %d to %d", startIndex, endIndex)))
	return nil
}

func runInsertAfter(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]
	sectionName := args[1]
	text := args[2]

	service, err := auth.GetDocsService(ctx)
	if err != nil {
		return err
	}

	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}

	section := document.FindSection(doc, sectionName)
	if section == nil {
		return fmt.Errorf("section not found: %s", sectionName)
	}

	requests := []*docs.Request{
		{
			InsertText: &docs.InsertTextRequest{
				Location: &docs.Location{Index: section.EndIndex},
				Text:     "\n" + text + "\n",
			},
		},
	}

	_, err = service.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()

	if err != nil {
		return fmt.Errorf("error inserting text: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", green("✅ Text inserted after section '"+sectionName+"'"))
	return nil
}

func runSetMarkdown(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]
	markdownFile := args[1]

	content, err := os.ReadFile(markdownFile)
	if err != nil {
		return fmt.Errorf("error reading markdown file: %w", err)
	}

	service, err := auth.GetDocsService(ctx)
	if err != nil {
		return err
	}

	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}

	endIndex := doc.Body.Content[len(doc.Body.Content)-1].EndIndex

	requests := []*docs.Request{}
	if endIndex > 2 {
		requests = append(requests, &docs.Request{
			DeleteContentRange: &docs.DeleteContentRangeRequest{
				Range: &docs.Range{
					EndIndex:   endIndex - 1,
					StartIndex: 1,
				},
			},
		})
	}

	markdownRequests := conversion.MarkdownToDocsRequests(string(content), 1)
	requests = append(requests, markdownRequests...)

	_, err = service.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()

	if err != nil {
		return fmt.Errorf("error updating document: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", green("✅ Document content updated from markdown"))
	return nil
}

func runUpdateSection(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]
	sectionName := args[1]
	markdownFile := args[2]

	content, err := os.ReadFile(markdownFile)
	if err != nil {
		return fmt.Errorf("error reading markdown file: %w", err)
	}

	service, err := auth.GetDocsService(ctx)
	if err != nil {
		return err
	}

	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}

	section := document.FindSection(doc, sectionName)
	if section == nil {
		return fmt.Errorf("section not found: %s", sectionName)
	}

	requests := []*docs.Request{
		{
			DeleteContentRange: &docs.DeleteContentRangeRequest{
				Range: &docs.Range{
					EndIndex:   section.EndIndex + 1,
					StartIndex: section.EndIndex,
				},
			},
		},
	}

	markdownRequests := conversion.MarkdownToDocsRequests(string(content), section.EndIndex)
	requests = append(requests, markdownRequests...)

	_, err = service.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()

	if err != nil {
		return fmt.Errorf("error updating section: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", green("✅ Section '"+sectionName+"' updated"))
	return nil
}
