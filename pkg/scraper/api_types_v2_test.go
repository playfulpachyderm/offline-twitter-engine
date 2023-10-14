package scraper_test

import (
	"testing"

	"encoding/json"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "gitlab.com/offline-twitter/twitter_offline_engine/pkg/scraper"
)

/**
 * Parse an APIV2User
 */
func TestAPIV2ParseUser(t *testing.T) {
	data, err := os.ReadFile("test_responses/api_v2/user_michael_malice.json")
	if err != nil {
		panic(err)
	}

	assert := assert.New(t)

	var user_result APIV2UserResult
	err = json.Unmarshal(data, &user_result)
	require.NoError(t, err)

	user := user_result.ToUser()

	assert.Equal(user.ID, UserID(44067298))
	assert.Equal(user.DisplayName, "Michael Malice")
	assert.Equal(user.Handle, UserHandle("michaelmalice"))
	assert.Equal(user.Bio, "Author of Dear Reader, The New Right & The Anarchist Handbook\nHost of \"YOUR WELCOME\" \nSubject of Ego & "+
		"Hubris by Harvey Pekar\nHe/Him ⚑\n@SheathUnderwear Model")
	assert.Equal(user.FollowingCount, 964)
	assert.Equal(user.FollowersCount, 334571)
	assert.Equal(user.Location, "Austin")
	assert.Equal(user.Website, "https://amzn.to/3oInafv")
	assert.Equal(user.JoinDate.Unix(), int64(1243920952))
	assert.Equal(user.IsPrivate, false)
	assert.Equal(user.IsVerified, true)
	assert.Equal(user.IsBanned, false)
	assert.Equal(user.ProfileImageUrl, "https://pbs.twimg.com/profile_images/1415820415314931715/_VVX4GI8.jpg")
	assert.Equal(user.BannerImageUrl, "https://pbs.twimg.com/profile_banners/44067298/1615134676")
	assert.Equal(user.PinnedTweetID, TweetID(1477347403023982596))
}

/**
 * Parse a plain text tweet
 */
func TestAPIV2ParseTweet(t *testing.T) {
	data, err := os.ReadFile("test_responses/api_v2/tweet_plaintext.json")
	if err != nil {
		panic(err)
	}
	assert := assert.New(t)

	var tweet_result APIV2Result
	err = json.Unmarshal(data, &tweet_result)
	assert.NoError(err)

	trove, err := tweet_result.ToTweetTrove()
	assert.NoError(err)

	assert.Equal(1, len(trove.Tweets))
	tweet, ok := trove.Tweets[1485708879174508550]
	assert.True(ok)
	assert.Equal(tweet.ID, TweetID(1485708879174508550))
	assert.Equal(tweet.UserID, UserID(44067298))
	assert.Equal(tweet.Text, "If Boris Johnson is driven out of office, it wouldn't mark the first time the Tories had four PMs in a "+
		"row\nThey had previously governed the UK for 13 years with 4 PMs, from 1951-1964")
	assert.Equal(tweet.PostedAt.Unix(), int64(1643055574))
	assert.Equal(tweet.QuotedTweetID, TweetID(0))
	assert.Equal(tweet.InReplyToID, TweetID(0))
	assert.Equal(tweet.NumLikes, 38)
	assert.Equal(tweet.NumRetweets, 2)
	assert.Equal(tweet.NumReplies, 2)
	assert.Equal(tweet.NumQuoteTweets, 1)
	assert.Equal(0, len(tweet.Images))
	assert.Equal(0, len(tweet.Videos))
	assert.Equal(0, len(tweet.Polls))
	assert.Equal(0, len(tweet.Mentions))
	assert.Equal(0, len(tweet.ReplyMentions))
	assert.Equal(0, len(tweet.Hashtags))
	assert.Equal("", tweet.TombstoneType)
	assert.False(tweet.IsStub)

	assert.Equal(1, len(trove.Users))
	user, ok := trove.Users[44067298]
	assert.True(ok)
	assert.Equal(UserID(44067298), user.ID)
	assert.Equal(UserHandle("michaelmalice"), user.Handle)

	assert.Equal(0, len(trove.Retweets))
}

/**
 * Parse a tweet with a quoted tweet
 */
