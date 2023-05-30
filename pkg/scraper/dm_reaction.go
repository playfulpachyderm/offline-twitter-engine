package scraper

type DMReactionID int

type DMReaction struct {
	ID             DMReactionID `db:"id"`
	Time           int
	ConversationID ConversationID
	MessageID      DMID
	ReactionKey    string
	SenderID       UserID
}
