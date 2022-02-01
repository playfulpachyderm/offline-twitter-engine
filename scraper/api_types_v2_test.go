package scraper_test

import (
	"testing"
	"io/ioutil"
	"encoding/json"
	"fmt"

	"github.com/stretchr/testify/assert"

	. "offline_twitter/scraper"
)

/**
 * Parse an  APIV2User
 */
func TestAPIV2ParseUser(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/api_v2/user_michael_malice.json")
	if err != nil {
		panic(err)
	}

	assert := assert.New(t)

	var user_result APIV2UserResult
	err = json.Unmarshal(data, &user_result)
	if err != nil {
		t.Errorf(err.Error())
	}

	user := user_result.ToUser()

	assert.Equal(user.ID, UserID(44067298))
	assert.Equal(user.DisplayName, "Michael Malice")
	assert.Equal(user.Handle, UserHandle("michaelmalice"))
	assert.Equal(user.Bio, "Author of Dear Reader, The New Right & The Anarchist Handbook\nHost of \"YOUR WELCOME\" \nSubject of Ego & Hubris by Harvey Pekar\nHe/Him ⚑\n@SheathUnderwear Model")
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
	data, err := ioutil.ReadFile("test_responses/api_v2/tweet_plaintext.json")
	if err != nil {
		panic(err)
	}
	assert := assert.New(t)

	var tweet_result APIV2Result
	err = json.Unmarshal(data, &tweet_result)
	assert.NoError(err)

	trove := tweet_result.ToTweetTrove()

	assert.Equal(1, len(trove.Tweets))
	tweet, ok := trove.Tweets[1485708879174508550]
	assert.True(ok)
	assert.Equal(tweet.ID, TweetID(1485708879174508550))
	assert.Equal(tweet.UserID, UserID(44067298))
	assert.Equal(tweet.Text, "If Boris Johnson is driven out of office, it wouldn't mark the first time the Tories had four PMs in a row\nThey had previously governed the UK for 13 years with 4 PMs, from 1951-1964")
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
	assert.Equal(0, len(tweet.Polls))
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
	data, err := ioutil.ReadFile("test_responses/api_v2/tweet_with_quoted_tweet.json")
	if err != nil {
		panic(err)
	}

	var tweet_result APIV2Result
	err = json.Unmarshal(data, &tweet_result)
	assert.NoError(err)

	trove := tweet_result.ToTweetTrove()

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
	assert.Contains(quoted_tweet.ReplyMentions, UserHandle("michaelmalice"))
	assert.Equal(1, quoted_tweet.NumReplies)
	assert.Equal(12, quoted_tweet.NumLikes)

	quote_tweet, ok := trove.Tweets[1485690410899021826]
	assert.True(ok)
	assert.Equal(TweetID(1485690410899021826), quote_tweet.ID)
	assert.Equal(TweetID(1485690069079846915), quote_tweet.QuotedTweetID)
	assert.Equal("Hatred is powerless in and of itself despite all the agitprop to the contrary\nHatred didnt stop Trump's election, for example", quote_tweet.Text)

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
	data, err := ioutil.ReadFile("test_responses/api_v2/retweet.json")
	if err != nil {
		panic(err)
	}

	var tweet_result APIV2Result
	err = json.Unmarshal(data, &tweet_result)
	assert.NoError(err)

	trove := tweet_result.ToTweetTrove()

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
	assert.Equal(0, v.ViewCount)  // TODO: make this work
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
	data, err := ioutil.ReadFile("test_responses/api_v2/retweet_with_quote_tweet.json")
	if err != nil {
		panic(err)
	}

	var tweet_result APIV2Result
	err = json.Unmarshal(data, &tweet_result)
	assert.NoError(err)

	trove := tweet_result.ToTweetTrove()

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


// Check a plain old tweet
func TestAPIV2FeedSimpleTweet(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/api_v2/feed_simple_tweet.json")
	if err != nil {
		panic(err)
	}
	var feed APIV2Response
	err = json.Unmarshal(data, &feed)
	if err != nil {
		t.Errorf(err.Error())
	}

	tweet_trove, err := feed.ToTweetTrove()
	if err != nil {
		panic(err)
	}

	if len(tweet_trove.Users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(tweet_trove.Users))
	}
	user := tweet_trove.Users[44067298]
	if user.ID != 44067298 {
		t.Errorf("Expected ID %d, got %d", 44067298, user.ID)
	}
	if user.DisplayName != "Michael Malice" {
		t.Errorf("Expected display name %q, got %q", "Michael Malice", user.DisplayName)
	}


	if len(tweet_trove.Tweets) != 1 {
		t.Errorf("Expected %d tweets, got %d", 1, len(tweet_trove.Tweets))
	}
	tweet := tweet_trove.Tweets[1485708879174508550]
	if tweet.ID != 1485708879174508550 {
		t.Errorf("Expected ID 1485708879174508550, got %d", tweet.ID)
	}
	if tweet.UserID != UserID(44067298) {
		t.Errorf("Expected user ID 44067298, got %d", tweet.UserID)
	}
	expected_text := "If Boris Johnson is driven out of office, it wouldn't mark the first time the Tories had four PMs in a row\nThey had previously governed the UK for 13 years with 4 PMs, from 1951-1964"
	if tweet.Text != expected_text {
		t.Errorf("Expected text: %q, got: %q", expected_text, tweet.Text)
	}

	if len(tweet_trove.Retweets) != 0 {
		t.Errorf("Shouldn't be any retweets")
	}
}


