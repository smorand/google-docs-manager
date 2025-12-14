package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/api/docs/v1"
)

// formatTextCmd formats text with bold, italic, underline, color, size
var formatTextCmd = &cobra.Command{
	Use:   "format-text <document-id> <start-index> <end-index>",
	Short: "Format text (bold, italic, underline, color, size)",
	Args:  cobra.ExactArgs(3),
	RunE:  runFormatText,
}

// alignParagraphCmd aligns a paragraph
var alignParagraphCmd = &cobra.Command{
	Use:   "align-paragraph <document-id> <start-index> <end-index> <alignment>",
	Short: "Align paragraph (START, CENTER, END, JUSTIFIED)",
	Args:  cobra.ExactArgs(4),
	RunE:  runAlignParagraph,
}

// createBulletsCmd creates a bulleted list
var createBulletsCmd = &cobra.Command{
	Use:   "create-bullets <document-id> <start-index> <end-index>",
	Short: "Create bulleted list",
	Args:  cobra.ExactArgs(3),
	RunE:  runCreateBullets,
}

// createNumberedCmd creates a numbered list
var createNumberedCmd = &cobra.Command{
	Use:   "create-numbered <document-id> <start-index> <end-index>",
	Short: "Create numbered list",
	Args:  cobra.ExactArgs(3),
	RunE:  runCreateNumbered,
}

// removeBulletsCmd removes bullets/numbering from a list
var removeBulletsCmd = &cobra.Command{
	Use:   "remove-bullets <document-id> <start-index> <end-index>",
	Short: "Remove bullets/numbering from list",
	Args:  cobra.ExactArgs(3),
	RunE:  runRemoveBullets,
}

func initFormattingCommands() {
	formatTextCmd.Flags().Bool("bold", false, "Make text bold")
	formatTextCmd.Flags().Bool("italic", false, "Make text italic")
	formatTextCmd.Flags().Bool("underline", false, "Underline text")
	formatTextCmd.Flags().String("color", "", "Text color (hex, e.g., #FF0000)")
	formatTextCmd.Flags().Float64("size", 0, "Font size in points")
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

	service, err := getDocsService(ctx)
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
		textStyle.ForegroundColor = parseColor(textColor)
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
				Range: &docs.Range{
					StartIndex: startIndex,
					EndIndex:   endIndex,
				},
				TextStyle: textStyle,
				Fields:    strings.Join(fields, ","),
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
		"START":     true,
		"CENTER":    true,
		"END":       true,
		"JUSTIFIED": true,
	}

	if !validAlignments[alignment] {
		return fmt.Errorf("invalid alignment: %s (must be START, CENTER, END, or JUSTIFIED)", alignment)
	}

	service, err := getDocsService(ctx)
	if err != nil {
		return err
	}

	requests := []*docs.Request{
		{
			UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
				Range: &docs.Range{
					StartIndex: startIndex,
					EndIndex:   endIndex,
				},
				ParagraphStyle: &docs.ParagraphStyle{
					Alignment: alignment,
				},
				Fields: "alignment",
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

	service, err := getDocsService(ctx)
	if err != nil {
		return err
	}

	requests := []*docs.Request{
		{
			CreateParagraphBullets: &docs.CreateParagraphBulletsRequest{
				Range: &docs.Range{
					StartIndex: startIndex,
					EndIndex:   endIndex,
				},
				BulletPreset: listType + "_DISC_CIRCLE_SQUARE",
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

	service, err := getDocsService(ctx)
	if err != nil {
		return err
	}

	requests := []*docs.Request{
		{
			DeleteParagraphBullets: &docs.DeleteParagraphBulletsRequest{
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
		return fmt.Errorf("error removing bullets: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", green("✅ Bullets/numbering removed"))
	return nil
}
