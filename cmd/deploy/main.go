package main

import (
	"os"

	"github.com/disgoorg/disgo"
	"github.com/makeitchaccha/ringring/internal/pkg/icommand"
	"github.com/makeitchaccha/ringring/internal/pkg/locale"
	"github.com/makeitchaccha/ringring/pkg/command"
)

func main() {
	// login discord bot and deploy commands
	locale.Init("./locales")
	client, err := disgo.New(os.Getenv("DISCORD_TOKEN"))

	if err != nil {
		panic(err)
	}

	// deploy commands
	commandManager := command.NewManager()
	commandManager.Register(&icommand.Settings{})

	if err := commandManager.Deploy(client); err != nil {
		panic(err)
	}

}
