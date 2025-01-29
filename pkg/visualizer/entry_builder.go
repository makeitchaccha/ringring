package visualizer

import (
	"image"
	"image/color"
	"time"
)

type EntryBuilder struct {
	Entry
}

func NewEntryBuilder(avatar image.Image, color color.Color) *EntryBuilder {
	b := &EntryBuilder{}
	b.Avatar = avatar
	b.Color = color
	return b
}

func (b *EntryBuilder) AddSection(start, end time.Time) *EntryBuilder {
	b.Sections = append(b.Sections, Section{Start: start, End: end})
	return b
}

func (b *EntryBuilder) Build() Entry {
	return b.Entry
}
