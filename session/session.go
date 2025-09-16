package session

import (
	"weak"

	"github.com/disgoorg/snowflake/v2"
)

type Status string

// Lifecycle of a session
const (
	StatusActive   Status = "active"   // Session is currently active: Someone is in the call
	StatusInactive Status = "inactive" // Session is inactive: No one is in the call, but session exists.
	StatusEnded    Status = "ended"    // Session has ended: No one is in the call, and session is waiting to be disposed.
)

type Session struct {
	ChannelID     snowflake.ID
	SessionStatus Status

	Members   []*Member
	MemberMap map[snowflake.ID]weak.Pointer[*Member]
}
