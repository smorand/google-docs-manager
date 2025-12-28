package conversion

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"google.golang.org/api/docs/v1"
)

// DocsToMarkdown converts a Google Doc to markdown format
func DocsToMarkdown(doc *docs.Document) string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("# %s\n\n", doc.Title))

	for _, element := range doc.Body.Content {
		if element.Paragraph != nil {
			paragraph := element.Paragraph

			if paragraph.ParagraphStyle != nil && paragraph.ParagraphStyle.NamedStyleType != "" {
				styleType := paragraph.ParagraphStyle.NamedStyleType
				text := GetParagraphText(paragraph)

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
					md.WriteString(formatParagraphAsMarkdown(paragraph))
				}
			} else {
				md.WriteString(formatParagraphAsMarkdown(paragraph))
			}
		} else if element.Table != nil {
			md.WriteString(tableToMarkdown(element.Table))
		}
	}

	return md.String()
}

// GetParagraphText extracts text from a paragraph
func GetParagraphText(paragraph *docs.Paragraph) string {
	var text strings.Builder
	for _, element := range paragraph.Elements {
		if element.TextRun != nil {
			text.WriteString(element.TextRun.Content)
		}
	}
	return strings.TrimSpace(text.String())
}

func formatParagraphAsMarkdown(paragraph *docs.Paragraph) string {
	var text strings.Builder

	for _, element := range paragraph.Elements {
		if element.TextRun != nil {
			content := element.TextRun.Content
			style := element.TextRun.TextStyle

			if style != nil {
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

func tableToMarkdown(table *docs.Table) string {
	var md strings.Builder

	for rowIdx, row := range table.TableRows {
		md.WriteString("|")
		for _, cell := range row.TableCells {
			cellText := getTableCellText(cell)
			md.WriteString(fmt.Sprintf(" %s |", cellText))
		}
		md.WriteString("\n")

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

func getTableCellText(cell *docs.TableCell) string {
	var text strings.Builder
	for _, element := range cell.Content {
		if element.Paragraph != nil {
			text.WriteString(GetParagraphText(element.Paragraph))
		}
	}
	return strings.TrimSpace(text.String())
}

// MarkdownToDocsRequests converts markdown to Docs API requests
func MarkdownToDocsRequests(markdown string, startIndex int64) []*docs.Request {
	var requests []*docs.Request
	lines := strings.Split(markdown, "\n")
	currentIndex := startIndex

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			requests = append(requests, &docs.Request{
				InsertText: &docs.InsertTextRequest{
					Location: &docs.Location{Index: currentIndex},
					Text:     "\n",
				},
			})
			currentIndex++
			continue
		}

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

			requests = append(requests, &docs.Request{
				InsertText: &docs.InsertTextRequest{
					Location: &docs.Location{Index: currentIndex},
					Text:     text + "\n",
				},
			})

			styleType := fmt.Sprintf("HEADING_%d", level)
			textLen := int64(utf8.RuneCountInString(text))
			requests = append(requests, &docs.Request{
				UpdateParagraphStyle: &docs.UpdateParagraphStyleRequest{
					Range: &docs.Range{
						StartIndex: currentIndex,
						EndIndex:   currentIndex + textLen + 1,
					},
					ParagraphStyle: &docs.ParagraphStyle{
						NamedStyleType: styleType,
					},
					Fields: "namedStyleType",
				},
			})

			currentIndex += textLen + 1
		} else {
			processedText, formatRequests := parseInlineFormatting(line, currentIndex)

			requests = append(requests, &docs.Request{
				InsertText: &docs.InsertTextRequest{
					Location: &docs.Location{Index: currentIndex},
					Text:     processedText + "\n",
				},
			})

			requests = append(requests, formatRequests...)

			textLen := int64(utf8.RuneCountInString(processedText))
			currentIndex += textLen + 1
		}
	}

	return requests
}

func parseInlineFormatting(text string, startIndex int64) (string, []*docs.Request) {
	var requests []*docs.Request

	boldRegex := regexp.MustCompile(`\*\*([^*]+)\*\*|__([^_]+)__`)
	matches := boldRegex.FindAllStringSubmatchIndex(text, -1)

	offset := int64(0)
	for _, match := range matches {
		matchStart := match[0]
		matchEnd := match[1]

		var content string
		if match[2] != -1 && match[3] != -1 {
			content = text[match[2]:match[3]]
		} else if match[4] != -1 && match[5] != -1 {
			content = text[match[4]:match[5]]
		}

		contentRuneCount := int64(utf8.RuneCountInString(content))
		matchRuneCount := int64(utf8.RuneCountInString(text[matchStart:matchEnd]))
		markerLength := matchRuneCount - contentRuneCount

		matchStartRunes := int64(utf8.RuneCountInString(text[:matchStart]))
		adjustedStart := matchStartRunes - offset
		adjustedEnd := adjustedStart + contentRuneCount

		requests = append(requests, &docs.Request{
			UpdateTextStyle: &docs.UpdateTextStyleRequest{
				Range: &docs.Range{
					StartIndex: startIndex + adjustedStart,
					EndIndex:   startIndex + adjustedEnd,
				},
				TextStyle: &docs.TextStyle{
					Bold: true,
				},
				Fields: "bold",
			},
		})

		offset += markerLength
	}

	plainText := boldRegex.ReplaceAllString(text, "$1$2")

	italicRegex := regexp.MustCompile(`\*([^*]+)\*|_([^_]+)_`)
	italicMatches := italicRegex.FindAllStringSubmatchIndex(plainText, -1)

	offset = int64(0)
	for _, match := range italicMatches {
		matchStart := match[0]
		matchEnd := match[1]

		var content string
		if match[2] != -1 && match[3] != -1 {
			content = plainText[match[2]:match[3]]
		} else if match[4] != -1 && match[5] != -1 {
			content = plainText[match[4]:match[5]]
		}

		contentRuneCount := int64(utf8.RuneCountInString(content))
		matchRuneCount := int64(utf8.RuneCountInString(plainText[matchStart:matchEnd]))
		markerLength := matchRuneCount - contentRuneCount

		matchStartRunes := int64(utf8.RuneCountInString(plainText[:matchStart]))
		adjustedStart := matchStartRunes - offset
		adjustedEnd := adjustedStart + contentRuneCount

		requests = append(requests, &docs.Request{
			UpdateTextStyle: &docs.UpdateTextStyleRequest{
				Range: &docs.Range{
					StartIndex: startIndex + adjustedStart,
					EndIndex:   startIndex + adjustedEnd,
				},
				TextStyle: &docs.TextStyle{
					Italic: true,
				},
				Fields: "italic",
			},
		})

		offset += markerLength
	}

	plainText = italicRegex.ReplaceAllString(plainText, "$1$2")

	return plainText, requests
}
