package scraper_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "offline_twitter/scraper"
)

func load_tweet_from_file(filename string) Tweet {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var apitweet APITweet
	err = json.Unmarshal(data, &apitweet)
	if err != nil {
		panic(err)
	}
	tweet, err := ParseSingleTweet(apitweet)
	if err != nil {
		panic(err)
	}
	return tweet
}

func TestParseSingleTweet(t *testing.T) {
	assert := assert.New(t)
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_unicode_chars.json")

	assert.Equal("The fact that @michaelmalice new book ‘The Anarchist Handbook’ is just absolutely destroying on the charts is the "+
		"largest white pill I’ve swallowed in years.", tweet.Text)
	assert.Len(tweet.Mentions, 1)
	assert.Contains(tweet.Mentions, UserHandle("michaelmalice"))
	assert.Empty(tweet.Urls)
	assert.Equal(int64(1621639105), tweet.PostedAt.Unix())
	assert.Zero(tweet.QuotedTweetID)
	assert.Empty(tweet.Polls)
}

func TestParseTweetWithImage(t *testing.T) {
	assert := assert.New(t)
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_image.json")

	assert.Equal("this saddens me every time", tweet.Text)
	assert.Len(tweet.Images, 1)
}

/**
 * Ensure the fake url (link to the quoted tweet) is not parsed as a URL; it should just be ignored
 */
func TestParseTweetWithQuotedTweetAsLink(t *testing.T) {
	assert := assert.New(t)
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_quoted_tweet_as_link2.json")

	assert.Equal("sometimes they're too dimwitted to even get the wrong title right", tweet.Text)
	assert.Equal(TweetID(1395882872729477131), tweet.InReplyToID)
	assert.Equal(TweetID(1396194494710788100), tweet.QuotedTweetID)
	assert.Empty(tweet.ReplyMentions)
	assert.Empty(tweet.Polls)
	assert.Empty(tweet.Urls)
}

/**
 * Quote-tweets with links should work properly
 */
func TestParseTweetWithQuotedTweetAndLink(t *testing.T) {
	assert := assert.New(t)
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_quoted_tweet_and_url.json")

	assert.Equal("This is video he’s talking about. Please watch. Is there a single US politician capable of doing this with the "+
		"weasels and rats running American industry today?", tweet.Text)
	assert.Equal(TweetID(1497997890999898115), tweet.QuotedTweetID)

	assert.Len(tweet.Urls, 1)
	url := tweet.Urls[0]
	assert.Equal(url.Text, "https://youtu.be/VjrlTMvirVo")
}

func TestParseTweetWithVideo(t *testing.T) {
	assert := assert.New(t)
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_video.json")
	assert.Empty(tweet.Images)
	assert.Len(tweet.Videos, 1)

	v := tweet.Videos[0]
	assert.Equal("https://video.twimg.com/ext_tw_video/1418951950020845568/pu/vid/720x1280/sm4iL9_f8Lclh0aa.mp4?tag=12", v.RemoteURL)
	assert.False(v.IsGif)
}

func TestParseTweetWith2Videos(t *testing.T) {
	assert := assert.New(t)
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_2_videos.json")
	assert.Empty(tweet.Images)
	assert.Len(tweet.Videos, 2)

	v1 := tweet.Videos[0]
	assert.Equal("https://video.twimg.com/ext_tw_video/1579701730148847617/pu/vid/576x576/ghA0fyf58v-2naWR.mp4?tag=12", v1.RemoteURL)
	assert.False(v1.IsGif)
	assert.Equal("gh/ghA0fyf58v-2naWR.mp4", v1.LocalFilename)
	assert.Equal("xU/xUlghaCXbPOVN7vI.jpg", v1.ThumbnailLocalPath)

	v2 := tweet.Videos[1]
	assert.Equal("https://video.twimg.com/ext_tw_video/1579701730157252608/pu/vid/480x480/VQ69Ut84XT2BgIzX.mp4?tag=12", v2.RemoteURL)
	assert.False(v2.IsGif)
	assert.Equal("VQ/VQ69Ut84XT2BgIzX.mp4", v2.LocalFilename)
	assert.Equal("dY/dYN55HDytKvM1Bi8.jpg", v2.ThumbnailLocalPath)
}

func TestParseTweetWithImageAndVideo(t *testing.T) {
	assert := assert.New(t)
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_image_and_video.json")
	assert.Len(tweet.Images, 1)
	assert.Len(tweet.Videos, 1)

	img := tweet.Images[0]
	assert.Equal(img.ID, ImageID(1579292192580911104))
	assert.Equal(img.RemoteURL, "https://pbs.twimg.com/media/FerF4bdVQAAKeYJ.jpg")

	vid := tweet.Videos[0]
	assert.Equal(vid.ID, VideoID(1579292197752430592))
	assert.Equal(vid.ThumbnailRemoteUrl, "https://pbs.twimg.com/ext_tw_video_thumb/1579292197752430592/pu/img/soG4wMWOy3AVpllM.jpg")
	assert.Equal(vid.RemoteURL, "https://video.twimg.com/ext_tw_video/1579292197752430592/pu/vid/640x750/UE-PSqG2EE5N2dN8.mp4?tag=12")
}

