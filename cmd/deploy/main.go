package main

import (
	"log"
	"os"

	"github.com/disgoorg/disgo"
	"github.com/makeitchaccha/ringring/internal/pkg/i18n"
	"github.com/makeitchaccha/ringring/internal/pkg/icommand"
	"github.com/makeitchaccha/ringring/pkg/command"
)

func main() {
	translator, err := i18n.NewTranslator("./locales", "en-US")
	if err != nil {
		log.Fatalf("failed to initialize translator: %v", err)
	}

	// login discord bot and deploy commands
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
