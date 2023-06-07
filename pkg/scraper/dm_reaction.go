package scraper

type DMReaction struct {
	ID          DMMessageID `db:"id"`
	DMMessageID DMMessageID
	SenderID    UserID
	SentAt      Timestamp
	Emoji       string
}

func ParseAPIDMReaction(reacc APIDMReaction) DMReaction {
	ret := DMReaction{}
	ret.ID = DMMessageID(reacc.ID)
	ret.SenderID = UserID(reacc.SenderID)
	ret.SentAt = TimestampFromUnix(int64(reacc.Time))
	ret.Emoji = reacc.Emoji
	return ret
}
