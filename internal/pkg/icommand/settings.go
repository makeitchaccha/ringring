package icommand

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/json"
	"github.com/yuyaprgrm/ringring/internal/pkg/iform"
	"github.com/yuyaprgrm/ringring/internal/pkg/locale"
	"github.com/yuyaprgrm/ringring/internal/pkg/rule"
	"github.com/yuyaprgrm/ringring/pkg/command"
	"github.com/yuyaprgrm/ringring/pkg/form"
)

var _ command.Command = (*Settings)(nil)

type Settings struct {
	Form form.Manager
	Rule rule.Repository
}

func (s *Settings) Name() string {
	return "ringring"
}

func (s *Settings) Create() discord.ApplicationCommandCreate {
	return discord.SlashCommandCreate{
		Name:        s.Name(),
		Description: locale.Get(discord.LocaleEnglishUS).Command.Settings.Description,
		DescriptionLocalizations: locale.Localizations(func(entry locale.Entry) string {
			return entry.Command.Settings.Description
		}),
		DefaultMemberPermissions: json.NewNullablePtr(discord.PermissionManageGuild),

		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "guild",
				Description: locale.Get(discord.LocaleEnglishUS).Command.Settings.SubCommands.Guild.Description,
				DescriptionLocalizations: locale.Localizations(func(entry locale.Entry) string {
					return entry.Command.Settings.SubCommands.Guild.Description
				}),
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "category",
				Description: locale.Get(discord.LocaleEnglishUS).Command.Settings.SubCommands.Category.Description,
				DescriptionLocalizations: locale.Localizations(func(entry locale.Entry) string {
					return entry.Command.Settings.SubCommands.Category.Description
				}),
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionChannel{
						Name:        "category",
						Description: locale.Get(discord.LocaleEnglishUS).Command.Settings.SubCommands.Category.Options.Category.Description,
						DescriptionLocalizations: locale.Localizations(func(entry locale.Entry) string {
							return entry.Command.Settings.SubCommands.Category.Options.Category.Description
						}),
						Required: true,
						ChannelTypes: []discord.ChannelType{
							discord.ChannelTypeGuildCategory,
						},
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "channel",
				Description: locale.Get(discord.LocaleEnglishUS).Command.Settings.SubCommands.Channel.Description,
				DescriptionLocalizations: locale.Localizations(func(entry locale.Entry) string {
					return entry.Command.Settings.SubCommands.Channel.Description
				}),
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionChannel{
						Name:        "channel",
						Description: locale.Get(discord.LocaleEnglishUS).Command.Settings.SubCommands.Channel.Options.Channel.Description,
						DescriptionLocalizations: locale.Localizations(func(entry locale.Entry) string {
							return entry.Command.Settings.SubCommands.Channel.Options.Channel.Description
						}),
						Required: true,
						ChannelTypes: []discord.ChannelType{
							discord.ChannelTypeGuildVoice, discord.ChannelTypeGuildStageVoice,
						},
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "preview",
				Description: "preview the voice channels with how the bot would works",
			},
		},
	}
}

func (s *Settings) Execute(event *events.ApplicationCommandInteractionCreate) error {

	data := event.ApplicationCommandInteraction.SlashCommandInteractionData()

	if data.SubCommandName == nil {
		return fmt.Errorf("subcommand not found")
	}

	if event.GuildID() == nil {
		return fmt.Errorf("command is not available in DM")
	}

	if *data.SubCommandName == "preview" {
		// just send a preview message
		embeds := s.generatePreview(event)

		if len(embeds) == 0 {
			return event.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent("No voice channels found").
				SetEphemeral(true).
				Build(),
			)
		}

		// split the embeds into 10 embeds per message
		err := event.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("Preview of how the bot would works").
			SetEphemeral(true).
			Build(),
		)

		if err != nil {
			return err
		}

		for i := 0; i < len(embeds); i += 10 {
			end := i + 10
			if end > len(embeds) {
				end = len(embeds)
			}

			_, err := event.Client().Rest().CreateMessage(
				event.Channel().ID(),
				discord.NewMessageCreateBuilder().
					SetEmbeds(embeds[i:end]...).
					Build(),
			)

			if err != nil {
				return err
			}
		}

		return nil
	}

	var form *iform.Rule

	switch *data.SubCommandName {
	case "guild":
		form = iform.GuildRule(event.User().ID, s.Rule, event.Locale(), *event.GuildID())
		if rule, ok := s.Rule.FindGuildRule(*event.GuildID()); ok {
			form.HasDeleteButton = true
			form.Apply(rule)
		}

	case "category":
		category := data.Channel("category")
		form = iform.CategoryRule(event.User().ID, s.Rule, event.Locale(), category.ID)
		if rule, ok := s.Rule.FindCategoryRule(category.ID); ok {
			form.HasDeleteButton = true
			form.Apply(rule)
		}

	case "channel":
		channel := data.Channel("channel")
		form = iform.ChannelRule(event.User().ID, s.Rule, event.Locale(), channel.ID)
		if rule, ok := s.Rule.FindChannelRule(channel.ID); ok {
			form.HasDeleteButton = true
			form.Apply(rule)
		}
	}

	err := s.Form.Send(event.Channel().ID(), form)

	if err != nil {
		return err
	}

	return event.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent(locale.Get(event.Locale()).Command.Settings.Response.ShowForm).
		SetEphemeral(true).
		Build(),
	)
}

func (s *Settings) generatePreview(event *events.ApplicationCommandInteractionCreate) []discord.Embed {
	f := locale.Get(event.Locale()).Form.Settings.Fields
	channels, err := event.Client().Rest().GetGuildChannels(*event.GuildID())
	if err != nil {
		return []discord.Embed{
			discord.NewEmbedBuilder().
				SetTitle("failed to get channels, make sure the bot has permission to view channels").
				SetDescription(err.Error()).
				Build(),
		}
	}

	embeds := make([]discord.Embed, 0)
	for _, channel := range channels {
		if channel.Type() != discord.ChannelTypeGuildVoice && channel.Type() != discord.ChannelTypeGuildStageVoice {
			continue
		}

		builder := discord.NewEmbedBuilder()
		rule, scope := s.Rule.ScopedEffectiveRule(*event.GuildID(), channel.ParentID(), channel.ID())

		if !rule.Enabled {
			builder.SetTitlef("❌ %s", channel.Name())
			builder.SetDescription("通知が無効化されています")
			builder.SetColor(0xff0000)
			builder.AddField("スコープ", scope.String(), true)
		} else {
			builder.SetTitlef("✅ %s", channel.Name())
			builder.SetDescription("通知が有効化されています")
			builder.SetColor(0x00ff00)
			builder.AddField("スコープ", scope.String(), true)
			builder.AddField(f.NotificationChannel.Title, discord.ChannelMention(rule.NotificationChannel), true)
			builder.AddField(f.ChannelFormat.Title, f.ChannelFormat.Values[rule.ChannelFormat.String()], true)
			builder.AddField(f.History.Title, f.History.Values[rule.History.String()], true)
			if rule.History.ShouldDisplayName() {
				builder.AddField(f.UsernameFormat.Title, f.UsernameFormat.Values[rule.UserFormat.String()], true)
			}
		}

		embeds = append(embeds, builder.Build())
	}

	return embeds
}
