package iform

import (
	"errors"
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/makeitchaccha/ringring/internal/pkg/i18n"
	"github.com/makeitchaccha/ringring/internal/pkg/rule"
	"github.com/makeitchaccha/ringring/pkg/extstd"
	"github.com/makeitchaccha/ringring/pkg/form"
)

var _ form.Form = (*Rule)(nil)

type confirm int

const (
	confirmNone confirm = iota
	confirmSave
	confirmDelete
)

type Rule struct {
	owner       snowflake.ID
	ruleManager rule.Repository
	locale      discord.Locale

	HasDeleteButton bool
	Finalized       bool

	confirm confirm

	Scope           rule.Scope
	ScopeIdentifier snowflake.ID

	Enabled             form.Bool
	NotificationChannel extstd.Option[snowflake.ID]
	ChannelFormat       extstd.Option[rule.ChannelFormat]

	Privacy        extstd.Option[rule.History]
	UsernameFormat extstd.Option[rule.UserFormat]
}

const (
	settingKeyEnabled             = "e"
	settingKeyNotificationChannel = "nc"
	settingKeyUsernameFormat      = "uf"
	settingKeyChannelFormat       = "cf"
	settingKeyPrivacy             = "p"

	settingButtonSave    = "bs"
	settingButtonDiscard = "bdc"
	settingButtonDelete  = "bdl"

	settingButtonConfirmSave   = "bcs"
	settingButtonConfirmDelete = "bcd"
	settingButtonCancel        = "bc"
)

func GuildRule(owner snowflake.ID, ruleManager rule.Repository, locale discord.Locale, guildID snowflake.ID) *Rule {
	return &Rule{
		owner:           owner,
		ruleManager:     ruleManager,
		locale:          locale,
		Scope:           rule.ScopeGuild,
		ScopeIdentifier: guildID,

		Enabled:             true,
		NotificationChannel: extstd.None[snowflake.ID](),
		UsernameFormat:      extstd.None[rule.UserFormat](),
		ChannelFormat:       extstd.None[rule.ChannelFormat](),
		Privacy:             extstd.None[rule.History](),
	}
}

func CategoryRule(owner snowflake.ID, ruleManager rule.Repository, locale discord.Locale, categoryId snowflake.ID) *Rule {
	return &Rule{
		owner:           owner,
		ruleManager:     ruleManager,
		locale:          locale,
		Scope:           rule.ScopeCategory,
		ScopeIdentifier: categoryId,

		Enabled:             true,
		NotificationChannel: extstd.None[snowflake.ID](),
		UsernameFormat:      extstd.None[rule.UserFormat](),
		ChannelFormat:       extstd.None[rule.ChannelFormat](),
		Privacy:             extstd.None[rule.History](),
	}
}

func ChannelRule(owner snowflake.ID, ruleManager rule.Repository, locale discord.Locale, channelId snowflake.ID) *Rule {
	return &Rule{
		owner:           owner,
		ruleManager:     ruleManager,
		locale:          locale,
		Scope:           rule.ScopeChannel,
		ScopeIdentifier: channelId,

		Enabled:             true,
		NotificationChannel: extstd.None[snowflake.ID](),
		UsernameFormat:      extstd.None[rule.UserFormat](),
		ChannelFormat:       extstd.None[rule.ChannelFormat](),
		Privacy:             extstd.None[rule.History](),
	}
}

func (s *Rule) Apply(rule rule.Rule) {
	s.Enabled = form.Bool(rule.Enabled)
	s.NotificationChannel = extstd.Some(rule.NotificationChannel)
	s.ChannelFormat = extstd.Some(rule.ChannelFormat)
	s.Privacy = extstd.Some(rule.History)
	s.UsernameFormat = extstd.Some(rule.UserFormat)
}

func (s *Rule) Create() discord.MessageCreate {
	builder := discord.NewMessageCreateBuilder().
		SetEmbeds(
			s.buildEmbed(""),
		)

	if !s.Finalized {
		builder.AddContainerComponents(s.buildComponents()...)
	}

	return builder.Build()
}

