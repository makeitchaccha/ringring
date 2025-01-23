package rule

type Scope int

const (
	ScopeGuild Scope = iota
	ScopeCategory
	ScopeChannel
)

func (s Scope) String() string {
	switch s {
	case ScopeGuild:
		return "guild"
	case ScopeCategory:
		return "category"
	case ScopeChannel:
		return "channel"
	default:
		return "unknown"
	}
}
