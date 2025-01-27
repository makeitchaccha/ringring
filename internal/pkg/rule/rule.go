package rule

import "github.com/disgoorg/snowflake/v2"

type Rule struct {
	Enabled             bool
	NotificationChannel snowflake.ID
	History             History
	UserFormat          UserFormat
	ChannelFormat       ChannelFormat
}
