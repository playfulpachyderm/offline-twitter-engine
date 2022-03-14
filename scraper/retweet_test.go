package scraper_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "offline_twitter/scraper"
)

func TestParseSingleRetweet(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/tweet_that_is_a_retweet.json")
	if err != nil {
		panic(err)
	}
	var api_tweet APITweet
	err = json.Unmarshal(data, &api_tweet)
	require.NoError(t, err)

	retweet, err := ParseSingleRetweet(api_tweet)
	require.NoError(t, err)

	assert.Equal(TweetID(1404270043018448896), retweet.RetweetID)
	assert.Equal(TweetID(1404269989646028804), retweet.TweetID)
	assert.Equal(UserID(44067298), retweet.RetweetedByID)
	assert.Equal(int64(1623639042), retweet.RetweetedAt.Unix())
}
