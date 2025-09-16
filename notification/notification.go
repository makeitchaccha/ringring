package notification

// Notification represents a notification in the system.
// It is
type Notification struct {
	// VoiceChannelName is the name of the channel where the call is taking place.
	// It can be like plaintext of channel name "general" or channel mention "<#123456789>".
	VoiceChannelName string `json:"call_channel_name"`
}