func (s *Rule) update(status string) discord.MessageUpdate {
	builder := discord.NewMessageUpdateBuilder().
		SetEmbeds(
			s.buildEmbed(status),
		)

	if !s.Finalized {
		builder.SetContainerComponents(s.buildComponents()...)
	} else {
		builder.SetContainerComponents()
	}

	return builder.Build()
}

func (s *Rule) buildEmbed(status string) discord.Embed {
	e := i18n.Get(s.locale).Form.Settings.Fields
	builder := discord.NewEmbedBuilder().
		SetTitle(s.title()).
		SetDescription(s.description()).
		AddField(e.Notification.Title, e.Notification.Values[s.Enabled.String()], true).
		AddField(e.NotificationChannel.Title, discord.ChannelMention(s.NotificationChannel.UnwrapOr(0)), true).
		AddField(e.ChannelFormat.Title, e.ChannelFormat.Values[s.ChannelFormat.UnwrapOr(-1).String()], true).
		AddField(e.History.Title, e.History.Values[s.Privacy.UnwrapOr(-1).String()], true).
		AddField(e.UsernameFormat.Title, e.UsernameFormat.Values[s.UsernameFormat.UnwrapOr(-1).String()], true)

	if status != "" {
		builder.SetFooterText(status)
	}

	return builder.Build()
}

func (s *Rule) title() string {
	switch s.Scope {
	case rule.ScopeGuild:
		return i18n.Get(s.locale).Form.Settings.Title.Guild
	case rule.ScopeCategory:
		return fmt.Sprintf(i18n.Get(s.locale).Form.Settings.Title.Category, discord.ChannelMention(s.ScopeIdentifier))
	case rule.ScopeChannel:
		return fmt.Sprintf(i18n.Get(s.locale).Form.Settings.Title.Channel, discord.ChannelMention(s.ScopeIdentifier))
	}
	return ""
}

func (s *Rule) description() string {
	switch s.Scope {
	case rule.ScopeGuild:
		return i18n.Get(s.locale).Form.Settings.Description.Guild
	case rule.ScopeCategory:
		return i18n.Get(s.locale).Form.Settings.Description.Category
	case rule.ScopeChannel:
		return i18n.Get(s.locale).Form.Settings.Description.Channel
	}
	return ""
}

