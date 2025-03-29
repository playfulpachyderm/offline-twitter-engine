package persistence_test

import (
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/persistence"
)

func TestTokenizeSearchString(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	c, err := NewCursorFromSearchQuery("think")
	require.NoError(err)
	assert.Len(c.Keywords, 1)
	assert.Equal(c.Keywords[0], "think")
}

func TestTokenizeSearchStringMultipleWords(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	c, err := NewCursorFromSearchQuery("think tank")
	require.NoError(err)
	assert.Len(c.Keywords, 2)
	assert.Equal(c.Keywords[0], "think")
	assert.Equal(c.Keywords[1], "tank")
}

func TestTokenizeSearchStringQuotedTokens(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	c, err := NewCursorFromSearchQuery("\"think tank\"")
	require.NoError(err)
	assert.Len(c.Keywords, 1)
	assert.Equal("think tank", c.Keywords[0])
}

func TestTokenizeSearchStringFromUser(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	c, err := NewCursorFromSearchQuery("from:cernovich retweeted_by:blehbleh to:somebody")
	require.NoError(err)
	assert.Len(c.Keywords, 0)
	assert.Equal(c.FromUserHandle, UserHandle("cernovich"))
	assert.Equal(c.RetweetedByUserHandle, UserHandle("blehbleh"))
	assert.Equal(c.ToUserHandles, []UserHandle{"somebody"})
}

func TestTokenizeTokenizeComplexSearchString(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	c, err := NewCursorFromSearchQuery("stupid \"think tank\" from:kashi")
	require.NoError(err)
	assert.Len(c.Keywords, 2)
	assert.Equal("stupid", c.Keywords[0])
	assert.Equal("think tank", c.Keywords[1])
	assert.Equal(c.FromUserHandle, UserHandle("kashi"))
}

func TestTokenizeSearchStringBadQuotes(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	_, err := NewCursorFromSearchQuery("asdf \"fjk")
	require.Error(err)
	assert.ErrorIs(err, ErrUnmatchedQuotes)
	assert.ErrorIs(err, ErrInvalidQuery)
}

func TestTokenizeSearchWithDates(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	c, err := NewCursorFromSearchQuery("since:2020-01-01 until:2020-05-01")
	require.NoError(err)
	assert.Equal(c.SinceTimestamp.Time, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	assert.Equal(c.UntilTimestamp.Time, time.Date(2020, 5, 1, 0, 0, 0, 0, time.UTC))
}

func TestTokenizeSearchWithInvalidDates(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	_, err := NewCursorFromSearchQuery("since:fawejk")
	require.Error(err)
	assert.ErrorIs(err, ErrInvalidQuery)
}

func TestTokenizeSearchContentFilters(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	c, err := NewCursorFromSearchQuery("filter:links filter:videos filter:images filter:polls filter:spaces")
	require.NoError(err)
	assert.Equal(c.FilterLinks, REQUIRE)
	assert.Equal(c.FilterVideos, REQUIRE)
	assert.Equal(c.FilterImages, REQUIRE)
	assert.Equal(c.FilterPolls, REQUIRE)
	assert.Equal(c.FilterSpaces, REQUIRE)
}

func TestTokenizeSearchQuoteTweets(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	c, err := NewCursorFromSearchQuery("quoted_tweet_id:12345")
	require.NoError(err)
	assert.Equal(c.QuotedTweetID, TweetID(12345))

	c, err = NewCursorFromSearchQuery("quoted_tweet_id:1234d5")
	require.Error(err)
}
