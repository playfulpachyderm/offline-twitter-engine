package scraper

type DMID string

type DM struct {
	ID               DMID `db:"id"`
	Time             int
	Request          int
	ConversationID   ConversationID
	RecipientID      UserID
	SenderID         UserID
	Text             string
	MessageReactions []DMReaction
}