func TestAPIV2ParseTweetWithQuotedTweet(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/tweet_with_quoted_tweet.json")
	if err != nil {
		panic(err)
	}

	var tweet_result APIV2Result
	err = json.Unmarshal(data, &tweet_result)
	assert.NoError(err)

	trove, err := tweet_result.ToTweetTrove()
	assert.NoError(err)

	// Should be 2 tweets: quote-tweet and quoted-tweet
	assert.Equal(2, len(trove.Tweets))

	quoted_tweet, ok := trove.Tweets[1485690069079846915]
	assert.True(ok)
	assert.Equal(TweetID(1485690069079846915), quoted_tweet.ID)
	assert.Equal(UserID(892155218292617217), quoted_tweet.UserID)
	assert.Equal("The Left hates the Right so much that they won't let them leave the Union. I don't get it.", quoted_tweet.Text)
	assert.Equal(int64(1643051089), quoted_tweet.PostedAt.Unix())
	assert.Equal(TweetID(1485689207435710464), quoted_tweet.InReplyToID)
	assert.Equal(TweetID(0), quoted_tweet.QuotedTweetID)
	assert.Equal(1, len(quoted_tweet.ReplyMentions))
	assert.Contains(quoted_tweet.ReplyMentions, "michaelmalice")
	assert.Equal(1, quoted_tweet.NumReplies)
	assert.Equal(12, quoted_tweet.NumLikes)

	quote_tweet, ok := trove.Tweets[1485690410899021826]
	assert.True(ok)
	assert.Equal(TweetID(1485690410899021826), quote_tweet.ID)
	assert.Equal(TweetID(1485690069079846915), quote_tweet.QuotedTweetID)
	assert.Equal("Hatred is powerless in and of itself despite all the agitprop to the contrary\nHatred didnt stop Trump's election, "+
		"for example", quote_tweet.Text)

	// Should be 2 users: quoter and quoted
	assert.Equal(2, len(trove.Users))

	user_quoting, ok := trove.Users[44067298]
	assert.True(ok)
	assert.Equal(UserHandle("michaelmalice"), user_quoting.Handle)

	user_quoted, ok := trove.Users[892155218292617217]
	assert.True(ok)
	assert.Equal(UserHandle("baalzimon"), user_quoted.Handle)

	// No retweets
	assert.Equal(0, len(trove.Retweets))
}

/**
 * Parse a retweet
 */
func TestAPIV2ParseRetweet(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/retweet.json")
	if err != nil {
		panic(err)
	}

	var tweet_result APIV2Result
	err = json.Unmarshal(data, &tweet_result)
	assert.NoError(err)

	trove, err := tweet_result.ToTweetTrove()
	assert.NoError(err)

	// Should only be 1 tweet, the retweeted one
	assert.Equal(1, len(trove.Tweets))
	tweet, ok := trove.Tweets[1485694028620316673]
	assert.True(ok)
	assert.Equal(TweetID(1485694028620316673), tweet.ID)
	assert.Equal(UserID(1326229737551912960), tweet.UserID)
	assert.Equal("More mask madness, this time in an elevator. The mask police are really nuts https://t.co/3BpvLjdJwD", tweet.Text)
	assert.Equal(int64(1643052033), tweet.PostedAt.Unix())
	assert.Equal(5373, tweet.NumLikes)
	assert.Equal(TweetID(0), tweet.InReplyToID)
	assert.Equal(1, len(tweet.Videos))

	// Check the video
	v := tweet.Videos[0]
	assert.Equal("https://pbs.twimg.com/ext_tw_video_thumb/1485627274594590721/pu/img/O6mMKrsqWl8WcMy1.jpg", v.ThumbnailRemoteUrl)
	assert.Equal(0, v.ViewCount) // TODO: make this work
	assert.Equal(720, v.Height)
	assert.Equal(720, v.Width)
	assert.Equal(30066, v.Duration)

	// Should fetch both the retweeting and retweeted users
	assert.Equal(2, len(trove.Users))

	retweeted_user, ok := trove.Users[1326229737551912960]
	assert.True(ok)
	assert.Equal(UserID(1326229737551912960), retweeted_user.ID)
	assert.Equal(UserHandle("libsoftiktok"), retweeted_user.Handle)

	retweeting_user, ok := trove.Users[44067298]
	assert.True(ok)
	assert.Equal(UserID(44067298), retweeting_user.ID)
	assert.Equal(UserHandle("michaelmalice"), retweeting_user.Handle)

	// Should be 1 retweet
	assert.Equal(1, len(trove.Retweets))
	retweet, ok := trove.Retweets[1485699748514476037]
	assert.True(ok)
	assert.Equal(TweetID(1485699748514476037), retweet.RetweetID)
	assert.Equal(TweetID(1485694028620316673), retweet.TweetID)
	assert.Equal(int64(1643053397), retweet.RetweetedAt.Unix())
	assert.Equal(UserID(44067298), retweet.RetweetedByID)
}

