package scraper_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "offline_twitter/scraper"
)

func TestParseAPIDMMessage(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/dms/dm_message.json")
	if err != nil {
		panic(err)
	}
	var api_message APIDMMessage
	err = json.Unmarshal(data, &api_message)
	require.NoError(t, err)

	message := ParseAPIDMMessage(api_message)
	assert.Equal(message.ID, DMMessageID(1663623203644751885))
	assert.Equal(message.SentAt, TimestampFromUnix(1685473655064))
	assert.Equal(message.DMChatRoomID, DMChatRoomID("1458284524761075714-1488963321701171204"))
	assert.Equal(message.SenderID, UserID(1458284524761075714))
	assert.Equal(message.Text, "Yeah i know who you are lol")
	assert.Equal(message.InReplyToID, DMMessageID(0))
	assert.Len(message.Reactions, 0)
}

func TestParseAPIDMMessageWithReaction(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/dms/dm_message_with_reacc.json")
	if err != nil {
		panic(err)
	}
	var api_message APIDMMessage
	err = json.Unmarshal(data, &api_message)
	require.NoError(t, err)

	message := ParseAPIDMMessage(api_message)
	assert.Equal(message.ID, DMMessageID(1663623062195957773))
	require.Len(t, message.Reactions, 1)

	reacc := message.Reactions[0]
	assert.Equal(reacc.ID, DMMessageID(1665914315742781440))
	assert.Equal(reacc.SentAt, TimestampFromUnix(1686019898732))
	assert.Equal(reacc.DMMessageID, DMMessageID(1663623062195957773))
	assert.Equal(reacc.SenderID, UserID(1458284524761075714))
	assert.Equal(reacc.Emoji, "😂")
}

func TestParseAPIDMConversation(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/dms/dm_chat_room.json")
	if err != nil {
		panic(err)
	}
	var api_room APIDMConversation
	err = json.Unmarshal(data, &api_room)
	require.NoError(t, err)

	// Simulate one of the participants being logged in
	InitApi(API{UserID: 1458284524761075714})

	chat_room := ParseAPIDMChatRoom(api_room)
	assert.Equal(DMChatRoomID("1458284524761075714-1488963321701171204"), chat_room.ID)
	assert.Equal("ONE_TO_ONE", chat_room.Type)
	assert.Equal(TimestampFromUnix(1686025129086), chat_room.LastMessagedAt)
	assert.False(chat_room.IsNSFW)

	assert.Len(chat_room.Participants, 2)

	p1 := chat_room.Participants[1458284524761075714]
	assert.Equal(UserID(1458284524761075714), p1.UserID)
	assert.Equal(DMMessageID(1665936253483614212), p1.LastReadEventID)
	assert.True(p1.IsChatSettingsValid)
	assert.False(p1.IsNotificationsDisabled)
	assert.False(p1.IsReadOnly)
	assert.True(p1.IsTrusted)
	assert.False(p1.IsMuted)
	assert.Equal("AT_END", p1.Status)

	p2 := chat_room.Participants[1488963321701171204]
	assert.Equal(UserID(1488963321701171204), p2.UserID)
	assert.Equal(DMMessageID(1663623062195957773), p2.LastReadEventID)
	assert.False(p2.IsChatSettingsValid)
}
