package format

type Participant int

const (
	ParticipantUsername Participant = iota
	ParticipantDisplay
	ParticipantMention
)

func (p Participant) String() string {
	switch p {
	case ParticipantUsername:
		return "username"
	case ParticipantDisplay:
		return "display"
	case ParticipantMention:
		return "mention"
	default:
		return "unknown"
	}
}