/**
 * Parse a retweeted quote tweet
 */
func TestAPIV2ParseRetweetedQuoteTweet(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/retweet_with_quote_tweet.json")
	if err != nil {
		panic(err)
	}

	var tweet_result APIV2Result
	err = json.Unmarshal(data, &tweet_result)
	assert.NoError(err)

	trove, err := tweet_result.ToTweetTrove()
	assert.NoError(err)

	// Quoted tweet and quoting tweet
	assert.Equal(2, len(trove.Tweets))
	quoted_tweet, ok := trove.Tweets[1484900469482962944]
	assert.True(ok)
	assert.Equal(TweetID(1484900469482962944), quoted_tweet.ID)
	assert.Equal(UserID(14347972), quoted_tweet.UserID)
	assert.Equal(TweetID(1484643409130397702), quoted_tweet.QuotedTweetID)

	quoting_tweet, ok := trove.Tweets[1485272859102621697]
	assert.True(ok)
	assert.Equal(TweetID(1485272859102621697), quoting_tweet.ID)
	assert.Equal(UserID(1434720042193760256), quoting_tweet.UserID)
	assert.Equal(TweetID(1484900469482962944), quoting_tweet.QuotedTweetID)
	assert.Equal(200, quoting_tweet.NumLikes)

	// 3 Users: quoted, quoter, and retweeter
	assert.Equal(3, len(trove.Users))

	retweeting_user, ok := trove.Users[599817378]
	assert.True(ok)
	assert.Equal(UserID(599817378), retweeting_user.ID)
	assert.Equal(UserHandle("ScottMGreer"), retweeting_user.Handle)

	retweeted_user, ok := trove.Users[1434720042193760256]
	assert.True(ok)
	assert.Equal(UserID(1434720042193760256), retweeted_user.ID)
	assert.Equal(UserHandle("LatinxPutler"), retweeted_user.Handle)

	quoted_user, ok := trove.Users[14347972]
	assert.True(ok)
	assert.Equal(UserID(14347972), quoted_user.ID)
	assert.Equal(UserHandle("Heminator"), quoted_user.Handle)

	// Should be 1 retweet
	assert.Equal(1, len(trove.Retweets))
	retweet, ok := trove.Retweets[1485273090665984000]
	assert.True(ok)
	assert.Equal(TweetID(1485273090665984000), retweet.RetweetID)
	assert.Equal(TweetID(1485272859102621697), retweet.TweetID)
	assert.Equal(int64(1642951674), retweet.RetweetedAt.Unix())
	assert.Equal(UserID(599817378), retweet.RetweetedByID)
}

/**
 * Parse tweet with quoted tombstone
 */
func TestAPIV2ParseTweetWithQuotedTombstone(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/tweet_with_quoted_tombstone.json")
	if err != nil {
		panic(err)
	}

	var tweet_result APIV2Result
	err = json.Unmarshal(data, &tweet_result)
	assert.NoError(err)

	trove, err := tweet_result.ToTweetTrove()
	assert.NoError(err)

	assert.Equal(1, len(trove.Users))
	user, ok := trove.Users[44067298]
	assert.True(ok)
	assert.Equal(UserHandle("michaelmalice"), user.Handle)

	assert.Equal(1, len(trove.TombstoneUsers))
	assert.Contains(trove.TombstoneUsers, UserHandle("coltnkat"))

	assert.Equal(2, len(trove.Tweets))
	tombstoned_tweet, ok := trove.Tweets[1485774025347371008]
	assert.True(ok)
	assert.Equal(TweetID(1485774025347371008), tombstoned_tweet.ID)
	assert.Equal("no longer exists", tombstoned_tweet.TombstoneType)
	assert.True(tombstoned_tweet.IsStub)
	assert.Equal(UserHandle("coltnkat"), tombstoned_tweet.UserHandle)

	assert.Equal(0, len(trove.Retweets))
}

/**
 * Parse a tweet with a link
 */
