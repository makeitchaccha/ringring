package call

import (
	"fmt"
	"time"

	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
)

// Manager is used for managing ongoing calls
type Manager interface {
	Add(call *Call, now time.Time) (Handler, error)
	Remove(channelID snowflake.ID)
	Get(channelID snowflake.ID) (Handler, bool)
}

var _ Manager = (*managerImpl)(nil)

type managerImpl struct {
	rest     rest.Rest
	handlers map[snowflake.ID]Handler
}

// NewManager creates a new Manager
func NewManager(rest rest.Rest) Manager {
	return &managerImpl{
		rest:     rest,
		handlers: make(map[snowflake.ID]Handler),
	}
}

func (m *managerImpl) Add(call *Call, now time.Time) (Handler, error) {
	handler, err := NewHandler(call, m.rest, now)

	if err != nil {
		return nil, fmt.Errorf("failed to create handler: %w", err)
	}

	m.handlers[call.ChannelID] = handler
	return handler, nil

}

func (m *managerImpl) Remove(channelID snowflake.ID) {
	delete(m.handlers, channelID)
}

func (m *managerImpl) Get(channelID snowflake.ID) (Handler, bool) {
	call, ok := m.handlers[channelID]
	return call, ok
}
