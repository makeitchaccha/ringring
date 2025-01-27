package command

import (
	"errors"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

var (
	ErrCommandAlreadyExists = errors.New("command already exists")
)

type Manager interface {
	Register(command Command) error
	Deploy(client bot.Client) error

	// OnCommandInteractionCreate should be called when ApplicationCommandInteractionCreate event is received
	// This is used to handle the command interaction
	OnCommandInteractionCreate(event *events.ApplicationCommandInteractionCreate)
}

type managerImpl struct {
	commands map[string]Command
}

func NewManager() Manager {
	return &managerImpl{commands: make(map[string]Command)}
}

func (m *managerImpl) Register(command Command) error {
	if _, ok := m.commands[command.Name()]; ok {
		return ErrCommandAlreadyExists
	}
	m.commands[command.Name()] = command
	return nil
}

func (m *managerImpl) Deploy(client bot.Client) error {
	for _, command := range m.commands {
		if _, err := client.Rest().CreateGlobalCommand(client.ApplicationID(), command.Create()); err != nil {
			return err
		}
	}
	return nil
}

func (m *managerImpl) OnCommandInteractionCreate(event *events.ApplicationCommandInteractionCreate) {
	command, ok := m.commands[event.Data.CommandName()]

	if !ok {
		// just ignore, maybe other bot should handle this command
		return
	}

	if err := command.Execute(event); err != nil {
		m.handleCommandError(event, err)
	}
}

func (m *managerImpl) handleCommandError(event *events.ApplicationCommandInteractionCreate, err error) {
	event.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent("Failed to execute command: " + err.Error()).
		SetEphemeral(true).
		Build())
}