func (s *Rule) buildComponents() []discord.ContainerComponent {

	switch s.confirm {
	case confirmSave:
		return s.buildConfirmSaveComponents()
	case confirmDelete:
		return s.buildConfirmDeleteComponents()
	}

	f := i18n.Get(s.locale).Form.Settings.Fields

	channel := discord.NewChannelSelectMenu(settingKeyNotificationChannel, f.NotificationChannel.Title).
		WithChannelTypes(discord.ChannelTypeGuildText).
		WithMaxValues(1).
		WithMinValues(1)
	if s.NotificationChannel.IsSome() {
		channel = channel.AddDefaultValue(s.NotificationChannel.Unwrap())
	}

	localizedOption := func(values map[string]string) func(key string) discord.StringSelectMenuOption {
		return func(key string) discord.StringSelectMenuOption {
			return discord.NewStringSelectMenuOption(values[key], key)
		}
	}

	markAsDefault := func(value string) func(option discord.StringSelectMenuOption) discord.StringSelectMenuOption {
		return func(option discord.StringSelectMenuOption) discord.StringSelectMenuOption {
			if option.Value == value {
				return option.WithDefault(true)
			}
			return option
		}
	}

	markAsDefaultPrivacy := markAsDefault(s.Privacy.UnwrapOr(-1).String())
	privacyLocalizedOption := localizedOption(f.History.Values)
	history := discord.
		NewStringSelectMenu(
			settingKeyPrivacy, f.History.Title,
			markAsDefaultPrivacy(privacyLocalizedOption(rule.HistoryNone.String())),
			markAsDefaultPrivacy(privacyLocalizedOption(rule.HistoryNameOnly.String())),
			markAsDefaultPrivacy(privacyLocalizedOption(rule.HistoryNameWithDuration.String())),
			markAsDefaultPrivacy(privacyLocalizedOption(rule.HistoryNameWithDurationAndTimeline.String())),
		).
		WithMinValues(1).
		WithMaxValues(1)

	markAsDefaultUsernameFormat := markAsDefault(s.UsernameFormat.UnwrapOr(-1).String())
	usernameLocalizedOption := localizedOption(f.UsernameFormat.Values)

	usernameFormat := discord.
		NewStringSelectMenu(
			settingKeyUsernameFormat, f.UsernameFormat.Title,
			markAsDefaultUsernameFormat(usernameLocalizedOption(rule.UserFormatUsername.String())),
			markAsDefaultUsernameFormat(usernameLocalizedOption(rule.UserFormatDisplay.String())),
			markAsDefaultUsernameFormat(usernameLocalizedOption(rule.UserFormatMention.String())),
		).
		WithMinValues(1).
		WithMaxValues(1)

	if !s.Privacy.UnwrapOr(-1).ShouldDisplayName() {
		usernameFormat = usernameFormat.AsDisabled()
	}

	markAsDefaultChannelFormat := markAsDefault(s.ChannelFormat.UnwrapOr(-1).String())
	channelFormatLocalizedOption := localizedOption(f.ChannelFormat.Values)

	channelFormat := discord.
		NewStringSelectMenu(
			settingKeyChannelFormat, f.ChannelFormat.Title,
			markAsDefaultChannelFormat(channelFormatLocalizedOption(rule.ChannelFormatDisplay.String())),
			markAsDefaultChannelFormat(channelFormatLocalizedOption(rule.ChannelFormatMention.String())),
		).
		WithMinValues(1).
		WithMaxValues(1)

	if !s.Enabled {
		channel = channel.AsDisabled()
		history = history.AsDisabled()
		usernameFormat = usernameFormat.AsDisabled()
		channelFormat = channelFormat.AsDisabled()
	}

	menus := []discord.ContainerComponent{
		discord.NewActionRow(channel),
		discord.NewActionRow(channelFormat),
		discord.NewActionRow(history),
		discord.NewActionRow(usernameFormat),
	}

	// enable/disable, save, discard, delete buttons
	b := i18n.Get(s.locale).Form.Settings.Buttons

	toggle := discord.NewPrimaryButton(b.ToggleEnability[(!s.Enabled).String()], settingKeyEnabled)
	save := discord.NewSuccessButton(b.Save.Primary, settingButtonSave)
	discard := discord.NewSecondaryButton(b.Discard, settingButtonDiscard)
	delete := discord.NewDangerButton(b.Delete.Primary, settingButtonDelete)

	buttonRow := discord.NewActionRow().
		AddComponents(toggle, save, discard)

	if s.HasDeleteButton {
		buttonRow = buttonRow.AddComponents(delete)
	}

	return append([]discord.ContainerComponent{
		buttonRow,
	}, menus...)

}

func (s *Rule) buildConfirmSaveComponents() []discord.ContainerComponent {
	b := i18n.Get(s.locale).Form.Settings.Buttons.Save

	confirm := discord.NewSuccessButton(b.Confirm, settingButtonConfirmSave)
	cancel := discord.NewSecondaryButton(b.Cancel, settingButtonCancel)

	return []discord.ContainerComponent{
		discord.NewActionRow(confirm, cancel),
	}
}

func (s *Rule) buildConfirmDeleteComponents() []discord.ContainerComponent {
	b := i18n.Get(s.locale).Form.Settings.Buttons.Delete

	confirm := discord.NewDangerButton(b.Confirm, settingButtonConfirmDelete)
	cancel := discord.NewSecondaryButton(b.Cancel, settingButtonCancel)

	return []discord.ContainerComponent{
		discord.NewActionRow(confirm, cancel),
	}
}

