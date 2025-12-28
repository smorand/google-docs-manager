package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"google-docs-manager/internal/auth"
	"google-docs-manager/internal/conversion"

	"github.com/spf13/cobra"
	"google.golang.org/api/docs/v1"
)

var (
	alignParagraphCmd = &cobra.Command{
		Args:  cobra.ExactArgs(4),
		RunE:  runAlignParagraph,
		Short: "Align paragraph (START, CENTER, END, JUSTIFIED)",
		Use:   "align-paragraph <document-id> <start-index> <end-index> <alignment>",
	}

	createBulletsCmd = &cobra.Command{
		Args:  cobra.ExactArgs(3),
		RunE:  runCreateBullets,
		Short: "Create bulleted list",
		Use:   "create-bullets <document-id> <start-index> <end-index>",
	}

	createNumberedCmd = &cobra.Command{
		Args:  cobra.ExactArgs(3),
		RunE:  runCreateNumbered,
		Short: "Create numbered list",
		Use:   "create-numbered <document-id> <start-index> <end-index>",
	}

	formatTextCmd = &cobra.Command{
		Args:  cobra.ExactArgs(3),
		RunE:  runFormatText,
		Short: "Format text (bold, italic, underline, color, size)",
		Use:   "format-text <document-id> <start-index> <end-index>",
	}

	removeBulletsCmd = &cobra.Command{
		Args:  cobra.ExactArgs(3),
		RunE:  runRemoveBullets,
		Short: "Remove bullets/numbering from list",
		Use:   "remove-bullets <document-id> <start-index> <end-index>",
	}
)

func initFormattingCommands() {
	formatTextCmd.Flags().Bool("bold", false, "Make text bold")
	formatTextCmd.Flags().Bool("italic", false, "Make text italic")
	formatTextCmd.Flags().Bool("underline", false, "Underline text")
	formatTextCmd.Flags().String("color", "", "Text color (hex, e.g., #FF0000)")
	formatTextCmd.Flags().Float64("size", 0, "Font size in points")
}

func runAlignParagraph(cmd *cobra.Command, args []string) error {
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

	alignment := strings.ToUpper(args[3])
	validAlignments := map[string]bool{
		"CENTER":    true,
		"END":       true,
		"JUSTIFIED": true,
		"START":     true,
	}

	if !validAlignments[alignment] {
		return fmt.Errorf("invalid alignment: %s (must be START, CENTER, END, or JUSTIFIED)", alignment)
	}

	service, err := auth.GetDocsService(ctx)
	if err != nil {
		return err
	}

	requests := []*docs.Request{
		{
			UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
				Fields: "alignment",
				ParagraphStyle: &docs.ParagraphStyle{
					Alignment: alignment,
				},
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
		return fmt.Errorf("error aligning paragraph: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", green("✅ Paragraph aligned to "+alignment))
	return nil
}

func runCreateBullets(cmd *cobra.Command, args []string) error {
	return createList(args, "BULLET")
}

func runCreateNumbered(cmd *cobra.Command, args []string) error {
	return createList(args, "NUMBER")
}

func createList(args []string, listType string) error {
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
			CreateParagraphBullets: &docs.CreateParagraphBulletsRequest{
				BulletPreset: listType + "_DISC_CIRCLE_SQUARE",
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
		return fmt.Errorf("error creating list: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", green("✅ "+listType+" list created"))
	return nil
}

func runFormatText(cmd *cobra.Command, args []string) error {
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

	textStyle := &docs.TextStyle{}
	fields := []string{}

	bold, _ := cmd.Flags().GetBool("bold")
	if bold {
		textStyle.Bold = true
		fields = append(fields, "bold")
	}

	italic, _ := cmd.Flags().GetBool("italic")
	if italic {
		textStyle.Italic = true
		fields = append(fields, "italic")
	}

	underline, _ := cmd.Flags().GetBool("underline")
	if underline {
		textStyle.Underline = true
		fields = append(fields, "underline")
	}

	textColor, _ := cmd.Flags().GetString("color")
	if textColor != "" {
		textStyle.ForegroundColor = conversion.ParseColor(textColor)
		fields = append(fields, "foregroundColor")
	}

	fontSize, _ := cmd.Flags().GetFloat64("size")
	if fontSize > 0 {
		textStyle.FontSize = &docs.Dimension{
			Magnitude: fontSize,
			Unit:      "PT",
		}
		fields = append(fields, "fontSize")
	}

	if len(fields) == 0 {
		return fmt.Errorf("no formatting options specified")
	}

	requests := []*docs.Request{
		{
			UpdateTextStyle: &docs.UpdateTextStyleRequest{
				Fields: strings.Join(fields, ","),
				Range: &docs.Range{
					EndIndex:   endIndex,
					StartIndex: startIndex,
				},
				TextStyle: textStyle,
			},
		},
	}

	_, err = service.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()

	if err != nil {
		return fmt.Errorf("error formatting text: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", green("✅ Text formatted"))
	return nil
}

func runRemoveBullets(cmd *cobra.Command, args []string) error {
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
			DeleteParagraphBullets: &docs.DeleteParagraphBulletsRequest{
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
		return fmt.Errorf("error removing bullets: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", green("✅ Bullets/numbering removed"))
	return nil
}
