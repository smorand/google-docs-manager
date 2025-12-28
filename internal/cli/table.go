package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"google-docs-manager/internal/auth"
	"google-docs-manager/internal/conversion"

	"github.com/spf13/cobra"
	"google.golang.org/api/docs/v1"
)

var (
	insertTableCmd = &cobra.Command{
		Args:  cobra.ExactArgs(4),
		RunE:  runInsertTable,
		Short: "Insert a table at index",
		Use:   "insert-table <document-id> <index> <rows> <cols>",
	}

	styleTableCellCmd = &cobra.Command{
		Args:  cobra.ExactArgs(4),
		RunE:  runStyleTableCell,
		Short: "Style table cell (background color)",
		Use:   "style-table-cell <document-id> <table-start-index> <row> <col>",
	}

	updateTableCellCmd = &cobra.Command{
		Args:  cobra.ExactArgs(5),
		RunE:  runUpdateTableCell,
		Short: "Update table cell content",
		Use:   "update-table-cell <document-id> <table-start-index> <row> <col> <text>",
	}
)

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

	service, err := auth.GetDocsService(ctx)
	if err != nil {
		return err
	}

	requests := []*docs.Request{
		{
			InsertTable: &docs.InsertTableRequest{
				Columns:  int64(cols),
				Location: &docs.Location{Index: index},
				Rows:     int64(rows),
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

	service, err := auth.GetDocsService(ctx)
	if err != nil {
		return err
	}

	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}

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
				Fields: "backgroundColor",
				TableCellStyle: &docs.TableCellStyle{
					BackgroundColor: conversion.ParseColor(bgColor),
				},
				TableRange: &docs.TableRange{
					ColumnSpan: 1,
					RowSpan:    1,
					TableCellLocation: &docs.TableCellLocation{
						ColumnIndex:        int64(col),
						RowIndex:           int64(row),
						TableStartLocation: &docs.Location{Index: tableStartIndex},
					},
				},
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

	service, err := auth.GetDocsService(ctx)
	if err != nil {
		return err
	}

	doc, err := service.Documents.Get(documentID).Do()
	if err != nil {
		return fmt.Errorf("error getting document: %w", err)
	}

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

	cell := table.TableRows[row].TableCells[col]
	startIdx := cell.Content[0].StartIndex
	endIdx := cell.Content[len(cell.Content)-1].EndIndex

	requests := []*docs.Request{
		{
			DeleteContentRange: &docs.DeleteContentRangeRequest{
				Range: &docs.Range{
					EndIndex:   endIdx - 1,
					StartIndex: startIdx,
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
