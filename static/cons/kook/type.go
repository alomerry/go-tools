package kook

const (
	EventText      = 1
	EventImage     = 2
	EventVideo     = 3
	EventFile      = 4
	EventAudio     = 8
	EventKMarkdown = 9
	EventCard      = 10
	EventSystem    = 255
	SyStemEvent    = 255 // Deprecated: use EventSystem instead
)

const (
	ChannelTypePerson           = "PERSON"
	ChannelTypeGroup            = "GROUP"
	ChannelTypeBroadcast        = "BROADCAST"
	ChannelTypeWebhookChallenge = "WEBHOOK_CHALLENGE"
)