// Check a retweet
func TestAPIV2FeedRetweet(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/api_v2/feed_simple_retweet.json")
	if err != nil {
		panic(err)
	}
	var feed APIV2Response
	err = json.Unmarshal(data, &feed)
	if err != nil {
		t.Errorf(err.Error())
	}

	tweet_trove, err := feed.ToTweetTrove()
	if err != nil {
		panic(err)
	}

	// Should fetch both the retweeting and retweeted users
	if len(tweet_trove.Users) != 2 {
		t.Errorf("Expected %d users, got %d", 2, len(tweet_trove.Users))
	}
	user := tweet_trove.Users[44067298]
	if user.ID != 44067298 {
		t.Errorf("Expected ID %d, got %d", 44067298, user.ID)
	}
	if user.DisplayName != "Michael Malice" {
		t.Errorf("Expected display name %q, got %q", "Michael Malice", user.DisplayName)
	}

	retweeted_user := tweet_trove.Users[1326229737551912960]
	if retweeted_user.ID != 1326229737551912960 {
		t.Errorf("Expected ID %d, got %d", 1326229737551912960, retweeted_user.ID)
	}
	if retweeted_user.Handle != "libsoftiktok" {
		t.Errorf("Expected handle %q, got %q", "libsoftiktok", retweeted_user.Handle)
	}

	// Should only be 1 tweet, the retweeted one
	if len(tweet_trove.Tweets) != 1 {
		t.Errorf("Expected %d tweets, got %d", 1, len(tweet_trove.Tweets))
	}
	tweet, ok := tweet_trove.Tweets[1485694028620316673]
	if !ok {
		t.Fatalf("Didn't get the tweet")
	}
	if tweet.ID != 1485694028620316673 {
		t.Errorf("Expected ID %d, got %d", 1485694028620316673, tweet.ID)
	}
	if tweet.UserID != UserID(1326229737551912960) {
		t.Errorf("Expected user ID %d, got %d", 1326229737551912960, tweet.UserID)
	}
	expected_text := "More mask madness, this time in an elevator. The mask police are really nuts https://t.co/3BpvLjdJwD"
	if tweet.Text != expected_text {
		t.Errorf("Expected text: %q, got: %q", expected_text, tweet.Text)
	}

	// Should be 1 retweet
	if len(tweet_trove.Retweets) != 1 {
		t.Errorf("Expected %d retweets, got %d", 1, len(tweet_trove.Retweets))
	}
	retweet := tweet_trove.Retweets[1485699748514476037]
	if retweet.RetweetID != 1485699748514476037 {
		t.Errorf("Expected RetweetID %d, got %d", 1485699748514476037, retweet.RetweetID)
	}
	if retweet.TweetID != 1485694028620316673 {
		t.Errorf("Expected TweetID 1485694028620316673, got %d", retweet.TweetID)
	}
	if retweet.RetweetedAt.Unix() != 1643053397 {
		t.Errorf("Expected retweeted_at %d, got %d", 1643053397, retweet.RetweetedAt.Unix())
	}
	if retweet.RetweetedByID != UserID(44067298) {
		t.Errorf("Expected retweeted_by 44067298, got %d", retweet.RetweetedByID)
	}
}


