package visualizer

import (
	"image"
	"image/color"
	"time"
)

type Entry struct {
	Avatar   image.Image
	Color    color.Color
	Sections []section
}

// section represents a time section in a timeline entry.
// It is defined by a start and end time.
type section struct {
	Start time.Time
	End   time.Time
}
