package form

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

type Manager interface {
	// TODO: better name
	Send(channelID snowflake.ID, form Form) error

	// as same as the command manager, we need to handle the interaction
	OnComponentInteractionCreate(event *events.ComponentInteractionCreate)
}

var _ Manager = (*managerImpl)(nil)

type managerImpl struct {
	rest  rest.Channels
	forms map[snowflake.ID]Form // map of message ID to form
}

func NewManager(rest rest.Channels) Manager {
	return &managerImpl{
		rest:  rest,
		forms: make(map[snowflake.ID]Form),
	}
}

func (m *managerImpl) Send(channelID snowflake.ID, form Form) error {
	msg, err := m.rest.CreateMessage(channelID, form.Create())
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	m.forms[msg.ID] = form
	return nil
}

func (m *managerImpl) OnComponentInteractionCreate(event *events.ComponentInteractionCreate) {
	form, ok := m.forms[event.Message.ID]
	if !ok {
		return
	}

	if err := form.Handle(event); err != nil {
		event.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Failed to handle interaction: %v", err),
		})
	}
}
