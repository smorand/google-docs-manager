package cli

import (
	"context"
	"fmt"
	"os"

	"google-docs-manager/internal/auth"

	"github.com/spf13/cobra"
	"google.golang.org/api/docs/v1"
)

var (
	addFooterCmd = &cobra.Command{
		Args:  cobra.ExactArgs(2),
		RunE:  runAddFooter,
		Short: "Add footer to document",
		Use:   "add-footer <document-id> <text>",
	}

	addHeaderCmd = &cobra.Command{
		Args:  cobra.ExactArgs(2),
		RunE:  runAddHeader,
		Short: "Add header to document",
		Use:   "add-header <document-id> <text>",
	}
)

func runAddFooter(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]
	footerText := args[1]

	service, err := auth.GetDocsService(ctx)
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

func runAddHeader(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]
	headerText := args[1]

	service, err := auth.GetDocsService(ctx)
	if err != nil {
		return err
	}

	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}

	headerID := ""
	if doc.Headers != nil && len(doc.Headers) > 0 {
		for id := range doc.Headers {
			headerID = id
			break
		}
	}

	requests := []*docs.Request{}

	if headerID == "" {
		requests = append(requests, &docs.Request{
			CreateHeader: &docs.CreateHeaderRequest{
				Type: "DEFAULT",
			},
		})
	}

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
