package main

import (
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/api/docs/v1"
)

// Document structure types
type Section struct {
	Title      string
	StartIndex int64
	EndIndex   int64
	Level      int
}

// docsToMarkdown converts a Google Doc to markdown format
func docsToMarkdown(doc *docs.Document) string {
	var md strings.Builder

	// Get document title
	md.WriteString(fmt.Sprintf("# %s\n\n", doc.Title))

	// Process content
	for _, element := range doc.Body.Content {
		if element.Paragraph != nil {
			paragraph := element.Paragraph

			// Check for heading style
			if paragraph.ParagraphStyle != nil && paragraph.ParagraphStyle.NamedStyleType != "" {
				styleType := paragraph.ParagraphStyle.NamedStyleType
				text := getParagraphText(paragraph)

				switch styleType {
				case "HEADING_1":
					md.WriteString(fmt.Sprintf("# %s\n\n", text))
				case "HEADING_2":
					md.WriteString(fmt.Sprintf("## %s\n\n", text))
				case "HEADING_3":
					md.WriteString(fmt.Sprintf("### %s\n\n", text))
				case "HEADING_4":
					md.WriteString(fmt.Sprintf("#### %s\n\n", text))
				case "HEADING_5":
					md.WriteString(fmt.Sprintf("##### %s\n\n", text))
				case "HEADING_6":
					md.WriteString(fmt.Sprintf("###### %s\n\n", text))
				default:
					// Normal paragraph
					md.WriteString(formatParagraphAsMarkdown(paragraph))
				}
			} else {
				// Normal paragraph
				md.WriteString(formatParagraphAsMarkdown(paragraph))
			}
		} else if element.Table != nil {
			// Handle tables
			md.WriteString(tableToMarkdown(element.Table))
		}
	}

	return md.String()
}

// getParagraphText extracts text from a paragraph
func getParagraphText(paragraph *docs.Paragraph) string {
	var text strings.Builder
	for _, element := range paragraph.Elements {
		if element.TextRun != nil {
			text.WriteString(element.TextRun.Content)
		}
	}
	return strings.TrimSpace(text.String())
}

// formatParagraphAsMarkdown formats a paragraph with inline formatting
func formatParagraphAsMarkdown(paragraph *docs.Paragraph) string {
	var text strings.Builder

	for _, element := range paragraph.Elements {
		if element.TextRun != nil {
			content := element.TextRun.Content
			style := element.TextRun.TextStyle

			if style != nil {
				// Apply formatting
				if style.Bold {
					content = fmt.Sprintf("**%s**", strings.TrimSpace(content))
				}
				if style.Italic {
					content = fmt.Sprintf("*%s*", strings.TrimSpace(content))
				}
				if style.Link != nil {
					url := style.Link.Url
					content = fmt.Sprintf("[%s](%s)", strings.TrimSpace(content), url)
				}
			}

			text.WriteString(content)
		}
	}

	result := text.String()
	if strings.TrimSpace(result) != "" {
		return result + "\n\n"
	}
	return ""
}

// tableToMarkdown converts a table to markdown format
func tableToMarkdown(table *docs.Table) string {
	var md strings.Builder

	for rowIdx, row := range table.TableRows {
		md.WriteString("|")
		for _, cell := range row.TableCells {
			cellText := getTableCellText(cell)
			md.WriteString(fmt.Sprintf(" %s |", cellText))
		}
		md.WriteString("\n")

		// Add separator after header row
		if rowIdx == 0 {
			md.WriteString("|")
			for range row.TableCells {
				md.WriteString(" --- |")
			}
			md.WriteString("\n")
		}
	}

	md.WriteString("\n")
	return md.String()
}

// getTableCellText extracts text from a table cell
func getTableCellText(cell *docs.TableCell) string {
	var text strings.Builder
	for _, element := range cell.Content {
		if element.Paragraph != nil {
			text.WriteString(getParagraphText(element.Paragraph))
		}
	}
	return strings.TrimSpace(text.String())
}

