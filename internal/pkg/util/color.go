package util

import (
	"image"
	"image/color"

	"github.com/lucasb-eyer/go-colorful"
)

func ExtractMainColor(img image.Image) color.Color {
	bounds := img.Bounds()
	colorCount := make(map[color.Color]int)

	// HACK: Prevent all white or transparent images from being considered
	colorCount[color.Black] = 1

	// Iterate over each pixel
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			// Ignore white like colors
			if r > 0xBFFF && g > 0xBFFF && b > 0xBFFF {
				continue
			}
			// also ignore transparent pixels
			if a == 0 {
				continue
			}
			colorCount[c]++
		}
	}

	// Find the most frequent color
	var mainColor color.Color
	maxCount := 0
	for c, count := range colorCount {
		if count > maxCount {
			maxCount = count
			mainColor = c
		}
	}

	return mainColor
}

func TransformColorWithSpecificLuminance(c color.Color, targetLuminance float64) color.Color {
	original, successful := colorful.MakeColor(c)

	if !successful {
		return c // Return the original color if conversion fails
	}

	_, a, b := original.Lab()
	transformed := colorful.Lab(targetLuminance, a, b)

	return transformed.Clamped()
}
