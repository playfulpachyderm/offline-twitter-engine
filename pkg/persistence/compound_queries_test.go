package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

// A feed should load
func TestBuildUserFeed(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	c := NewUserFeedCursor(UserHandle("cernovich"))
	c.PageSize = 2

	feed, err := profile.NextPage(c, UserID(0))
	require.NoError(err)

	assert.Len(feed.Retweets, 2)
	_, is_ok := feed.Retweets[1490135787144237058]
	assert.True(is_ok)
	_, is_ok = feed.Retweets[1490119308692766723]
	assert.True(is_ok)

	assert.Len(feed.Tweets, 2)
	_, is_ok = feed.Tweets[1490120332484972549]
	assert.True(is_ok)
	_, is_ok = feed.Tweets[1490116725395927042]
	assert.True(is_ok)

	assert.Len(feed.Users, 2)
	_, is_ok = feed.Users[358545917]
	assert.True(is_ok)
	_, is_ok = feed.Users[18812728]
	assert.True(is_ok)

	require.Len(feed.Items, 2)
	assert.Equal(feed.Items[0].TweetID, TweetID(1490120332484972549))
	assert.Equal(feed.Items[0].RetweetID, TweetID(1490135787144237058))
	assert.Equal(feed.Items[1].TweetID, TweetID(1490116725395927042))
	assert.Equal(feed.Items[1].RetweetID, TweetID(1490119308692766723))

	assert.Equal(feed.CursorBottom.CursorValue, 1644107102000)
}

// Should load a feed in the middle (i.e., after some timestamp)
func TestBuildUserFeedPage2(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	c := NewUserFeedCursor(UserHandle("cernovich"))
	c.PageSize = 2
	c.CursorPosition = CURSOR_MIDDLE
	c.CursorValue = 1644107102000
	feed, err := profile.NextPage(c, UserID(0))
	require.NoError(err)

	assert.Len(feed.Retweets, 1)
	_, is_ok := feed.Retweets[1490100255987171332]
	assert.True(is_ok)

	assert.Len(feed.Tweets, 2)
	_, is_ok = feed.Tweets[1489944024278523906]
	assert.True(is_ok)
	_, is_ok = feed.Tweets[1453461248142495744]
	assert.True(is_ok)

	assert.Len(feed.Users, 2)
	_, is_ok = feed.Users[358545917]
	assert.True(is_ok)
	_, is_ok = feed.Users[96906231]
	assert.True(is_ok)

	require.Len(feed.Items, 2)
	assert.Equal(feed.Items[0].TweetID, TweetID(1489944024278523906))
	assert.Equal(feed.Items[0].RetweetID, TweetID(1490100255987171332))
	assert.Equal(feed.Items[1].TweetID, TweetID(1453461248142495744))
	assert.Equal(feed.Items[1].RetweetID, TweetID(0))

	assert.Equal(feed.CursorBottom.CursorValue, 1635367140000)
}

// When the end of the feed is reached, an "End of feed" error should be raised
func TestBuildUserFeedEnd(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	c := NewUserFeedCursor(UserHandle("cernovich"))
	c.PageSize = 2
	c.CursorPosition = CURSOR_MIDDLE
	c.CursorValue = 1 // Won't be anything
	feed, err := profile.NextPage(c, UserID(0))
	require.NoError(err)

	assert.Len(feed.Retweets, 0)
	assert.Len(feed.Tweets, 0)
	assert.Len(feed.Users, 0)
	require.Len(feed.Items, 0)

	assert.Equal(feed.CursorBottom.CursorPosition, CURSOR_END)
}

func TestUserFeedHasLikesInfo(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	// Fetch @Peter_Nimitz user feed while logged in as @MysteryGrove
	c := NewUserFeedCursor(UserHandle("Peter_Nimitz"))
	feed, err := profile.NextPage(c, UserID(1178839081222115328))
	require.NoError(err)

	// Should have "liked" 1 tweet
	for _, t := range feed.Tweets {
		assert.Equal(t.IsLikedByCurrentUser, t.ID == TweetID(1413646595493568516))
	}
}

func TestUserFeedWithTombstone(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	c := NewUserFeedCursor(UserHandle("Heminator"))
	feed, err := profile.NextPage(c, UserID(0))
	require.NoError(err)
	tombstone_tweet := feed.Tweets[TweetID(31)]
	assert.Equal(tombstone_tweet.TombstoneText, "This Tweet was deleted by the Tweet author")
}

