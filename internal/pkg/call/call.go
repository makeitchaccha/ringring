package call

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/golang/freetype/truetype"
	"github.com/makeitchaccha/design/timeline"
	"github.com/yuyaprgrm/ringring/internal/pkg/cache"
	"github.com/yuyaprgrm/ringring/internal/pkg/locale"
	"github.com/yuyaprgrm/ringring/internal/pkg/rule"
)

type Call struct {
	Locale      discord.Locale
	Font        *truetype.Font
	Rule        rule.Rule
	ChannelID   snowflake.ID
	ChannelName string
	Start       time.Time
	End         time.Time
	Members     []*Member
	MemberMap   map[snowflake.ID]*Member
	Onlines     int
}

func New(locale discord.Locale, rule rule.Rule, channel discord.Channel, font *truetype.Font) *Call {
	return &Call{
		Locale:      locale,
		Font:        font,
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

	// testing ...
	if c.Rule.History.ShouldDisplayTimeline() {
		builder.SetImage("attachment://thumbnail.png")
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

type GenerateOptions func(b *timeline.TimelineBuilder)

func WithIndicator(indicator time.Time) GenerateOptions {
	return func(b *timeline.TimelineBuilder) {
		b.SetIndicator(indicator)
	}
}

func (c *Call) GenerateTimeline(rest rest.Rest, now time.Time, frame time.Time, opts ...GenerateOptions) (*discord.File, error) {

	// generate timeline
	builder := timeline.NewTimelineBuilder(c.Start, frame)

	for _, opt := range opts {
		opt(builder)
	}

	for _, m := range c.Members {
		avatar, err := cache.GetAvatar(rest, m.id)
		if err != nil {
			return nil, err
		}
		e := timeline.NewEntryBuilder(avatar)
		onlineSeries := timeline.NewSeriesBuilder(2.0/7.0, nil)
		for _, log := range m.onlineSections {
			if log.end.IsZero() {
				log.end = now
			}
			alpha := 1.0
			if log.mute {
				alpha = 0.8
			}
			if log.deaf {
				alpha = 0.6
			}
			onlineSeries.AddSection(log.start, log.end, timeline.WithAlpha(alpha))
		}
		e.AddSeries(onlineSeries.Build())
		if m.HasStreamed() {
			streamingSeries := timeline.NewSeriesBuilder(0.5/7.0, color.RGBA64{R: 0xB000, G: 0x0000, B: 0x0000, A: 0xFFFF})
			for _, log := range m.streamingSections {
				if log.end.IsZero() {
					log.end = now
				}
				streamingSeries.AddSection(log.start, log.end)
			}
			e.AddSeries(streamingSeries.Build())
		}
		builder.AddEntries(e.Build())
	}

	// prevent generated timeline from being too long in the vertical direction
	// w:h = 4:3 is used in the thumbnail
	layout := timeline.DefaultLayout()
	if h, w := builder.Height(), builder.Width(); 4*h > 3*w {
		// adjust width to fit the timeline
		dw := 4*h/3 - w
		layout.TimelineWidth += dw
	}

	builder.SetLayout(layout)

	// font modification
	ticsFont := truetype.NewFace(c.Font, &truetype.Options{
		Size: 15,
	})

	builder.MainTics.Label.Font = ticsFont
	builder.SubTics.Label.Font = ticsFont

	r := builder.Build().Generate()

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