func TestAPIV2ParseTweetWithURL(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/tweet_with_url.json")
	if err != nil {
		panic(err)
	}

	var tweet_result APIV2Result
	err = json.Unmarshal(data, &tweet_result)
	assert.NoError(err)

	trove, err := tweet_result.ToTweetTrove()
	assert.NoError(err)

	assert.Equal(1, len(trove.Tweets))
	tweet, ok := trove.Tweets[1485695695025803264]
	assert.True(ok)
	assert.Equal("This led to what I discussed as \"anguish signaling,\" where progs competed in proclaiming their distress both to "+
		"show they were the Good Guys but also to get the pack to regroup, akin to wolves howling.", tweet.Text)

	assert.Equal(1, len(tweet.Urls))
	url := tweet.Urls[0]
	assert.Equal(tweet.ID, url.TweetID)
	assert.Equal("observer.com", url.Domain)
	assert.Equal("Why Evangelical Progressives Need to Demonstrate Anguish Publicly", url.Title)
	assert.Equal("https://observer.com/2016/12/why-evangelical-progressives-need-to-demonstrate-anguish-publicly/", url.Text)
	assert.Equal("The concept of “virtue signaling” gained a great deal of currency in this past year. It’s a way to demonstrate to "+
		"others that one is a good person without having to do anything", url.Description)
	assert.Equal("https://pbs.twimg.com/card_img/1485694664640507911/WsproWyP?format=jpg&name=600x600", url.ThumbnailRemoteUrl)
	assert.Equal(600, url.ThumbnailWidth)
	assert.Equal(300, url.ThumbnailHeight)
	assert.Equal(UserID(15738599), url.SiteID)
	assert.Equal(UserID(15738599), url.CreatorID)
}

/**
 * Parse a tweet with a link with a "player" card
 */
func TestAPIV2ParseTweetWithURLPlayerCard(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/tweet_with_url_player_card.json")
	if err != nil {
		panic(err)
	}

	var tweet_result APIV2Result
	err = json.Unmarshal(data, &tweet_result)
	assert.NoError(err)

	trove, err := tweet_result.ToTweetTrove()
	assert.NoError(err)

	assert.Equal(1, len(trove.Tweets))
	tweet, ok := trove.Tweets[1485504913614327808]
	assert.True(ok)
	assert.Equal("i'll just leave this here", tweet.Text)

	assert.Equal(1, len(tweet.Urls))
	url := tweet.Urls[0]
	assert.Equal(tweet.ID, url.TweetID)
	assert.Equal("www.youtube.com", url.Domain)
	assert.Equal("Michael Malice on Kennedy Nov. 15, 2016", url.Title)
	assert.Equal("https://www.youtube.com/watch?v=c9TypEM1ik4&t=9s", url.Text)
	assert.Equal("Steve Bannon;", url.Description)
	assert.Equal("https://pbs.twimg.com/card_img/1485504774233415680/fsbK59th?format=jpg&name=800x320_1", url.ThumbnailRemoteUrl)
	assert.Equal(UserID(10228272), url.SiteID)
}

/**
 * Parse a tweet with a link with a "player" card
 */
func TestAPIV2ParseTweetWithURLRetweet(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/retweet_with_url.json")
	if err != nil {
		panic(err)
	}

	var tweet_result APIV2Result
	err = json.Unmarshal(data, &tweet_result)
	assert.NoError(err)

	trove, err := tweet_result.ToTweetTrove()
	assert.NoError(err)

	assert.Equal(1, len(trove.Tweets))
	tweet, ok := trove.Tweets[1488605073588559873]
	assert.True(ok)
	assert.Equal("REJOICE", tweet.Text)

	assert.Equal(1, len(tweet.Urls))
	url := tweet.Urls[0]
	assert.Equal(tweet.ID, url.TweetID)
	assert.Equal("tippinsights.com", url.Domain)
}

/**
 * Parse a tweet with a poll
 */
func TestAPIV2ParseTweetWithPoll(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/tweet_with_poll.json")
	if err != nil {
		panic(err)
	}

	var tweet_result APIV2Result
	err = json.Unmarshal(data, &tweet_result)
	assert.NoError(err)

	trove, err := tweet_result.ToTweetTrove()
	assert.NoError(err)

	assert.Len(trove.Tweets, 1)
	tweet, ok := trove.Tweets[1485692111106285571]
	assert.True(ok)
	assert.Len(tweet.Polls, 1)

	poll := tweet.Polls[0]
	assert.Equal(TweetID(1485692111106285571), poll.TweetID)
	assert.Equal(poll.NumChoices, 4)

	assert.Equal("1", poll.Choice1)
	assert.Equal("2", poll.Choice2)
	assert.Equal("C", poll.Choice3)
	assert.Equal("E", poll.Choice4)
	assert.Equal(891, poll.Choice1_Votes)
	assert.Equal(702, poll.Choice2_Votes)
	assert.Equal(459, poll.Choice3_Votes)
	assert.Equal(1801, poll.Choice4_Votes)

	assert.Equal(int64(1643137976), poll.VotingEndsAt.Unix())
	assert.Equal(int64(1643055638), poll.LastUpdatedAt.Unix())
	assert.Equal(1440*60, poll.VotingDuration)
}

