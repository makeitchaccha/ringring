package rule

import (
	"github.com/disgoorg/snowflake/v2"
	"gorm.io/gorm"
)

type Manager interface {
	// UpdateRule updates the rule for the given guild, category or channel
	SetRule(scope Scope, id snowflake.ID, rule Rule)
	RemoveRule(scope Scope, id snowflake.ID)

	// ScopedEffectiveRule returns the rule for the specifier, and the kind of scope it was found in
	ScopedEffectiveRule(guildID snowflake.ID, categoryID *snowflake.ID, channelID snowflake.ID) (Rule, Scope)
	// EffectiveRule returns the rule for the given specifier.
	EffectiveRule(guildID snowflake.ID, categoryID *snowflake.ID, channelID snowflake.ID) Rule

	Rule(scope Scope, id snowflake.ID) (Rule, bool)
	GuildRule(guildID snowflake.ID) (Rule, bool)
	CategoryRule(categoryID snowflake.ID) (Rule, bool)
	ChannelRule(channelID snowflake.ID) (Rule, bool)
}

var _ Manager = (*managerImpl)(nil)

type managerImpl struct {
	db         *gorm.DB
	guilds     map[snowflake.ID]Rule
	categories map[snowflake.ID]Rule
	channels   map[snowflake.ID]Rule
}

func NewManager(db *gorm.DB) Manager {
	mgr := &managerImpl{
		db:         db,
		guilds:     make(map[snowflake.ID]Rule),
		categories: make(map[snowflake.ID]Rule),
		channels:   make(map[snowflake.ID]Rule),
	}

	db.AutoMigrate(&RuleModel{})

	var rules []RuleModel
	mgr.db.Find(&rules)

	for _, rule := range rules {
		scope, id, r := rule.toRule()
		mgr.setRuleInternal(scope, id, r)
	}

	return mgr
}

func (m *managerImpl) SetRule(scope Scope, id snowflake.ID, rule Rule) {
	m.setRuleInternal(scope, id, rule)
	model := newModel(scope, id, rule)
	m.db.Save(&model)
}

func (m *managerImpl) setRuleInternal(scope Scope, id snowflake.ID, rule Rule) {
	switch scope {
	case ScopeGuild:
		m.guilds[id] = rule
	case ScopeCategory:
		m.categories[id] = rule
	case ScopeChannel:
		m.channels[id] = rule
	}
}

func (m *managerImpl) RemoveRule(scope Scope, id snowflake.ID) {
	m.removeRuleInternal(scope, id)
	m.db.Delete(&RuleModel{}, "scope = ? AND identifier = ?", int(scope), id)
}

func (m *managerImpl) removeRuleInternal(scope Scope, id snowflake.ID) {
	switch scope {
	case ScopeGuild:
		delete(m.guilds, id)
	case ScopeCategory:
		delete(m.categories, id)
	case ScopeChannel:
		delete(m.channels, id)
	}
	m.db.Delete(&RuleModel{}, "scope = ? AND identifier = ?", int(scope), id)
}

func (m *managerImpl) ScopedEffectiveRule(guildID snowflake.ID, categoryID *snowflake.ID, channelID snowflake.ID) (Rule, Scope) {
	if rule, ok := m.channels[channelID]; ok {
		return rule, ScopeChannel
	}
	if categoryID != nil {
		if rule, ok := m.categories[*categoryID]; ok {
			return rule, ScopeCategory
		}
	}
	if rule, ok := m.guilds[guildID]; ok {
		return rule, ScopeGuild
	}
	return Rule{Enabled: false}, ScopeGuild
}

func (m *managerImpl) EffectiveRule(guildID snowflake.ID, categoryID *snowflake.ID, channelID snowflake.ID) Rule {
	rule, _ := m.ScopedEffectiveRule(guildID, categoryID, channelID)
	return rule
}

func (m *managerImpl) Rule(scope Scope, id snowflake.ID) (Rule, bool) {
	switch scope {
	case ScopeGuild:
		return m.GuildRule(id)
	case ScopeCategory:
		return m.CategoryRule(id)
	case ScopeChannel:
		return m.ChannelRule(id)
	}
	return Rule{}, false
}

func (m *managerImpl) GuildRule(guildID snowflake.ID) (Rule, bool) {
	rule, ok := m.guilds[guildID]
	return rule, ok
}

func (m *managerImpl) CategoryRule(categoryID snowflake.ID) (Rule, bool) {
	rule, ok := m.categories[categoryID]
	return rule, ok
}

func (m *managerImpl) ChannelRule(channelID snowflake.ID) (Rule, bool) {
	rule, ok := m.channels[channelID]
	return rule, ok
}
