package call

import (
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/makeitchaccha/design/timeline"
)

// handler helps to update the call status
type Handler interface {
	RegisterMember(userID snowflake.ID, member *discord.Member)
	IsRegistered(userID snowflake.ID) bool
	MemberJoin(userID snowflake.ID, now time.Time)
	MemberLeave(userID snowflake.ID, now time.Time) (isEmpty bool)
	MemberStartStreaming(userID snowflake.ID, time time.Time)
	MemberStopStreaming(userID snowflake.ID, time time.Time)

	Update() error
	Close(t time.Time) error

	IsClosed() bool
}

type handlerImpl struct {
	call           *Call
	rest           rest.Rest
	channelID      snowflake.ID
	messageID      snowflake.ID
	updateCooldown time.Time
	closed         bool
}

func NewHandler(call *Call, rest rest.Rest, now time.Time) (Handler, error) {
	call.OnStart(now)
	message, err := rest.CreateMessage(
		call.Rule.NotificationChannel,
		discord.MessageCreate{
			Embeds: []discord.Embed{call.OngoingEmbed(time.Now())},
		},
	)

	if err != nil {
		return nil, err
	}

	return &handlerImpl{
		call:      call,
		rest:      rest,
		channelID: call.Rule.NotificationChannel,
		messageID: message.ID,
		closed:    false,
	}, nil
}

func (h *handlerImpl) RegisterMember(userID snowflake.ID, member *discord.Member) {
	if h.IsRegistered(userID) {
		panic("member already registered")
	}

	m := NewMember(userID, h.call.Rule.UserFormat.Format(member))
	h.call.Members = append(h.call.Members, m)
	h.call.MemberMap[userID] = m
}

func (h *handlerImpl) IsRegistered(userID snowflake.ID) bool {
	_, ok := h.call.MemberMap[userID]
	return ok
}

func (h *handlerImpl) MemberJoin(userID snowflake.ID, now time.Time) {
	member, ok := h.call.MemberMap[userID]
	if !ok {
		panic("member not registered")
	}

	h.call.Onlines++
	member.MarkAsOnline(now)
}

func (h *handlerImpl) MemberLeave(userID snowflake.ID, now time.Time) bool {
	member, ok := h.call.MemberMap[userID]
	if !ok {
		panic("member not registered")
	}

	h.call.Onlines--
	member.UnmarkAsOnline(now)

	return h.call.Onlines == 0
}

func (h *handlerImpl) MemberStartStreaming(userID snowflake.ID, now time.Time) {
	member, ok := h.call.MemberMap[userID]
	if !ok {
		panic("member not registered")
	}
	member.MarkAsStreaming(now)
}

func (h *handlerImpl) MemberStopStreaming(userID snowflake.ID, now time.Time) {
	member, ok := h.call.MemberMap[userID]
	if !ok {
		panic("member not registered")
	}
	member.UnmarkAsStreaming(now)
}

func (h *handlerImpl) Update() error {
	if h.closed {
		return nil
	}

	now := time.Now()
	if h.updateCooldown.After(now) {
		// update too fast
		return nil
	}

	messageUpdate := discord.NewMessageUpdateBuilder().
		AddEmbeds(h.call.OngoingEmbed(now))

	if h.call.Rule.History.ShouldDisplayTimeline() {
		frame := timeline.GenFrame(h.call.Start, now)
		file, err := h.call.GenerateTimeline(h.rest, now, frame, WithIndicator(now))
		if err != nil {
			fmt.Println("failed to generate timeline:", err)
			return err
		}
		messageUpdate.AddFiles(file)
	}

	_, err := h.rest.UpdateMessage(
		h.channelID,
		h.messageID,
		messageUpdate.Build(),
	)
	h.updateCooldown = time.Now().Add(10 * time.Second)
	return err
}

func (h *handlerImpl) Close(currentTime time.Time) error {
	if h.closed {
		return nil
	}

	h.call.OnEnd(currentTime)

	// update the message to show the call has ended
	// this is IMPORTANT MESSAGE, so we should retry if failed
	go func(currentTime time.Time) {
		defer func() {
			h.call = nil
		}()
		retryInterval := 10 * time.Second
		for retry := 0; retry < 3; retry++ {
			messageUpdate := discord.NewMessageUpdateBuilder().
				AddEmbeds(h.call.EndedEmbed())

			if h.call.Rule.History.ShouldDisplayTimeline() {
				file, err := h.call.GenerateTimeline(h.rest, currentTime, currentTime)
				if err != nil {
					fmt.Println("failed to generate timeline:", err)
					return
				}
				messageUpdate.AddFiles(file)
			}

			_, err := h.rest.UpdateMessage(
				h.channelID,
				h.messageID,
				messageUpdate.Build(),
			)

			if err == nil {
				break
			}
			fmt.Println("failed to update message:", err)
			fmt.Println("retrying in", retryInterval, "s")
			time.Sleep(retryInterval)
			retryInterval *= 2
		}
	}(currentTime)

	h.closed = true

	return nil
}

func (h *handlerImpl) IsClosed() bool {
	return h.closed
}
