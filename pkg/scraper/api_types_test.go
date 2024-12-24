package scraper_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

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

func TestParseAPIMedia(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/tweet_content/image.json")
	if err != nil {
		panic(err)
	}
	var apimedia APIMedia
	err = json.Unmarshal(data, &apimedia)
	require.NoError(t, err)

	image := ParseAPIMedia(apimedia)
	assert.Equal(ImageID(1395882862289772553), image.ID)
	assert.Equal("https://pbs.twimg.com/media/E18sEUrWYAk8dBl.jpg", image.RemoteURL)
	assert.Equal(593, image.Width)
	assert.Equal(239, image.Height)
	assert.Equal("E1/E18sEUrWYAk8dBl.jpg", image.LocalFilename)
	assert.False(image.IsDownloaded)
}

func TestParsePoll2Choices(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/tweet_content/poll_card_2_options.json")
	if err != nil {
		panic(err)
	}
	var apiCard APICard
	err = json.Unmarshal(data, &apiCard)
	require.NoError(t, err)

	poll := ParseAPIPoll(apiCard)
	assert.Equal(PollID(1457419248461131776), poll.ID)
	assert.Equal(2, poll.NumChoices)
	assert.Equal(60*60*24, poll.VotingDuration)
	assert.Equal(int64(1636397201), poll.VotingEndsAt.Unix())
	assert.Equal(int64(1636318755), poll.LastUpdatedAt.Unix())

	assert.Less(poll.LastUpdatedAt.Unix(), poll.VotingEndsAt.Unix())
	assert.Equal("Yes", poll.Choice1)
	assert.Equal("No", poll.Choice2)
	assert.Equal(529, poll.Choice1_Votes)
	assert.Equal(2182, poll.Choice2_Votes)
}

func TestParsePoll4Choices(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/tweet_content/poll_card_4_options_ended.json")
	if err != nil {
		panic(err)
	}
	var apiCard APICard
	err = json.Unmarshal(data, &apiCard)
	require.NoError(t, err)

	poll := ParseAPIPoll(apiCard)
	assert.Equal(PollID(1455611588854140929), poll.ID)
	assert.Equal(4, poll.NumChoices)
	assert.Equal(60*60*24, poll.VotingDuration)
	assert.Equal(int64(1635966221), poll.VotingEndsAt.Unix())
	assert.Equal(int64(1635966226), poll.LastUpdatedAt.Unix())
	assert.Greater(poll.LastUpdatedAt.Unix(), poll.VotingEndsAt.Unix())

	assert.Equal("Alec Baldwin", poll.Choice1)
	assert.Equal(1669, poll.Choice1_Votes)

	assert.Equal("Andew Cuomo", poll.Choice2)
	assert.Equal(272, poll.Choice2_Votes)

	assert.Equal("George Floyd", poll.Choice3)
	assert.Equal(829, poll.Choice3_Votes)

	assert.Equal("Derek Chauvin", poll.Choice4)
	assert.Equal(2397, poll.Choice4_Votes)
}

func TestPollHelpers(t *testing.T) {
	assert := assert.New(t)
	p := Poll{
		Choice1_Votes: 1,
		Choice2_Votes: 2,
		Choice3_Votes: 3,
		Choice4_Votes: 4,
		VotingEndsAt:  Timestamp{Time: time.Now().Add(10 * time.Second)},
	}
	assert.Equal(p.TotalVotes(), 10)
	assert.Equal(p.VotePercentage(p.Choice3_Votes), 30.0)

	assert.True(p.IsOpen())
	assert.False(p.IsWinner(p.Choice4_Votes))

	// End the poll
	p.VotingEndsAt = Timestamp{Time: time.Now().Add(-10 * time.Second)}
	assert.False(p.IsOpen())
	assert.False(p.IsWinner(p.Choice2_Votes))
	assert.True(p.IsWinner(p.Choice4_Votes))
}

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

