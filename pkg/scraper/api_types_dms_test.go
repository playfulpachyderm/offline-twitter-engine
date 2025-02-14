package scraper_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
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

	trove := api_message.ToTweetTrove()
	message, is_ok := trove.Messages[DMMessageID(api_message.ID)]
	require.True(t, is_ok)
	assert.Equal(message.ID, DMMessageID(1663623203644751885))
	assert.Equal(message.SentAt, TimestampFromUnixMilli(1685473655064))
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

	trove := api_message.ToTweetTrove()
	message, is_ok := trove.Messages[DMMessageID(api_message.ID)]
	require.True(t, is_ok)
	assert.Equal(message.ID, DMMessageID(1663623062195957773))
	require.Len(t, message.Reactions, 1)

	reacc, is_ok := message.Reactions[UserID(1458284524761075714)]
	require.True(t, is_ok)
	assert.Equal(reacc.ID, DMMessageID(1665914315742781440))
	assert.Equal(reacc.SentAt, TimestampFromUnixMilli(1686019898732))
	assert.Equal(reacc.DMMessageID, DMMessageID(1663623062195957773))
	assert.Equal(reacc.SenderID, UserID(1458284524761075714))
	assert.Equal(reacc.Emoji, "😂")
}

func TestParseAPIDMMessageWithEmbeddedTweet(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/dms/dm_message_with_tweet_attachment.json")
	if err != nil {
		panic(err)
	}
	var api_message APIDMMessage
	err = json.Unmarshal(data, &api_message)
	require.NoError(t, err)

	trove := api_message.ToTweetTrove()

	assert.Len(trove.Messages, 1)
	m, is_ok := trove.Messages[DMMessageID(1665936253483614212)]
	assert.True(is_ok)
	assert.Equal(m.EmbeddedTweetID, TweetID(1665509126737129472))
	assert.Equal(m.Text, "Check this out")
	assert.Len(m.Urls, 0) // Embedded tweets should not parse as URLs

	assert.Len(trove.Tweets, 1)
	_, is_ok = trove.Tweets[TweetID(1665509126737129472)]
	assert.True(is_ok)

	assert.Len(trove.Users, 1)
	_, is_ok = trove.Users[UserID(1458284524761075714)]
	assert.True(is_ok)
}

func TestParseAPIDMMessageWithEmbeddedImage(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/dms/dm_message_with_image.json")
	if err != nil {
		panic(err)
	}
	var api_message APIDMMessage
	err = json.Unmarshal(data, &api_message)
	require.NoError(t, err)

	trove := api_message.ToTweetTrove()

	assert.Len(trove.Messages, 1)
	m, is_ok := trove.Messages[DMMessageID(1766224476729995648)]
	assert.True(is_ok)
	assert.Len(m.Urls, 0) // Embedded images should not parse as URLs

	// Check that the short-URL is stripped
	assert.Equal("A gastropub staffed by white college girls and the chefs are all Latino", m.Text)

	assert.Len(m.Images, 1)
	assert.Equal(m.ID, m.Images[0].DMMessageID)
	assert.Equal("https://ton.twitter.com/1.1/ton/data/dm/1766224476729995648/1766224374648958976/L4Ah1GSh.jpg", m.Images[0].RemoteURL)
}

func TestParseAPIDMMessageWithEmbeddedVideo(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/dms/dm_message_with_video.json")
	if err != nil {
		panic(err)
	}
	var api_message APIDMMessage
	err = json.Unmarshal(data, &api_message)
	require.NoError(t, err)

	trove := api_message.ToTweetTrove()

	assert.Len(trove.Messages, 1)
	m, is_ok := trove.Messages[DMMessageID(1766248283901776125)]
	assert.True(is_ok)
	assert.Len(m.Urls, 0) // Embedded videos should not parse as URLs

	// Check the short-URL is stripped
	assert.Equal("", m.Text)

	assert.Len(m.Videos, 1)
	assert.Equal(m.ID, m.Videos[0].DMMessageID)
	assert.Equal(
		"https://video.twimg.com/dm_video/1766248268416385024/vid/avc1/500x280/edFuZXtEVvem158AjvmJ3SZ_1DdG9cbSoW4fm6cDF1k.mp4?tag=1",
		m.Videos[0].RemoteURL)
}

func TestParseAPIDMMessageWithUrlCard(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/dms/dm_message_with_url_card.json")
	if err != nil {
		panic(err)
	}
	var api_message APIDMMessage
	err = json.Unmarshal(data, &api_message)
	require.NoError(t, err)

	trove := api_message.ToTweetTrove()

	assert.Len(trove.Messages, 1)
	m, is_ok := trove.Messages[DMMessageID(1766255994668191902)]
	assert.True(is_ok)
	assert.Len(m.Urls, 1)
	assert.Equal("You wrote this?", m.Text)
	url := m.Urls[0]
	assert.Equal(m.ID, url.DMMessageID)
	assert.Equal("https://offline-twitter.com/introduction/data-ownership-and-composability/", url.Text)
	assert.Equal("offline-twitter.com", url.Domain)
	assert.Equal("Data ownership and composability", url.Title)
}

