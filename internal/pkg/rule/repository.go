package rule

import (
	"github.com/disgoorg/snowflake/v2"
	"gorm.io/gorm"
)

type Repository interface {
	// UpdateRule updates the rule for the given guild, category or channel
	SaveRule(scope Scope, id snowflake.ID, rule Rule)
	DeleteRule(scope Scope, id snowflake.ID)

	FindRule(scope Scope, id snowflake.ID) (Rule, bool)
}

var _ Repository = (*repositoryImpl)(nil)

type repositoryImpl struct {
	db *gorm.DB
}

func CreateRepository(db *gorm.DB) Repository {
	mgr := &repositoryImpl{
		db: db,
	}

	db.AutoMigrate(&RuleModel{})

	return mgr
}

func (m *repositoryImpl) SaveRule(scope Scope, id snowflake.ID, rule Rule) {
	model := newModel(scope, id, rule)
	m.db.Save(&model)
}

func (m *repositoryImpl) DeleteRule(scope Scope, id snowflake.ID) {
	m.db.Delete(&RuleModel{}, "scope = ? AND identifier = ?", int(scope), id)
}

func (m *repositoryImpl) ScopedEffectiveRule(guildID snowflake.ID, categoryID *snowflake.ID, channelID snowflake.ID) (Rule, Scope) {
	if rule, ok := m.FindChannelRule(channelID); ok {
		return rule, ScopeChannel
	}
	if categoryID != nil { // categoryID is nil when the channel is not in a category
		if rule, ok := m.FindCategoryRule(*categoryID); ok {
			return rule, ScopeCategory
		}
	}
	if rule, ok := m.FindGuildRule(guildID); ok {
		return rule, ScopeGuild
	}
	return Rule{Enabled: false}, ScopeGuild
}

func (m *repositoryImpl) EffectiveRule(guildID snowflake.ID, categoryID *snowflake.ID, channelID snowflake.ID) Rule {
	rule, _ := m.ScopedEffectiveRule(guildID, categoryID, channelID)
	return rule
}

func (m *repositoryImpl) FindRule(scope Scope, id snowflake.ID) (Rule, bool) {
	switch scope {
	case ScopeGuild:
		return m.FindGuildRule(id)
	case ScopeCategory:
		return m.FindCategoryRule(id)
	case ScopeChannel:
		return m.FindChannelRule(id)
	}
	return Rule{}, false
}

func (m *repositoryImpl) FindGuildRule(guildID snowflake.ID) (Rule, bool) {
	var model RuleModel
	if err := m.db.First(&model, "scope = ? AND identifier = ?", int(ScopeGuild), guildID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return Rule{}, false
		}
		panic(err)
	}
	_, _, rule := model.toRule()
	return rule, true
}

func (m *repositoryImpl) FindCategoryRule(categoryID snowflake.ID) (Rule, bool) {
	var model RuleModel
	if err := m.db.First(&model, "scope = ? AND identifier = ?", int(ScopeCategory), categoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return Rule{}, false
		}
		panic(err)
	}
	_, _, rule := model.toRule()
	return rule, true
}

func (m *repositoryImpl) FindChannelRule(channelID snowflake.ID) (Rule, bool) {
	var model RuleModel
	if err := m.db.First(&model, "scope = ? AND identifier = ?", int(ScopeChannel), channelID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return Rule{}, false
		}
		panic(err)
	}
	_, _, rule := model.toRule()
	return rule, true
}