// Check a quote-tweet
func TestAPIV2FeedQuoteTweet(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/api_v2/feed_quote_tweet.json")
	if err != nil {
		panic(err)
	}
	var feed APIV2Response
	err = json.Unmarshal(data, &feed)
	if err != nil {
		t.Errorf(err.Error())
	}

	tweet_trove, err := feed.ToTweetTrove()
	if err != nil {
		panic(err)
	}

	// Should be 2 users: quoter and quoted
	if len(tweet_trove.Users) != 2 {
		t.Errorf("Expected %d users, got %d", 2, len(tweet_trove.Users))
	}
	quoting_user := tweet_trove.Users[44067298]
	if quoting_user.ID != 44067298 {
		t.Errorf("Expected quoting user ID %d, got %d", 44067298, quoting_user.ID)
	}
	quoted_user := tweet_trove.Users[892155218292617217]
	if quoted_user.ID != 892155218292617217 {
		t.Errorf("Expected quoted user ID %d, got %d", 892155218292617217, quoted_user.ID)
	}
	expected_quoted_bio := "Creator of Little Homes and Mooncars"
	if quoted_user.Bio != expected_quoted_bio {
		t.Errorf("Expected bio %q, got %q", expected_quoted_bio, quoted_user.Bio)
	}


	// Should be 2 tweets: quote-tweet and quoted-tweet
	if len(tweet_trove.Tweets) != 2 {
		t.Errorf("Expected %d tweets, got %d", 2, len(tweet_trove.Tweets))
	}
	quoted_tweet := tweet_trove.Tweets[1485690069079846915]
	if quoted_tweet.ID != 1485690069079846915 {
		t.Errorf("Expected quoted ID %d, got %d", 1485690069079846915, quoted_tweet.ID)
	}
	expected_quoted_text := "The Left hates the Right so much that they won't let them leave the Union. I don't get it."
	if quoted_tweet.Text != expected_quoted_text {
		t.Errorf("Expected text %q, got %q", expected_quoted_text, quoted_tweet.Text)
	}
	quote_tweet := tweet_trove.Tweets[1485690410899021826]
	if quote_tweet.ID != 1485690410899021826 {
		t.Errorf("Expected quoting ID %d, got %d", 1485690410899021826, quote_tweet.ID)
	}
	if quote_tweet.QuotedTweetID != 1485690069079846915 {
		t.Errorf("Expected to be quoting tweet ID %d, got %d", 1485690069079846915, quote_tweet.QuotedTweetID)
	}


	// No retweets
	if len(tweet_trove.Retweets) != 0 {
		t.Errorf("Shouldn't be any retweets")
	}
}