/**
 * Parse a tweet with a Space
 */
func TestAPIV2ParseTweetWithSpace(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/api_v2/tweet_with_audiospaces_card.json")
	if err != nil {
		panic(err)
	}

	var tweet_result APIV2Result
	err = json.Unmarshal(data, &tweet_result)
	assert.NoError(err)

	trove, err := tweet_result.ToTweetTrove()
	assert.NoError(err)

	assert.Len(trove.Tweets, 1)
	tweet, ok := trove.Tweets[1497647006445146113]
	require.True(ok)
	assert.Len(tweet.Urls, 0)
	assert.Equal(tweet.SpaceID, SpaceID("1dRJZlRNZDzKB"))

	require.Len(trove.Spaces, 1)
	s, is_ok := trove.Spaces["1dRJZlRNZDzKB"]
	require.True(is_ok)

	assert.Equal(SpaceID("1dRJZlRNZDzKB"), s.ID)
	assert.Equal("https://t.co/5RLbEwQgvH", s.ShortUrl)
	assert.False(s.IsDetailsFetched)
}

func TestParseAPIV2UserFeed(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/api_v2/user_feed_apiv2.json")
	require.NoError(err)

	var feed APIV2Response
	err = json.Unmarshal(data, &feed)
	require.NoError(err)

	tweet_trove, err := feed.ToTweetTrove()
	require.NoError(err)

	// Check users
	user := tweet_trove.Users[44067298]
	assert.Equal(UserID(44067298), user.ID)
	assert.Equal("Michael Malice", user.DisplayName)

	retweeted_user := tweet_trove.Users[1326229737551912960]
	assert.Equal(UserID(1326229737551912960), retweeted_user.ID)
	assert.Equal(UserHandle("libsoftiktok"), retweeted_user.Handle)

	quote_tweeted_user := tweet_trove.Users[892155218292617217]
	assert.Equal(UserID(892155218292617217), quote_tweeted_user.ID)

	// Check retweets
	assert.Len(tweet_trove.Retweets, 2)

	// Test cursor-bottom
	bottom_cursor := feed.GetCursorBottom()
	assert.Equal("HBaYgL2Fp/T7nCkAAA==", bottom_cursor)
}

/**
 * Should correctly identify an "empty" response
 */
func TestAPIV2FeedIsEmpty(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/empty_response.json")
	if err != nil {
		panic(err)
	}
	var feed APIV2Response
	err = json.Unmarshal(data, &feed)
	require.NoError(t, err)

	assert.True(feed.IsEmpty())

	// Make sure parsing it doesn't cause an error
	trove, err := feed.ToTweetTrove()
	if err != nil {
		panic(err)
	}
	assert.Len(trove.Tweets, 0)
	assert.Len(trove.Users, 0)
	assert.Len(trove.Retweets, 0)
}

/**
 * Should get the right Instruction element
 */
func TestAPIV2GetMainInstructionFromFeed(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/user_feed_apiv2.json")
	if err != nil {
		panic(err)
	}
	var feed APIV2Response
	err = json.Unmarshal(data, &feed)
	require.NoError(t, err)

	assert.Equal(len(feed.GetMainInstruction().Entries), 41)

	// Check that they have OriginalJSON filled out
	for _, entry := range feed.GetMainInstruction().Entries {
		assert.True(len(entry.OriginalJSON) > 0)
	}

	// Test that this is a writable version
	feed.GetMainInstruction().Entries = append(feed.GetMainInstruction().Entries, APIV2Entry{
		InnerAPIV2Entry: InnerAPIV2Entry{EntryID: "asdf"},
	})
	assert.Equal(len(feed.GetMainInstruction().Entries), 42)
	assert.Equal(feed.GetMainInstruction().Entries[41].EntryID, "asdf")
}

func TestAPIV2TombstoneEntry(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/tombstone_tweet.json")
	require.NoError(t, err)

	var entry APIV2Entry
	err = json.Unmarshal(data, &entry)
	require.NoError(t, err)

	trove := entry.ToTweetTrove()
	assert.NoError(err)
	assert.Len(trove.Tweets, 1)
	assert.Len(trove.Users, 0)
	assert.Len(trove.Retweets, 0)

	tweet, is_ok := trove.Tweets[1454515503242829830]
	assert.True(is_ok)
	assert.Equal(tweet.ID, TweetID(1454515503242829830))
}

