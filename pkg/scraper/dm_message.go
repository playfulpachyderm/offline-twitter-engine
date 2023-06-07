package scraper

type DMMessageID int

type DMMessage struct {
	ID           DMMessageID `db:"id"`
	DMChatRoomID DMChatRoomID
	SenderID     UserID
	SentAt       Timestamp
	RequestID    string
	Text         string
	InReplyToID  DMMessageID
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