// markdownToDocsRequests converts markdown to Docs API requests
func markdownToDocsRequests(markdown string, startIndex int64) []*docs.Request {
	var requests []*docs.Request
	lines := strings.Split(markdown, "\n")
	currentIndex := startIndex

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			// Insert newline
			requests = append(requests, &docs.Request{
				InsertText: &docs.InsertTextRequest{
					Location: &docs.Location{Index: currentIndex},
					Text:     "\n",
				},
			})
			currentIndex++
			continue
		}

		// Check for headings
		if strings.HasPrefix(line, "#") {
			level := 0
			for _, c := range line {
				if c == '#' {
					level++
				} else {
					break
				}
			}
			text := strings.TrimSpace(line[level:])

			// Insert text
			requests = append(requests, &docs.Request{
				InsertText: &docs.InsertTextRequest{
					Location: &docs.Location{Index: currentIndex},
					Text:     text + "\n",
				},
			})

			// Apply heading style
			styleType := fmt.Sprintf("HEADING_%d", level)
			requests = append(requests, &docs.Request{
				UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
					Range: &docs.Range{
						StartIndex: currentIndex,
						EndIndex:   currentIndex + int64(len(text)) + 1,
					},
					ParagraphStyle: &docs.ParagraphStyle{
						NamedStyleType: styleType,
					},
					Fields: "namedStyleType",
				},
			})

			currentIndex += int64(len(text)) + 1
		} else {
			// Regular paragraph - handle inline formatting
			processedText, formatRequests := parseInlineFormatting(line, currentIndex)

			// Insert text
			requests = append(requests, &docs.Request{
				InsertText: &docs.InsertTextRequest{
					Location: &docs.Location{Index: currentIndex},
					Text:     processedText + "\n",
				},
			})

			// Add formatting requests
			requests = append(requests, formatRequests...)

			currentIndex += int64(len(processedText)) + 1
		}
	}

	return requests
}

// parseInlineFormatting parses markdown inline formatting
func parseInlineFormatting(text string, startIndex int64) (string, []*docs.Request) {
	var requests []*docs.Request
	plainText := text

	// Bold (**text** or __text__)
	boldRegex := regexp.MustCompile(`\*\*([^*]+)\*\*|__([^_]+)__`)
	matches := boldRegex.FindAllStringSubmatchIndex(plainText, -1)
	for _, match := range matches {
		start := match[0]
		end := match[1]
		content := plainText[match[2]:match[3]]
		if content == "" {
			content = plainText[match[4]:match[5]]
		}

		requests = append(requests, &docs.Request{
			UpdateTextStyle: &docs.UpdateTextStyleRequest{
				Range: &docs.Range{
					StartIndex: startIndex + int64(start),
					EndIndex:   startIndex + int64(end),
				},
				TextStyle: &docs.TextStyle{
					Bold: true,
				},
				Fields: "bold",
			},
		})
	}

	// Remove markdown syntax for plain text
	plainText = boldRegex.ReplaceAllString(plainText, "$1$2")

	// Italic (*text* or _text_)
	italicRegex := regexp.MustCompile(`\*([^*]+)\*|_([^_]+)_`)
	plainText = italicRegex.ReplaceAllString(plainText, "$1$2")

	return plainText, requests
}

// getDocumentStructure extracts the structure of a document
func getDocumentStructure(doc *docs.Document) []Section {
	var sections []Section

	for _, element := range doc.Body.Content {
		if element.Paragraph != nil {
			paragraph := element.Paragraph
			if paragraph.ParagraphStyle != nil && paragraph.ParagraphStyle.NamedStyleType != "" {
				styleType := paragraph.ParagraphStyle.NamedStyleType

				// Check if it's a heading
				if strings.HasPrefix(styleType, "HEADING_") {
					level := 0
					fmt.Sscanf(styleType, "HEADING_%d", &level)

					text := getParagraphText(paragraph)
					startIndex := element.StartIndex
					endIndex := element.EndIndex

					sections = append(sections, Section{
						Title:      text,
						StartIndex: startIndex,
						EndIndex:   endIndex,
						Level:      level,
					})
				}
			}
		}
	}

	return sections
}

// findSection finds a section by name
func findSection(doc *docs.Document, sectionName string) *Section {
	sections := getDocumentStructure(doc)

	for _, section := range sections {
		if strings.EqualFold(section.Title, sectionName) {
			return &section
		}
	}

	return nil
}

// parseColor parses a hex color to RGB values
func parseColor(hexColor string) *docs.OptionalColor {
	hexColor = strings.TrimPrefix(hexColor, "#")

	if len(hexColor) != 6 {
		return nil
	}

	var r, g, b int
	fmt.Sscanf(hexColor, "%02x%02x%02x", &r, &g, &b)

	return &docs.OptionalColor{
		Color: &docs.Color{
			RgbColor: &docs.RgbColor{
				Red:   float64(r) / 255.0,
				Green: float64(g) / 255.0,
				Blue:  float64(b) / 255.0,
			},
		},
	}
}
