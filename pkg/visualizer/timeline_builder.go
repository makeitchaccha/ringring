package visualizer

import "time"

type TimelineBuilder struct {
	Timeline
}

func NewTimelineBuilder(start, end time.Time) *TimelineBuilder {
	b := &TimelineBuilder{}
	b.StartTime = start
	b.EndTime = end
	return b
}

func (b *TimelineBuilder) AddEntries(entries ...Entry) *TimelineBuilder {
	b.Entries = append(b.Entries, entries...)
	return b
}

func (b *TimelineBuilder) Build() Timeline {
	return b.Timeline
}
