package notification

import "github.com/disgoorg/disgo/discord"

type UserFormat int

const (
	UserFormatUsername UserFormat = iota
	UserFormatDisplay
	UserFormatMention
)

func (f UserFormat) String() string {
	switch f {
	case UserFormatUsername:
		return "username"
	case UserFormatDisplay:
		return "display"
	case UserFormatMention:
		return "mention"
	default:
		return "unknown"
	}
}

func (f UserFormat) Format(member *discord.Member) string {
	switch f {
	case UserFormatUsername:
		return member.User.Username
	case UserFormatDisplay:
		return member.EffectiveName()
	case UserFormatMention:
		return member.Mention()
	default:
		return "unknown"
	}
}

func ParseUserFormat(s string) UserFormat {
	switch s {
	case "username":
		return UserFormatUsername
	case "display":
		return UserFormatDisplay
	case "mention":
		return UserFormatMention
	default:
		return UserFormat(-1)
	}
}

type ChannelFormat int

const (
	ChannelFormatDisplay ChannelFormat = iota
	ChannelFormatMention
)

func (f ChannelFormat) String() string {
	switch f {
	case ChannelFormatDisplay:
		return "display"
	case ChannelFormatMention:
		return "mention"
	default:
		return "unknown"
	}
}

func (f ChannelFormat) Format(channel discord.Channel) string {
	switch f {
	case ChannelFormatDisplay:
		return channel.Name()
	case ChannelFormatMention:
		return discord.ChannelMention(channel.ID())
	default:
		return "unknown"
	}
}

func ParseChannelFormat(s string) ChannelFormat {
	switch s {
	case "display":
		return ChannelFormatDisplay
	case "mention":
		return ChannelFormatMention
	default:
		return ChannelFormat(-1)
	}
}
