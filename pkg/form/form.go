package form

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type Form interface {
	Create() discord.MessageCreate
	Handle(event *events.ComponentInteractionCreate) error
}

type Bool bool

const (
	True  = Bool(true)
	False = Bool(false)
)

func (b Bool) String() string {
	if b {
		return "true"
	}
	return "false"
}

func ParseBool(s string) Bool {
	return s == "true"
}
