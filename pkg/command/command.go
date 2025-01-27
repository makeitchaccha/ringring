package command

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type Command interface {
	Name() string
	Create() discord.ApplicationCommandCreate
	Execute(event *events.ApplicationCommandInteractionCreate) error
}
