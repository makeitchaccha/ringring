package main

import (
	"os"

	"github.com/disgoorg/disgo"
	"github.com/yuyaprgrm/ringring/internal/pkg/icommand"
	"github.com/yuyaprgrm/ringring/internal/pkg/locale"
	"github.com/yuyaprgrm/ringring/pkg/command"
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