// Check a retweeted quote-tweet
func TestAPIV2FeedRetweetedQuoteTweet(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/api_v2/feed_retweeted_quote_tweet.json")
	if err != nil {
		panic(err)
	}
	var feed APIV2Response
	err = json.Unmarshal(data, &feed)
	if err != nil {
		t.Errorf(err.Error())
	}

	tweet_trove, err := feed.ToTweetTrove()
	if err != nil {
		panic(err)
	}

	// 3 Users: quoted, quoter, and retweeter
	if len(tweet_trove.Users) != 3 {
		t.Errorf("Expected %d users, got %d", 3, len(tweet_trove.Users))
	}
	retweeting_user := tweet_trove.Users[599817378]
	if retweeting_user.ID != 599817378 {
		t.Errorf("Expected retweeting user ID %d, got %d", 599817378, retweeting_user.ID)
	}
	if retweeting_user.Website != "https://www.youtube.com/highlyrespected" {
		t.Errorf("Expected RTing user website %q, got %q", "https://www.youtube.com/highlyrespected", retweeting_user.Website)
	}
	retweeted_user := tweet_trove.Users[1434720042193760256]
	if retweeted_user.ID != 1434720042193760256 {
		t.Errorf("Expected retweed user ID %d, got %d", 1434720042193760256, retweeted_user.ID)
	}
	if retweeted_user.FollowersCount != 17843 {
		t.Errorf("Expected %d followers, got %d", 17843, retweeted_user.FollowersCount)
	}
	quoted_user := tweet_trove.Users[14347972]
	if quoted_user.ID != 14347972 {
		t.Errorf("Expected quoted user ID %d, got %d", 14347972, quoted_user.ID)
	}
	if quoted_user.IsVerified != true {
		t.Errorf("Expected quoted user to be verified")
	}


	// Quoted tweet and quoting tweet
	if len(tweet_trove.Tweets) != 2 {
		t.Errorf("Expected %d tweets, got %d", 2, len(tweet_trove.Tweets))
	}

	// The retweet
	if len(tweet_trove.Retweets) != 1 {
		t.Errorf("Expected %d retweets, got %d", 1, len(tweet_trove.Retweets))
	}
}



func TestParseAPIV2UserFeed(t *testing.T) {
	data, err := ioutil.ReadFile("test_responses/api_v2/user_feed_apiv2.json")
	if err != nil {
		panic(err)
	}
	var feed APIV2Response
	err = json.Unmarshal(data, &feed)
	if err != nil {
		t.Errorf(err.Error())
	}

	tweet_trove, err := feed.ToTweetTrove()
	if err != nil {
		panic(err)
	}

	// Check users
	user := tweet_trove.Users[44067298]
	if user.ID != 44067298 {
		t.Errorf("Expected ID %d, got %d", 44067298, user.ID)
	}
	if user.DisplayName != "Michael Malice" {
		t.Errorf("Expected display name %q, got %q", "Michael Malice", user.DisplayName)
	}

	retweeted_user := tweet_trove.Users[1326229737551912960]
	if retweeted_user.ID != 1326229737551912960 {
		t.Errorf("Expected ID %d, got %d", 1326229737551912960, retweeted_user.ID)
	}
	if retweeted_user.Handle != "libsoftiktok" {
		t.Errorf("Expected handle %q, got %q", "libsoftiktok", retweeted_user.Handle)
	}

	quote_tweeted_user := tweet_trove.Users[892155218292617217]
	if quote_tweeted_user.ID != 892155218292617217 {
		t.Errorf("Expected ID %d, got %d", 892155218292617217, quote_tweeted_user.ID)
	}

	// Check retweets
	if len(tweet_trove.Retweets) != 2 {
		t.Errorf("Expected %d retweets but got %d", 2, len(tweet_trove.Retweets))
	}

	// Test cursor-bottom
	bottom_cursor := feed.GetCursorBottom()
	if bottom_cursor != "HBaYgL2Fp/T7nCkAAA==" {
		t.Errorf("Expected cursor %q, got %q", "HBaYgL2Fp/T7nCkAAA==", bottom_cursor)
	}

	fmt.Printf("%d Users, %d Tweets, %d Retweets\n", len(tweet_trove.Users), len(tweet_trove.Tweets), len(tweet_trove.Retweets))
}
