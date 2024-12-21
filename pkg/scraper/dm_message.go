package scraper

type DMMessageID int

type DMReaction struct {
	ID          DMMessageID `db:"id"`
	DMMessageID DMMessageID `db:"message_id"`
	SenderID    UserID      `db:"sender_id"`
	SentAt      Timestamp   `db:"sent_at"`
	Emoji       string      `db:"emoji"`
}

type DMMessage struct {
	ID              DMMessageID  `db:"id"`
	DMChatRoomID    DMChatRoomID `db:"chat_room_id"`
	SenderID        UserID       `db:"sender_id"`
	SentAt          Timestamp    `db:"sent_at"`
	RequestID       string       `db:"request_id"`
	Text            string       `db:"text"`
	InReplyToID     DMMessageID  `db:"in_reply_to_id"`
	EmbeddedTweetID TweetID      `db:"embedded_tweet_id"`
	Reactions       map[UserID]DMReaction

	Images []Image
	Videos []Video
	Urls   []Url

	LastReadEventUserIDs []UserID // Used for rendering
}
