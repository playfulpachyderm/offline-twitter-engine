package persistence_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

// A feed should load
func TestBuildUserFeed(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	c := persistence.NewUserFeedCursor(UserHandle("cernovich"))
	c.PageSize = 2

	feed, err := profile.NextPage(c)
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

	assert.Equal(feed.CursorBottom.CursorValue, 1644107102)
}

// Should load a feed in the middle (i.e., after some timestamp)
func TestBuildUserFeedPage2(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	c := persistence.NewUserFeedCursor(UserHandle("cernovich"))
	c.PageSize = 2
	c.CursorPosition = persistence.CURSOR_MIDDLE
	c.CursorValue = 1644107102
	feed, err := profile.NextPage(c)
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

	assert.Equal(feed.CursorBottom.CursorValue, 1635367140)
}

// When the end of the feed is reached, an "End of feed" error should be raised
func TestBuildUserFeedEnd(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	c := persistence.NewUserFeedCursor(UserHandle("cernovich"))
	c.PageSize = 2
	c.CursorPosition = persistence.CURSOR_MIDDLE
	c.CursorValue = 1 // Won't be anything
	feed, err := profile.NextPage(c)
	require.NoError(err)

	assert.Len(feed.Retweets, 0)
	assert.Len(feed.Tweets, 0)
	assert.Len(feed.Users, 0)
	require.Len(feed.Items, 0)

	assert.Equal(feed.CursorBottom.CursorPosition, persistence.CURSOR_END)
}

func TestUserFeedWithTombstone(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	c := persistence.NewUserFeedCursor(UserHandle("Heminator"))
	feed, err := profile.NextPage(c)
	require.NoError(err)
	tombstone_tweet := feed.Tweets[TweetID(31)]
	assert.Equal(tombstone_tweet.TombstoneText, "This Tweet was deleted by the Tweet author")
}

func TestTweetDetailWithReplies(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	tweet_detail, err := profile.GetTweetDetail(TweetID(1413646595493568516))
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
		_, is_ok := tweet_detail.Tweets[id]
		assert.True(is_ok)
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

	require.Len(tweet_detail.ReplyChains, 4)
	assert.Len(tweet_detail.ReplyChains[0], 1)
	assert.Equal(tweet_detail.ReplyChains[0][0], TweetID(1413647919215906817))
	assert.Len(tweet_detail.ReplyChains[1], 2)
	assert.Equal(tweet_detail.ReplyChains[1][0], TweetID(1413657324267311104))
	assert.Equal(tweet_detail.ReplyChains[1][1], TweetID(1413658466795737091))
	assert.Len(tweet_detail.ReplyChains[2], 1)
	assert.Equal(tweet_detail.ReplyChains[2][0], TweetID(1413650853081276421))
	assert.Len(tweet_detail.ReplyChains[3], 2)
	assert.Equal(tweet_detail.ReplyChains[3][0], TweetID(1413772782358433792))
	assert.Equal(tweet_detail.ReplyChains[3][1], TweetID(1413773185296650241))
}

func TestTweetDetailWithParents(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	tweet_detail, err := profile.GetTweetDetail(TweetID(1413773185296650241))
	require.NoError(err)

	assert.Len(tweet_detail.Retweets, 0)

	assert.Len(tweet_detail.Tweets, 4)
	for _, id := range []TweetID{
		1413646309047767042,
		1413646595493568516,
		1413772782358433792,
		1413773185296650241,
	} {
		_, is_ok := tweet_detail.Tweets[id]
		assert.True(is_ok)
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
