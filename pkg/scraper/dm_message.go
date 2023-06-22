package scraper

type DMMessageID int

type DMReaction struct {
	ID          DMMessageID `db:"id"`
	DMMessageID DMMessageID `db:"message_id"`
	SenderID    UserID      `db:"sender_id"`
	SentAt      Timestamp   `db:"sent_at"`
	Emoji       string      `db:"emoji"`
}

func ParseAPIDMReaction(reacc APIDMReaction) DMReaction {
	ret := DMReaction{}
	ret.ID = DMMessageID(reacc.ID)
	ret.SenderID = UserID(reacc.SenderID)
	ret.SentAt = TimestampFromUnix(int64(reacc.Time))
	ret.Emoji = reacc.Emoji
	return ret
}

type DMMessage struct {
	ID           DMMessageID  `db:"id"`
	DMChatRoomID DMChatRoomID `db:"chat_room_id"`
	SenderID     UserID       `db:"sender_id"`
	SentAt       Timestamp    `db:"sent_at"`
	RequestID    string       `db:"request_id"`
	Text         string       `db:"text"`
	InReplyToID  DMMessageID  `db:"in_reply_to_id"`
	Reactions    []DMReaction
}

func ParseAPIDMMessage(message APIDMMessage) DMMessage {
	ret := DMMessage{}
	ret.ID = DMMessageID(message.ID)
	ret.SentAt = TimestampFromUnix(int64(message.Time))
	ret.DMChatRoomID = DMChatRoomID(message.ConversationID)
	ret.SenderID = UserID(message.MessageData.SenderID)
	ret.Text = message.MessageData.Text

	ret.InReplyToID = DMMessageID(message.MessageData.ReplyData.ID) // Will be "0" if not a reply

	ret.Reactions = []DMReaction{}
	for _, api_reacc := range message.MessageReactions {
		reacc := ParseAPIDMReaction(api_reacc)
		reacc.DMMessageID = ret.ID
		ret.Reactions = append(ret.Reactions, reacc)
	}
	return ret
}
