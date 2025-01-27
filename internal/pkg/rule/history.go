package rule

type History int

const (
	HistoryNone History = iota
	HistoryNameOnly
	HistoryNameWithDuration
	HistoryNameWithDurationAndTimeline
)

func (m History) String() string {
	switch m {
	case HistoryNone:
		return "none"
	case HistoryNameOnly:
		return "name"
	case HistoryNameWithDuration:
		return "name_with_duration"
	case HistoryNameWithDurationAndTimeline:
		return "name_with_duration_and_timeline"
	default:
		return "unknown"
	}
}

func ParseHistory(s string) History {
	switch s {
	case "none":
		return HistoryNone
	case "name":
		return HistoryNameOnly
	case "name_with_duration":
		return HistoryNameWithDuration
	case "name_with_duration_and_timeline":
		return HistoryNameWithDurationAndTimeline
	default:
		return History(-1)
	}
}

func (m History) ShouldDisplayName() bool {
	return m == HistoryNameOnly || m == HistoryNameWithDuration || m == HistoryNameWithDurationAndTimeline
}

func (m History) ShouldDisplayDuration() bool {
	return m == HistoryNameWithDuration || m == HistoryNameWithDurationAndTimeline
}

func (m History) ShouldDisplayTimeline() bool {
	return m == HistoryNameWithDurationAndTimeline
}
