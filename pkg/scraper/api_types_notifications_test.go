package scraper_test

import (
	"testing"

	"encoding/json"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func TestParseNotificationsPage(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/notifications/notifications_response_first_page.json")
	require.NoError(err)

	var resp TweetResponse
	err = json.Unmarshal(data, &resp)
	require.NoError(err)

	current_user_id := UserID(12345678)
	tweet_trove, err := resp.ToTweetTroveAsNotifications(current_user_id)
	require.NoError(err)

	notif1, is_ok := tweet_trove.Notifications["FKncQJGVgAQAAAABSQ3bEYsN6BFN3re-ZsU"]
	assert.True(is_ok)
	assert.Equal(9, notif1.Type) // login
	assert.Equal(current_user_id, notif1.UserID)

	notif2, is_ok := tweet_trove.Notifications["FKncQJGVgAQAAAABSQ3bEYsN6BFaOkNV8aw"]
	assert.True(is_ok)
	assert.Equal(2, notif2.Type) // retweet
	assert.Equal(current_user_id, notif2.UserID)
	assert.Equal(UserID(1458284524761075714), notif2.ActionUserID)
	assert.Equal(TweetID(1824915465275392037), notif2.ActionTweetID)
	assert.Equal(TimestampFromUnixMilli(1723928739342), notif2.SentAt)

	notif3, is_ok := tweet_trove.Notifications["FKncQJGVgAQAAAABSQ3bEYsN6BE-OY688aw"]
	assert.True(is_ok)
	assert.Equal(1, notif3.Type) // like
	assert.Equal(current_user_id, notif3.UserID)
	assert.Equal(UserID(1458284524761075714), notif3.ActionUserID)
	assert.Equal(TweetID(1824915465275392037), notif3.ActionTweetID)

	notif4, is_ok := tweet_trove.Notifications["FKncQJGVgAQAAAABSQ3bEYsN6BGLlh8UIQs"]
	assert.True(is_ok)
	assert.Equal(11, notif4.Type) // recommended
	assert.Equal(current_user_id, notif4.UserID)

	notif5, is_ok := tweet_trove.Notifications["FKncQJGVgAQAAAABSQ3bEYsN6BHS11EvITw"]
	assert.True(is_ok)
	assert.Equal(5, notif5.Type) // followed you
	assert.Equal(current_user_id, notif5.UserID)
	assert.Equal(UserID(28815778), notif5.ActionUserID)

	notif6, is_ok := tweet_trove.Notifications["FKncQJGVgAQAAAABSQ3bEYsN6BE5ujkCepo"]
	assert.True(is_ok)
	assert.Equal(1, notif6.Type)
	assert.Equal(current_user_id, notif6.UserID)
	assert.Equal(UserID(1458284524761075714), notif6.ActionUserID)
	assert.Equal(TweetID(1826778617705115868), notif6.ActionTweetID)
	assert.Contains(notif6.UserIDs, UserID(1458284524761075714))
	assert.Contains(notif6.UserIDs, UserID(2694459866))

	// Check users
	for _, u_id := range []UserID{1458284524761075714, 28815778} {
		_, is_ok := tweet_trove.Users[u_id]
		assert.True(is_ok)
	}

	// Check tweets
	for _, t_id := range []TweetID{1824915465275392037, 1826778617705115868} {
		_, is_ok := tweet_trove.Tweets[t_id]
		assert.True(is_ok)
	}

	// Test cursor-bottom
	bottom_cursor := resp.GetCursor()
	assert.Equal("DAACDAABCgABFKncQJGVgAQIAAIAAAABCAADSQ3bEQgABIsN6BEACwACAAAAC0FaRkxRSXFNLTJJAAA", bottom_cursor)
}
