package persistence_test

import (
	"fmt"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