func TestParseTweetWithGif(t *testing.T) {
	assert := assert.New(t)
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_that_is_a_reply_with_gif.json")
	assert.Len(tweet.Videos, 1)

	v := tweet.Videos[0]
	assert.Equal("https://video.twimg.com/tweet_video/E189-VhVoAYcrDv.mp4", v.RemoteURL)
	assert.True(v.IsGif)
}

func TestParseTweetWithUrl(t *testing.T) {
	assert := assert.New(t)
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_url_card.json")
	assert.Empty(tweet.Polls)
	assert.Len(tweet.Urls, 1)

	u := tweet.Urls[0]
	assert.Equal("https://reason.com/2021/08/30/la-teachers-union-cecily-myart-cruz-learning-loss/", u.Text)
	assert.Equal("https://t.co/Y1lWjNEiPK", u.ShortText)
	assert.True(u.HasCard)
	assert.Equal("reason.com", u.Domain)
}

func TestParseTweetWithUrlButNoCard(t *testing.T) {
	assert := assert.New(t)
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_url_but_no_card.json")
	assert.Len(tweet.Urls, 1)

	u := tweet.Urls[0]
	assert.Equal("https://www.politico.com/newsletters/west-wing-playbook/2021/09/16/the-jennifer-rubin-wh-symbiosis-494364", u.Text)
	assert.Equal("https://t.co/ZigZyLctwt", u.ShortText)
	assert.False(u.HasCard)
}

func TestParseTweetWithMultipleUrls(t *testing.T) {
	assert := assert.New(t)
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_multiple_urls.json")
	assert.Empty(tweet.Polls)
	assert.Len(tweet.Urls, 3)

	assert.False(tweet.Urls[0].HasCard)
	assert.False(tweet.Urls[1].HasCard)
	assert.True(tweet.Urls[2].HasCard)

	assert.Equal("Biden’s victory came from the suburbs", tweet.Urls[2].Title)
}

func TestTweetWithLotsOfReplyMentions(t *testing.T) {
	assert := assert.New(t)
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_at_mentions_in_front.json")
	assert.Len(tweet.ReplyMentions, 4)

	for i, v := range []UserHandle{"rob_mose", "primalpoly", "jmasseypoet", "SpaceX"} {
		assert.Equal(v, tweet.ReplyMentions[i])
	}
}

func TestTweetWithPoll(t *testing.T) {
	assert := assert.New(t)
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_poll_4_choices.json")
	assert.Len(tweet.Polls, 1)

	p := tweet.Polls[0]
	assert.Equal(tweet.ID, p.TweetID)
	assert.Equal(4, p.NumChoices)
	assert.Equal("Tribal armband", p.Choice1)
	assert.Equal("Marijuana leaf", p.Choice2)
	assert.Equal("Butterfly", p.Choice3)
	assert.Equal("Maple leaf", p.Choice4)
	assert.Equal(1593, p.Choice1_Votes)
	assert.Equal(624, p.Choice2_Votes)
	assert.Equal(778, p.Choice3_Votes)
	assert.Equal(1138, p.Choice4_Votes)
	assert.Equal(1440*60, p.VotingDuration)
	assert.Equal(int64(1638331934), p.VotingEndsAt.Unix())
	assert.Equal(int64(1638331935), p.LastUpdatedAt.Unix())
}

func TestTweetWithSpace(t *testing.T) {
	assert := assert.New(t)
	tweet := load_tweet_from_file("test_responses/single_tweets/tweet_with_space_card.json")
	assert.Len(tweet.Urls, 0)
	assert.Len(tweet.Spaces, 1)

	s := tweet.Spaces[0]
	assert.Equal(SpaceID("1YpKkZVyQjoxj"), s.ID)
}

func TestParseTweetResponse(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/michael_malice_feed.json")
	if err != nil {
		panic(err)
	}
	var tweet_resp TweetResponse
	err = json.Unmarshal(data, &tweet_resp)
	require.NoError(t, err)

	trove, err := ParseTweetResponse(tweet_resp)
	require.NoError(t, err)
	tweets, retweets, users := trove.Transform()

	assert.Len(tweets, 29-3)
	assert.Len(retweets, 3)
	assert.Len(users, 9)
}

func TestParseTweetResponseWithTombstones(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/tombstones/tombstone_deleted.json")
	if err != nil {
		panic(err)
	}
	var tweet_resp TweetResponse
	err = json.Unmarshal(data, &tweet_resp)
	require.NoError(t, err)

	extra_users := tweet_resp.HandleTombstones()
	assert.Len(extra_users, 1)

	trove, err := ParseTweetResponse(tweet_resp)
	require.NoError(t, err)
	tweets, retweets, users := trove.Transform()

	assert.Len(tweets, 2)
	assert.Len(retweets, 0)
	assert.Len(users, 1)
}
