package call

import (
	"time"

	"github.com/disgoorg/snowflake/v2"
)

type Member struct {
	id                snowflake.ID
	name              string
	online            bool
	lastUpdate        time.Time
	duration          time.Duration
	onlineSections    []sectionWithStatus
	streamingSections []section
}

type section struct {
	start time.Time
	end   time.Time
}

type sectionWithStatus struct {
	section
	mute bool
	deaf bool
}

func (s sectionWithStatus) IsSameStatus(other sectionWithStatus) bool {
	return s.mute == other.mute && s.deaf == other.deaf
}

func NewMember(userID snowflake.ID, name string) *Member {
	return &Member{
		id:                userID,
		name:              name,
		online:            false,
		lastUpdate:        time.Time{},
		duration:          0,
		onlineSections:    make([]sectionWithStatus, 0),
		streamingSections: make([]section, 0),
	}
}

func (m *Member) MarkAsOnline(now time.Time, muted, deaf bool) {
	if m.online {
		panic("member already online")
	}

	m.online = true
	m.lastUpdate = now

	section := sectionWithStatus{mute: muted, deaf: deaf}
	section.start = now
	m.onlineSections = append(m.onlineSections, section)
}

func (m *Member) UpdateStatus(now time.Time, muted, deaf bool) {
	if !m.online {
		panic("member not online")
	}

	l := len(m.onlineSections)

	after := sectionWithStatus{mute: muted, deaf: deaf}
	if m.onlineSections[l-1].IsSameStatus(after) {
		return
	}

	// calculate the duration of the last section
	m.duration += now.Sub(m.lastUpdate)

	m.onlineSections[l-1].end = now
	after.start = now
	m.onlineSections = append(m.onlineSections, after)
	m.lastUpdate = now
}

func (m *Member) UnmarkAsOnline(now time.Time) {
	if !m.online {
		panic("member already offline")
	}

	m.online = false
	m.duration += now.Sub(m.lastUpdate)

	l := len(m.onlineSections)
	m.onlineSections[l-1].end = now
}

func (m *Member) MarkAsStreaming(now time.Time) {
	m.streamingSections = append(m.streamingSections, section{start: now})
}

func (m *Member) UnmarkAsStreaming(now time.Time) {
	if len(m.streamingSections) == 0 {
		panic("member not streaming")
	}

	l := len(m.streamingSections)
	m.streamingSections[l-1].end = now
}

func (m *Member) HasStreamed() bool {
	return len(m.streamingSections) > 0
}

func (m Member) calculateDuration(now time.Time) time.Duration {
	if m.online {
		return m.duration + now.Sub(m.lastUpdate)
	}

	return m.duration
}
