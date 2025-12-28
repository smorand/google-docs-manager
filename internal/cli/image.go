package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"google-docs-manager/internal/auth"

	"github.com/spf13/cobra"
	"google.golang.org/api/docs/v1"
)

var insertImageCmd = &cobra.Command{
	Args:  cobra.ExactArgs(3),
	RunE:  runInsertImage,
	Short: "Insert an image at index",
	Use:   "insert-image <document-id> <index> <image-url>",
}

func initImageCommands() {
	insertImageCmd.Flags().Float64("height", 0, "Image height in points")
	insertImageCmd.Flags().Float64("width", 0, "Image width in points")
}

func runInsertImage(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]

	index, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid index: %w", err)
	}

	imageURL := args[2]

	service, err := auth.GetDocsService(ctx)
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
			Height: &docs.Dimension{
				Magnitude: imageHeight,
				Unit:      "PT",
			},
			Width: &docs.Dimension{
				Magnitude: imageWidth,
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
