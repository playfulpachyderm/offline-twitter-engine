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
	ret.SentAt = TimestampFromUnixMilli(int64(reacc.Time))
	ret.Emoji = reacc.Emoji
	return ret
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
}

func ParseAPIDMMessage(message APIDMMessage) DMMessage {
	ret := DMMessage{}
	ret.ID = DMMessageID(message.ID)
	ret.SentAt = TimestampFromUnixMilli(int64(message.Time))
	ret.DMChatRoomID = DMChatRoomID(message.ConversationID)
	ret.SenderID = UserID(message.MessageData.SenderID)
	ret.Text = message.MessageData.Text

	ret.InReplyToID = DMMessageID(message.MessageData.ReplyData.ID) // Will be "0" if not a reply

	ret.Reactions = make(map[UserID]DMReaction)
	for _, api_reacc := range message.MessageReactions {
		reacc := ParseAPIDMReaction(api_reacc)
		reacc.DMMessageID = ret.ID
		ret.Reactions[reacc.SenderID] = reacc
	}
	if message.MessageData.Attachment.Photo.ID != 0 {
		new_image := ParseAPIMedia(message.MessageData.Attachment.Photo)
		new_image.DMMessageID = ret.ID
		ret.Images = []Image{new_image}
	}
	if message.MessageData.Attachment.Video.ID != 0 {
		entity := message.MessageData.Attachment.Video
		if entity.Type == "video" || entity.Type == "animated_gif" {
			new_video := ParseAPIVideo(entity)
			new_video.DMMessageID = ret.ID
			ret.Videos = append(ret.Videos, new_video)
		}
	}

	// Process URLs and link previews
	for _, url := range message.MessageData.Entities.URLs {
		// Skip it if it's an embedded tweet
		_, id, is_ok := TryParseTweetUrl(url.ExpandedURL)
		if is_ok && id == TweetID(message.MessageData.Attachment.Tweet.Status.ID) {
			continue
		}
		// Skip it if it's an embedded image
		if message.MessageData.Attachment.Photo.URL == url.ShortenedUrl {
			continue
		}
		// Skip it if it's an embedded video
		if message.MessageData.Attachment.Video.URL == url.ShortenedUrl {
			continue
		}

		var new_url Url
		if message.MessageData.Attachment.Card.ShortenedUrl == url.ShortenedUrl {
			if message.MessageData.Attachment.Card.Name == "3691233323:audiospace" {
				// This "url" is just a link to a Space.  Don't process it as a Url
				continue
			}
			new_url = ParseAPIUrlCard(message.MessageData.Attachment.Card)
		}
		new_url.Text = url.ExpandedURL
		new_url.ShortText = url.ShortenedUrl
		new_url.DMMessageID = ret.ID
		ret.Urls = append(ret.Urls, new_url)
	}

	return ret
}
