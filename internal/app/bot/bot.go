package bot

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
	"github.com/golang/freetype/truetype"
	"github.com/yuyaprgrm/ringring/internal/pkg/call"
	"github.com/yuyaprgrm/ringring/internal/pkg/icommand"
	"github.com/yuyaprgrm/ringring/internal/pkg/locale"
	"github.com/yuyaprgrm/ringring/internal/pkg/rule"
	"github.com/yuyaprgrm/ringring/pkg/command"
	"github.com/yuyaprgrm/ringring/pkg/form"
	"golang.org/x/image/font/gofont/goregular"
	"gorm.io/gorm"
)

type Bot interface {
	Start(ctx context.Context) error
	Close(ctx context.Context)
}

var _ Bot = (*botImpl)(nil)

type botImpl struct {
	client bot.Client
	font   *truetype.Font

	callManager    call.Manager
	formManager    form.Manager
	ruleManager    rule.Manager
	commandManager command.Manager

	cancelClose map[snowflake.ID]chan<- struct{}
}

type ConfigOpt func(*botImpl)

func WithFont(font *truetype.Font) ConfigOpt {
	return func(b *botImpl) {
		b.font = font
	}
}

func New(token string, db *gorm.DB, opts ...ConfigOpt) (Bot, error) {

	locale.Init("./locales")

	client, err := disgo.New(token,
		bot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagVoiceStates, cache.FlagMembers, cache.FlagGuilds, cache.FlagChannels),
		),
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(gateway.IntentGuilds, gateway.IntentGuildVoiceStates),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	// initialize form manager
	formManager := form.NewManager(client.Rest())
	client.AddEventListeners(bot.NewListenerFunc(formManager.OnComponentInteractionCreate))

	// initialize rule manager
	ruleManager := rule.NewManager(db)

	// initialize command for bot
	commandManager := command.NewManager()
	commandManager.Register(&icommand.Settings{Form: formManager, Rule: ruleManager})
	client.AddEventListeners(bot.NewListenerFunc(commandManager.OnCommandInteractionCreate))

	// initialize call manager
	callManager := call.NewManager(client.Rest())

	font, err := truetype.Parse(goregular.TTF)

	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %w", err)
	}

	b := &botImpl{
		client:         client,
		font:           font,
		callManager:    callManager,
		formManager:    formManager,
		ruleManager:    ruleManager,
		commandManager: commandManager,
		cancelClose:    make(map[snowflake.ID]chan<- struct{}),
	}

	for _, opt := range opts {
		opt(b)
	}

	client.AddEventListeners(bot.NewListenerFunc(b.onVoiceStateUpdate))
	client.AddEventListeners(bot.NewListenerFunc(b.onGuildsReady))
	client.AddEventListeners(bot.NewListenerFunc(b.onGuildJoin))

	return b, nil
}

func (b *botImpl) Start(ctx context.Context) error {
	return b.client.OpenGateway(ctx)
}

func (b *botImpl) Close(ctx context.Context) {
	b.client.Close(ctx)
}

func (b *botImpl) onGuildJoin(event *events.GuildJoin) {
	b.client.Caches().VoiceStatesForEach(event.Guild.ID, func(voiceState discord.VoiceState) {
		member, ok := b.client.Caches().Member(event.Guild.ID, voiceState.UserID)
		if !ok {
			fmt.Fprintln(os.Stderr, "failed to get member")
			return
		}
		b.onJoinVoiceChannel(*voiceState.ChannelID, &member)
	})
}

func (b *botImpl) onGuildsReady(event *events.GuildsReady) {
	fmt.Println("guilds ready")
	b.client.Caches().GuildsForEach(func(guild discord.Guild) {
		fmt.Println("guild:", guild.ID)
		b.client.Caches().VoiceStatesForEach(guild.ID, func(voiceState discord.VoiceState) {
			member, ok := b.client.Caches().Member(guild.ID, voiceState.UserID)
			if !ok {
				fmt.Fprintln(os.Stderr, "failed to get member")
				return
			}
			b.onJoinVoiceChannel(*voiceState.ChannelID, &member)
		})
	})
}

func (b *botImpl) onVoiceStateUpdate(event *events.GuildVoiceStateUpdate) {
	// scenarios:
	// 1. user leaves voice channel (nil <- before id)
	// 2. user joins voice channel (after id <- nil)
	// 3. user moves to another voice channel (after id <- before id, after id != before id)

	fmt.Println("voice state update")
	if event.VoiceState.ChannelID == nil {
		fmt.Println("leave voice channel")
		b.onLeaveVoiceChannel(*event.OldVoiceState.ChannelID, &event.Member)
		return
	}

	if event.OldVoiceState.ChannelID == nil {
		fmt.Println("join voice channel")
		b.onJoinVoiceChannel(*event.VoiceState.ChannelID, &event.Member)
		return
	}

	if *event.VoiceState.ChannelID != *event.OldVoiceState.ChannelID {
		fmt.Println("move voice channel")
		b.onLeaveVoiceChannel(*event.OldVoiceState.ChannelID, &event.Member)
		b.onJoinVoiceChannel(*event.VoiceState.ChannelID, &event.Member)
		return
	}
}

func (b *botImpl) onJoinVoiceChannel(channelID snowflake.ID, member *discord.Member) {
	now := time.Now()
	handler, ok := b.callManager.Get(channelID)
	if !ok {
		fmt.Println("new call candidate detected")
		channel, err := b.client.Rest().GetChannel(channelID)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to get channel:", err)
			return
		}
		guildChannel, ok := channel.(discord.GuildChannel)
		if !ok {
			fmt.Fprintln(os.Stderr, "channel is supposed to be a guild channel")
			return
		}
		rule, scope := b.ruleManager.ScopedEffectiveRule(guildChannel.GuildID(), guildChannel.ParentID(), guildChannel.ID())
		fmt.Println("rule:", rule, "scope:", scope)
		if !rule.Enabled {
			fmt.Println("rule is not enabled, skip")
			return
		}
		handler, err = b.callManager.Add(call.New(discord.LocaleJapanese, rule, channel, b.font), now)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to create call:", err)
			return
		}

		// update message
		go func() {
			for !handler.IsClosed() {
				handler.Update()
				time.Sleep(1 * time.Minute)
			}
		}()
	}
	if !handler.IsRegistered(member.User.ID) {
		handler.RegisterMember(member.User.ID, member)
	}
	handler.MemberJoin(member.User.ID, now)
	handler.Update()
	if cancel, ok := b.cancelClose[channelID]; ok {
		cancel <- struct{}{} // notify to cancel shutdown sequence
		delete(b.cancelClose, channelID)
	}
}

func (b *botImpl) onLeaveVoiceChannel(channelID snowflake.ID, member *discord.Member) {
	now := time.Now()
	handler, ok := b.callManager.Get(channelID)
	if !ok {
		return
	}
	isEmpty := handler.MemberLeave(member.User.ID, now)
	handler.Update()

	if isEmpty {
		// start shutdown sequence
		// the purpose of this is to prevent too many call logs created in a short period
		// if the user rejoins the voice channel, the call will be recreated

		cancel := make(chan struct{})
		b.cancelClose[channelID] = cancel
		go func() {
			channelID := channelID
			now := now
			handler := handler
			cancel := cancel
			wait := 1 * time.Minute
			select {
			case <-time.After(wait):
				handler.Close(now)
				b.callManager.Remove(channelID)
			case <-cancel:
			}

			delete(b.cancelClose, channelID)
		}()
	}
}
