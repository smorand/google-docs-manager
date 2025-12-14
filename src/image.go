package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"google.golang.org/api/docs/v1"
)

// insertImageCmd inserts an image
var insertImageCmd = &cobra.Command{
	Use:   "insert-image <document-id> <index> <image-url>",
	Short: "Insert an image at index",
	Args:  cobra.ExactArgs(3),
	RunE:  runInsertImage,
}

func initImageCommands() {
	insertImageCmd.Flags().Float64("width", 0, "Image width in points")
	insertImageCmd.Flags().Float64("height", 0, "Image height in points")
}

func runInsertImage(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]

	index, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid index: %w", err)
	}

	imageURL := args[2]

	service, err := getDocsService(ctx)
	if err != nil {
		return err
	}

	request := &docs.InsertInlineImageRequest{
		Location: &docs.Location{Index: index},
		Uri:      imageURL,
	}

	imageWidth, _ := cmd.Flags().GetFloat64("width")
	imageHeight, _ := cmd.Flags().GetFloat64("height")

	if imageWidth > 0 && imageHeight > 0 {
		request.ObjectSize = &docs.Size{
			Width: &docs.Dimension{
				Magnitude: imageWidth,
				Unit:      "PT",
			},
			Height: &docs.Dimension{
				Magnitude: imageHeight,
				Unit:      "PT",
			},
		}
	}

	requests := []*docs.Request{
		{
			InsertInlineImage: request,
		},
	}

	_, err = service.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()

	if err != nil {
		return fmt.Errorf("error inserting image: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", green("✅ Image inserted"))
	return nil
}