func TestUserLikesFeed(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	// Fetch @Peter_Nimitz user feed while logged in as @MysteryGrove
	c := NewUserFeedLikesCursor(UserHandle("MysteryGrove"))
	require.Equal(c.SortOrder, SORT_ORDER_LIKED_AT)
	c.PageSize = 2
	feed, err := profile.NextPage(c, UserID(0))
	require.NoError(err)

	require.Len(feed.Tweets, 2)
	for i, expected_tweet_id := range []TweetID{1698765208393576891, 1426669666928414720} {
		assert.Equal(feed.Items[i].TweetID, expected_tweet_id)
		_, is_ok := feed.Tweets[expected_tweet_id]
		assert.True(is_ok)
	}

	require.Equal(feed.CursorBottom.CursorValue, 4)
	feed, err = profile.NextPage(feed.CursorBottom, UserID(0))
	require.NoError(err)

	require.Len(feed.Tweets, 2)
	for i, expected_tweet_id := range []TweetID{1343633011364016128, 1513313535480287235} {
		assert.Equal(feed.Items[i].TweetID, expected_tweet_id)
		_, is_ok := feed.Tweets[expected_tweet_id]
		assert.True(is_ok)
	}

	assert.Equal(feed.CursorBottom.CursorValue, 2)
	feed, err = profile.NextPage(feed.CursorBottom, UserID(0))
	require.NoError(err)

	require.Len(feed.Tweets, 1)
	for i, expected_tweet_id := range []TweetID{1413646595493568516} {
		assert.Equal(feed.Items[i].TweetID, expected_tweet_id)
		_, is_ok := feed.Tweets[expected_tweet_id]
		assert.True(is_ok)
	}
	assert.Equal(feed.CursorBottom.CursorPosition, CURSOR_END)
}

func TestTweetDetailWithReplies(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	tweet_detail, err := profile.GetTweetDetail(TweetID(1413646595493568516), UserID(1178839081222115328))
	require.NoError(err)

	assert.Len(tweet_detail.Retweets, 0)

	assert.Len(tweet_detail.Tweets, 8)
	for _, id := range []TweetID{
		1413646309047767042,
		1413646595493568516,
		1413647919215906817,
		1413657324267311104,
		1413658466795737091,
		1413650853081276421,
		1413772782358433792,
		1413773185296650241,
	} {
		t, is_ok := tweet_detail.Tweets[id]
		assert.True(is_ok)
		assert.Equal(t.IsLikedByCurrentUser, id == 1413646595493568516)
	}

	assert.Len(tweet_detail.Users, 4)
	for _, id := range []UserID{
		1032468021485293568,
		1372116552942764034,
		1067869346775646208,
		1304281147074064385,
	} {
		_, is_ok := tweet_detail.Users[id]
		assert.True(is_ok)
	}

	require.Len(tweet_detail.ParentIDs, 1)
	assert.Equal(tweet_detail.ParentIDs[0], TweetID(1413646309047767042))

	require.Len(tweet_detail.ReplyChains, 3)
	assert.Len(tweet_detail.ReplyChains[0], 2)
	assert.Equal(tweet_detail.ReplyChains[0][0], TweetID(1413657324267311104))
	assert.Equal(tweet_detail.ReplyChains[0][1], TweetID(1413658466795737091))
	assert.Len(tweet_detail.ReplyChains[1], 1)
	assert.Equal(tweet_detail.ReplyChains[1][0], TweetID(1413650853081276421))
	assert.Len(tweet_detail.ReplyChains[2], 2)
	assert.Equal(tweet_detail.ReplyChains[2][0], TweetID(1413772782358433792))
	assert.Equal(tweet_detail.ReplyChains[2][1], TweetID(1413773185296650241))
}

func TestTweetDetailWithParents(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	tweet_detail, err := profile.GetTweetDetail(TweetID(1413773185296650241), UserID(1178839081222115328))
	require.NoError(err)

	assert.Len(tweet_detail.Retweets, 0)

	assert.Len(tweet_detail.Tweets, 4)
	for _, id := range []TweetID{
		1413646309047767042,
		1413646595493568516,
		1413772782358433792,
		1413773185296650241,
	} {
		t, is_ok := tweet_detail.Tweets[id]
		assert.True(is_ok)
		assert.Equal(t.IsLikedByCurrentUser, id == 1413646595493568516)
	}

	assert.Len(tweet_detail.Users, 2)
	_, is_ok := tweet_detail.Users[1032468021485293568]
	assert.True(is_ok)
	_, is_ok = tweet_detail.Users[1372116552942764034]
	assert.True(is_ok)

	require.Len(tweet_detail.ParentIDs, 3)
	assert.Equal(tweet_detail.ParentIDs[0], TweetID(1413646309047767042))
	assert.Equal(tweet_detail.ParentIDs[1], TweetID(1413646595493568516))
	assert.Equal(tweet_detail.ParentIDs[2], TweetID(1413772782358433792))

	require.Len(tweet_detail.ReplyChains, 0)
}

