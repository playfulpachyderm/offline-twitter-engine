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

	feed, err := profile.GetUserFeed(358545917, 2, TimestampFromUnix(0))
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

	assert.Equal(feed.BottomTimestamp(), TimestampFromUnix(1644107102))
}

// Should load a feed in the middle (i.e., after some timestamp)
func TestBuildUserFeedPage2(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	feed, err := profile.GetUserFeed(358545917, 2, TimestampFromUnix(1644107102))
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

	assert.Equal(feed.BottomTimestamp(), TimestampFromUnix(1635367140))
}

// When the end of the feed is reached, an "End of feed" error should be raised
func TestBuildUserFeedEnd(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	profile, err := persistence.LoadProfile("../../sample_data/profile")
	require.NoError(err)

	feed, err := profile.GetUserFeed(358545917, 2, TimestampFromUnix(1)) // Won't be anything after "1"
	require.Error(err)
	require.ErrorIs(err, persistence.ErrEndOfFeed)

	assert.Len(feed.Retweets, 0)
	assert.Len(feed.Tweets, 0)
	assert.Len(feed.Users, 0)
	require.Len(feed.Items, 0)
}
