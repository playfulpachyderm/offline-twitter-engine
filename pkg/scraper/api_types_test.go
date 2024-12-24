package scraper_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

func TestNormalizeContent(t *testing.T) {
	assert := assert.New(t)
	test_cases := []struct {
		filename            string
		eventual_full_text  string
		quoted_status_id    TweetID
		in_reply_to_id      TweetID
		retweeted_status_id TweetID
		reply_mentions      string
	}{
		{"test_responses/single_tweets/tweet_that_is_a_reply_with_gif.json", "", 0, 1395882872729477131, 0, "@michaelmalice"},
		{"test_responses/single_tweets/tweet_with_image.json", "this saddens me every time", 0, 0, 0, ""},
		{"test_responses/single_tweets/tweet_that_is_a_reply.json", "Noted", 0, 1396194494710788100, 0, "@RvaTeddy @michaelmalice"},
		{"test_responses/single_tweets/tweet_with_4_images.json", "These are public health officials who are making decisions about " +
			"your lifestyle because they know more about health, fitness and well-being than you do", 0, 0, 0, ""},
		{"test_responses/single_tweets/tweet_with_at_mentions_in_front.json", "It always does, doesn't it?", 0, 1428907275532476416, 0,
			"@rob_mose @primalpoly @jmasseypoet @SpaceX"},
		{"test_responses/single_tweets/tweet_with_unicode_chars.json", "The fact that @michaelmalice new book ‘The Anarchist Handbook’ " +
			"is just absolutely destroying on the charts is the largest white pill I’ve swallowed in years.", 0, 0, 0, ""},
		{"test_responses/single_tweets/tweet_with_quoted_tweet_as_link.json", "", 1422680899670274048, 0, 0, ""},
		{"test_responses/single_tweets/tweet_with_quoted_tweet_as_link2.json", "sometimes they're too dimwitted to even get the wrong " +
			"title right", 1396194494710788100, 1395882872729477131, 0, ""},
		{"test_responses/single_tweets/tweet_with_quoted_tweet_as_link3.json", "I was using an analogy about creating out-groups but " +
			"the Germans sure love their literalism", 1442092399358930946, 1335678942020300802, 0, ""},
		{"test_responses/single_tweets/tweet_with_html_entities.json", "By the 1970s  the elite consensus was that \"the hunt for " +
			"atomic spies\" had been a grotesque over-reaction to minor leaks that cost the lives of the Rosenbergs & ruined many " +
			"innocents. Only when the USSR fell was it discovered that they & other spies had given away ALL the secrets", 0, 0, 0, ""},
	}

	for _, v := range test_cases {
		data, err := os.ReadFile(v.filename)
		if err != nil {
			panic(err)
		}
		var tweet APITweet
		err = json.Unmarshal(data, &tweet)
		assert.NoError(err, "Failed at "+v.filename)

		tweet.NormalizeContent()

		assert.Equal(v.eventual_full_text, tweet.FullText, "Tweet text")
		assert.Equal(int64(v.quoted_status_id), tweet.QuotedStatusID, "Quoted status ID")
		assert.Equal(int64(v.in_reply_to_id), tweet.InReplyToStatusID, "In reply to ID")
		assert.Equal(int64(v.retweeted_status_id), tweet.RetweetedStatusID, "Retweeted status ID")
		assert.Equal(v.reply_mentions, tweet.Entities.ReplyMentions, "Reply mentions")
	}
}

func TestGetCursorBottom(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/midriffs_anarchist_cookbook.json")
	if err != nil {
		panic(err)
	}
	var tweet_resp APIv1Response
	err = json.Unmarshal(data, &tweet_resp)
	assert.NoError(err)

	assert.Equal("LBmGhsC+ibH1peAmgICjpbS0m98mgICj7a2lmd8mhsC4rbmsmN8mgMCqkbT1p+AmgsC4ucv4o+AmhoCyrf+nlt8mhMC9qfOwlt8mJQISAAA=",
		tweet_resp.GetCursorBottom())
}

func TestIsEndOfFeed(t *testing.T) {
	assert := assert.New(t)
	test_cases := []struct {
		filename       string
		is_end_of_feed bool
	}{
		{"test_responses/michael_malice_feed.json", false},
		{"test_responses/kwiber_end_of_feed.json", true},
	}
	for _, v := range test_cases {
		data, err := os.ReadFile(v.filename)
		if err != nil {
			panic(err)
		}
		var tweet_resp APIv1Response
		err = json.Unmarshal(data, &tweet_resp)
		assert.NoError(err)
		assert.Equal(v.is_end_of_feed, tweet_resp.IsEndOfFeed())
	}
}