func TestParseAPIDMConversation(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/dms/dm_chat_room.json")
	require.NoError(t, err)

	var api_room APIDMConversation
	err = json.Unmarshal(data, &api_room)
	require.NoError(t, err)

	// Simulate one of the participants being logged in
	chat_room := ParseAPIDMChatRoom(api_room, UserID(1458284524761075714))
	assert.Equal(DMChatRoomID("1458284524761075714-1488963321701171204"), chat_room.ID)
	assert.Equal("ONE_TO_ONE", chat_room.Type)
	assert.Equal(TimestampFromUnixMilli(1686025129086), chat_room.LastMessagedAt)
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

func TestParseAPIDMGroupChat(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/dms/chat_room_group_chat.json")
	require.NoError(t, err)

	var api_room APIDMConversation
	err = json.Unmarshal(data, &api_room)
	require.NoError(t, err)

	// Simulate one of the participants being logged in
	chat_room := ParseAPIDMChatRoom(api_room, UserID(1458284524761075714))
	assert.Equal(DMChatRoomID("1710215025518948715"), chat_room.ID)
	assert.Equal("GROUP_DM", chat_room.Type)
	assert.Equal(TimestampFromUnixMilli(1700112789457), chat_room.LastMessagedAt)
	assert.False(chat_room.IsNSFW)

	// Group DM settings
	assert.Equal(chat_room.CreatedAt, TimestampFromUnixMilli(1696582011037))
	assert.Equal(chat_room.CreatedByUserID, UserID(2694459866))
	assert.Equal(chat_room.Name, "Schön ist die Welt")
	assert.Equal(chat_room.AvatarImageRemoteURL,
		"https://pbs.twimg.com/dm_group_img/1722785857403240448/3Wt_yJEq6i_G-kAT2rXheTojjhqkYE3okoW5JGUUHY7J9D8O9o?format=jpg&name=orig")
	assert.Equal(chat_room.AvatarImageLocalPath, "1710215025518948715_avatar_3Wt_yJEq6i_G-kAT2rXheTojjhqkYE3okoW5JGUUHY7J9D8O9o.jpg")

	assert.Len(chat_room.Participants, 5)
}

func TestParseInbox(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/dms/inbox.json")
	require.NoError(t, err)

	var inbox APIDMResponse
	err = json.Unmarshal(data, &inbox)
	require.NoError(t, err)

	trove := inbox.InboxInitialState.ToTweetTrove(UserID(0))

	for _, id := range []DMMessageID{1663623062195957773, 1663623203644751885, 1665922180176044037, 1665936253483614212} {
		m, is_ok := trove.Messages[id]
		assert.True(is_ok, "Message with ID %d not in the trove!")
		assert.Equal(m.ID, id)
	}
	for _, id := range []UserID{1458284524761075714, 1488963321701171204} {
		u, is_ok := trove.Users[id]
		assert.True(is_ok, "User with ID %d not in the trove!")
		assert.Equal(u.ID, id)
	}
	room_id := DMChatRoomID("1458284524761075714-1488963321701171204")
	room, is_ok := trove.Rooms[room_id]
	assert.True(is_ok)
	assert.Equal(room.ID, room_id)
}

func TestParseDMRoomResponse(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/dms/dm_conversation_response.json")
	require.NoError(t, err)

	var inbox APIDMResponse
	err = json.Unmarshal(data, &inbox)
	require.NoError(t, err)

	trove := inbox.ConversationTimeline.ToTweetTrove(UserID(0))

	for _, id := range []DMMessageID{
		1663623062195957773,
		1663623203644751885,
		1665922180176044037,
		1665936253483614212,
		1726009944393372005,
	} {
		m, is_ok := trove.Messages[id]
		assert.True(is_ok, "Message with ID %d not in the trove!")
		assert.Equal(m.ID, id)
	}
	for _, id := range []UserID{1458284524761075714, 1488963321701171204} {
		u, is_ok := trove.Users[id]
		assert.True(is_ok, "User with ID %d not in the trove!")
		assert.Equal(u.ID, id)
	}
	room_id := DMChatRoomID("1458284524761075714-1488963321701171204")
	room, is_ok := trove.Rooms[room_id]
	assert.True(is_ok)
	assert.Equal(room.ID, room_id)
	assert.Equal(trove.GetOldestMessage(room_id), DMMessageID(1663623062195957773))
}

func TestParseInboxUpdates(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/dms/user_updates_simulated.json")
	require.NoError(t, err)

	var inbox APIDMResponse
	err = json.Unmarshal(data, &inbox)
	require.NoError(t, err)

	trove := inbox.UserEvents.ToTweetTrove(UserID(0))

	assert.Len(trove.Messages, 2) // Should ignore stuff that isn't a message

	_, is_ok := trove.Messages[1725969457464447135]
	assert.True(is_ok)

	message_receiving_a_reacc, is_ok := trove.Messages[1725980964718100721]
	assert.True(is_ok)
	assert.Len(message_receiving_a_reacc.Reactions, 1)
}
