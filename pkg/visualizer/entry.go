package visualizer

import (
	"image"
	"image/color"
	"time"
)

type Entry struct {
	Avatar   image.Image
	Color    color.Color
	Sections []Section
}

// Section represents a time Section in a timeline entry.
// It is defined by a start and end time.
type Section struct {
	Start time.Time
	End   time.Time
}