func TestParseAPIUrlCard(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/tweet_content/url_card.json")
	if err != nil {
		panic(err)
	}
	var apiCard APICard
	err = json.Unmarshal(data, &apiCard)
	require.NoError(t, err)

	url := ParseAPIUrlCard(apiCard)
	assert.Equal("reason.com", url.Domain)
	assert.Equal("L.A. Teachers Union Leader: 'There's No Such Thing As Learning Loss'", url.Title)
	assert.Equal("\"It’s OK that our babies may not have learned all their times tables,\" says Cecily Myart-Cruz. \"They learned "+
		"resilience.\"", url.Description)
	assert.Equal(600, url.ThumbnailWidth)
	assert.Equal(315, url.ThumbnailHeight)
	assert.Equal("https://pbs.twimg.com/card_img/1434998862305968129/odDi9EqO?format=jpg&name=600x600", url.ThumbnailRemoteUrl)
	assert.Equal("od/odDi9EqO_600x600.jpg", url.ThumbnailLocalPath)
	assert.Equal(UserID(155581583), url.CreatorID)
	assert.Equal(UserID(16467567), url.SiteID)
	assert.True(url.HasThumbnail)
	assert.False(url.IsContentDownloaded)
}

func TestParseAPIUrlCardWithPlayer(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/tweet_content/url_card_with_player.json")
	if err != nil {
		panic(err)
	}
	var apiCard APICard
	err = json.Unmarshal(data, &apiCard)
	require.NoError(t, err)

	url := ParseAPIUrlCard(apiCard)
	assert.Equal("www.youtube.com", url.Domain)
	assert.Equal("The Politically Incorrect Guide to the Constitution (Starring Tom...", url.Title)
	assert.Equal("Watch this episode on LBRY/Odysee: https://odysee.com/@capitalresearch:5/the-politically-incorrect-guide-to-the:8"+
		"Watch this episode on Rumble: https://rumble...", url.Description)
	assert.Equal("https://pbs.twimg.com/card_img/1437849456423194639/_1t0btyt?format=jpg&name=800x320_1", url.ThumbnailRemoteUrl)
	assert.Equal("_1/_1t0btyt_800x320_1.jpg", url.ThumbnailLocalPath)
	assert.Equal(UserID(10228272), url.SiteID)
	assert.True(url.HasThumbnail)
	assert.False(url.IsContentDownloaded)
}

func TestParseAPIUrlCardWithPlayerAndPlaceholderThumbnail(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/tweet_content/url_card_with_player_placeholder_image.json")
	if err != nil {
		panic(err)
	}
	var apiCard APICard
	err = json.Unmarshal(data, &apiCard)
	require.NoError(t, err)

	url := ParseAPIUrlCard(apiCard)
	assert.Equal("www.youtube.com", url.Domain)
	assert.Equal("Did Michael Malice Turn Me into an Anarchist? | Ep 181", url.Title)
	assert.Equal("SUBSCRIBE TO THE NEW SHOW W/ ELIJAH & SYDNEY: \"YOU ARE HERE\"YT: https://www.youtube.com/youareheredaily____________"+
		"__________________________________________...", url.Description)
	assert.Equal("https://pbs.twimg.com/cards/player-placeholder.png", url.ThumbnailRemoteUrl)
	assert.Equal("player-placeholder.png", url.ThumbnailLocalPath)
	assert.Equal(UserID(10228272), url.SiteID)
	assert.True(url.HasThumbnail)
	assert.False(url.IsContentDownloaded)
}

func TestParseAPIUrlCardWithoutThumbnail(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/tweet_content/url_card_without_thumbnail.json")
	if err != nil {
		panic(err)
	}
	var apiCard APICard
	err = json.Unmarshal(data, &apiCard)
	require.NoError(t, err)

	url := ParseAPIUrlCard(apiCard)
	assert.Equal("en.m.wikipedia.org", url.Domain)
	assert.Equal("Entryism - Wikipedia", url.Title)
	assert.Equal("", url.Description)
	assert.True(url.HasCard)
	assert.False(url.HasThumbnail)
}

