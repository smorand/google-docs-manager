package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"google.golang.org/api/docs/v1"
)

// insertTableCmd inserts a table
var insertTableCmd = &cobra.Command{
	Use:   "insert-table <document-id> <index> <rows> <cols>",
	Short: "Insert a table at index",
	Args:  cobra.ExactArgs(4),
	RunE:  runInsertTable,
}

// updateTableCellCmd updates a table cell
var updateTableCellCmd = &cobra.Command{
	Use:   "update-table-cell <document-id> <table-start-index> <row> <col> <text>",
	Short: "Update table cell content",
	Args:  cobra.ExactArgs(5),
	RunE:  runUpdateTableCell,
}

// styleTableCellCmd styles a table cell
var styleTableCellCmd = &cobra.Command{
	Use:   "style-table-cell <document-id> <table-start-index> <row> <col>",
	Short: "Style table cell (background color)",
	Args:  cobra.ExactArgs(4),
	RunE:  runStyleTableCell,
}

func initTableCommands() {
	styleTableCellCmd.Flags().String("bg-color", "", "Background color (hex, e.g., #FF0000)")
}

func runInsertTable(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]

	index, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid index: %w", err)
	}

	rows, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid rows: %w", err)
	}

	cols, err := strconv.Atoi(args[3])
	if err != nil {
		return fmt.Errorf("invalid cols: %w", err)
	}

	service, err := getDocsService(ctx)
	if err != nil {
		return err
	}

	requests := []*docs.Request{
		{
			InsertTable: &docs.InsertTableRequest{
				Location: &docs.Location{Index: index},
				Rows:     int64(rows),
				Columns:  int64(cols),
			},
		},
	}

	_, err = service.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()

	if err != nil {
		return fmt.Errorf("error inserting table: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", green(fmt.Sprintf("✅ Table inserted (%dx%d)", rows, cols)))
	return nil
}

func runUpdateTableCell(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]

	tableStartIndex, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid table start index: %w", err)
	}

	row, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid row: %w", err)
	}

	col, err := strconv.Atoi(args[3])
	if err != nil {
		return fmt.Errorf("invalid col: %w", err)
	}

	text := args[4]

	service, err := getDocsService(ctx)
	if err != nil {
		return err
	}

	// Get document to find table structure
	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}

	// Find table at the specified index
	var table *docs.Table
	for _, element := range doc.Body.Content {
		if element.Table != nil && element.StartIndex == tableStartIndex {
			table = element.Table
			break
		}
	}

	if table == nil {
		return fmt.Errorf("table not found at index %d", tableStartIndex)
	}

	if row >= len(table.TableRows) || col >= len(table.TableRows[0].TableCells) {
		return fmt.Errorf("row/col out of bounds")
	}

	// Get cell content start index
	cell := table.TableRows[row].TableCells[col]
	startIdx := cell.Content[0].StartIndex
	endIdx := cell.Content[len(cell.Content)-1].EndIndex

	requests := []*docs.Request{
		{
			DeleteContentRange: &docs.DeleteContentRangeRequest{
				Range: &docs.Range{
					StartIndex: startIdx,
					EndIndex:   endIdx - 1,
				},
			},
		},
		{
			InsertText: &docs.InsertTextRequest{
				Location: &docs.Location{Index: startIdx},
				Text:     text,
			},
		},
	}

	_, err = service.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()

	if err != nil {
		return fmt.Errorf("error updating cell: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", green(fmt.Sprintf("✅ Table cell updated (row %d, col %d)", row, col)))
	return nil
}

func runStyleTableCell(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	documentID := args[0]

	tableStartIndex, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid table start index: %w", err)
	}

	row, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid row: %w", err)
	}

	col, err := strconv.Atoi(args[3])
	if err != nil {
		return fmt.Errorf("invalid col: %w", err)
	}

	service, err := getDocsService(ctx)
	if err != nil {
		return err
	}

	// Get document to find table structure
	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}

	// Find table
	var table *docs.Table
	for _, element := range doc.Body.Content {
		if element.Table != nil && element.StartIndex == tableStartIndex {
			table = element.Table
			break
		}
	}

	if table == nil {
		return fmt.Errorf("table not found at index %d", tableStartIndex)
	}

	if row >= len(table.TableRows) || col >= len(table.TableRows[0].TableCells) {
		return fmt.Errorf("row/col out of bounds")
	}

	bgColor, _ := cmd.Flags().GetString("bg-color")

	requests := []*docs.Request{
		{
			UpdateTableCellStyle: &docs.UpdateTableCellStyleRequest{
				TableRange: &docs.TableRange{
					TableCellLocation: &docs.TableCellLocation{
						TableStartLocation: &docs.Location{Index: tableStartIndex},
						RowIndex:           int64(row),
						ColumnIndex:        int64(col),
					},
					RowSpan:    1,
					ColumnSpan: 1,
				},
				TableCellStyle: &docs.TableCellStyle{
					BackgroundColor: parseColor(bgColor),
				},
				Fields: "backgroundColor",
			},
		},
	}

	_, err = service.Documents.BatchUpdate(documentID, &docs.BatchUpdateDocumentRequest{
		Requests: requests,
	}).Do()

	if err != nil {
		return fmt.Errorf("error styling cell: %w", err)
	}

	fmt.Fprintf(os.Stderr, "%s\n", green(fmt.Sprintf("✅ Table cell styled (row %d, col %d)", row, col)))
	return nil
}
