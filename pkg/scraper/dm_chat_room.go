package scraper

import "fmt"

type DMChatRoomID string

type DMChatParticipant struct {
	UserID          UserID
	DMChatRoomID    DMChatRoomID
	LastReadEventID DMMessageID

	IsChatSettingsValid     bool
	IsNotificationsDisabled bool
	IsReadOnly              bool
	IsTrusted               bool
	IsMuted                 bool
	Status                  string
}

type DMChatRoom struct {
	ID             DMChatRoomID
	Type           string
	LastMessagedAt Timestamp
	IsNSFW         bool

	Participants []DMChatParticipant
}

func ParseAPIDMChatRoom(api_room APIDMConversation) DMChatRoom {
	fmt.Printf("%#v\n", api_room)
	ret := DMChatRoom{}
	ret.ID = DMChatRoomID(api_room.ConversationID)
	ret.Type = api_room.Type
	ret.LastMessagedAt = TimestampFromUnix(int64(api_room.SortTimestamp))
	ret.IsNSFW = api_room.NSFW

	ret.Participants = []DMChatParticipant{}
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
		ret.Participants = append(ret.Participants, participant)
	}
	return ret
}