func TestAPIV2UserFeedTombstoneEntry(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/user_feed_tombstone_entry.json")
	require.NoError(t, err)

	var entry APIV2Entry
	err = json.Unmarshal(data, &entry)
	require.NoError(t, err)

	trove := entry.ToTweetTrove()
	// assert.Len(trove.Tweets, 1)
	// assert.Len(trove.Users, 1)
	assert.Len(trove.Retweets, 0)

	// tweet, is_ok := trove.Tweets[1670122897912786944]
	// assert.True(is_ok)
	// assert.Equal(tweet.ID, TweetID(1670122897912786944))
	// assert.Equal("age-restricted", tweet.TombstoneType)
	// assert.NotEqual(UserID(0), tweet.UserID)
	// user, is_ok := trove.Users[tweet.UserID]
	// assert.True(user.IsIdFake)
}

func TestTweetWithNewerVersion(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/tweet_with_newer_version.json")
	require.NoError(t, err)

	var entry APIV2Result
	err = json.Unmarshal(data, &entry)
	require.NoError(t, err)

	trove, err := entry.ToTweetTrove()
	assert.NoError(err)

	assert.Len(trove.Tweets, 1)
	tweet, is_ok := trove.Tweets[1653413433461579783]
	assert.True(is_ok)
	assert.Equal(tweet.TombstoneType, "newer-version-available")

	assert.Len(trove.Users, 0)
	assert.Len(trove.Retweets, 0)
}

func TestAPIV2ConversationThreadWithTombstones(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/conversation_thread_with_tombstones.json")
	require.NoError(t, err)

	var resp APIV2Response
	err = json.Unmarshal(data, &resp)
	require.NoError(t, err)

	trove, err := resp.ToTweetTrove()
	require.NoError(t, err)

	assert.Len(trove.Tweets, 4)
	t1, is_ok := trove.Tweets[1454515503242829830]
	assert.True(is_ok)
	assert.True(t1.IsStub)
	assert.Equal(TweetID(0), t1.InReplyToID)
	assert.NotEqual(UserID(0), t1.UserID)
	t1_user, is_ok := trove.Users[t1.UserID]
	assert.True(is_ok)
	assert.True(t1_user.IsIdFake)
	assert.Equal(UserHandle("<UNKNOWN USER>"), t1_user.Handle)

	t2, is_ok := trove.Tweets[1454521424144654344]
	assert.True(is_ok)
	assert.True(t2.IsStub)
	assert.Equal(TweetID(1454515503242829830), t2.InReplyToID)
	assert.NotEqual(UserID(0), t2.UserID)
	t2_user, is_ok := trove.Users[t2.UserID]
	assert.True(is_ok)
	assert.True(t2_user.IsIdFake)
	assert.Equal(UserHandle("<UNKNOWN USER>"), t2_user.Handle)

	t3, is_ok := trove.Tweets[1454522147750260742]
	assert.True(is_ok)
	assert.True(t3.IsStub)
	assert.Equal(TweetID(1454521424144654344), t3.InReplyToID)
	assert.Equal(UserID(1365863538393309184), t3.UserID)
	t3_user, is_ok := trove.Users[t3.UserID]
	assert.True(is_ok)
	assert.False(t3_user.IsIdFake)
	assert.Equal(UserHandle("itsbackwereover"), t3_user.Handle)
	assert.Contains(trove.TombstoneUsers, UserHandle("itsbackwereover"))

	t4, is_ok := trove.Tweets[1454526270809726977]
	assert.True(is_ok)
	assert.False(t4.IsStub)
	assert.Equal(TweetID(1454522147750260742), t4.InReplyToID)
	t4_user, is_ok := trove.Users[t4.UserID]
	assert.True(is_ok)
	assert.False(t4_user.IsIdFake)

	assert.Len(trove.Retweets, 0)
}

func TestTweetWithWarning(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/tweet_with_warning.json")
	require.NoError(t, err)
	var tweet_result APIV2Result
	err = json.Unmarshal(data, &tweet_result)
	require.NoError(t, err)

	trove, err := tweet_result.ToTweetTrove()
	assert.NoError(err)

	assert.Len(trove.Retweets, 1)
	assert.Len(trove.Tweets, 2)
	assert.Len(trove.Users, 3)
}