func TestHandleTombstonesHidden(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/tombstones/tombstone_hidden_1.json")
	if err != nil {
		panic(err)
	}
	var tweet_resp APIv1Response
	err = json.Unmarshal(data, &tweet_resp)
	require.NoError(t, err)
	assert.Equal(2, len(tweet_resp.GlobalObjects.Tweets), "Before tombstone handling")

	tweet_resp.HandleTombstones()

	assert.Equal(4, len(tweet_resp.GlobalObjects.Tweets), "After tombstone handling")

	first_tombstone, ok := tweet_resp.GlobalObjects.Tweets["1454522147750260742"]
	if assert.True(ok, "Missing tombstone") {
		assert.Equal(int64(1454522147750260742), first_tombstone.ID)
		assert.Equal(int64(1365863538393309184), first_tombstone.UserID)
		assert.Equal("hidden", first_tombstone.TombstoneText)
	}

	second_tombstone, ok := tweet_resp.GlobalObjects.Tweets["1454515503242829830"]
	if assert.True(ok, "Missing tombstone") {
		assert.Equal(int64(1454515503242829830), second_tombstone.ID)
		assert.Equal(int64(1365863538393309184), second_tombstone.UserID)
		assert.Equal("hidden", second_tombstone.TombstoneText)
	}
}

func TestHandleTombstonesDeleted(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/tombstones/tombstone_deleted.json")
	if err != nil {
		panic(err)
	}
	var tweet_resp APIv1Response
	err = json.Unmarshal(data, &tweet_resp)
	require.NoError(t, err)
	assert.Equal(1, len(tweet_resp.GlobalObjects.Tweets), "Before tombstone handling")

	tweet_resp.HandleTombstones()

	assert.Equal(2, len(tweet_resp.GlobalObjects.Tweets), "After tombstone handling")

	tombstone, ok := tweet_resp.GlobalObjects.Tweets["1454521654781136902"]
	if assert.True(ok) {
		assert.Equal(int64(1454521654781136902), tombstone.ID)
		assert.Equal(int64(1218687933391298560), tombstone.UserID)
		assert.Equal("deleted", tombstone.TombstoneText)
	}
}

func TestHandleTombstonesUnavailable(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/tombstones/tombstone_unavailable.json")
	if err != nil {
		panic(err)
	}
	var tweet_resp APIv1Response
	err = json.Unmarshal(data, &tweet_resp)
	require.NoError(t, err)
	assert.Equal(2, len(tweet_resp.GlobalObjects.Tweets), "Before tombstone handling")

	tweet_resp.HandleTombstones()

	assert.Equal(3, len(tweet_resp.GlobalObjects.Tweets), "After tombstone handling")

	tombstone, ok := tweet_resp.GlobalObjects.Tweets["1452686887651532809"]
	if assert.True(ok) {
		assert.Equal(int64(1452686887651532809), tombstone.ID)
		assert.Equal(int64(1241389617502445569), tombstone.UserID)
		assert.Equal("unavailable", tombstone.TombstoneText)
	}
}

// Should extract a user handle from a shortened tweet URL
func TestParseHandleFromShortenedTweetUrl(t *testing.T) {
	assert := assert.New(t)

	short_url := "https://t.co/rZVrNGJyDe"
	expanded_url := "https://twitter.com/MarkSnyderJr1/status/1460857606147350529"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", short_url, func(req *http.Request) (*http.Response, error) {
		header := http.Header{}
		header.Set("Location", expanded_url)
		return &http.Response{StatusCode: 301, Header: header}, nil
	})

	// Check the httpmock interceptor is working correctly
	require.Equal(t, expanded_url, ExpandShortUrl(short_url), "httpmock didn't intercept the request")

	result, err := ParseHandleFromTweetUrl(short_url)
	require.NoError(t, err)
	assert.Equal(UserHandle("MarkSnyderJr1"), result)
}

// Should compute tiny profile image URLs correctly, and fix local paths if needed (e.g., "_normal" and no file extension)
func TestGetTinyURLs(t *testing.T) {
	assert := assert.New(t)
	u := User{
		ProfileImageUrl: "https://pbs.twimg.com/profile_images/1208124284/iwRReicO.jpg",
		Handle:          "testUser",
	}
	assert.Equal(u.GetTinyProfileImageUrl(), "https://pbs.twimg.com/profile_images/1208124284/iwRReicO_normal.jpg")
	assert.Equal(u.GetTinyProfileImageLocalPath(), "testUser_profile_iwRReicO_normal.jpg")

	// User with poorly formed profile image URL
	u.ProfileImageUrl = "https://pbs.twimg.com/profile_images/1208124284/iwRReicO_normal"
	assert.Equal(u.GetTinyProfileImageUrl(), "https://pbs.twimg.com/profile_images/1208124284/iwRReicO_normal")
	assert.Equal(u.GetTinyProfileImageLocalPath(), "testUser_profile_iwRReicO_normal.jpg")
}
