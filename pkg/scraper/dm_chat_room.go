package scraper

type DMChatRoomID string

type DMChatParticipant struct {
	DMChatRoomID    DMChatRoomID `db:"chat_room_id"`
	UserID          UserID       `db:"user_id"`
	LastReadEventID DMMessageID  `db:"last_read_event_id"`

	IsChatSettingsValid            bool   `db:"is_chat_settings_valid"`
	IsNotificationsDisabled        bool   `db:"is_notifications_disabled"`
	IsMentionNotificationsDisabled bool   `db:"is_mention_notifications_disabled"`
	IsReadOnly                     bool   `db:"is_read_only"`
	IsTrusted                      bool   `db:"is_trusted"`
	IsMuted                        bool   `db:"is_muted"`
	Status                         string `db:"status"`
}

type DMChatRoom struct {
	ID             DMChatRoomID `db:"id"`
	Type           string       `db:"type"`
	LastMessagedAt Timestamp    `db:"last_messaged_at"`
	IsNSFW         bool         `db:"is_nsfw"`

	Participants map[UserID]DMChatParticipant
}

func ParseAPIDMChatRoom(api_room APIDMConversation) DMChatRoom {
	ret := DMChatRoom{}
	ret.ID = DMChatRoomID(api_room.ConversationID)
	ret.Type = api_room.Type
	ret.LastMessagedAt = TimestampFromUnix(int64(api_room.SortTimestamp))
	ret.IsNSFW = api_room.NSFW

	ret.Participants = make(map[UserID]DMChatParticipant)
	for _, api_participant := range api_room.Participants {
		participant := DMChatParticipant{}
		participant.UserID = UserID(api_participant.UserID)
		participant.DMChatRoomID = ret.ID
		participant.LastReadEventID = DMMessageID(api_participant.LastReadEventID)

		// Process chat settings if this is the logged-in user
		if participant.UserID == the_api.UserID {
			participant.IsNotificationsDisabled = api_room.NotificationsDisabled
			participant.IsReadOnly = api_room.ReadOnly
			participant.IsTrusted = api_room.Trusted
			participant.IsMuted = api_room.Muted
			participant.Status = api_room.Status
			participant.IsChatSettingsValid = true
		}
		ret.Participants[participant.UserID] = participant
	}
	return ret
}