func (s *Rule) Handle(event *events.ComponentInteractionCreate) error {

	if s.owner != event.User().ID {
		event.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent(i18n.Get(s.locale).Form.Settings.Error.NotOwner).
			SetEphemeral(true).
			Build(),
		)
	}

	e := i18n.Get(s.locale).Form.Settings.Fields

	switch event.Data.CustomID() {

	case settingKeyEnabled:
		s.Enabled = !s.Enabled
		if err := event.UpdateMessage(s.update(fmt.Sprintf(e.Notification.Update, e.Notification.Values[s.Enabled.String()]))); err != nil {
			return err
		}

	case settingKeyNotificationChannel:
		value := event.ChannelSelectMenuInteractionData().Values[0]
		s.NotificationChannel = extstd.Some(value)
		channel, err := event.Client().Rest().GetChannel(value)
		if err != nil {
			return err
		}
		if err := event.UpdateMessage(s.update(fmt.Sprintf(e.NotificationChannel.Update, "#"+channel.Name()))); err != nil {
			return err
		}

	case settingKeyUsernameFormat:
		value := event.StringSelectMenuInteractionData().Values[0]
		s.UsernameFormat = extstd.Some(rule.ParseUserFormat(value))
		if err := event.UpdateMessage(s.update(fmt.Sprintf(e.UsernameFormat.Update, e.UsernameFormat.Values[value]))); err != nil {
			return err
		}

	case settingKeyChannelFormat:
		value := event.StringSelectMenuInteractionData().Values[0]
		s.ChannelFormat = extstd.Some(rule.ParseChannelFormat(value))
		if err := event.UpdateMessage(s.update(fmt.Sprintf(e.ChannelFormat.Update, e.ChannelFormat.Values[value]))); err != nil {
			return err
		}

	case settingKeyPrivacy:
		value := event.StringSelectMenuInteractionData().Values[0]
		s.Privacy = extstd.Some(rule.ParseHistory(value))
		if err := event.UpdateMessage(s.update(fmt.Sprintf(e.History.Update, e.History.Values[value]))); err != nil {
			return err
		}

		// button interaction
	case settingButtonSave:
		if err := s.validate(); err != nil {
			return event.UpdateMessage(s.update(err.Error()))
		}
		s.confirm = confirmSave
		return event.UpdateMessage(s.update(i18n.Get(s.locale).Form.Settings.Buttons.Save.ConfirmStatus))
	case settingButtonConfirmSave:
		if err := s.validate(); err != nil {
			return event.UpdateMessage(s.update(err.Error()))
		}

		s.Finalized = true
		s.confirm = confirmNone

		r := rule.Rule{}

		if s.Enabled {
			r = rule.Rule{
				Enabled:             true,
				NotificationChannel: s.NotificationChannel.Unwrap(),
				ChannelFormat:       s.ChannelFormat.Unwrap(),
				History:             s.Privacy.Unwrap(),
				UserFormat:          s.UsernameFormat.Unwrap(),
			}
		}

		s.ruleManager.SaveRule(
			s.Scope,
			s.ScopeIdentifier,
			r,
		)
		return event.UpdateMessage(s.update(i18n.Get(s.locale).Form.Settings.Validate.Success))

	case settingButtonDiscard:
		return event.UpdateMessage(discord.NewMessageUpdateBuilder().SetContent("discard").SetEmbeds().SetContainerComponents().Build())

	case settingButtonDelete:
		s.confirm = confirmDelete
		return event.UpdateMessage(s.update(i18n.Get(s.locale).Form.Settings.Buttons.Delete.ConfirmStatus))
	case settingButtonConfirmDelete:
		s.ruleManager.DeleteRule(s.Scope, s.ScopeIdentifier)
		return event.UpdateMessage(discord.NewMessageUpdateBuilder().SetContent("deleted").SetEmbeds().SetContainerComponents().Build())

	case settingButtonCancel:
		s.confirm = confirmNone
		return event.UpdateMessage(s.update(""))
	}

	return nil
}

func (s *Rule) validate() error {
	if !s.Enabled {
		// if disabled, no need to validate more
		return nil
	}

	messages := []string{}

	if s.NotificationChannel.IsNone() {
		messages = append(messages, i18n.Get(s.locale).Form.Settings.Validate.Error.NoNotificationChannel)
	}

	if s.ChannelFormat.IsNone() {
		messages = append(messages, i18n.Get(s.locale).Form.Settings.Validate.Error.NoChannelFormat)
	}

	if s.Privacy.IsNone() {
		messages = append(messages, i18n.Get(s.locale).Form.Settings.Validate.Error.NoPrivacy)
	} else {
		if s.Privacy.Unwrap().ShouldDisplayName() && s.UsernameFormat.IsNone() {
			messages = append(messages, i18n.Get(s.locale).Form.Settings.Validate.Error.NoUsernameFormat)
		}
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, "\n"))
	}

	return nil
}