// Should check if a url is a tweet url, and if so, parse it
func TestParseTweetUrl(t *testing.T) {
	assert := assert.New(t)

	// Test valid tweet url
	url := "https://twitter.com/kanesays23/status/1429583672827465730"
	handle, id, is_ok := TryParseTweetUrl(url)
	assert.True(is_ok)
	assert.Equal(UserHandle("kanesays23"), handle)
	assert.Equal(TweetID(1429583672827465730), id)

	// Test url with GET params
	handle, id, is_ok = TryParseTweetUrl("https://twitter.com/NerdNoticing/status/1263192389050654720?s=20")
	assert.True(is_ok)
	assert.Equal(UserHandle("NerdNoticing"), handle)
	assert.Equal(TweetID(1263192389050654720), id)

	// Test a `mobile.twitter.com` url
	handle, id, is_ok = TryParseTweetUrl("https://mobile.twitter.com/APhilosophae/status/1497720548540964864")
	assert.True(is_ok)
	assert.Equal(UserHandle("APhilosophae"), handle)
	assert.Equal(TweetID(1497720548540964864), id)

	// Test a `x.com` url
	handle, id, is_ok = TryParseTweetUrl("https://x.com/brutedeforce/status/1579695139425222657?s=46")
	assert.True(is_ok)
	assert.Equal(UserHandle("brutedeforce"), handle)
	assert.Equal(TweetID(1579695139425222657), id)

	// Test invalid url
	_, _, is_ok = TryParseTweetUrl("https://twitter.com/NerdNoticing/status/1263192389050654720s=20")
	assert.False(is_ok)

	// Test empty string
	_, _, is_ok = TryParseTweetUrl("")
	assert.False(is_ok)
}

// Should extract a user handle from a tweet URL, or fail if URL is invalid
func TestParseHandleFromTweetUrl(t *testing.T) {
	assert := assert.New(t)

	// Test valid tweet url
	url := "https://twitter.com/kanesays23/status/1429583672827465730"
	result, err := ParseHandleFromTweetUrl(url)
	assert.NoError(err)
	assert.Equal(UserHandle("kanesays23"), result)

	// Test url with GET params
	result, err = ParseHandleFromTweetUrl("https://twitter.com/NerdNoticing/status/1263192389050654720?s=20")
	assert.NoError(err)
	assert.Equal(UserHandle("NerdNoticing"), result)

	// Test invalid url
	_, err = ParseHandleFromTweetUrl("https://twitter.com/NerdNoticing/status/1263192389050654720s=20")
	assert.Error(err)

	// Test empty string
	_, err = ParseHandleFromTweetUrl("")
	assert.Error(err)
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

func TestParseAPIVideo(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/tweet_content/video.json")
	require.NoError(err)

	var apivideo APIExtendedMedia
	err = json.Unmarshal(data, &apivideo)
	require.NoError(err)

	video := ParseAPIVideo(apivideo)
	assert.Equal(VideoID(1418951950020845568), video.ID)
	assert.Equal(1280, video.Height)
	assert.Equal(720, video.Width)
	assert.Equal("https://video.twimg.com/ext_tw_video/1418951950020845568/pu/vid/720x1280/sm4iL9_f8Lclh0aa.mp4?tag=12", video.RemoteURL)
	assert.Equal("sm/sm4iL9_f8Lclh0aa.mp4", video.LocalFilename)
	assert.Equal("https://pbs.twimg.com/ext_tw_video_thumb/1418951950020845568/pu/img/eUTaYYfuAJ8FyjUi.jpg", video.ThumbnailRemoteUrl)
	assert.Equal("eU/eUTaYYfuAJ8FyjUi.jpg", video.ThumbnailLocalPath)
	assert.Equal(275952, video.ViewCount)
	assert.Equal(88300, video.Duration)
	assert.False(video.IsDownloaded)
}

func TestParseGeoblockedVideo(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/tweet_content/video_geoblocked.json")
	require.NoError(err)

	var apivideo APIExtendedMedia
	err = json.Unmarshal(data, &apivideo)
	require.NoError(err)

	video := ParseAPIVideo(apivideo)
	assert.True(video.IsGeoblocked)
}
