package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"google.golang.org/api/docs/v1"
)

// setMarkdownCmd sets document content from markdown
var setMarkdownCmd = &cobra.Command{
	Use:   "set-markdown <document-id> <markdown-file>",
	Short: "Set document content from markdown file",
	Args:  cobra.ExactArgs(2),
	RunE:  runSetMarkdown,
}

// updateSectionCmd updates a specific section
var updateSectionCmd = &cobra.Command{
	Use:   "update-section <document-id> <section-name> <markdown-file>",
	Short: "Update a specific section with markdown content",
	Args:  cobra.ExactArgs(3),
	RunE:  runUpdateSection,
}

// insertAfterCmd inserts content after a section
var insertAfterCmd = &cobra.Command{
	Use:   "insert-after <document-id> <section-name> <text>",
	Short: "Insert text after a section",
	Args:  cobra.ExactArgs(3),
	RunE:  runInsertAfter,
}

// deleteTextCmd deletes text in a range
var deleteTextCmd = &cobra.Command{
	Use:   "delete-text <document-id> <start-index> <end-index>",
	Short: "Delete text in a range",
	Args:  cobra.ExactArgs(3),
	RunE:  runDeleteText,
}

func runSetMarkdown(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]
	markdownFile := args[1]

	// Read markdown file
	content, err := os.ReadFile(markdownFile)
	if err != nil {
		return fmt.Errorf("error reading markdown file: %w", err)
	}

	service, err := getDocsService(ctx)
	if err != nil {
		return err
	}

	// Get current document to find end index
	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}

	endIndex := doc.Body.Content[len(doc.Body.Content)-1].EndIndex

	// Clear existing content (except title)
	requests := []*docs.Request{
		{
			DeleteContentRange: &docs.DeleteContentRangeRequest{
				Range: &docs.Range{
					StartIndex: 1,
					EndIndex:   endIndex - 1,
				},
			},
		},
	}

	// Convert markdown to requests
	markdownRequests := markdownToDocsRequests(string(content), 1)
	requests = append(requests, markdownRequests...)

	// Execute batch update
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

	// Read markdown file
	content, err := os.ReadFile(markdownFile)
	if err != nil {
		return fmt.Errorf("error reading markdown file: %w", err)
	}

	service, err := getDocsService(ctx)
	if err != nil {
		return err
	}

	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}

	section := findSection(doc, sectionName)
	if section == nil {
		return fmt.Errorf("section not found: %s", sectionName)
	}

	// Delete existing section content
	requests := []*docs.Request{
		{
			DeleteContentRange: &docs.DeleteContentRangeRequest{
				Range: &docs.Range{
					StartIndex: section.EndIndex,
					EndIndex:   section.EndIndex + 1,
				},
			},
		},
	}

	// Insert new content
	markdownRequests := markdownToDocsRequests(string(content), section.EndIndex)
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

func runInsertAfter(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]
	sectionName := args[1]
	text := args[2]

	service, err := getDocsService(ctx)
	if err != nil {
		return err
	}

	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}

	section := findSection(doc, sectionName)
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

	service, err := getDocsService(ctx)
	if err != nil {
		return err
	}

	requests := []*docs.Request{
		{
			DeleteContentRange: &docs.DeleteContentRangeRequest{
				Range: &docs.Range{
					StartIndex: startIndex,
					EndIndex:   endIndex,
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
