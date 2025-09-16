package command

import "github.com/disgoorg/disgo/discord"

func Commands() []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		settingsCmd,
	}
}
