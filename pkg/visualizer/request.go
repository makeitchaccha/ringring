package visualizer

type Request struct {
	Users  []User  `json:"users"`
	Layout *Layout `json:"layout,omitempty"`
}

type User struct {
	AvatarLocation string    `json:"avatar_location"`
	Sections       []Section `json:"sections"`
}

type Section struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type Layout struct {
	Padding        TRBL     `json:"padding"`
	EntryHeight    int      `json:"entry_height"`
	HeadlineWidth  int      `json:"headline_width"`
	TimelineWidth  int      `json:"timeline_width"`
	OnlineBarWidth int      `json:"online_bar_width"`
	MainTics       *float64 `json:"main_tics,omitempty"`
	SubTics        *float64 `json:"sub_tics,omitempty"`
}

type TRBL struct {
	Top    int `json:"top"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
	Left   int `json:"left"`
}
