package call

import (
	"time"

	"github.com/disgoorg/snowflake/v2"
)

type Member struct {
	id       snowflake.ID
	name     string
	online   bool
	lastJoin time.Time
	duration time.Duration
	logs     []log
}

type log struct {
	join  time.Time
	leave time.Time
}

func NewMember(userID snowflake.ID, name string) *Member {
	return &Member{
		id:       userID,
		name:     name,
		online:   false,
		lastJoin: time.Time{},
		duration: 0,
		logs:     make([]log, 0),
	}
}

func (m *Member) MarkAsJoin(now time.Time) {
	if m.online {
		panic("member already online")
	}

	m.online = true
	m.lastJoin = now

	m.logs = append(m.logs, log{join: now})
}

func (m *Member) MarkAsLeave(now time.Time) {
	if !m.online {
		panic("member already offline")
	}

	m.online = false
	m.duration += now.Sub(m.lastJoin)

	l := len(m.logs)
	m.logs[l-1].leave = now
}

func (m Member) calculateDuration(now time.Time) time.Duration {
	if m.online {
		return m.duration + now.Sub(m.lastJoin)
	}

	return m.duration
}
