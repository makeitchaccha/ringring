package call

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/yuyaprgrm/ringring/internal/pkg/locale"
	"github.com/yuyaprgrm/ringring/internal/pkg/rule"
	"github.com/yuyaprgrm/ringring/pkg/visualizer"
)

type Call struct {
	Locale      discord.Locale
	Rule        rule.Rule
	ChannelID   snowflake.ID
	ChannelName string
	Start       time.Time
	End         time.Time
	Members     []*Member
	MemberMap   map[snowflake.ID]*Member
	Onlines     int
}

func New(locale discord.Locale, rule rule.Rule, channel discord.Channel) *Call {
	return &Call{
		Locale:      locale,
		Rule:        rule,
		ChannelID:   channel.ID(),
		ChannelName: rule.ChannelFormat.Format(channel),
		Members:     make([]*Member, 0),
		MemberMap:   make(map[snowflake.ID]*Member),
		Onlines:     0,
	}
}

func (c *Call) OnStart(now time.Time) {
	c.Start = now
}

func (c *Call) OnEnd(now time.Time) {
	c.End = now
}

func (c *Call) elapsed(now time.Time) time.Duration {
	return now.Sub(c.Start)
}

func (c *Call) OngoingEmbed(now time.Time) discord.Embed {
	n := locale.Get(c.Locale).Notification
	builder := discord.NewEmbedBuilder().
		SetTitle(n.Ongoing.Title).
		SetDescriptionf(n.Ongoing.Description, c.ChannelName).
		SetColor(0x547443).
		AddField(n.Common.StartTime, discord.FormattedTimestampMention(c.Start.Unix(), discord.TimestampStyleShortTime), true).
		AddField(n.Common.TimeElapsed, localizeDuration(c.Locale, c.elapsed(now), false), true)

	if c.Rule.History.ShouldDisplayName() {
		builder.AddField(n.Common.History, c.history(now), false)
	}

	return builder.Build()
}

func (c *Call) EndedEmbed() discord.Embed {
	n := locale.Get(c.Locale).Notification
	builder := discord.NewEmbedBuilder().
		SetTitle(n.Ended.Title).
		SetDescriptionf(n.Ended.Description, c.ChannelName).
		SetColor(0x547443).
		AddField(n.Common.StartTime, discord.FormattedTimestampMention(c.Start.Unix(), discord.TimestampStyleShortTime), true).
		AddField(n.Common.EndTime, discord.FormattedTimestampMention(c.End.Unix(), discord.TimestampStyleShortTime), true).
		AddField(n.Common.TimeElapsed, localizeDuration(c.Locale, c.elapsed(c.End), false), true)

	if c.Rule.History.ShouldDisplayName() {
		builder.AddField(n.Common.History, c.history(c.End), false)
	}

	if c.Rule.History.ShouldDisplayTimeline() {
		builder.SetImage("attachment://thumbnail.png")
	}

	return builder.Build()
}

func (c *Call) GenerateTimeline(rest rest.Rest) (*discord.File, error) {
	// fetch user avatars
	// we should cache the avatars to avoid rate limit
	// but done is better than perfect.
	for userID := range c.MemberMap {
		user, _ := rest.GetUser(userID)
		resp, err := http.Get(user.EffectiveAvatarURL(discord.WithSize(64), discord.WithFormat(discord.FileFormatPNG)))
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to fetch avatar:", err)
			continue
		}
		defer resp.Body.Close()
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to read body:", err)
			continue
		}
		os.WriteFile(filepath.Join("avatars", user.ID.String()+".png"), buf, 0644)
	}

	// generate timeline
	total := c.End.Sub(c.Start).Seconds()

	main := (24 * time.Hour).Seconds() / total
	sub := (1 * time.Hour).Seconds() / total

	request := visualizer.Request{}
	request.Layout = &visualizer.Layout{
		Padding:        visualizer.TRBL{Top: 10, Right: 10, Bottom: 10, Left: 10},
		EntryHeight:    70,
		HeadlineWidth:  100,
		TimelineWidth:  900,
		OnlineBarWidth: 20,
		MainTics:       &main,
		SubTics:        &sub,
	}
	for _, u := range c.Members {
		user := visualizer.User{}
		user.AvatarLocation = filepath.Join("avatars", u.id.String()+".png")
		for _, log := range u.logs {
			section := visualizer.Section{
				Start: log.join.Sub(c.Start).Seconds() / total,
				End:   log.leave.Sub(c.Start).Seconds() / total,
			}
			user.Sections = append(user.Sections, section)
		}
		request.Users = append(request.Users, user)
	}

	filename := fmt.Sprintf("timelines/%s.png", snowflake.New(time.Now()))
	err := visualizer.Generate(request, filename)

	if err != nil {
		return nil, fmt.Errorf("failed to generate timeline: %w", err)
	}

	r, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return &discord.File{
		Name:   "thumbnail.png",
		Reader: r,
	}, nil
}

func (c *Call) history(now time.Time) string {
	var sb strings.Builder
	for _, m := range c.Members {
		sb.WriteString(m.name)
		if c.Rule.History.ShouldDisplayDuration() {
			sb.WriteString(" (")
			sb.WriteString(localizeDuration(c.Locale, m.calculateDuration(now), true))
			sb.WriteString(")")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func localizeDuration(l discord.Locale, d time.Duration, withSecond bool) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	t := locale.Get(l).Notification.Common.Timeformat

	var sb strings.Builder
	if days > 0 {
		sb.WriteString(fmt.Sprintf("%d%s", days, t.Days))
	}
	if days > 0 || hours > 0 {
		sb.WriteString(fmt.Sprintf("%d%s", hours, t.Hours))
	}
	if days > 0 || hours > 0 || minutes > 0 || !withSecond {
		sb.WriteString(fmt.Sprintf("%d%s", minutes, t.Minutes))
	}
	if withSecond {
		sb.WriteString(fmt.Sprintf("%d%s", seconds, t.Seconds))
	}

	return sb.String()
}
