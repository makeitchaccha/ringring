package rule

import (
	"github.com/disgoorg/snowflake/v2"
	"gorm.io/gorm"
)

type RuleModel struct {
	gorm.Model
	Scope               int    `gorm:"primary_key"`
	Identifier          uint64 `gorm:"primary_key"`
	Enabled             bool
	NotificationChannel uint64
	History             string
	UserFormat          string
	ChannelFormat       string
}

func (m RuleModel) toRule() (Scope, snowflake.ID, Rule) {
	return Scope(m.Scope), snowflake.ID(m.Identifier), Rule{
		Enabled:             m.Enabled,
		NotificationChannel: snowflake.ID(m.NotificationChannel),
		History:             ParseHistory(m.History),
		UserFormat:          ParseUserFormat(m.UserFormat),
		ChannelFormat:       ParseChannelFormat(m.ChannelFormat),
	}
}

func newModel(scope Scope, id snowflake.ID, rule Rule) RuleModel {
	return RuleModel{
		Scope:               int(scope),
		Identifier:          uint64(id),
		Enabled:             rule.Enabled,
		NotificationChannel: uint64(rule.NotificationChannel),
		History:             rule.History.String(),
		UserFormat:          rule.UserFormat.String(),
		ChannelFormat:       rule.ChannelFormat.String(),
	}
}
