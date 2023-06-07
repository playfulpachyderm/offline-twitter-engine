package scraper

type APIDMReaction struct {
	ID       int    `json:"id,string"`
	Time     int    `json:"time,string"`
	SenderID int    `json:"sender_id,string"`
	Emoji    string `json:"emoji_reaction"`
}

type APIDMMessage struct {
	ID             int    `json:"id,string"`
	Time           int    `json:"time,string"`
	ConversationID string `json:"conversation_id"`
	MessageData    struct {
		ID        int    `json:"id,string"`
		Time      int    `json:"time,string"`
		SenderID  int    `json:"sender_id,string"`
		Text      string `json:"text"`
		ReplyData struct {
			ID int `json:"id,string"`
		} `json:"reply_data"`
		Attachment struct {
			Tweet APITweet `json:"tweet"`
		} `json:"attachment"`
	} `json:"message_data"`
	MessageReactions []APIDMReaction `json:"message_reactions"`
}

type APIDMConversation struct {
	ConversationID string `json:"conversation_id"`
	Type           string `json:"type"`
	SortTimestamp  int    `json:"sort_timestamp,string"`
	Participants   []struct {
		UserID          int `json:"user_id,string"`
		LastReadEventID int `json:"last_read_event_id,string"`
	}
	NSFW                  bool   `json:"nsfw"`
	NotificationsDisabled bool   `json:"notifications_disabled"`
	ReadOnly              bool   `json:"read_only"`
	Trusted               bool   `json:"trusted"`
	Muted                 bool   `json:"muted"`
	Status                string `json:"status"`
}

type APIInbox struct {
	LastSeenEventID int    `json:"last_seen_event_id,string"`
	Cursor          string `json:"cursor"`
	Entries         []struct {
		Message APIDMMessage `json:"message"`
	} `json:"entries"`
	Users         map[string]APIUser           `json:"users"`
	Conversations map[string]APIDMConversation `json:"conversations"`
}

type APIDMResponse struct {
	InboxInitialState APIInbox `json:"inbox_initial_state"`
}
