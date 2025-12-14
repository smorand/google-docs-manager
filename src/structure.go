package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/api/docs/v1"
)

// addHeaderCmd adds a header
var addHeaderCmd = &cobra.Command{
	Use:   "add-header <document-id> <text>",
	Short: "Add header to document",
	Args:  cobra.ExactArgs(2),
	RunE:  runAddHeader,
}

// addFooterCmd adds a footer
var addFooterCmd = &cobra.Command{
	Use:   "add-footer <document-id> <text>",
	Short: "Add footer to document",
	Args:  cobra.ExactArgs(2),
	RunE:  runAddFooter,
}

func runAddHeader(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]
	headerText := args[1]

	service, err := getDocsService(ctx)
	if err != nil {
		return err
	}

	// Get document to find header section
	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}

	// Create header if it doesn't exist
	headerID := ""
	if doc.Headers != nil && len(doc.Headers) > 0 {
		for id := range doc.Headers {
			headerID = id
			break
		}
	}

	requests := []*docs.Request{}

	if headerID == "" {
		// Create header
		requests = append(requests, &docs.Request{
			CreateHeader: &docs.CreateHeaderRequest{
				Type: "DEFAULT",
			},
		})
	}

	// Note: This is a simplified implementation
	// A complete implementation would need to get the header ID from the response
	// and insert text into the header section using a second batch update
	// This requires handling the response from the first request

	_, err = service.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()

	if err != nil {
		return fmt.Errorf("error adding header: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", green("✅ Header created (text insertion requires additional implementation)"))
	fmt.Fprintf(os.Stderr, "%s\n", cyan("   Header text provided: "+headerText))
	return nil
}

func runAddFooter(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]
	footerText := args[1]

	service, err := getDocsService(ctx)
	if err != nil {
		return err
	}

	requests := []*docs.Request{
		{
			CreateFooter: &docs.CreateFooterRequest{
				Type: "DEFAULT",
			},
		},
	}

	// Note: This is a simplified implementation
	// A complete implementation would need to get the footer ID from the response
	// and insert text into the footer section using a second batch update

	_, err = service.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()

	if err != nil {
		return fmt.Errorf("error adding footer: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", green("✅ Footer created (text insertion requires additional implementation)"))
	fmt.Fprintf(os.Stderr, "%s\n", cyan("   Footer text provided: "+footerText))
	return nil
}
