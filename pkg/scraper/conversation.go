package scraper

type ConversationID string

type Conversation struct {
	ID                    ConversationID
	Type                  string
	SortEventID           int
	SortTimestamp         int
	Participants          []User
	Nsfw                  bool
	NotificationsDisabled bool
	LastReadEventId       int
	ReadOnly              bool
	Trusted               bool
	LowQuality            bool
	Muted                 bool
}