func TestRetweetWithVisibilityResults(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/api_v2/retweet_with_visibility_results.json")
	require.NoError(err)
	var tweet_result APIV2Result
	err = json.Unmarshal(data, &tweet_result)
	require.NoError(err)

	trove, err := tweet_result.ToTweetTrove()
	assert.NoError(err)

	assert.Len(trove.Retweets, 1)
	assert.Len(trove.Tweets, 1)
	assert.Len(trove.Users, 2)

	rt, is_ok := trove.Retweets[1595999178593034241]
	require.True(is_ok)
	assert.Equal(rt.TweetID, TweetID(1595973736833892356))
}

func TestExpandableTweet(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/api_v2/expandable_tweet.json")
	require.NoError(err)
	var tweet_result APIV2Result
	err = json.Unmarshal(data, &tweet_result)
	require.NoError(err)

	trove, err := tweet_result.ToTweetTrove()
	assert.NoError(err)
	main_tweet, is_ok := trove.Tweets[TweetID(1649600354747572225)]
	require.True(is_ok)

	assert.True(main_tweet.IsExpandable)
	assert.Equal(main_tweet.Text, "This entire millenial media era has come and gone. Where are the lindy articles from all these websites? The ideas? \n\nIt was just a decade and a half of nothing. \n\na complete waste of time. \n\nAnd it ends with the blue checks being stripped. \n\nA fitting ending to a time not worth saving") //nolint:lll // It's a string
}

// In a user feed, an "entry" can contain multiple tweets when making authenticated requests.
// They should parse out as all the tweets.
func TestEntryWithConversationThread(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/api_v2/home_conversation_entry.json")
	require.NoError(err)
	var entry_result APIV2Entry
	err = json.Unmarshal(data, &entry_result)
	require.NoError(err)

	trove := entry_result.ToTweetTrove()

	assert.Len(trove.Tweets, 4) // 3 tweets in the thread plus the quoted tweet
	t1, is_ok := trove.Tweets[1624966566264680448]
	assert.True(is_ok)
	assert.Equal(TweetID(1624953636253208578), t1.QuotedTweetID)
	assert.Equal(13, t1.NumLikes)

	_, is_ok = trove.Tweets[1624953636253208578] // Quoted tweet
	assert.True(is_ok)
	_, is_ok = trove.Tweets[1624989022589427712] // Tweet 2
	assert.True(is_ok)
	_, is_ok = trove.Tweets[1624990170670850053] // Tweet 3
	assert.True(is_ok)
}

func TestProfileConversationEntry(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/api_v2/user_feed_profile_conversation_module_entry.json")
	require.NoError(err)
	var entry_result APIV2Entry
	err = json.Unmarshal(data, &entry_result)
	require.NoError(err)

	trove := entry_result.ToTweetTrove()

	assert.Len(trove.Tweets, 2)
	_, is_ok := trove.Tweets[1671870056953790465]
	assert.True(is_ok)
	_, is_ok = trove.Users[593289567]
	assert.True(is_ok)
	_, is_ok = trove.Tweets[1671918780044435456]
	assert.True(is_ok)
	_, is_ok = trove.Users[358545917]
	assert.True(is_ok)
}

// On a Tweet Detail page, there's a thread of replies, and then it says "Show more..." underneath
// to extend the conversation.  This is different from the "Show more..." button to load more
// replies to the original tweet!
func TestConversationThreadEntryWithShowMoreButton(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/api_v2/conversation_thread_entry_with_show_more_button.json")
	require.NoError(err)
	var entry_result APIV2Entry
	err = json.Unmarshal(data, &entry_result)
	require.NoError(err)

	trove := entry_result.ToTweetTrove()

	assert.Len(trove.Tweets, 1)
	t1, is_ok := trove.Tweets[1649803385485377536]
	assert.True(is_ok)
	assert.Equal(TweetID(1649600354747572225), t1.InReplyToID)

	assert.Len(trove.Users, 1)
}

func TestTweetWithInlineLabel(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/response_with_inline_label.json")
	require.NoError(t, err)
	var resp APIV2Response

	err = json.Unmarshal(data, &resp)
	require.NoError(t, err)

	// Test the "Show additional replies, including offensive" cursor
	assert.Equal("NwAAAPAoHBlWiICz7Y27o4kploCzmY7CyYgphoC-7d65sYgphsCj7cml4ogpjMC4ue3954gpJQYRFRAAAA", resp.GetCursorBottom())

	trove, err := resp.ToTweetTrove()
	assert.NoError(err)

	assert.Len(trove.Retweets, 0)
	assert.Len(trove.Tweets, 1)
	assert.Len(trove.Users, 1)
}

