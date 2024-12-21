package scraper_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func TestParseSingleRetweet(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/tweet_that_is_a_retweet.json")
	if err != nil {
		panic(err)
	}
	var api_tweet APITweet
	err = json.Unmarshal(data, &api_tweet)
	require.NoError(err)

	trove, err := api_tweet.ToTweetTrove()
	require.NoError(err)

	require.Len(trove.Tweets, 0)
	require.Len(trove.Retweets, 1)

	retweet, is_ok := trove.Retweets[TweetID(1404270043018448896)]
	require.True(is_ok)

	assert.Equal(TweetID(1404270043018448896), retweet.RetweetID)
	assert.Equal(TweetID(1404269989646028804), retweet.TweetID)
	assert.Equal(UserID(44067298), retweet.RetweetedByID)
	assert.Equal(int64(1623639042), retweet.RetweetedAt.Unix())
}
