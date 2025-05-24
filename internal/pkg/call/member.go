package call

import (
	"time"

	"github.com/disgoorg/snowflake/v2"
)

type Member struct {
	id                snowflake.ID
	name              string
	online            bool
	lastJoin          time.Time
	duration          time.Duration
	onlineSections    []section
	streamingSections []section
}

type section struct {
	start time.Time
	end   time.Time
}

func NewMember(userID snowflake.ID, name string) *Member {
	return &Member{
		id:                userID,
		name:              name,
		online:            false,
		lastJoin:          time.Time{},
		duration:          0,
		onlineSections:    make([]section, 0),
		streamingSections: make([]section, 0),
	}
}

func (m *Member) MarkAsOnline(now time.Time) {
	if m.online {
		panic("member already online")
	}

	m.online = true
	m.lastJoin = now

	m.onlineSections = append(m.onlineSections, section{start: now})
}

func (m *Member) UnmarkAsOnline(now time.Time) {
	if !m.online {
		panic("member already offline")
	}

	m.online = false
	m.duration += now.Sub(m.lastJoin)

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
		return m.duration + now.Sub(m.lastJoin)
	}

	return m.duration
}
