package conversion

import (
	"fmt"
	"strings"

	"google.golang.org/api/docs/v1"
)

// ParseColor parses a hex color to RGB values for Google Docs API
func ParseColor(hexColor string) *docs.OptionalColor {
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
