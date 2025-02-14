package persistence

type DMChatRoomID string

// A participant in a chat room.
//
// Most settings will only be fetched for the logged-in user (other chat members will just be "false" for
// everything).  The "IsChatSettingsValid" flag indicates whether this is the case.
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

// A chat room. Stores a map of chat participants and a reference to the most recent message,
// for preview purposes.
type DMChatRoom struct {
	ID             DMChatRoomID `db:"id"`
	Type           string       `db:"type"`
	LastMessagedAt Timestamp    `db:"last_messaged_at"` // Used for ordering the chats in the UI
	IsNSFW         bool         `db:"is_nsfw"`

	// GROUP_DM rooms
	CreatedAt            Timestamp `db:"created_at"`
	CreatedByUserID      UserID    `db:"created_by_user_id"`
	Name                 string    `db:"name"`
	AvatarImageRemoteURL string    `db:"avatar_image_remote_url"`
	AvatarImageLocalPath string    `db:"avatar_image_local_path"`

	LastMessageID DMMessageID `db:"last_message_id"` // Not stored, but used to generate preview
	Participants  map[UserID]DMChatParticipant
}

// TODO: view-layer
// - view helpers should go in a view layer
func (r DMChatRoom) GetParticipantIDs() []UserID {
	ret := []UserID{}
	for user_id := range r.Participants {
		ret = append(ret, user_id)
	}
	return ret
}