func TestTweetDetailWithShowMoreButton(t *testing.T) {
	assert := assert.New(t)
	data, err := os.ReadFile("test_responses/api_v2/tweet_detail_with_show_more_replies_button.json")
	require.NoError(t, err)
	var resp APIV2Response

	err = json.Unmarshal(data, &resp)
	require.NoError(t, err)

	// Test the "Show more replies" cursor
	assert.Equal("NwAAAPANHBlWhMC--YvZ3a8ugsDS1aGWqbAugIDQ6dCa3xIAYL314NGqsAkA0OGFzYqwLiUEERUOAAA", resp.GetCursorBottom())
}

func TestConversationThreadWithTombstoneReplies(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/api_v2/tweet_detail_with_tombstone_replies.json")
	require.NoError(err)
	var resp APIV2Response
	err = json.Unmarshal(data, &resp)
	require.NoError(err)

	trove, err := resp.ToTweetTrove()
	require.NoError(err)

	t1, is_ok := trove.Tweets[1433713164546293767] // Main tweet
	assert.True(is_ok)
	assert.False(t1.IsStub)
}

func TestTweetDetailWithUnjoinedNontombstoneTweet(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/api_v2/tweet_detail_with_unjoined_nontombstone_tweet.json")
	require.NoError(err)
	var resp APIV2Response
	err = json.Unmarshal(data, &resp)
	require.NoError(err)

	trove, err := resp.ToTweetTrove()
	require.NoError(err)

	assert.Len(trove.Tweets, 3)
	t1, is_ok := trove.Tweets[1481999034328006662]
	assert.True(is_ok)
	assert.True(t1.IsStub)

	t2, is_ok := trove.Tweets[1481999536918831107]
	assert.True(is_ok)
	assert.True(t2.IsStub)
	assert.Equal(t1.ID, t2.InReplyToID)

	t3, is_ok := trove.Tweets[1482000048447705090] // Main tweet
	assert.True(is_ok)
	assert.False(t3.IsStub)
	assert.Equal(t2.ID, t3.InReplyToID)
}

func TestParseHomeTimeline(t *testing.T) {
	require := require.New(t)
	data, err := os.ReadFile("test_responses/api_v2/home_timeline.json")
	require.NoError(err)
	var response_result APIV2Response
	err = json.Unmarshal(data, &response_result)
	require.NoError(err)

	trove, err := response_result.ToTweetTrove()
	require.NoError(err)

	require.Len(trove.Tweets, 13)
	require.Len(trove.Users, 11)
}

func TestParseHomeTimelineForYou(t *testing.T) {
	require := require.New(t)
	data, err := os.ReadFile("test_responses/api_v2/home_timeline_for_you.json")
	require.NoError(err)
	var response_result APIV2Response
	err = json.Unmarshal(data, &response_result)
	require.NoError(err)

	trove, err := response_result.ToTweetTrove()
	require.NoError(err)
	require.Len(trove.Tweets, 13)
	require.Len(trove.Retweets, 1)
	require.Len(trove.Users, 8)
}

func TestParseResultAsLikes(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/api_v2/likes_feed.json")
	require.NoError(err)
	var response_result APIV2Response
	err = json.Unmarshal(data, &response_result)
	require.NoError(err)

	trove, err := response_result.ToTweetTroveAsLikes()
	require.NoError(err)

	assert.Len(trove.Retweets, 0)
	assert.True(len(trove.Likes) == 20)
	for _, l := range trove.Likes {
		_, is_ok := trove.Tweets[l.TweetID]
		assert.True(is_ok, "Like (%#v) didn't have its Tweet in the trove", l)
	}
}

func TestTweetWithImplicitQuotedTombstone(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	data, err := os.ReadFile("test_responses/api_v2/tweet_with_implicit_quoted_tombstone.json")
	require.NoError(err)
	var entry_result APIV2Entry
	err = json.Unmarshal(data, &entry_result)
	require.NoError(err)

	trove := entry_result.ToTweetTrove()

	assert.Len(trove.Tweets, 2)

	t1, is_ok := trove.Tweets[TweetID(1586033916367904768)]
	assert.True(is_ok)
	assert.False(t1.IsStub)
	t2, is_ok := trove.Tweets[TweetID(1586033437806305280)]
	assert.True(is_ok)
	assert.True(t2.IsStub)
	assert.Equal(t2.TombstoneType, "unavailable")
}
