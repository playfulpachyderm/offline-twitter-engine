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
	new_message, err := profile.GetChatMessage(message.ID)
	require.NoError(err)

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
	new_message, err := profile.GetChatMessage(message.ID)
	require.NoError(err)

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
	assert.Equal(room.LastMessageID, DMMessageID(1665936253483614212))

	msg, is_ok := chat_view.Messages[room.LastMessageID]
	require.True(is_ok)
	assert.Equal(msg.Text, "Check this out\nhttps://t.co/rHeWGgNIZ1")

	require.Len(room.Participants, 2)
	for _, user_id := range []UserID{1458284524761075714, 1488963321701171204} {
		participant, is_ok := room.Participants[user_id]
		require.True(is_ok)
		assert.Equal(participant.IsChatSettingsValid, participant.UserID == 1488963321701171204)
		_, is_ok = chat_view.Users[user_id]
		require.True(is_ok)
	}
}

func TestGetChatRoomContents(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	room_id := DMChatRoomID("1458284524761075714-1488963321701171204")
	chat_view := profile.GetChatRoomContents(room_id)
	assert.Len(chat_view.Rooms, 1)
	room, is_ok := chat_view.Rooms[room_id]
	require.True(is_ok)

	// Participants
	require.Len(room.Participants, 2)
	for _, user_id := range []UserID{1458284524761075714, 1488963321701171204} {
		participant, is_ok := room.Participants[user_id]
		require.True(is_ok)
		assert.Equal(participant.IsChatSettingsValid, participant.UserID == 1488963321701171204)
		_, is_ok = chat_view.Users[user_id]
		require.True(is_ok)
	}

	// Messages
	require.Equal(chat_view.MessageIDs, []DMMessageID{1663623062195957773, 1663623203644751885, 1665922180176044037, 1665936253483614212})
	require.Len(chat_view.Messages, 4)
	for _, msg_id := range chat_view.MessageIDs {
		msg, is_ok := chat_view.Messages[msg_id]
		assert.True(is_ok)
		assert.Equal(msg.ID, msg_id)
	}

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
