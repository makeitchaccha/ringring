package visualizer

import (
	"image"
	"image/color"
	"io"
	"time"

	"github.com/fogleman/gg"
	"github.com/makeitchaccha/rendering/chart/timeline"
	"github.com/makeitchaccha/rendering/layout"
)

type Timeline struct {
	StartTime time.Time
	EndTime   time.Time
	Entries   []Entry
}

func (t Timeline) Generate() io.Reader {
	nEntries := len(t.Entries)

	const TIMELINE_WIDTH = 900.0
	const HEADLINE_WIDTH = 100.0
	const ENTRY_HEIGHT = 70.0
	const ONLINE_BAR_WIDTH = 20.0
	const ONLINE_BAR_FILLING_FACTOR = ONLINE_BAR_WIDTH / ENTRY_HEIGHT
	const PADDING = 10.0
	const PADDING_TOP = 40.0

	width := TIMELINE_WIDTH + HEADLINE_WIDTH + ONLINE_BAR_WIDTH + 2*PADDING
	height := ENTRY_HEIGHT*float64(nEntries) + PADDING_TOP + PADDING

	dc := gg.NewContext(int(width), int(height))
	dc.SetFontFace(fontFace)
	dc.SetColor(color.White)
	dc.Clear()

	cellHeights := make([]float64, nEntries)
	for i := range t.Entries {
		cellHeights[i] = ENTRY_HEIGHT
	}
	grid := layout.NewGrid(PADDING, PADDING_TOP, []float64{HEADLINE_WIDTH, TIMELINE_WIDTH}, cellHeights)

	headerGrid, _ := grid.ColAsSubgrid(0)
	for idx, f := range headerGrid.ForEachCellRenderFunc {
		entry := t.Entries[idx.Row]
		f(dc, func(dc *gg.Context, x, y, w, h float64) error {
			dc.Push()
			dc.DrawCircle(x+w/2, y+h/2, float64(entry.Avatar.Bounds().Dx())/2)
			dc.Clip()
			dc.DrawImageAnchored(entry.Avatar, int(x+w/2), int(y+h/2), 0.5, 0.5)
			dc.ResetClip()
			dc.Pop()
			return nil
		})
	}

	timelineGrid, _ := grid.ColAsSubgrid(1)

	// draw tics on an hour intervals

	timelineBounds := timelineGrid.Bounds()

	total := t.EndTime.Sub(t.StartTime).Seconds()

	main, sub := CalculateTics(t.EndTime.Sub(t.StartTime))
	for i, tics := range []Tics{main, sub} {
		_, offset := t.StartTime.Zone()
		dOffset := time.Duration(offset) * time.Second
		current := t.StartTime.Add(dOffset).Truncate(tics.interval).Add(-dOffset)
		if current.Before(t.StartTime) {
			current = current.Add(tics.interval) // move to the next hour to avoid drawing a tic at the start
		}
		for ; current.Before(t.EndTime); current = current.Add(tics.interval) {
			x := timelineBounds.Min.X + (timelineBounds.Dx() * current.Sub(t.StartTime).Seconds() / total)
			// draw a tic and label on the top

			dc.SetColor(color.RGBA{66, 66, 66, 255})
			dc.DrawStringAnchored(current.Format(tics.format), x, timelineBounds.Min.Y-5, 0.5, -float64(i))

			dc.SetColor(tics.color)
			dc.DrawLine(x, timelineBounds.Min.Y, x, timelineBounds.Max.Y)
			dc.Stroke()
		}

	}

	builder := timeline.NewTimelineBuilder().
		SetFillingFactor(ONLINE_BAR_FILLING_FACTOR)

	for _, entry := range t.Entries {
		if entry.Color == nil {
			entry.Color = extractMainColor(entry.Avatar)
		}

		entryBuilder := timeline.NewEntryBuilder(entry.Color)
		for _, section := range entry.Sections {
			s := section.Start.Sub(t.StartTime).Seconds() / total
			e := section.End.Sub(t.StartTime).Seconds() / total
			entryBuilder.AddSection(s, e)
		}

		builder.AddEntry(entryBuilder.Build())
	}

	builder.Build().RenderInGrid(dc, timelineGrid)

	// Draw start and end time vertical lines
	dc.SetColor(color.RGBA{0, 105, 92, 255})
	dc.DrawLine(timelineBounds.Min.X, timelineBounds.Min.Y, timelineBounds.Min.X, timelineBounds.Max.Y)
	dc.DrawLine(timelineBounds.Max.X, timelineBounds.Min.Y, timelineBounds.Max.X, timelineBounds.Max.Y)
	dc.Stroke()

	r, w := io.Pipe()
	go func() {
		dc.EncodePNG(w)
		w.Close()
	}()

	return r
}

func extractMainColor(img image.Image) color.Color {
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
