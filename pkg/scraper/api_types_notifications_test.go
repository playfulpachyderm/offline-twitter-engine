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
	assert.Equal(NOTIFICATION_TYPE_LOGIN, notif1.Type)
	assert.Equal(int64(1723851817578), notif1.SortIndex)
	assert.Equal(current_user_id, notif1.UserID)

	// Simple retweet: 1 user retweets 1 tweet
	notif2, is_ok := tweet_trove.Notifications["FKncQJGVgAQAAAABSQ3bEYsN6BFaOkNV8aw"]
	assert.True(is_ok)
	assert.Equal(NOTIFICATION_TYPE_RETWEET, notif2.Type)
	assert.Equal(current_user_id, notif2.UserID)
	assert.Equal(UserID(1458284524761075714), notif2.ActionUserID)
	assert.Equal(TweetID(1824915465275392037), notif2.ActionTweetID)
	assert.Equal(TimestampFromUnixMilli(1723928739342), notif2.SentAt)
	assert.Len(notif2.UserIDs, 1)
	assert.Contains(notif2.UserIDs, UserID(1458284524761075714))
	assert.Len(notif2.TweetIDs, 1)
	assert.Contains(notif2.TweetIDs, TweetID(1824915465275392037))
	assert.Len(notif2.RetweetIDs, 0)

	// Simple like: 1 user likes 1 tweet
	notif3, is_ok := tweet_trove.Notifications["FKncQJGVgAQAAAABSQ3bEYsN6BE-OY688aw"]
	assert.True(is_ok)
	assert.Equal(NOTIFICATION_TYPE_LIKE, notif3.Type)
	assert.Equal(current_user_id, notif3.UserID)
	assert.Equal(UserID(1458284524761075714), notif3.ActionUserID)
	assert.Equal(TweetID(1824915465275392037), notif3.ActionTweetID)
	assert.Len(notif2.UserIDs, 1)
	assert.Contains(notif2.UserIDs, UserID(1458284524761075714))
	assert.Len(notif2.TweetIDs, 1)
	assert.Contains(notif2.TweetIDs, TweetID(1824915465275392037))
	assert.Len(notif2.RetweetIDs, 0)

	notif4, is_ok := tweet_trove.Notifications["FKncQJGVgAQAAAABSQ3bEYsN6BGLlh8UIQs"]
	assert.True(is_ok)
	assert.Equal(NOTIFICATION_TYPE_RECOMMENDED_POST, notif4.Type)
	assert.Equal(current_user_id, notif4.UserID)

	notif5, is_ok := tweet_trove.Notifications["FKncQJGVgAQAAAABSQ3bEYsN6BHS11EvITw"]
	assert.True(is_ok)
	assert.Equal(NOTIFICATION_TYPE_FOLLOW, notif5.Type)
	assert.Equal(current_user_id, notif5.UserID)
	assert.Equal(UserID(28815778), notif5.ActionUserID)

	// 2 users like 1 tweet
	notif6, is_ok := tweet_trove.Notifications["FKncQJGVgAQAAAABSQ3bEYsN6BE5ujkCepo"]
	assert.True(is_ok)
	assert.Equal(NOTIFICATION_TYPE_LIKE, notif6.Type)
	assert.Equal(current_user_id, notif6.UserID)
	assert.Equal(UserID(2694459866), notif6.ActionUserID) // Most recent user
	assert.Equal(TweetID(1826778617705115868), notif6.ActionTweetID)
	assert.Len(notif6.UserIDs, 2)
	assert.Contains(notif6.UserIDs, UserID(1458284524761075714))
	assert.Contains(notif6.UserIDs, UserID(2694459866))
	assert.Len(notif6.TweetIDs, 1)
	assert.Contains(notif6.TweetIDs, TweetID(1826778617705115868))
	assert.Len(notif6.RetweetIDs, 0)

	notif7, is_ok := tweet_trove.Notifications["FKncQJGVgAQAAAABSQ3bEYsN6BGJjUVEd8Y"]
	assert.True(is_ok)
	assert.Equal(NOTIFICATION_TYPE_QUOTE_TWEET, notif7.Type)
	assert.Equal(TweetID(1817720429941059773), notif7.ActionTweetID) // Not in the trove (using fake data)

	notif8, is_ok := tweet_trove.Notifications["FKncQJGVgAQAAAABSQ3bEYsN6BG1nnPGJlQ"]
	assert.True(is_ok)
	assert.Equal(NOTIFICATION_TYPE_MENTION, notif8.Type)
	assert.Equal(TweetID(1814349573847982537), notif8.ActionTweetID)

	// User "liked" your retweet
	notif9, is_ok := tweet_trove.Notifications["FDzeDIfVUAIAAAABiJONco_yJREwmpDdUTQ"]
	assert.True(is_ok)
	assert.Equal(NOTIFICATION_TYPE_LIKE, notif9.Type)
	assert.Equal(TweetID(1826778771686392312), notif9.ActionRetweetID)
	assert.Equal(TweetID(0), notif9.ActionTweetID) // Tweet is not set
	assert.Equal(UserID(1633158398555353096), notif9.ActionUserID)
	assert.Len(notif9.TweetIDs, 0)
	assert.Len(notif9.UserIDs, 1)
	assert.Contains(notif9.UserIDs, UserID(1633158398555353096))
	assert.Len(notif9.RetweetIDs, 1)
	assert.Contains(notif9.RetweetIDs, TweetID(1826778771686392312))

	// Retweet of a retweet
	notif10, is_ok := tweet_trove.Notifications["FDzeDIfVUAIAAAABiJONco_yJRGACovgUTQ"]
	assert.True(is_ok)
	assert.Equal(NOTIFICATION_TYPE_RETWEET, notif10.Type)
	assert.Equal(TweetID(1827183097382654351), notif10.ActionRetweetID) // the retweet that he retweeted
	assert.Equal(TweetID(0), notif10.ActionTweetID)
	assert.Len(notif10.UserIDs, 1)
	assert.Contains(notif10.UserIDs, UserID(1678546445002059781))
	assert.Len(notif10.TweetIDs, 0)
	assert.Len(notif10.RetweetIDs, 1)
	assert.Contains(notif10.RetweetIDs, TweetID(1827183097382654351))

	notif11, is_ok := tweet_trove.Notifications["FDzeDIfVUAIAAAABiJONco_yJRHyMqRjxDY"]
	assert.True(is_ok)
	assert.Equal(NOTIFICATION_TYPE_USER_IS_LIVE, notif11.Type)
	assert.Equal(UserID(277536867), notif11.ActionUserID)

	// 1 user liked multiple posts
	notif12, is_ok := tweet_trove.Notifications["FDzeDIfVUAIAAAABiJONco_yJRESfwtSqvg"]
	assert.True(is_ok)
	assert.True(notif12.HasDetail)

	// TODO: communities
	// notif12, is_ok := tweet_trove.Notifications["FDzeDIfVUAIAAAABiJONco_yJRHPBNsDH88"]
	// assert.True(is_ok)
	// assert.Equal(NOTIFICATION_TYPE_COMMUNITY_PINNED_POST, notif12.Type)

	// Check users
	for _, u_id := range []UserID{1458284524761075714, 28815778, 1633158398555353096} {
		_, is_ok := tweet_trove.Users[u_id]
		assert.True(is_ok)
	}

	// Check tweets
	for _, t_id := range []TweetID{1824915465275392037, 1826778617705115868} {
		_, is_ok := tweet_trove.Tweets[t_id]
		assert.True(is_ok)
	}

	// Check retweets
	for _, r_id := range []TweetID{1826778771686392312} {
		_, is_ok := tweet_trove.Retweets[r_id]
		assert.True(is_ok)
	}

	// Test unread notifs
	assert.Equal(int64(1724566381021), resp.CheckUnreadNotifications())

	// Test cursor-bottom
	bottom_cursor := resp.GetCursor()
	assert.Equal("DAACDAABCgABFKncQJGVgAQIAAIAAAABCAADSQ3bEQgABIsN6BEACwACAAAAC0FaRkxRSXFNLTJJAAA", bottom_cursor)
	assert.False(resp.IsEndOfFeed())
}

func TestParseNotificationsEndOfFeed(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/notifications/notifications_end_of_feed.json")
	require.NoError(err)

	var resp TweetResponse
	err = json.Unmarshal(data, &resp)
	require.NoError(err)

	assert.True(resp.IsEndOfFeed())
}

func TestParseNotificationDetail(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/notifications/notification_detail.json")
	require.NoError(err)

	var resp TweetResponse
	err = json.Unmarshal(data, &resp)
	require.NoError(err)

	trove, ids, err := resp.ToTweetTroveAsNotificationDetail()
	require.NoError(err)
	assert.Len(ids, 2)
	assert.Contains(ids, TweetID(1827544032714633628))
	assert.Contains(ids, TweetID(1826743131108487390))

	_, is_ok := trove.Tweets[1826743131108487390]
	assert.True(is_ok)
	_, is_ok = trove.Retweets[1827544032714633628]
	assert.True(is_ok)
}
