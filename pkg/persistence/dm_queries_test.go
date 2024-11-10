package persistence_test

import (
	"fmt"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func TestSaveAndLoadChatRoom(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestDMs"
	profile := create_or_load_profile(profile_path)

	chat_room := create_dummy_chat_room()
	chat_room.Type = "fnort"
	primary_user, is_ok := chat_room.Participants[UserID(-1)]
	require.True(is_ok)
	primary_user.Status = fmt.Sprintf("status for %s", chat_room.ID)
	chat_room.Participants[primary_user.UserID] = primary_user

	// Save it
	err := profile.SaveChatRoom(chat_room)
	require.NoError(err)

	// Reload it
	new_chat_room, err := profile.GetChatRoom(chat_room.ID)
	require.NoError(err)

	if diff := deep.Equal(chat_room, new_chat_room); diff != nil {
		t.Error(diff)
	}
}

func TestModifyChatRoom(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestDMs"
	profile := create_or_load_profile(profile_path)

	// Save it
	chat_room := create_dummy_chat_room()
	chat_room.LastMessagedAt = TimestampFromUnix(2)
	err := profile.SaveChatRoom(chat_room)
	require.NoError(err)

	// Modify it
	chat_room.LastMessagedAt = TimestampFromUnix(35)
	err = profile.SaveChatRoom(chat_room)
	require.NoError(err)

	// Reload it
	new_chat_room, err := profile.GetChatRoom(chat_room.ID)
	require.NoError(err)

	assert.Equal(t, new_chat_room.LastMessagedAt, TimestampFromUnix(35))
}

func TestModifyChatParticipant(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestDMs"
	profile := create_or_load_profile(profile_path)

	// Save it
	chat_room := create_dummy_chat_room()
	err := profile.SaveChatRoom(chat_room)
	require.NoError(err)

	// Add a participant and modify the existing one
	primary_user, is_ok := chat_room.Participants[UserID(-1)]
	require.True(is_ok)
	primary_user.IsReadOnly = true
	primary_user.LastReadEventID = DMMessageID(1500)
	chat_room.Participants[primary_user.UserID] = primary_user
	new_user := create_dummy_user()
	chat_room.Participants[new_user.ID] = DMChatParticipant{
		DMChatRoomID:        chat_room.ID,
		UserID:              new_user.ID,
		LastReadEventID:     DMMessageID(0),
		IsChatSettingsValid: false,
	}

	// Save again
	err = profile.SaveUser(&new_user)
	require.NoError(err)
	err = profile.SaveChatRoom(chat_room)
	require.NoError(err)

	// Reload it
	new_chat_room, err := profile.GetChatRoom(chat_room.ID)
	require.NoError(err)

	if diff := deep.Equal(chat_room, new_chat_room); diff != nil {
		t.Error(diff)
	}
}

func TestSaveAndLoadChatMessage(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestDMs"
	profile := create_or_load_profile(profile_path)
	message := create_dummy_chat_message()

	// Save it
	err := profile.SaveChatMessage(message)
	require.NoError(err)

	// Reload it
	trove_with_new_message, err := profile.GetChatMessage(message.ID)
	require.NoError(err)
	new_message, is_ok := trove_with_new_message.Messages[message.ID]
	require.True(is_ok)

	if diff := deep.Equal(message, new_message); diff != nil {
		t.Error(diff)
	}

	// Scraping the same message again shouldn't break
	err = profile.SaveChatMessage(message)
	require.NoError(err)
}

func TestAddReactionToChatMessage(t *testing.T) {
	require := require.New(t)
	profile_path := "test_profiles/TestDMs"
	profile := create_or_load_profile(profile_path)
	message := create_dummy_chat_message()

	// Save it
	err := profile.SaveChatMessage(message)
	require.NoError(err)

	// Add a reaction
	new_user := create_dummy_user()
	message.Reactions[new_user.ID] = DMReaction{
		ID:          DMMessageID(message.ID + 10),
		DMMessageID: message.ID,
		SenderID:    new_user.ID,
		SentAt:      TimestampFromUnix(51000),
		Emoji:       "🅱",
	}
	require.NoError(profile.SaveUser(&new_user))
	require.NoError(profile.SaveChatMessage(message))

	// Reload it
	trove_with_new_message, err := profile.GetChatMessage(message.ID)
	require.NoError(err)
	new_message, is_ok := trove_with_new_message.Messages[message.ID]
	require.True(is_ok)

	if diff := deep.Equal(message, new_message); diff != nil {
		t.Error(diff)
	}
}

func TestGetChatRoomsPreview(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	chat_view := profile.GetChatRoomsPreview(UserID(1458284524761075714))
	assert.Len(chat_view.Rooms, 1)
	assert.Len(chat_view.RoomIDs, 1)
	assert.Equal(chat_view.RoomIDs, []DMChatRoomID{"1458284524761075714-1488963321701171204"})

	room, is_ok := chat_view.Rooms[chat_view.RoomIDs[0]]
	require.True(is_ok)
	assert.Equal(room.LastMessageID, DMMessageID(1766595519000760325))

	msg, is_ok := chat_view.Messages[room.LastMessageID]
	require.True(is_ok)
	assert.Equal(msg.Text, "This looks pretty good huh")

	// Participants
	require.Len(room.Participants, 2)
	for _, user_id := range []UserID{1458284524761075714, 1488963321701171204} {
		participant, is_ok := room.Participants[user_id]
		require.True(is_ok)
		assert.Equal(participant.IsChatSettingsValid, participant.UserID == 1488963321701171204)
		u, is_ok := chat_view.Users[user_id]
		require.True(is_ok)
		assert.Equal(u.ID, user_id)
		assert.NotEqual(u.Handle, "") // Make sure it's filled out
	}
}

func TestGetChatRoomContents(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	room_id := DMChatRoomID("1458284524761075714-1488963321701171204")
	chat_view := profile.GetChatRoomContents(room_id, -1)
	assert.Len(chat_view.Rooms, 1)
	room, is_ok := chat_view.Rooms[room_id]
	require.True(is_ok)

	// Participants
	require.Len(room.Participants, 2)
	for _, user_id := range []UserID{1458284524761075714, 1488963321701171204} {
		participant, is_ok := room.Participants[user_id]
		require.True(is_ok)
		assert.Equal(participant.IsChatSettingsValid, participant.UserID == 1488963321701171204)
		u, is_ok := chat_view.Users[user_id]
		require.True(is_ok)
		assert.Equal(u.ID, user_id)
		assert.NotEqual(u.Handle, "") // Make sure it's filled out
	}

	// Messages
	expected_message_ids := []DMMessageID{
		1663623062195957773, 1663623203644751885, 1665922180176044037, 1665936253483614212,
		1766248283901776125, 1766255994668191902, 1766595519000760325,
	}
	require.Equal(chat_view.MessageIDs, expected_message_ids)
	require.Len(chat_view.Messages, len(expected_message_ids))
	for _, msg_id := range chat_view.MessageIDs {
		msg, is_ok := chat_view.Messages[msg_id]
		assert.True(is_ok)
		assert.Equal(msg.ID, msg_id)
	}

	// Attachments
	m_img := chat_view.Messages[DMMessageID(1766595519000760325)]
	require.Len(m_img.Images, 1)
	assert.Equal(m_img.Images[0].RemoteURL,
		"https://ton.twitter.com/1.1/ton/data/dm/1766595519000760325/1766595500407459840/ML6pC79A.png")
	m_vid := chat_view.Messages[DMMessageID(1766248283901776125)]
	require.Len(m_vid.Videos, 1)
	assert.Equal(m_vid.Videos[0].RemoteURL,
		"https://video.twimg.com/dm_video/1766248268416385024/vid/avc1/500x280/edFuZXtEVvem158AjvmJ3SZ_1DdG9cbSoW4fm6cDF1k.mp4?tag=1")
	m_url := chat_view.Messages[DMMessageID(1766255994668191902)]
	require.Len(m_url.Urls, 1)
	assert.Equal(m_url.Urls[0].Text, "https://offline-twitter.com/introduction/data-ownership-and-composability/")

	// Reactions
	msg_with_reacc := chat_view.Messages[DMMessageID(1663623062195957773)]
	require.Len(msg_with_reacc.Reactions, 1)
	reacc, is_ok := msg_with_reacc.Reactions[UserID(1458284524761075714)]
	require.True(is_ok)
	assert.Equal(reacc.Emoji, "😂")

	// Embedded tweets
	require.Len(chat_view.Tweets, 1)
	twt, is_ok := chat_view.Tweets[TweetID(1665509126737129472)]
	require.True(is_ok)
	assert.Equal(twt.InReplyToID, TweetID(1665505986184900611))
	assert.Equal(twt.NumLikes, 7)
	u, is_ok := chat_view.Users[twt.UserID]
	require.True(is_ok)
	assert.Equal(u.Location, "on my computer")
}

func TestGetChatRoomContentsAfterTimestamp(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	room_id := DMChatRoomID("1488963321701171204-1178839081222115328")
	chat_view := profile.GetChatRoomContents(room_id, 1686025129141)

	// MessageIDs should just be the ones in the thread
	require.Equal(chat_view.MessageIDs, []DMMessageID{1665936253483614215, 1665936253483614216, 1665937253483614217})

	// Replied messages should be available, but not in the list of MessageIDs
	require.Len(chat_view.Messages, 4)
	msg, is_ok := chat_view.Messages[1665936253483614214]
	assert.True(is_ok)
	assert.Equal(msg.ID, DMMessageID(1665936253483614214))
	for _, msg_id := range chat_view.MessageIDs {
		msg, is_ok := chat_view.Messages[msg_id]
		assert.True(is_ok)
		assert.Equal(msg.ID, msg_id)
	}
}

func TestGetUnreadConversations(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	offline_twatter_unreads := profile.GetUnreadConversations(UserID(1488963321701171204))
	require.Len(offline_twatter_unreads, 1)
	assert.Equal(offline_twatter_unreads[0], DMChatRoomID("1488963321701171204-1178839081222115328"))
	mystery_unreads := profile.GetUnreadConversations(UserID(1178839081222115328))
	assert.Len(mystery_unreads, 0)
}