func TestTweetDetailWithThread(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	tweet_detail, err := profile.GetTweetDetail(TweetID(1698762403163304110), UserID(0))
	require.NoError(err)

	assert.Len(tweet_detail.Retweets, 0)

	assert.Len(tweet_detail.Tweets, 11)

	expected_thread := []TweetID{
		1698762405268902217, 1698762406929781161, 1698762408410390772, 1698762409974857832,
		1698762411853971851, 1698762413393236329, 1698762414957666416,
	}

	assert.Equal(expected_thread, tweet_detail.ThreadIDs)

	for _, id := range expected_thread {
		_, is_ok := tweet_detail.Tweets[id]
		assert.True(is_ok)
	}

	assert.Len(tweet_detail.Users, 2)
	_, is_ok := tweet_detail.Users[1458284524761075714]
	assert.True(is_ok)
	_, is_ok = tweet_detail.Users[534463724]
	assert.True(is_ok)

	require.Len(tweet_detail.ReplyChains, 1) // Should not include the Thread replies
	assert.Equal(tweet_detail.ReplyChains[0][0], TweetID(1698792233619562866))
}

func TestNotificationsFeed(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := LoadProfile("../../sample_data/profile")
	require.NoError(err)

	feed := profile.GetNotificationsForUser(UserID(1488963321701171204), 0, 6)
	assert.Len(feed.TweetTrove.Notifications, 6)
	assert.Len(feed.TweetTrove.Tweets, 3)
	assert.Len(feed.TweetTrove.Retweets, 1)
	assert.Len(feed.TweetTrove.Users, 6)

	// Check that Users were retrieved on the notification with detail
	notif, is_ok := feed.TweetTrove.Notifications["FKncQJGVgAQAAAABSQ3bEaTgXL8f40e77r4"]
	assert.True(is_ok)
	assert.Len(notif.UserIDs, 3)
	// Ensure they're also in the TweetTrove
	for _, u_id := range notif.UserIDs {
		_, is_ok := feed.TweetTrove.Users[u_id]
		assert.True(is_ok)
	}

	assert.Len(feed.Items, 6)
	assert.Equal(feed.Items[0].NotificationID, NotificationID("FDzeDIfVUAIAAAABiJONcqaBFAzeN-n-Luw"))
	assert.Equal(feed.Items[0].RetweetID, TweetID(1490135787124232223))
	assert.Equal(feed.Items[1].NotificationID, NotificationID("FDzeDIfVUAIAAvsBiJONcqYgiLgXOolO9t0"))
	assert.Equal(feed.Items[1].TweetID, TweetID(1826778617705115869))
	assert.Equal(feed.Items[2].NotificationID, NotificationID("FKncQJGVgAQAAAABSQ3bEaTgXL8VBxefepo"))
	assert.Equal(feed.Items[2].TweetID, TweetID(1826778617705115868))
	assert.Equal(feed.Items[3].NotificationID, NotificationID("FKncQJGVgAQAAAABSQ3bEaTgXL_S11Ev36g"))
	assert.Equal(feed.Items[4].NotificationID, NotificationID("FKncQJGVgAQAAAABSQ3bEaTgXL-G8wObqVY"))
	assert.Equal(feed.Items[5].NotificationID, NotificationID("FKncQJGVgAQAAAABSQ3bEaTgXL8f40e77r4"))
	assert.Equal(feed.Items[5].TweetID, TweetID(1826778617705115868))

	// Tweet should be "liked"
	liked_tweet, is_ok := feed.TweetTrove.Tweets[1826778617705115869]
	require.True(is_ok)
	assert.True(liked_tweet.IsLikedByCurrentUser)

	assert.Equal(feed.CursorBottom.CursorPosition, CURSOR_MIDDLE)
	assert.Equal(feed.CursorBottom.CursorValue, 1723494244885)

	// Paginated version
	// -----------------

	// Limit 3, after sort_index of the 1st one above
	feed = profile.GetNotificationsForUser(UserID(1488963321701171204), 1726604756351, 3)
	assert.Len(feed.TweetTrove.Notifications, 3)

	assert.Len(feed.Items, 3)
	assert.Equal(feed.Items[0].NotificationID, NotificationID("FDzeDIfVUAIAAvsBiJONcqYgiLgXOolO9t0"))
	assert.Equal(feed.Items[1].NotificationID, NotificationID("FKncQJGVgAQAAAABSQ3bEaTgXL8VBxefepo"))
	assert.Equal(feed.Items[2].NotificationID, NotificationID("FKncQJGVgAQAAAABSQ3bEaTgXL_S11Ev36g"))

	assert.Equal(feed.CursorBottom.CursorPosition, CURSOR_MIDDLE)
	assert.Equal(feed.CursorBottom.CursorValue, 1724251072880)

	// At end of feed
	// --------------

	// cursor = last notification's sort index
	feed = profile.GetNotificationsForUser(UserID(1488963321701171204), 1723494244885, 3)
	assert.Len(feed.Items, 0)
	assert.Equal(feed.CursorBottom.CursorPosition, CURSOR_END)
}
