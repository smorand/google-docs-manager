package document

import (
	"fmt"
	"strings"

	"google-docs-manager/internal/conversion"
	"google.golang.org/api/docs/v1"
)

// Section represents a document section/heading
type Section struct {
	EndIndex   int64  `json:"endIndex"`
	Level      int    `json:"level"`
	StartIndex int64  `json:"startIndex"`
	Title      string `json:"title"`
}

// GetStructure extracts the structure of a document
func GetStructure(doc *docs.Document) []Section {
	var sections []Section

	for _, element := range doc.Body.Content {
		if element.Paragraph != nil {
			paragraph := element.Paragraph
			if paragraph.ParagraphStyle != nil && paragraph.ParagraphStyle.NamedStyleType != "" {
				styleType := paragraph.ParagraphStyle.NamedStyleType

				if strings.HasPrefix(styleType, "HEADING_") {
					level := 0
					fmt.Sscanf(styleType, "HEADING_%d", &level)

					text := conversion.GetParagraphText(paragraph)
					startIndex := element.StartIndex
					endIndex := element.EndIndex

					sections = append(sections, Section{
						EndIndex:   endIndex,
						Level:      level,
						StartIndex: startIndex,
						Title:      text,
					})
				}
			}
		}
	}

	return sections
}

// FindSection finds a section by name
func FindSection(doc *docs.Document, sectionName string) *Section {
	sections := GetStructure(doc)

	for _, section := range sections {
		if strings.EqualFold(section.Title, sectionName) {
			return &section
		}
	}

	return nil
}
